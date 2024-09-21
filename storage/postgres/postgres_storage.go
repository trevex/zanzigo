package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"strconv"
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

func (s *PostgresStorage) CursorStart() zanzigo.Cursor {
	return uuid.Must(uuid.FromString("ffffffff-ffff-ffff-ffff-ffffffffffff")).Bytes()
}

func (s *PostgresStorage) List(ctx context.Context, t zanzigo.Tuple, p zanzigo.Pagination) ([]zanzigo.Tuple, zanzigo.Cursor, error) {
	args := make([]any, 0, 8)
	whereClauses := ""
	if t.ObjectType != "" {
		args = append(args, t.ObjectType)
		whereClauses += "object_type=$" + strconv.Itoa(len(args)) + " AND "
	}
	if t.ObjectID != "" {
		args = append(args, t.ObjectID)
		whereClauses += "object_id=$" + strconv.Itoa(len(args)) + " AND "
	}
	if t.ObjectRelation != "" {
		args = append(args, t.ObjectRelation)
		whereClauses += "object_relation=$" + strconv.Itoa(len(args)) + " AND "
	}
	if t.SubjectType != "" {
		args = append(args, t.SubjectType)
		whereClauses += "subject_type=$" + strconv.Itoa(len(args)) + " AND "
	}
	if t.SubjectID != "" {
		args = append(args, t.SubjectID)
		whereClauses += "subject_id=$" + strconv.Itoa(len(args)) + " AND "
	}
	if t.SubjectRelation != "" {
		args = append(args, t.SubjectRelation)
		whereClauses += "subject_relation=$" + strconv.Itoa(len(args)) + " AND "
	}

	cursor, err := uuid.FromBytes(p.Cursor)
	if err != nil {
		return nil, nil, err
	}
	args = append(args, cursor)
	whereClauses += "uuid<$" + strconv.Itoa(len(args))

	args = append(args, p.Limit)
	limit := "LIMIT $" + strconv.Itoa(len(args))

	rows, err := s.pool.Query(ctx, "SELECT uuid, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE "+whereClauses+" ORDER BY uuid DESC "+limit, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	tuples := make([]zanzigo.Tuple, 0, p.Limit)

	for rows.Next() {
		var t zanzigo.Tuple
		err := rows.Scan(&cursor, &t.ObjectType, &t.ObjectID, &t.ObjectRelation, &t.SubjectType, &t.SubjectID, &t.SubjectRelation)
		if err != nil {
			return nil, nil, err
		}
		tuples = append(tuples, t)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// We use UUIDv7, so we can directly use it as cursor as it is sequential
	return tuples, cursor.Bytes(), nil
}

func (s *PostgresStorage) PrepareRuleset(object, relation string, ruleset []zanzigo.InferredRule) (zanzigo.Userdata, error) {
	if !s.useFunctions {
		return SelectQueryFor(ruleset, true, "$%d")
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
	// TODO: current implementation could be more memory efficient by using buffer
	argNum := 1
	placesholders := make([]any, 0, len(checks)*6)
	args := make([]any, 0, len(checks)*4)
	queries := make([]string, 0, len(checks))

	// We iterate over all check and combine all the queries
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
	funcDecl, query, err := FunctionFor(fmt.Sprintf("zanzigo_%s_%s", object, relation), ruleset)
	if err != nil {
		return "", err
	}

	_, err = s.pool.Exec(context.Background(), funcDecl)
	return query, err
}
