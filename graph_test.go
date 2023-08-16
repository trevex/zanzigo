package zanzigo_test

import (
	"testing"

	"github.com/trevex/zanzigo"
)

func TestGraph(t *testing.T) {
	// ...
	graph := zanzigo.TypeMap{
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

	storage, err := zanzigo.NewPostgresStorage("")

	err := storage.Write(zanzigo.Tuple{
		SubjectType: "user",
		SubjectID:   "myuser",
		Relation:    "member",
		ObjectType:  "group",
		ObjectID:    "mygroup",
	})

	err := storage.Write(zanzigo.Tuple{
		SubjectType: "folder",
		SubjectID:   "myfolder",
		Relation:    "parent",
		ObjectType:  "doc",
		ObjectID:    "mydoc",
	})

	err := storage.Write(zanzigo.Tuple{
		SubjectType: "tupleset",
		SubjectID:   "group:mygroup#member",
		Relation:    "viewer",
		ObjectType:  "folder",
		ObjectID:    "myfolder",
	})

	resolver, err := graph.Resolver(storage)
	result, err := resolver.Check(zanzigo.Tuple{
		SubjectType: "user",
		SubjectID:   "myuser",
		Relation:    "viewer",
		ObjectType:  "doc", // Tuple set!?
		ObjectID:    "mydoc",
	})

}
