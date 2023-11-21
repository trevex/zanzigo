package zanzigo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTupleString(t *testing.T) {
	t1 := TupleString("doc:mydoc#viewer@user:myuser")
	require.Equal(t, Tuple{
		ObjectType:     "doc",
		ObjectID:       "mydoc",
		ObjectRelation: "viewer",
		SubjectType:    "user",
		SubjectID:      "myuser",
	}, t1)

	t2 := TupleString("doc:mydoc#editor@group:mygroup#member")
	require.Equal(t, Tuple{
		ObjectType:      "doc",
		ObjectID:        "mydoc",
		ObjectRelation:  "editor",
		SubjectType:     "group",
		SubjectID:       "mygroup",
		SubjectRelation: "member",
	}, t2)
}
