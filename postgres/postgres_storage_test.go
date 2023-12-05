package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/trevex/zanzigo"
	"github.com/trevex/zanzigo/testsuite"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
)

var (
	databaseURL = ""
	storage     zanzigo.Storage
)

func TestMain(m *testing.M) {
	var (
		pool     *dockertest.Pool
		resource *dockertest.Resource
		err      error
	)

	databaseURL = os.Getenv("TEST_POSTGRES_DATABASE_URL")

	if databaseURL == "" {
		pool, err = dockertest.NewPool("")
		if err != nil {
			log.Fatalf("Could not connect to docker: %s", err)
		}

		resource, err = pool.RunWithOptions(&dockertest.RunOptions{
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
		_ = resource.Expire(300) // In any case container should be killed in 5min

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
	}

	if err := RunMigrations(databaseURL); err != nil {
		log.Fatalf("Could not migrate db: %s", err)
	}

	storage, err = NewPostgresStorage(databaseURL)
	if err != nil {
		log.Fatalf("PostgresStorage creation failed: %v", err)
	}

	// Let's load the testsuite-data
	err = testsuite.Load(context.Background(), storage)
	if err != nil {
		log.Fatalf("Failed loading data into storage: %v", err)
	}

	code := m.Run()

	// os.Exit doesn't care for defer, so let's explicitly purge and close...
	storage.Close()
	if pool != nil {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}

	os.Exit(code)
}

func TestPostgresWithTestSuite(t *testing.T) {
	storageFunctions, err := NewPostgresStorage(databaseURL, UseFunctions())
	if err != nil {
		t.Fatalf("PostgresStorage creation failed: %v", err)
	}
	defer storageFunctions.Close()
	testsuite.RunTestAll(t, map[string]testsuite.TestConfig{
		"queries": {
			Storage: storage,
			Expectations: testsuite.Expectations{
				UserdataCheckQueryTuple: zanzigo.MarkedTuple{
					CheckIndex: 0,
					RuleIndex:  2,
					Tuple:      zanzigo.TupleString("doc:mydoc#parent@folder:myfolder"),
				},
			},
		},
		"functions": {
			Storage: storageFunctions,
			Expectations: testsuite.Expectations{
				UserdataCheckQueryTuple: zanzigo.MarkedTuple{
					CheckIndex: 0,
					RuleIndex:  0,
					Tuple:      zanzigo.TupleString("doc:mydoc#viewer@user:myuser"),
				},
			},
		},
	})
}

func BenchmarkPostgres(b *testing.B) {
	storageFunctions, err := NewPostgresStorage(databaseURL, UseFunctions())
	require.NoError(b, err)

	defer storageFunctions.Close()
	testsuite.RunBenchmarkAll(b, map[string]zanzigo.Storage{
		"queries":   storage,
		"functions": storageFunctions,
	})
}

func standardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
