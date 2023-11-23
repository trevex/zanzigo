package zanzigo_test

import (
	"cmp"
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	require.True(t, model.IsValid(zanzigo.TupleString("doc:mydoc#viewer@user:myuser")))
	require.True(t, model.IsValid(zanzigo.TupleString("doc:mydoc#viewer@group:mygroup#member")))
	require.False(t, model.IsValid(zanzigo.TupleString("wrong:mydoc#viewer@group:mygroup#member")))
	require.False(t, model.IsValid(zanzigo.TupleString("doc:mydoc#wrong@group:mygroup#member")))
	require.False(t, model.IsValid(zanzigo.TupleString("doc:mydoc#viewer@wrong:mygroup#member")))
	require.False(t, model.IsValid(zanzigo.TupleString("doc:mydoc#viewer@group:mygroup#wrong")))

	ruleset := model.RulesetFor("doc", "viewer")
	expected := []zanzigo.InferredRule{
		{Kind: zanzigo.KindDirect, Object: "doc", Relations: []string{"editor", "owner", "viewer"}},
		{Kind: zanzigo.KindDirectUserset, Object: "doc", Relations: []string{"editor", "owner", "viewer"}},
		{Kind: zanzigo.KindIndirect, Object: "doc", Relations: []string{"parent"}, Subject: "folder", WithRelationToSubject: []string{"editor", "owner", "viewer"}},
	}

	if slices.CompareFunc(ruleset, expected, func(a, b zanzigo.InferredRule) int {
		return cmp.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
	}) != 0 {
		t.Fatalf("Expected ruleset %v, but got %v instead", expected, ruleset)
	}
}
