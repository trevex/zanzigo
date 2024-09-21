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

	filepath = os.Getenv("TEST_PEBBLE_DIR")

	if filepath == "" {
		_ = os.RemoveAll("./pebble")
		filepath = "./pebble"
	}

	var err error
	storage, err = NewPebbleStorage(filepath)
	if err != nil {
		log.Fatalf("PebbleStorage creation failed: %v", err)
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

func TestPebbleWithTestSuite(t *testing.T) {
	testsuite.RunTestAll(t, map[string]testsuite.TestConfig{
		"pebble": {
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

func BenchmarkPebble(b *testing.B) {
	testsuite.RunBenchmarkAll(b, map[string]zanzigo.Storage{
		"pebble": storage,
	})
}
