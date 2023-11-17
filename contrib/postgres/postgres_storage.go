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
	"github.com/jackc/pgx/v5/pgtype/zeronull"
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

func NewPostgresStorage(databaseURL string) (zanzigo.Storage, error) {
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
	return &postgresStorage{pool}, nil
}

type postgresStorage struct {
	pool *pgxpool.Pool
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

func (s *postgresStorage) QueryForCommands(commands []zanzigo.Command) zanzigo.Query {
	return NewPostgresQuery(commands)
}

func (s *postgresStorage) RunQuery(ctx context.Context, qc zanzigo.QueryContext, q zanzigo.Query, commands []zanzigo.Command) ([]zanzigo.MarkedTuple, error) {
	args := make([]any, 0, q.NumArgs())
	for _, command := range commands {
		switch command.Kind() { // NEEDS TO BE IN SYNC WITH `NewPostgresQuery`
		case zanzigo.KindDirect:
			args = append(args, qc.Root.ObjectID, qc.Root.SubjectType, qc.Root.SubjectID, zeronull.Text(qc.Root.SubjectRelation))
		case zanzigo.KindDirectUserset:
			args = append(args, qc.Root.ObjectID)
		case zanzigo.KindIndirect:
			args = append(args, qc.Root.ObjectID)
		default:
			panic("unreachable")
		}
	}
	if len(args) != q.NumArgs() {
		return nil, errors.New("args for query do not match expected number of args")
	}
	rows, err := s.pool.Query(ctx, q.Get(), args...)
	if err != nil {
		return nil, err
	}
	tuples := []zanzigo.MarkedTuple{}
	for rows.Next() {
		t := zanzigo.MarkedTuple{}
		err := rows.Scan(&t.CommandID, &t.ObjectType, &t.ObjectID, &t.ObjectRelation, &t.SubjectType, &t.SubjectID, &t.SubjectRelation)
		if err != nil {
			return nil, err
		}
		tuples = append(tuples, t)
	}
	return tuples, nil
}

func NewPostgresQuery(commands []zanzigo.Command) zanzigo.Query {
	parts := []string{}
	j := 1
	for id, command := range commands {
		rule := command.Rule()
		relations := "(" + strings.Join(lo.Map(command.Rule().Relations, func(r string, _ int) string {
			return "object_relation='" + r + "'"
		}), " OR ") + ")"
		switch command.Kind() { // NEEDS TO BE IN SYNC WITH `postgresStorage.RunQuery`
		case zanzigo.KindDirect:
			// NOTE: we need to be careful with `subject_relation` to handle NULL properly!
			parts = append(parts, fmt.Sprintf("(SELECT %d AS command_id, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='%s' AND object_id=$%d AND %s AND subject_type=$%d AND subject_id=$%d AND subject_relation=$%d)", id, rule.Object, j, relations, j+1, j+2, j+3))
			j += 4
		case zanzigo.KindDirectUserset:
			parts = append(parts, fmt.Sprintf("(SELECT %d AS command_id, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='%s' AND object_id=$%d AND %s AND subject_relation IS NOT NULL)", id, rule.Object, j, relations))
			j += 1
		case zanzigo.KindIndirect:
			parts = append(parts, fmt.Sprintf("(SELECT %d AS command_id, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='%s' AND object_id=$%d AND %s AND subject_type='%s')", id, rule.Object, j, relations, rule.Subject))
		default:
			panic("unreachable")
		}
	}
	return &postgresQuery{
		query:   strings.Join(parts, " UNION ALL ") + " ORDER BY command_id",
		numArgs: j,
	}
}

type postgresQuery struct {
	query   string
	numArgs int
}

func (q *postgresQuery) Get() string {
	return q.query
}
func (q *postgresQuery) NumArgs() int {
	return q.numArgs
}
