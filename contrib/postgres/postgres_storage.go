package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/trevex/zanzigo"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	pgxuuid "github.com/jackc/pgx-gofrs-uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/lo"
)

//go:embed migrations/*.sql
var fs embed.FS

func RunMigrations(databaseURL string) error {
	driver, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}
	migrate, err := migrate.NewWithSourceInstance("iofs", driver, databaseURL)
	if err != nil {
		return err
	}
	err = migrate.Up()
	if err != nil {
		return err
	}
	return nil
}

type PostgresOption interface {
	do(*postgresConfig)
}

type postgresConfig struct {
	useFunctions bool
}

type postgresFunctionAdapter func(*postgresConfig)

func (fn postgresFunctionAdapter) do(c *postgresConfig) {
	fn(c)
}

// This was implemented as a test, but is significantly slower than raw queries.
// TODO: remove when considered obsolete!
func UseFunctions() PostgresOption {
	return postgresFunctionAdapter(func(c *postgresConfig) { c.useFunctions = true })
}

type postgresStorage struct {
	pool         *pgxpool.Pool
	useFunctions bool
}

func NewPostgresStorage(databaseURL string, options ...PostgresOption) (zanzigo.Storage, error) {
	opts := postgresConfig{}
	lo.ForEach(options, func(o PostgresOption, _ int) { o.do(&opts) })
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err // TODO: wrap?
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxuuid.Register(conn.TypeMap())
		return nil
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}
	return &postgresStorage{pool, opts.useFunctions}, nil
}

func (s *postgresStorage) Close() error {
	s.pool.Close()
	return nil
}

func (s *postgresStorage) Write(ctx context.Context, t zanzigo.Tuple) error {
	_, err := s.pool.Exec(ctx, "INSERT INTO tuples (object_type, object_id, object_relation, subject_type, subject_id, subject_relation) values($1, $2, $3, $4, $5, $6)", t.ObjectType, t.ObjectID, t.ObjectRelation, t.SubjectType, t.SubjectID, t.SubjectRelation)
	return err
}

func (s *postgresStorage) Read(ctx context.Context, t zanzigo.Tuple) (uuid.UUID, error) {
	uuid := uuid.UUID{}
	err := s.pool.QueryRow(ctx, "SELECT uuid FROM tuples WHERE object_type=$1 AND object_id=$2 AND object_relation=$3 AND subject_type=$4 AND subject_id=$5 AND subject_relation=$6", t.ObjectType, t.ObjectID, t.ObjectRelation, t.SubjectType, t.SubjectID, t.SubjectRelation).
		Scan(&uuid)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid, zanzigo.ErrNotFound
	}
	return uuid, err
}

func (s *postgresStorage) PrepareForChecks(object, relation string, commands []zanzigo.CheckCommand) (zanzigo.Userdata, error) {
	if !s.useFunctions {
		return newPostgresQuery(commands), nil
	}
	return s.newPostgresFunction(object, relation, commands)
}

func (s *postgresStorage) QueryChecks(ctx context.Context, checks []zanzigo.CheckRequest) ([]zanzigo.MarkedTuple, error) {
	if !s.useFunctions {
		return s.queryChecks(ctx, checks)
	}
	return s.queryChecksFunction(ctx, checks)
}

func (s *postgresStorage) queryChecks(ctx context.Context, checks []zanzigo.CheckRequest) ([]zanzigo.MarkedTuple, error) {
	argNum := 1
	args := make([]any, 0) // TODO: precalculate by adding q.numArgs
	checkQueries := make([]string, 0, len(checks))
	for i, check := range checks {
		q, ok := check.Userdata.(*postgresQuery)
		if !ok {
			panic("malformed query data")
		}
		for _, command := range check.Commands {
			switch command.Kind() { // NEEDS TO BE IN SYNC WITH `NewPostgresCheckQuery`
			case zanzigo.KindDirect:
				args = append(args, check.Tuple.ObjectID, check.Tuple.SubjectType, check.Tuple.SubjectID, check.Tuple.SubjectRelation)
			case zanzigo.KindDirectUserset:
				args = append(args, check.Tuple.ObjectID)
			case zanzigo.KindIndirect:
				args = append(args, check.Tuple.ObjectID)
			default:
				panic("unreachable")
			}
		}
		if len(args) != argNum+q.numArgs-1 {
			return nil, errors.New("args for query do not match expected number of args")
		}
		checkQueries = append(checkQueries, fmt.Sprintf("SELECT %d AS check_id", i)+", command_id, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM ("+fmt.Sprintf(q.query, makeArgsRange(argNum, argNum+q.numArgs-1)...)+fmt.Sprintf(") AS check%d", i))
		argNum += q.numArgs
	}

	fullQuery := strings.Join(checkQueries, " UNION ALL ") + " ORDER BY command_id"

	rows, err := s.pool.Query(ctx, fullQuery, args...)
	if err != nil {
		return nil, err
	}
	tuples := []zanzigo.MarkedTuple{}
	for rows.Next() {
		t := zanzigo.MarkedTuple{}
		err := rows.Scan(&t.CheckID, &t.CommandID, &t.ObjectType, &t.ObjectID, &t.ObjectRelation, &t.SubjectType, &t.SubjectID, &t.SubjectRelation)
		if err != nil {
			return nil, err
		}
		tuples = append(tuples, t)
	}
	return tuples, nil
}

// Will precompute the query for the commands
func newPostgresQuery(commands []zanzigo.CheckCommand) *postgresQuery {
	parts := []string{}
	j := 0
	for id, command := range commands {
		rule := command.Rule()
		relations := "(" + strings.Join(lo.Map(command.Rule().Relations, func(r string, _ int) string {
			return "object_relation='" + r + "'"
		}), " OR ") + ")"
		switch command.Kind() { // NEEDS TO BE IN SYNC WITH `postgresStorage.RunQuery`
		case zanzigo.KindDirect:
			// NOTE: we need to be careful with `subject_relation` to handle NULL properly!
			parts = append(parts, fmt.Sprintf("(SELECT %d AS command_id, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='%s' AND object_id=%%s AND %s AND subject_type=%%s AND subject_id=%%s AND subject_relation=%%s)", id, rule.Object, relations))
			j += 4
		case zanzigo.KindDirectUserset:
			parts = append(parts, fmt.Sprintf("(SELECT %d AS command_id, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='%s' AND object_id=%%s AND %s AND subject_relation <> '')", id, rule.Object, relations))
			j += 1
		case zanzigo.KindIndirect:
			parts = append(parts, fmt.Sprintf("(SELECT %d AS command_id, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='%s' AND object_id=%%s AND %s AND subject_type='%s')", id, rule.Object, relations, rule.Subject))
			j += 1
		default:
			panic("unreachable")
		}
	}
	return &postgresQuery{
		query:   strings.Join(parts, " UNION ALL "),
		numArgs: j,
	}
}

type postgresQuery struct {
	query   string
	numArgs int
}

func (s *postgresStorage) queryChecksFunction(ctx context.Context, checks []zanzigo.CheckRequest) ([]zanzigo.MarkedTuple, error) {
	if len(checks) != 1 {
		panic("unreachable")
	}
	check := checks[0]
	fn, ok := check.Userdata.(*postgresFunction)
	if !ok {
		panic("malformed query data")
	}
	result := false
	err := s.pool.QueryRow(ctx, fn.query, check.Tuple.ObjectID, check.Tuple.SubjectType, check.Tuple.SubjectID, check.Tuple.SubjectRelation).Scan(&result)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return []zanzigo.MarkedTuple{{CheckID: 0, CommandID: 0, Tuple: check.Tuple}}, nil
}

func (s *postgresStorage) newPostgresFunction(object, relation string, commands []zanzigo.CheckCommand) (*postgresFunction, error) {
	funcDecl, query := postgresFunctionFor(object, relation, commands)

	_, err := s.pool.Exec(context.Background(), funcDecl)
	return &postgresFunction{
		query: query,
	}, err
}

type postgresFunction struct {
	query string
}

// TODO: respect maxDepth!
func postgresFunctionFor(object, relation string, commands []zanzigo.CheckCommand) (string, string) {
	innerSelect := strings.Join(lo.Map(commands, func(command zanzigo.CheckCommand, id int) string {
		rule := command.Rule()
		relations := "(" + strings.Join(lo.Map(command.Rule().Relations, func(r string, _ int) string {
			return "object_relation='" + r + "'"
		}), " OR ") + ")"
		switch command.Kind() {
		case zanzigo.KindDirect:
			return fmt.Sprintf("(SELECT %d AS command_id, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='%s' AND object_id=$1 AND %s AND subject_type=$2 AND subject_id=$3 AND subject_relation=$4)", id, rule.Object, relations)
		case zanzigo.KindDirectUserset:
			return fmt.Sprintf("(SELECT %d AS command_id, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='%s' AND object_id=$1 AND %s AND subject_relation <> '')", id, rule.Object, relations)
		case zanzigo.KindIndirect:
			return fmt.Sprintf("(SELECT %d AS command_id, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='%s' AND object_id=$1 AND %s AND subject_type='%s')", id, rule.Object, relations, rule.Subject)
		default:
			panic("unreachable")
		}
	}), " UNION ALL ") + " ORDER BY command_id"
	conditions := ""
	for id, command := range commands {
		if id == 0 {
			conditions += "IF "
		} else {
			conditions += "ELSIF "
		}
		conditions += fmt.Sprintf("mt.command_id = %d THEN\n", id)
		switch command.Kind() {
		case zanzigo.KindDirect:
			conditions += "RETURN TRUE;\n"
		case zanzigo.KindDirectUserset:
			conditions += addResultCheck("EXECUTE FORMAT('SELECT zanzigo_%s_%s($1, $2, $3, $4)', mt.subject_type, mt.subject_relation) USING mt.subject_id, $2, $3, $4 INTO result;\n")
		case zanzigo.KindIndirect:
			relations := command.Rule().WithRelationToSubject
			for _, relation := range relations {
				conditions += addResultCheck(fmt.Sprintf("EXECUTE FORMAT('SELECT zanzigo_%%s_%s($1, $2, $3, $4)', mt.subject_type) USING mt.subject_id, $2, $3, $4 INTO result;\n", relation))
			}
		default:
			panic("unreachable")
		}
		if id == len(commands)-1 {
			conditions += "END IF;"
		}
	}
	funcName := fmt.Sprintf("zanzigo_%s_%s", object, relation)
	funcDecl := fmt.Sprintf(`CREATE OR REPLACE FUNCTION %s(TEXT, TEXT, TEXT, TEXT) RETURNS BOOLEAN LANGUAGE 'plpgsql' AS $$
DECLARE
mt RECORD;
result BOOLEAN;
BEGIN
FOR mt IN
%s
LOOP
%s
END LOOP;
RETURN FALSE;
END;
$$;`, funcName, innerSelect, conditions)

	return funcDecl, "SELECT " + funcName + "($1, $2, $3, $4)"
}

func addResultCheck(in string) string {
	return in + "IF result = TRUE THEN\nRETURN TRUE;\nEND IF;\n"
}

// Create functions for the set of commands

func makeArgsRange(min, max int) []any {
	a := make([]any, max-min+1)
	for i := range a {
		a[i] = fmt.Sprintf("$%d", min+i)
	}
	return a
}
