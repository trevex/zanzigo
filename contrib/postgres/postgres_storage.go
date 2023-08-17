package postgres

import (
	"embed"

	"github.com/trevex/zanzigo"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
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

func NewPostgresStorage(connStr string) (zanzigo.Storage, error) {
	return &postgresStorage{}, nil
}

type postgresStorage struct{}

func (s *postgresStorage) Write(t zanzigo.Tuple) error {
	return nil
}
