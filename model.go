package zanzigo

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"
)

const (
	// We need a way to combine rules, we do that by setting 'anyOf' to [Rule.InheritIf].
	anyOfPlaceholder = "anyOf"
)

var (
	// TODO: doc
	ErrTypeUnknown = errors.New("Unknown type used in tuple")
	// TODO: doc
	ErrRelationUnknown = errors.New("Unknown relation used in tuple")
)

// A [Rule] is associated with a relationship of an authorization [Model] and defines when the requirement of the relationship is met.
// Without any fields specified the [Rule] will still be met by direct relations between an object and a subject.
type Rule struct {
	// If InheritIf is set the relation is inheritable from the specified relation,
	// e.g. `viewer` relationship inherited if subject is `editor`.
	InheritIf string `json:"inheritIf"`
	// If OfType is set, the relation specified by InheritIf needs to exist between the subject and the specified object-type.
	// This requires WithRelation to be set as there needs to be a WithRelation between object and an instance of OfType.
	OfType string `json:"ofType,omitempty"`
	// WithRelation defines, which relation needs to exist between OfType and the object to inherit the relationship status.
	WithRelation string `json:"withRelation,omitempty"`
	// Rules should not be set directly, but are public to make serializing rules easier.
	// The purpose of Rules is to allow combining rules. This should be done with functions such as [AnyOf] to properly mark the rule.
	Rules []Rule `json:"rules,omitempty"`
}

// AnyOf combines multiple rules into one rule, which when applied to a relation will
// result in a relation when any of the specified rules applies.
func AnyOf(rules ...Rule) Rule {
	return Rule{
		InheritIf: anyOfPlaceholder,
		Rules:     rules,
	}
}

// [InferredRule]s are precomputed for a given [Model] based on the [ObjectMap] and specified [Rule]s.
// Several types of rules exist, which require different traversals of the authorization model.
type Kind int

const (
	// Should never be used, but is used as a default value to make sure [Kind] is always specified.
	KindUnknown Kind = iota
	// A direct relationship between object and subject exists, user has direct access to object.
	KindDirect
	// A direct relationship between object and usersets exists, e.g. user is part of a group with access to the desired resource.
	KindDirectUserset
	// An indirect relationship between object and subject exists through another nested object, e.g. user has access to a folder containing a document.
	KindIndirect
)

// An InferredRule is the result of [Rule]s being prepared when a model is instiantiated via [NewModel].
// It is a flattened and preprocessed form of rules that is directly used to interact with [Storage]-implementations.
// It merges relations, splits them if multiple [Kind]s apply and it is important to note,
// that they are also sorted by [InferredRule.Kind] in [Model.InferredRules].
type InferredRule struct {
	Kind                  Kind
	Object                string
	Subject               string
	Relations             []string
	WithRelationToSubject []string
}

// A map of objects to map of relations to sorted rulesets of [InferredRule].
type InferredRuleMap map[string]map[string][]InferredRule

// The ObjectMap is the primary input of a model and is required to create a model
// and compute the inferred rules. The key is expected to be the object-type.
//
// The structure is inspired by [warrant].
//
// [warrant]: https://docs.warrant.dev/concepts/object-types/
type ObjectMap map[string]RelationMap

// RelationMap maps relationship-names to rules.
type RelationMap map[string]Rule

type validationMap map[string]map[string]struct{}

// A Model is the authorization model created from an [ObjectMap].
// During creation the model-definition provided by an [ObjectMap] is computed
// into a lower-lever ruleset of [InferredRule]s.
type Model struct {
	InferredRules InferredRuleMap
	validations   validationMap
}

// NewModel checks the [ObjectMap] for correctness and will infer the rules and
// prepare them for check-resolution.
func NewModel(objects ObjectMap) (*Model, error) {
	if err := validateObjects(objects); err != nil {
		return nil, err
	}
	return &Model{
		InferredRules: inferRules(objects),
		validations:   validations(objects),
	}, nil
}

// Rules are sorted direct first, indirect last.
// Returns the Rulset for a particular object-type and relation.
// If the object-type or relation does not exist, nil will be returned.
func (m *Model) RulesetFor(object, relation string) []InferredRule {
	relations, ok := m.InferredRules[object]
	if !ok {
		return nil
	}
	return relations[relation]
}

// TODO: add input validation to examples and document properly
func (m *Model) IsValid(t Tuple) bool {
	ors, ok := m.validations[t.ObjectType]
	if !ok {
		return false
	}
	if _, ok := ors[t.ObjectRelation]; !ok {
		return false
	}

	srs, ok := m.validations[t.SubjectType]
	if !ok {
		return false
	}
	if t.SubjectRelation != "" {
		if _, ok := srs[t.SubjectRelation]; !ok {
			return false
		}
	}
	return true
}

func validations(objects ObjectMap) validationMap {
	vs := validationMap{}
	for object, relations := range objects {
		vs[object] = map[string]struct{}{}
		for relation := range relations {
			vs[object][relation] = struct{}{}
		}
	}
	return vs
}

func validateObjects(objects ObjectMap) error {
	for object, relations := range objects {
		for relation, rule := range relations {
			err := validateRule(objects, object, relation, rule)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// TODO: add object, relation to every error
func validateRule(objects ObjectMap, object, relation string, rule Rule) error {
	// For AnyOf-rules we iterate over the subrules
	if rule.InheritIf == anyOfPlaceholder {
		if rule.OfType != "" {
			return fmt.Errorf("Invalid Rule: AnyOf-rules should not specify a OfType!")
		}
		if rule.WithRelation != "" {
			return fmt.Errorf("Invalid Rule: AnyOf-rules should not specify a WithRelation!")
		}
		for _, subrule := range rule.Rules {
			err := validateRule(objects, object, relation, subrule)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if (rule.OfType != "" && rule.WithRelation == "") || (rule.OfType == "" && rule.WithRelation != "") {
		return fmt.Errorf("Invalid Rule: Both OfType and WithRelation need to be specified or left empty.")
	}
	if rule.OfType != "" {
		if rule.InheritIf == "" {
			return fmt.Errorf("Invalid Rule: InheritIf is mandatory for when OfType is specified")
		}
		relations, ok := objects[rule.OfType]
		if !ok {
			return fmt.Errorf("Invalid Rule: Object type '%s' specified by OfType does not exist", rule.OfType)
		}
		_, ok = relations[rule.InheritIf]
		if !ok {
			return fmt.Errorf("InvalidRule: Relation '%s' of '%s' referenced by InheritIf does not exist", rule.WithRelation, rule.OfType)
		}

		_, ok = objects[object][rule.WithRelation]
		if !ok {
			return fmt.Errorf("InvalidRule: Relation '%s' of '%s' referenced by WithRelation does not exist", rule.WithRelation, object)
		}
		return nil // NO ADDITIONAL CHECKS REQUIRED
	}

	if rule.InheritIf != "" {
		_, ok := objects[object][rule.InheritIf]
		if !ok {
			return fmt.Errorf("InvalidRule: Relation '%s' of '%s' referenced by InheritIf does not exist", rule.InheritIf, object)
		}
	}

	return nil
}

func inferRules(objects ObjectMap) InferredRuleMap {
	inferredRules := InferredRuleMap{}
	// For each object and relation:
	for object, relations := range objects {
		inferredRules[object] = map[string][]InferredRule{}
		for relation, rule := range relations {
			// Infer the rules
			ruleset := inferRule(objects, object, relation, rule)
			// Remove duplicates by first sorting and then compacting
			slices.SortFunc(ruleset, func(a, b InferredRule) int {
				return cmp.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
			})
			ruleset = slices.CompactFunc(ruleset, func(a, b InferredRule) bool {
				return a.Object == b.Object && a.Subject == b.Subject &&
					strings.Join(a.Relations, "|") == strings.Join(b.Relations, "|") &&
					strings.Join(a.WithRelationToSubject, "|") == strings.Join(b.WithRelationToSubject, "|")
			})
			// Merge indirect rules into the minimal set required
			ruleset = mergeRulesWithRelationToSubject(ruleset)
			// Merge direct rules to minimize the amount of inferRules
			ruleset = mergeRulesRelations(ruleset)
			// Properly set the Kind based on the rule
			ruleset = expandRuleKinds(ruleset)
			// Sort the rules to make sure direct comes first
			ruleset = sortInferredRulesByKind(ruleset)

			inferredRules[object][relation] = ruleset
		}
	}
	return inferredRules
}

func inferRule(objects ObjectMap, object, relation string, rule Rule) []InferredRule {
	// Let's unfold the rules if AnyOf
	if rule.InheritIf == anyOfPlaceholder {
		rules := []InferredRule{}
		for _, subrule := range rule.Rules {
			rules = append(rules, inferRule(objects, object, relation, subrule)...)
		}
		return rules
	}

	// Always include direct-relationship
	rules := []InferredRule{
		InferredRule{
			Object:    object,
			Relations: []string{relation},
		},
	}

	// Inherit from current object
	if rule.InheritIf != "" && rule.OfType == "" {
		rules = append(rules, inferRule(objects, object, rule.InheritIf, objects[object][rule.InheritIf])...)
		// Inherit from other object type
	} else if rule.InheritIf != "" && rule.OfType != "" && rule.WithRelation != "" {
		rules = append(rules, InferredRule{
			Object:                object,
			Relations:             []string{rule.WithRelation},
			Subject:               rule.OfType,
			WithRelationToSubject: []string{rule.InheritIf},
		})
	}

	return rules
}

type replacement struct {
	i    int
	j    int
	rule InferredRule
}

// NOTE: Requires rules to be sorted and rule.Relations to NOT be merged yet!
func mergeRulesWithRelationToSubject(rules []InferredRule) []InferredRule {
	replacing := false
	replacements := []replacement{}
	current := replacement{}
	// We create a list of replacements to run through slices.Replace
	for i, rule := range rules {
		if !replacing && len(rule.WithRelationToSubject) > 0 {
			replacing = true
			current.i = i
			current.j = i + 1
			current.rule = rule
		} else if replacing && len(rule.WithRelationToSubject) > 0 && rule.Object == current.rule.Object && rule.Subject == current.rule.Subject && slices.Compare(rule.Relations, current.rule.Relations) == 0 {
			current.j = i + 1
			current.rule.WithRelationToSubject = append(current.rule.WithRelationToSubject, rule.WithRelationToSubject...)
		} else {
			replacing = false
			if len(current.rule.WithRelationToSubject) > 0 {
				replacements = append(replacements, current)
			}
		}
	}
	if replacing {
		replacements = append(replacements, current)
	}
	// Replace the rules as specified by replacements
	for _, r := range replacements {
		rules = slices.Replace(rules, r.i, r.j, r.rule)
	}
	return rules
}

// NOTE: Requires rules to be sorted!
func mergeRulesRelations(rules []InferredRule) []InferredRule {
	replacing := false
	replacements := []replacement{}
	current := replacement{}
	// We create a list of replacements to run through slices.Replace
	for i, rule := range rules {
		if !replacing && len(rule.WithRelationToSubject) == 0 && rule.Subject == "" {
			replacing = true
			current.i = i
			current.j = i + 1
			current.rule = rule
		} else if replacing && len(rule.WithRelationToSubject) == 0 && rule.Object == current.rule.Object && rule.Subject == "" {
			current.j = i + 1
			current.rule.Relations = append(current.rule.Relations, rule.Relations...)
		} else {
			replacing = false
			if len(current.rule.Relations) > 0 {
				replacements = append(replacements, current)
			}
		}
	}
	if replacing {
		replacements = append(replacements, current)
	}
	// Replace the rules as specified by replacements
	for _, r := range replacements {
		rules = slices.Replace(rules, r.i, r.j, r.rule)
	}
	return rules
}

// Adds Kind to rules based on the shape. Two rules will be create for a "direct-shaped" rule to also check usersets.
func expandRuleKinds(rules []InferredRule) []InferredRule {
	expanded := []InferredRule{}
	for _, rule := range rules {
		if len(rule.WithRelationToSubject) > 0 { // INDIRECT
			rule.Kind = KindIndirect
			expanded = append(expanded, rule)
		} else { // DIRECT
			rule.Kind = KindDirect
			ruleUserset := rule
			ruleUserset.Kind = KindDirectUserset
			expanded = append(expanded, rule, ruleUserset)
		}
	}
	return expanded
}

func sortInferredRulesByKind(rules []InferredRule) []InferredRule {
	slices.SortFunc(rules, func(a, b InferredRule) int {
		return int(a.Kind) - int(b.Kind)
	})
	return rules
}
