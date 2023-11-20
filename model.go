package zanzigo

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"
)

const (
	AnyOfPlaceholder = "anyOf"
)

var (
	ErrTypeUnknown     = errors.New("Unknown type used in tuple")
	ErrRelationUnknown = errors.New("Unknown relation used in tuple")
)

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

type Kind int

const (
	KindUnknown Kind = iota
	KindDirect
	KindDirectUserset
	KindIndirect
)

type InferredRule struct {
	Kind                  Kind
	Object                string
	Subject               string // Order of fields important to ensure proper ordering in mergeRules
	Relations             []string
	WithRelationToSubject []string
}

type InferredRuleMap map[string]map[string][]InferredRule

// Inspired by https://docs.warrant.dev/concepts/object-types/
type ObjectMap map[string]RelationMap

type RelationMap map[string]Rule

type Model struct {
	Objects       ObjectMap
	InferredRules InferredRuleMap
}

func NewModel(objects ObjectMap) (*Model, error) {
	// TODO: check objects for correctness
	return &Model{
		Objects:       objects,
		InferredRules: inferRules(objects),
	}, nil
}

// Rules are sorted direct first, indirect last.
func (m *Model) RulesetFor(object, relation string) []InferredRule {
	relations, ok := m.InferredRules[object]
	if !ok {
		return nil
	}
	return relations[relation]
}

func inferRules(objects ObjectMap) InferredRuleMap {
	mergedRules := InferredRuleMap{}
	for object, relations := range objects {
		mergedRules[object] = map[string][]InferredRule{}
		for relation, rule := range relations {
			rules := inferRule(objects, object, relation, rule)
			// Remove duplicates
			slices.SortFunc(rules, func(a, b InferredRule) int {
				return cmp.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
			})
			rules = slices.CompactFunc(rules, func(a, b InferredRule) bool {
				return a.Object == b.Object && a.Subject == b.Subject &&
					strings.Join(a.Relations, "|") == strings.Join(b.Relations, "|") &&
					strings.Join(a.WithRelationToSubject, "|") == strings.Join(b.WithRelationToSubject, "|")
			})
			rules = mergeRulesWithRelationToSubject(rules)
			rules = mergeRulesRelations(rules)
			rules = expandRuleKinds(rules)
			rules = sortInferredRulesByKind(rules)

			mergedRules[object][relation] = rules
		}
	}
	return mergedRules
}

func inferRule(objects ObjectMap, object, relation string, rule Rule) []InferredRule {
	// Let's unfold the rules if AnyOf
	if rule.InheritIf == AnyOfPlaceholder {
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

// NOTE: Requires rules to be sorted and rule.Relations to NOT be merged yet
func mergeRulesWithRelationToSubject(rules []InferredRule) []InferredRule {
	replacing := false
	replacements := []replacement{}
	current := replacement{}
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
	for _, r := range replacements {
		rules = slices.Replace(rules, r.i, r.j, r.rule)
	}
	return rules
}

// NOTE: Requires rules to be sorted
// TODO: Does it require another deduplication? What about merging relations of WithRelationToSubject?
func mergeRulesRelations(rules []InferredRule) []InferredRule {
	replacing := false
	replacements := []replacement{}
	current := replacement{}
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
	for _, r := range replacements {
		rules = slices.Replace(rules, r.i, r.j, r.rule)
	}
	return rules
}

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
