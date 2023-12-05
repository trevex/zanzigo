package sqlite3

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/trevex/zanzigo"
	testsuite "github.com/trevex/zanzigo/storage"
)

var (
	filepath = ""
	storage  zanzigo.Storage
)

func TestMain(m *testing.M) {

	filepath = os.Getenv("TEST_SQLITE_FILE")

	if filepath == "" {
		_ = os.Remove("./test.db")
		filepath = "./test.db"
	}

	if err := RunMigrations(filepath); err != nil {
		log.Fatalf("Could not migrate db: %s", err)
	}

	var err error
	storage, err = NewSQLite3Storage(filepath)
	if err != nil {
		log.Fatalf("SQLite3Storage creation failed: %v", err)
	}

	// Let's load the testsuite-data
	err = testsuite.Load(context.Background(), storage)
	if err != nil {
		log.Fatalf("Failed loading data into storage: %v", err)
	}

	code := m.Run()

	// os.Exit doesn't care for defer, so let's explicitly purge and close...
	storage.Close()

	os.Exit(code)
}

func TestSQLite3WithTestSuite(t *testing.T) {
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
	})
}

func BenchmarkSQLite3(b *testing.B) {
	testsuite.RunBenchmarkAll(b, map[string]zanzigo.Storage{
		"queries": storage,
	})
}
