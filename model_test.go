package zanzigo_test

import (
	"cmp"
	"fmt"
	"slices"
	"testing"

	"github.com/trevex/zanzigo"
)

func TestModel(t *testing.T) {
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

	rules := model.RulesFor("doc", "viewer")
	expected := []zanzigo.MergedRule{
		{Object: "doc", Relations: []string{"editor", "owner", "viewer"}},
		{Object: "doc", Relations: []string{"parent"}, Subject: "folder", WithRelationToSubject: []string{"editor", "owner", "viewer"}},
	}
	if slices.CompareFunc(rules, expected, func(a, b zanzigo.MergedRule) int {
		return cmp.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
	}) != 0 {
		t.Fatalf("Expected rules %v, but got %v instead", expected, rules)
	}
}
