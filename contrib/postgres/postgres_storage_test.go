package postgres

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/trevex/zanzigo"
	"golang.org/x/exp/slices"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	databaseURL = ""
	storage     zanzigo.Storage
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

	if err := RunMigrations(databaseURL); err != nil {
		log.Fatalf("Could not migrate db: %s", err)
	}

	storage, err = NewPostgresStorage(databaseURL)
	if err != nil {
		log.Fatalf("PostgresStorage creation failed: %v", err)
	}
	defer storage.Close()

	DefaultData(storage)

	code := m.Run()

	// os.Exit doesn't care for defer, so let's explicitly purge...
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func DefaultModel(t *testing.T) *zanzigo.Model {
	model, err := zanzigo.NewModel(zanzigo.ObjectMap{
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
	})
	if err != nil {
		t.Fatalf("Model creation failed: %v", err)
	}
	return model
}

func DefaultData(storage zanzigo.Storage) {
	ctx := context.Background()

	err := storage.Write(ctx, zanzigo.Tuple{
		ObjectType:     "group",
		ObjectID:       "mygroup",
		ObjectRelation: "member",
		SubjectType:    "user",
		SubjectID:      "myuser",
	})
	if err != nil {
		log.Fatalf("Expected storage.Write not to fail: %v", err)
	}

	err = storage.Write(ctx, zanzigo.Tuple{
		ObjectType:     "doc",
		ObjectID:       "mydoc",
		ObjectRelation: "parent",
		SubjectType:    "folder",
		SubjectID:      "myfolder",
	})
	if err != nil {
		log.Fatalf("Expected storage.Write not to fail: %v", err)
	}

	err = storage.Write(ctx, zanzigo.Tuple{
		ObjectType:      "folder",
		ObjectID:        "myfolder",
		ObjectRelation:  "viewer",
		SubjectType:     "group",
		SubjectID:       "mygroup",
		SubjectRelation: "member",
	})
	if err != nil {
		log.Fatalf("Expected storage.Write not to fail: %v", err)
	}
}

func TestPostgresWithResolver(t *testing.T) {
	model := DefaultModel(t)

	storage, err := NewPostgresStorage(databaseURL)
	if err != nil {
		t.Fatalf("PostgresStorage creation failed: %v", err)
	}
	defer storage.Close()

	resolver, err := zanzigo.NewResolver(model, storage, 16)
	if err != nil {
		t.Fatalf("Expected Resolver creation to not error on: %v", err)
	}

	result, err := resolver.Check(context.Background(), zanzigo.Tuple{
		ObjectType:     "doc",
		ObjectID:       "mydoc",
		ObjectRelation: "viewer",
		SubjectType:    "user",
		SubjectID:      "myuser",
	})
	if !result || err != nil {
		t.Fatalf("Expected resolver.Check to return true, nil, but got %v, %v instead", result, err)
	}

}

func TestPostgresQueryBuilding(t *testing.T) {
	model := DefaultModel(t)

	storage, err := NewPostgresStorage(databaseURL)
	if err != nil {
		t.Fatalf("PostgresStorage creation failed: %v", err)
	}
	defer storage.Close()

	resolver, err := zanzigo.NewResolver(model, storage, 16)
	commands := resolver.CheckCommandsFor("doc", "viewer")
	query := NewPostgresCheckQuery(commands).(*postgresCheckQuery)
	expectedQueryString := `(SELECT 0 AS command_id, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='doc' AND object_id=$%d AND (object_relation='editor' OR object_relation='owner' OR object_relation='viewer') AND subject_type=$%d AND subject_id=$%d AND subject_relation=$%d) UNION ALL (SELECT 1 AS command_id, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='doc' AND object_id=$%d AND (object_relation='editor' OR object_relation='owner' OR object_relation='viewer') AND subject_relation <> '') UNION ALL (SELECT 2 AS command_id, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='doc' AND object_id=$%d AND (object_relation='parent') AND subject_type='folder')`
	if query.query != expectedQueryString {
		t.Fatalf("Expected computed query for commands to be `%s`, but got: %s", expectedQueryString, query.query)
	}

	ctx := context.Background()
	tuples, err := storage.QueryChecks(ctx, []zanzigo.CheckPayload{{
		Tuple: zanzigo.Tuple{
			ObjectType:     "doc",
			ObjectID:       "mydoc",
			ObjectRelation: "viewer",
			SubjectType:    "user",
			SubjectID:      "myuser",
		},
		Query:    query,
		Commands: commands,
	}})
	if err != nil {
		t.Fatalf("Expected query to not err: %v", err)
	}

	expectedTuples := []zanzigo.MarkedTuple{
		zanzigo.MarkedTuple{0, 2, zanzigo.Tuple{"doc", "mydoc", "parent", "folder", "myfolder", ""}},
	}
	if slices.CompareFunc(tuples, expectedTuples, func(a, b zanzigo.MarkedTuple) int {
		return cmp.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
	}) != 0 {
		t.Fatalf("Expected tuples %v, but got %v instead", expectedTuples, tuples)
	}
}
