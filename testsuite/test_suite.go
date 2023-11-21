package testsuite

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trevex/zanzigo"
	"golang.org/x/exp/slices"
)

var Model = func() *zanzigo.Model {
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
		log.Fatalf("Expected storage.Write not to fail: %v", err)
	}
	return model
}()

func Load(ctx context.Context, storage zanzigo.Storage) error {
	_, err := storage.Read(ctx, zanzigo.Tuple{
		ObjectType:     "doc",
		ObjectID:       "mydoc",
		ObjectRelation: "owner",
		SubjectType:    "user",
		SubjectID:      "myowner",
	})
	if err == nil {
		fmt.Println(">>> WARN  Last tuple already exists skipping loading of test data!")
		return nil
	}

	err = storage.Write(ctx, zanzigo.Tuple{
		ObjectType:     "group",
		ObjectID:       "mygroup",
		ObjectRelation: "member",
		SubjectType:    "user",
		SubjectID:      "myuser",
	})
	if err != nil {
		return err
	}

	err = storage.Write(ctx, zanzigo.Tuple{
		ObjectType:     "doc",
		ObjectID:       "mydoc",
		ObjectRelation: "parent",
		SubjectType:    "folder",
		SubjectID:      "myfolder",
	})
	if err != nil {
		return err
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
		return err
	}

	err = storage.Write(ctx, zanzigo.Tuple{
		ObjectType:     "doc",
		ObjectID:       "mydoc",
		ObjectRelation: "owner",
		SubjectType:    "user",
		SubjectID:      "myowner",
	})
	if err != nil {
		return err
	}

	return nil
}

type TestConfig struct {
	Storage      zanzigo.Storage
	Expectations Expectations
}

func RunTestAll(t *testing.T, configs map[string]TestConfig) {
	for name, config := range configs {
		t.Run(name, func(t *testing.T) {
			RunTest(t, config.Storage, config.Expectations)
		})
	}
}

type Expectations struct {
	UserdataCheckQueryTuple zanzigo.MarkedTuple
}

func RunTest(t *testing.T, storage zanzigo.Storage, expectations Expectations) {
	resolver, err := zanzigo.NewSequentialResolver(Model, storage, 16)
	require.NoError(t, err)

	t.Run("checks", func(t *testing.T) {
		result, err := resolver.Check(context.Background(), zanzigo.Tuple{
			ObjectType:     "doc",
			ObjectID:       "mydoc",
			ObjectRelation: "viewer",
			SubjectType:    "user",
			SubjectID:      "myuser",
		})
		require.NoError(t, err)
		require.True(t, result)

		result, err = resolver.Check(context.Background(), zanzigo.Tuple{
			ObjectType:     "doc",
			ObjectID:       "mydoc",
			ObjectRelation: "editor",
			SubjectType:    "user",
			SubjectID:      "myuser",
		})
		require.NoError(t, err)
		require.False(t, result)
	})

	t.Run("userdata", func(t *testing.T) {
		ruleset := resolver.RulesetFor("doc", "viewer")
		userdata, err := storage.PrepareRuleset("doc", "viewer", ruleset)
		require.NoError(t, err)

		ctx := context.Background()
		tuples, err := storage.QueryChecks(ctx, []zanzigo.Check{{
			Tuple: zanzigo.Tuple{
				ObjectType:     "doc",
				ObjectID:       "mydoc",
				ObjectRelation: "viewer",
				SubjectType:    "user",
				SubjectID:      "myuser",
			},
			Userdata: userdata,
			Ruleset:  ruleset,
		}})
		require.NoError(t, err)

		expectedTuples := []zanzigo.MarkedTuple{expectations.UserdataCheckQueryTuple}
		if slices.CompareFunc(tuples, expectedTuples, func(a, b zanzigo.MarkedTuple) int {
			return cmp.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
		}) != 0 {
			t.Fatalf("Expected tuples %v, but got %v instead", expectedTuples, tuples)
		}
	})

}

func RunBenchmarkAll(b *testing.B, storages map[string]zanzigo.Storage) {
	for name, storage := range storages {
		b.Run(name, func(b *testing.B) {
			RunBenchmark(b, storage)
		})
	}
}

func RunBenchmark(b *testing.B, storage zanzigo.Storage) {
	resolver, err := zanzigo.NewSequentialResolver(Model, storage, 16)
	require.NoError(b, err)

	// TODO: generate junk data?

	b.Run("indirect_nested_4", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := resolver.Check(context.Background(), zanzigo.Tuple{
				ObjectType:     "doc",
				ObjectID:       "mydoc",
				ObjectRelation: "viewer",
				SubjectType:    "user",
				SubjectID:      "myuser",
			})
			require.NoError(b, err)
		}
	})
	b.Run("direct", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := resolver.Check(context.Background(), zanzigo.Tuple{
				ObjectType:     "doc",
				ObjectID:       "mydoc",
				ObjectRelation: "viewer",
				SubjectType:    "user",
				SubjectID:      "myowner",
			})
			require.NoError(b, err)
		}
	})
}
