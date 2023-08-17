package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/trevex/zanzigo"
	"github.com/trevex/zanzigo/contrib/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	databaseURL = ""
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15.4",
		Env: []string{
			"POSTGRES_PASSWORD=zanzigo",
			"POSTGRES_USER=zanzigo",
			"POSTGRES_DB=zanzigo",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true // Stopped container should be removed
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	resource.Expire(300) // In any case container should be killed in 5min

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseURL = fmt.Sprintf("postgres://zanzigo:zanzigo@%s/zanzigo?sslmode=disable", hostAndPort)

	// We connect with exponential backoff (maximum wait 2min)
	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		db, err := sql.Open("pgx", databaseURL)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to postgres: %s", err)
	}

	if err := postgres.RunMigrations(databaseURL); err != nil {
		log.Fatalf("Could not migrate db: %s", err)
	}

	code := m.Run()

	// os.Exit doesn't care for defer, so let's explicitly purge...
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestPostgres(t *testing.T) {
	model := zanzigo.Model{
		"user": zanzigo.RelationMap{},
		"group": zanzigo.RelationMap{
			"member": zanzigo.Rule{},
		},
		"folder": zanzigo.RelationMap{
			"owner": zanzigo.Rule{},
			"editor": zanzigo.Rule{
				InheritIf: "owner",
			},
			"viewer": zanzigo.Rule{
				InheritIf: "editor",
			},
		},
		"doc": zanzigo.RelationMap{
			"parent": zanzigo.Rule{},
			"owner": zanzigo.Rule{
				InheritIf:    "owner",
				OfType:       "folder",
				WithRelation: "parent",
			},
			"editor": zanzigo.AnyOf(
				zanzigo.Rule{InheritIf: "owner"},
				zanzigo.Rule{
					InheritIf:    "editor",
					OfType:       "folder",
					WithRelation: "parent",
				},
			),
			"viewer": zanzigo.AnyOf(
				zanzigo.Rule{InheritIf: "editor"},
				zanzigo.Rule{
					InheritIf:    "viewer",
					OfType:       "folder",
					WithRelation: "parent",
				},
			),
		},
	}

	storage, err := postgres.NewPostgresStorage(databaseURL)
	if err != nil {
		t.Fatalf("Expected PostgresStorage to not error during creation: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	err = storage.Write(ctx, zanzigo.Tuple{
		Object:   "group:mygroup",
		Relation: "member",
		User:     "user:myuser",
	})
	if err != nil {
		t.Fatalf("Expected storage.Write to not error: %v", err)
	}

	err = storage.Write(ctx, zanzigo.Tuple{
		Object:   "doc:mydoc",
		Relation: "parent",
		User:     "folder:myfolder",
	})
	if err != nil {
		t.Fatalf("Expected storage.Write to not error: %v", err)
	}

	err = storage.Write(ctx, zanzigo.Tuple{
		Object:    "folder:myfolder",
		Relation:  "viewer",
		IsUserset: true,
		User:      "group:mygroup#member",
	})
	if err != nil {
		t.Fatalf("Expected storage.Write to not error: %v", err)
	}

	executer := zanzigo.NewSequentialExecuter()

	resolver, err := model.Resolver(storage, executer)
	if err != nil {
		t.Fatalf("Expected Resolver creation to not error on: %v", err)
	}

	result, err := resolver.Check(ctx, zanzigo.Tuple{
		Object:   "doc:mydoc",
		Relation: "viewer",
		User:     "user:myuser",
	})
	if !result || err != nil {
		t.Fatalf("Expected resolver.Check to return true, nil, but got %v, %v instead", result, err)
	}

}
