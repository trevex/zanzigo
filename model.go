package zanzigo

import "context"

const (
	AnyOfPlaceholder = "anyOf"
)

// Inspired by https://docs.warrant.dev/concepts/object-types/
type Model map[string]RelationMap

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

func (model Model) Resolver(storage Storage, executer Executer) (*Resolver, error) {
	// TODO: Check TypeMap for errors
	return &Resolver{
		model, storage, executer,
	}, nil
}

type Resolver struct {
	model    Model
	storage  Storage
	executer Executer
}

func (r *Resolver) Check(ctx context.Context, t Tuple) (bool, error) {
	return true, nil
}
