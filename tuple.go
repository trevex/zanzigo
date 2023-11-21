package zanzigo

import "strings"

// / ⟨tuple⟩ ::= ⟨object⟩‘#’⟨relation⟩‘@’⟨user⟩
type Tuple struct {
	/// ⟨object⟩ ::= ⟨namespace⟩‘:’⟨object id⟩
	ObjectType string `json:"object_type"`
	ObjectID   string `json:"object_id"`
	/// ⟨relation⟩
	ObjectRelation string `json:"relation"`
	/// ⟨user⟩ ::= ⟨namespace⟩‘:’⟨user id⟩ | ⟨userset⟩
	SubjectType string `json:"user_type"`
	SubjectID   string `json:"user_id"`
	/// ⟨userset⟩ ::= ⟨object⟩‘#’⟨relation⟩
	SubjectRelation string `json:"user_relation"`
}

var EmptyTuple = Tuple{}

// Parses a string in Zanzibar-format and returns the resulting tuple.
// If the string is malformed, EmptyTuple will be returned.
//
// Examples for input are: 'doc:mydoc#viewer@user:myuser' or 'doc:mydoc#editor@group:mygroup#member'
func TupleString(s string) Tuple {
	splits := strings.Split(s, "@")
	if len(splits) != 2 {
		return EmptyTuple
	}
	object := splits[0]
	subject := splits[1]

	// Process object-part first
	splits = strings.Split(object, "#")
	if len(splits) != 2 {
		return EmptyTuple
	}
	objectRelation := splits[1]
	object = splits[0]
	splits = strings.Split(object, ":")
	if len(splits) != 2 {
		return EmptyTuple
	}
	objectType := splits[0]
	objectID := splits[1]

	// Next process subject-part, but relation is optional
	subjectRelation := ""
	splits = strings.Split(subject, "#")
	if len(splits) == 2 {
		subjectRelation = splits[1]
	}
	subject = splits[0]
	splits = strings.Split(subject, ":")
	subjectType := splits[0]
	subjectID := splits[1]

	return Tuple{
		ObjectType:      objectType,
		ObjectID:        objectID,
		ObjectRelation:  objectRelation,
		SubjectType:     subjectType,
		SubjectID:       subjectID,
		SubjectRelation: subjectRelation,
	}
}
