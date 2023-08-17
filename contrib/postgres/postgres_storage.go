package postgres

import (
	"context"
	"embed"

	"github.com/trevex/zanzigo"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
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
	pool, err := pgxpool.Connect(context.Background(), databaseURL)
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
	_, err := s.pool.Exec(ctx, "INSERT INTO tuples (object, relation, is_userset, user_) values($1, $2, $3, $4)", t.Object, t.Relation, t.IsUserset, t.User)
	return err
}
