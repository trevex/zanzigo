package zanzigo

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
