package zanzigo

const (
	AnyOfPlaceholder = "anyOf"
)

type TypeMap map[string]RelationMap

type RelationMap map[string]Rule

type Rule struct {
	InheritIf    string
	OfType       string
	WithRelation string
	Rules        []Rule
}

func AnyOf(rules ...Rule) Rule {
	return Rule{
		InheritIf: AnyOfPlaceholder,
		Rules:     rules,
	}
}

func (tm TypeMap) Validate() error {
	return nil
}
