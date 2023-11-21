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
	migrations, err := migrate.NewWithSourceInstance("iofs", driver, databaseURL)
	if err != nil {
		return err
	}
	err = migrations.Up()
	if err != nil && err != migrate.ErrNoChange {
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

func UseFunctions() PostgresOption {
	return postgresFunctionAdapter(func(c *postgresConfig) { c.useFunctions = true })
}

type PostgresStorage struct {
	pool         *pgxpool.Pool
	useFunctions bool
}

func NewPostgresStorage(databaseURL string, options ...PostgresOption) (*PostgresStorage, error) {
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
	return &PostgresStorage{pool, opts.useFunctions}, nil
}

func (s *PostgresStorage) Close() error {
	s.pool.Close()
	return nil
}

func (s *PostgresStorage) Write(ctx context.Context, t zanzigo.Tuple) error {
	_, err := s.pool.Exec(ctx, "INSERT INTO tuples (object_type, object_id, object_relation, subject_type, subject_id, subject_relation) values($1, $2, $3, $4, $5, $6)", t.ObjectType, t.ObjectID, t.ObjectRelation, t.SubjectType, t.SubjectID, t.SubjectRelation)
	return err
}

func (s *PostgresStorage) Read(ctx context.Context, t zanzigo.Tuple) (uuid.UUID, error) {
	uuid := uuid.UUID{}
	err := s.pool.QueryRow(ctx, "SELECT uuid FROM tuples WHERE object_type=$1 AND object_id=$2 AND object_relation=$3 AND subject_type=$4 AND subject_id=$5 AND subject_relation=$6", t.ObjectType, t.ObjectID, t.ObjectRelation, t.SubjectType, t.SubjectID, t.SubjectRelation).
		Scan(&uuid)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid, zanzigo.ErrNotFound
	}
	return uuid, err
}

func (s *PostgresStorage) PrepareRuleset(object, relation string, ruleset []zanzigo.InferredRule) (zanzigo.Userdata, error) {
	if !s.useFunctions {
		return queryWithPlaceholdersFor(ruleset), nil
	}
	return s.createOrReplaceFunctionFor(object, relation, ruleset)
}

func (s *PostgresStorage) QueryChecks(ctx context.Context, crs []zanzigo.Check) ([]zanzigo.MarkedTuple, error) {
	if !s.useFunctions {
		return s.queryChecksWithQuery(ctx, crs)
	}
	return s.queryChecksWithFunction(ctx, crs)
}

///////////////////////////////////////////////////////////////////////////////
// QUERY-BASED IMPLEMENTATION
///////////////////////////////////////////////////////////////////////////////

func (s *PostgresStorage) queryChecksWithQuery(ctx context.Context, checks []zanzigo.Check) ([]zanzigo.MarkedTuple, error) {
	argNum := 1
	placesholders := make([]any, 0, len(checks)*6)
	args := make([]any, 0, len(checks)*4)
	queries := make([]string, 0, len(checks))
	for i, check := range checks {
		query, ok := check.Userdata.(string)
		if !ok {
			panic("malformed query data")
		}
		for _, rule := range check.Ruleset {
			switch rule.Kind { // NEEDS TO BE IN SYNC WITH `NewPostgresCheckQuery`
			case zanzigo.KindDirect:
				placesholders = append(placesholders, argNum, argNum+1, argNum+2, argNum+3)
			case zanzigo.KindDirectUserset:
				placesholders = append(placesholders, argNum)
			case zanzigo.KindIndirect:
				placesholders = append(placesholders, argNum)
			default:
				panic("unreachable")
			}
		}
		// Every ruleset always contains a direct relationship, so we can safely add all values
		args = append(args, check.Tuple.ObjectID, check.Tuple.SubjectType, check.Tuple.SubjectID, check.Tuple.SubjectRelation)
		queries = append(queries, fmt.Sprintf("SELECT %d AS check_index", i)+", rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM ("+query+fmt.Sprintf(") AS cr%d", i))
		argNum += 4
	}

	// We append the query and replace all the $%d with the proper placeholder numbers corresponding to args
	fullQuery := fmt.Sprintf(strings.Join(queries, " UNION ALL ")+" ORDER BY rule_index", placesholders...)

	// Let's fetch all the rows
	rows, err := s.pool.Query(ctx, fullQuery, args...)
	if err != nil {
		return nil, err
	}
	tuples := []zanzigo.MarkedTuple{}
	for rows.Next() {
		t := zanzigo.MarkedTuple{}
		err := rows.Scan(&t.CheckIndex, &t.RuleIndex, &t.ObjectType, &t.ObjectID, &t.ObjectRelation, &t.SubjectType, &t.SubjectID, &t.SubjectRelation)
		if err != nil {
			return nil, err
		}
		tuples = append(tuples, t)
	}
	return tuples, nil
}

// precompute the query for the specific ruleset
func queryWithPlaceholdersFor(ruleset []zanzigo.InferredRule) string {
	parts := make([]string, 0, len(ruleset))
	for id, rule := range ruleset {
		relations := "(" + strings.Join(lo.Map(rule.Relations, func(r string, _ int) string {
			return "object_relation='" + r + "'"
		}), " OR ") + ")"
		switch rule.Kind {
		case zanzigo.KindDirect:
			// Has 4 substitutions still to do: $1, $2, $3, $4 (if we assume it is first ruleset of a batch)
			parts = append(parts, fmt.Sprintf("(SELECT %d AS rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='%s' AND object_id=$%%d AND %s AND subject_type=$%%d AND subject_id=$%%d AND subject_relation=$%%d)", id, rule.Object, relations))
		case zanzigo.KindDirectUserset:
			// Only requires one additional substitions $1 (only first arg, which is .ObjectID)
			parts = append(parts, fmt.Sprintf("(SELECT %d AS rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='%s' AND object_id=$%%d AND %s AND subject_relation <> '')", id, rule.Object, relations))
		case zanzigo.KindIndirect:
			// Only requires one additional substitions $1 (only first arg, which is .ObjectID)
			parts = append(parts, fmt.Sprintf("(SELECT %d AS rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='%s' AND object_id=$%%d AND %s AND subject_type='%s')", id, rule.Object, relations, rule.Subject))
		default:
			panic("unreachable")
		}
	}
	return strings.Join(parts, " UNION ALL ")
}

///////////////////////////////////////////////////////////////////////////////
// FUNCTION-BASED IMPLEMENTATION
///////////////////////////////////////////////////////////////////////////////

func (s *PostgresStorage) queryChecksWithFunction(ctx context.Context, checks []zanzigo.Check) ([]zanzigo.MarkedTuple, error) {
	if len(checks) != 1 {
		panic("unreachable")
	}
	check := checks[0]
	query, ok := check.Userdata.(string)
	if !ok {
		panic("malformed query data")
	}
	result := false
	err := s.pool.QueryRow(ctx, query, check.Tuple.ObjectID, check.Tuple.SubjectType, check.Tuple.SubjectID, check.Tuple.SubjectRelation).Scan(&result)
	if err != nil {
		return nil, err
	}
	if !result {
		return nil, nil
	}

	return []zanzigo.MarkedTuple{{CheckIndex: 0, RuleIndex: 0, Tuple: check.Tuple}}, nil
}

func (s *PostgresStorage) createOrReplaceFunctionFor(object, relation string, ruleset []zanzigo.InferredRule) (string, error) {
	funcDecl, query := functionAndQueryFor(object, relation, ruleset)

	_, err := s.pool.Exec(context.Background(), funcDecl)
	return query, err
}

// TODO: respect maxDepth!
func functionAndQueryFor(object, relation string, ruleset []zanzigo.InferredRule) (string, string) {
	innerSelect := strings.Join(lo.Map(ruleset, func(rule zanzigo.InferredRule, id int) string {
		relations := "(" + strings.Join(lo.Map(rule.Relations, func(r string, _ int) string {
			return "object_relation='" + r + "'"
		}), " OR ") + ")"
		switch rule.Kind {
		case zanzigo.KindDirect:
			return fmt.Sprintf("(SELECT %d AS rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='%s' AND object_id=$1 AND %s AND subject_type=$2 AND subject_id=$3 AND subject_relation=$4)", id, rule.Object, relations)
		case zanzigo.KindDirectUserset:
			return fmt.Sprintf("(SELECT %d AS rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='%s' AND object_id=$1 AND %s AND subject_relation <> '')", id, rule.Object, relations)
		case zanzigo.KindIndirect:
			return fmt.Sprintf("(SELECT %d AS rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='%s' AND object_id=$1 AND %s AND subject_type='%s')", id, rule.Object, relations, rule.Subject)
		default:
			panic("unreachable")
		}
	}), " UNION ALL ") + " ORDER BY rule_index"
	conditions := ""
	for id, rule := range ruleset {
		if id == 0 {
			conditions += "IF "
		} else {
			conditions += "ELSIF "
		}
		conditions += fmt.Sprintf("mt.rule_index = %d THEN\n", id)
		switch rule.Kind {
		case zanzigo.KindDirect:
			conditions += "RETURN TRUE;\n"
		case zanzigo.KindDirectUserset:
			conditions += addResultCheck("EXECUTE FORMAT('SELECT zanzigo_%s_%s($1, $2, $3, $4)', mt.subject_type, mt.subject_relation) USING mt.subject_id, $2, $3, $4 INTO result;\n")
		case zanzigo.KindIndirect:
			relations := rule.WithRelationToSubject
			for _, relation := range relations {
				conditions += addResultCheck(fmt.Sprintf("SELECT zanzigo_%s_%s(mt.subject_id, $2, $3, $4) INTO result;\n", rule.Subject, relation))
			}
		default:
			panic("unreachable")
		}
		if id == len(ruleset)-1 {
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
