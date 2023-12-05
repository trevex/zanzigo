package sqlite3

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/trevex/zanzigo"
	"github.com/trevex/zanzigo/storage/postgres"
)

//go:embed migrations/*.sql
var fs embed.FS

func RunMigrations(filepath string) error {
	driver, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}
	migrations, err := migrate.NewWithSourceInstance("iofs", driver, "sqlite3://"+filepath)
	if err != nil {
		return err
	}
	err = migrations.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

type SQLite3Storage struct {
	db *sql.DB
}

func NewSQLite3Storage(filepath string) (*SQLite3Storage, error) {
	db, err := sql.Open("sqlite3", filepath)
	return &SQLite3Storage{db}, err
}

func (s *SQLite3Storage) Close() error {
	return s.db.Close()
}

func (s *SQLite3Storage) Write(ctx context.Context, t zanzigo.Tuple) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, "INSERT INTO tuples (uuid, object_type, object_id, object_relation, subject_type, subject_id, subject_relation) values(?, ?, ?, ?, ?, ?, ?)", id.String(), t.ObjectType, t.ObjectID, t.ObjectRelation, t.SubjectType, t.SubjectID, t.SubjectRelation)
	return err
}

func (s *SQLite3Storage) Read(ctx context.Context, t zanzigo.Tuple) (uuid.UUID, error) {
	id := uuid.UUID{}
	err := s.db.QueryRowContext(ctx, "SELECT uuid FROM tuples WHERE object_type=? AND object_id=? AND object_relation=? AND subject_type=? AND subject_id=? AND subject_relation=?", t.ObjectType, t.ObjectID, t.ObjectRelation, t.SubjectType, t.SubjectID, t.SubjectRelation).
		Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		return id, zanzigo.ErrNotFound
	}
	return id, err
}

func (s *SQLite3Storage) PrepareRuleset(object, relation string, ruleset []zanzigo.InferredRule) (zanzigo.Userdata, error) {
	// TODO: Checking the query plan reveals idx_tuples-index is used for all selects (this is not the case for Postgres and not expected).
	//       This can be changed using INDEXED BY, but rather it should be verified how SQLite is supposed to plan the queries.
	return postgres.SelectQueryFor(ruleset, false, "?")
}

func (s *SQLite3Storage) QueryChecks(ctx context.Context, checks []zanzigo.Check) ([]zanzigo.MarkedTuple, error) {
	// TODO: current implementation could be more memory efficient by using buffer
	argNum := 1
	args := make([]any, 0, len(checks)*6)
	queries := make([]string, 0, len(checks))

	// We iterate over all check and combine all the queries
	for i, check := range checks {
		query, ok := check.Userdata.(string)
		if !ok {
			panic("malformed query data")
		}
		for _, rule := range check.Ruleset {
			switch rule.Kind {
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
		queries = append(queries, fmt.Sprintf("SELECT %d AS check_index", i)+", rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM ("+query+fmt.Sprintf(") AS cr%d", i))
		argNum += 4
	}

	// Join all queries with UNION ALL and ORDER BY rule index
	fullQuery := strings.Join(queries, " UNION ALL ") + " ORDER BY rule_index"

	// Let's fetch all the rows
	rows, err := s.db.QueryContext(ctx, fullQuery, args...)
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
