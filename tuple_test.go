package zanzigo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTupleString(t *testing.T) {
	input1 := "doc:mydoc#viewer@user:myuser"
	t1 := TupleString(input1)
	require.Equal(t, Tuple{
		ObjectType:     "doc",
		ObjectID:       "mydoc",
		ObjectRelation: "viewer",
		SubjectType:    "user",
		SubjectID:      "myuser",
	}, t1)
	out1 := t1.ToString()
	require.Equal(t, input1, out1)

	input2 := "doc:mydoc#editor@group:mygroup#member"
	t2 := TupleString(input2)
	require.Equal(t, Tuple{
		ObjectType:      "doc",
		ObjectID:        "mydoc",
		ObjectRelation:  "editor",
		SubjectType:     "group",
		SubjectID:       "mygroup",
		SubjectRelation: "member",
	}, t2)
	out2 := t2.ToString()
	require.Equal(t, input2, out2)
}
