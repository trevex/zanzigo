package zanzigo

const (
	AnyOfPlaceholder = "anyOf"
)

// Inspired by https://docs.warrant.dev/concepts/object-types/
type TypeMap map[string]RelationMap

type RelationMap map[string]Rule

type Rule struct {
	InheritIf    string `json:"inheritIf"`
	OfType       string `json:"ofType,omitempty"`
	WithRelation string `json:"withRelation,omitempty"`
	Rules        []Rule `json:"rules,omitempty"`
}

func AnyOf(rules ...Rule) Rule {
	return Rule{
		InheritIf: AnyOfPlaceholder,
		Rules:     rules,
	}
}

func (tm TypeMap) Resolver(storage Storage, executer Executer) (*Resolver, error) {
	return &Resolver{}, nil
}

type Resolver struct{}

func (r *Resolver) Check(t Tuple) (bool, error) {
	return true, nil
}
