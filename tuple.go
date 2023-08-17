package zanzigo

// / ⟨tuple⟩ ::= ⟨object⟩‘#’⟨relation⟩‘@’⟨user⟩
type Tuple struct {
	/// ⟨object⟩ ::= ⟨namespace⟩‘:’⟨object id⟩
	Object string `json:"object"`
	/// ⟨relation⟩
	Relation string `json:"relation"`
	/// Flag to indicate whether `.User` is a user or userset
	IsUserset bool `json:"isUserset"`
	/// ⟨user⟩ ::= ⟨namespace⟩‘:’⟨user id⟩ | ⟨userset⟩
	/// ⟨userset⟩ ::= ⟨object⟩‘#’⟨relation⟩
	User string `json:"user"`
}
