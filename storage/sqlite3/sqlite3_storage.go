package sqlite3

import (
	"context"
	"embed"
	"fmt"
	"runtime"
	"strings"

	"github.com/trevex/zanzigo"
	"github.com/trevex/zanzigo/storage/postgres"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"zombiezen.com/go/sqlite/sqlitex"
)

//go:embed migrations/*.sql
var fs embed.FS

var (
	ErrUnableToGetConn = fmt.Errorf("unable to get connection from pool")
)

func RunMigrations(filepath string) error {
	driver, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}
	migrations, err := migrate.NewWithSourceInstance("iofs", driver, "sqlite://"+filepath)
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
	pool *sqlitex.Pool
}

func NewSQLite3Storage(filepath string) (*SQLite3Storage, error) {
	pool, err := sqlitex.Open(filepath, 0, max(4, runtime.NumCPU()))
	return &SQLite3Storage{pool}, err
}

func (s *SQLite3Storage) Close() error {
	return s.pool.Close()
}

func (s *SQLite3Storage) Write(ctx context.Context, t zanzigo.Tuple) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	conn := s.pool.Get(ctx)
	if conn == nil {
		return ErrUnableToGetConn
	}
	defer s.pool.Put(conn)

	stmt, err := conn.Prepare("INSERT INTO tuples (uuid, object_type, object_id, object_relation, subject_type, subject_id, subject_relation) values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	stmt.BindText(1, id.String())
	stmt.BindText(2, t.ObjectType)
	stmt.BindText(3, t.ObjectID)
	stmt.BindText(4, t.ObjectRelation)
	stmt.BindText(5, t.SubjectType)
	stmt.BindText(6, t.SubjectID)
	stmt.BindText(7, t.SubjectRelation)

	_, err = stmt.Step()
	return err
}

func (s *SQLite3Storage) Read(ctx context.Context, t zanzigo.Tuple) (uuid.UUID, error) {
	id := uuid.UUID{}

	conn := s.pool.Get(ctx)
	if conn == nil {
		return id, ErrUnableToGetConn
	}
	defer s.pool.Put(conn)

	stmt, err := conn.Prepare("SELECT uuid FROM tuples WHERE object_type=? AND object_id=? AND object_relation=? AND subject_type=? AND subject_id=? AND subject_relation=?")
	if err != nil {
		return id, err
	}
	stmt.BindText(1, t.ObjectType)
	stmt.BindText(2, t.ObjectID)
	stmt.BindText(3, t.ObjectRelation)
	stmt.BindText(4, t.SubjectType)
	stmt.BindText(5, t.SubjectID)
	stmt.BindText(6, t.SubjectRelation)
	hasRows, err := stmt.Step()

	if !hasRows {
		return id, zanzigo.ErrNotFound
	}

	return uuid.FromString(stmt.ColumnText(0))
}

func (s *SQLite3Storage) PrepareRuleset(object, relation string, ruleset []zanzigo.InferredRule) (zanzigo.Userdata, error) {
	// TODO: Checking the query plan reveals idx_tuples-index is used for all selects (this is not the case for Postgres and not expected).
	//       This can be changed using INDEXED BY, but rather it should be verified how SQLite is supposed to plan the queries.
	return postgres.SelectQueryFor(ruleset, false, "?")
}

func (s *SQLite3Storage) QueryChecks(ctx context.Context, checks []zanzigo.Check) ([]zanzigo.MarkedTuple, error) {
	// TODO: current implementation could be more memory efficient by using buffer
	argNum := 1
	args := make([]string, 0, len(checks)*6)
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

	conn := s.pool.Get(ctx)
	if conn == nil {
		return nil, ErrUnableToGetConn
	}
	defer s.pool.Put(conn)

	// Let's fetch all the rows
	stmt, err := conn.Prepare(fullQuery)
	if err != nil {
		return nil, err
	}
	for i, arg := range args {
		stmt.BindText(i+1, arg)
	}

	tuples := []zanzigo.MarkedTuple{}
	for {
		if hasRow, err := stmt.Step(); err != nil {
			return nil, err
		} else if !hasRow {
			break
		}

		t := zanzigo.MarkedTuple{}
		t.CheckIndex = stmt.ColumnInt(0)
		t.RuleIndex = stmt.ColumnInt(1)
		t.ObjectType = stmt.ColumnText(2)
		t.ObjectID = stmt.ColumnText(3)
		t.ObjectRelation = stmt.ColumnText(4)
		t.SubjectType = stmt.ColumnText(5)
		t.SubjectID = stmt.ColumnText(6)
		t.SubjectRelation = stmt.ColumnText(7)
		tuples = append(tuples, t)
	}

	return tuples, nil
}
