package zanzigo

import (
	"context"
	"errors"
	"fmt"
)

type UserdataMap map[string]map[string]Userdata

// During creation a set of static commands are precomputed which will also be passed on to the [Storage]-backend via [Storage.PrepareForCheckCommands].
type Resolver struct {
	storage  Storage
	userdata UserdataMap
	rules    InferredRuleMap
	maxDepth int
}

// NewResolver creates a new resolver for the particular [Model] using the designated [Storage]-implementation.
// The main purpose of the [Resolver] is to traverse the ReBAC-policies and check whether a [Tuple] is authorized or not.
// During creation a set of static commands are precomputed which will also be passed on to the [Storage]-backend via [Storage.PrepareForCheckCommands].
func NewResolver(model *Model, storage Storage, maxDepth int) (*Resolver, error) {
	userdata, err := prepareUserdataForRules(storage, model.InferredRules)
	return &Resolver{
		storage, userdata, model.InferredRules, maxDepth,
	}, err
}

func (r *Resolver) Check(ctx context.Context, t Tuple) (bool, error) {
	ruleset, ok := r.rules[t.ObjectType][t.ObjectRelation]
	if !ok {
		return false, fmt.Errorf("failed to find %s > %s in query map", t.ObjectType, t.ObjectRelation)
	}
	// needs to exist, otherwise .NewResolver would have failed
	userdata := r.userdata[t.ObjectType][t.ObjectRelation]
	depth := 0
	return r.check(ctx, []Check{{
		Tuple:    t,
		Ruleset:  ruleset,
		Userdata: userdata,
	}}, depth)
}

func (r *Resolver) check(ctx context.Context, checks []Check, depth int) (bool, error) {
	if len(checks) == 0 {
		return false, nil
	}
	if depth > r.maxDepth {
		return false, errors.New("max depth exceeded")
	}
	depth += 1

	markedTuples, err := r.storage.QueryChecks(ctx, checks)
	if err != nil {
		return false, err
	}

	nextChecks := []Check{}
	// Returned marked tuples are ordered by .CommandID and commands are ordered with directs first,
	// so we can exit early if we find a direct relationship.
	for _, mt := range markedTuples {
		check := checks[mt.CheckID]
		rule := check.Ruleset[mt.RuleID]
		switch rule.Kind {
		case KindDirect:
			return true, nil
		case KindDirectUserset:
			ruleset, ok := r.rules[mt.SubjectType][mt.SubjectRelation]
			if !ok {
				return false, fmt.Errorf("failed to find %s > %s in query map", mt.SubjectType, mt.SubjectRelation)
			}
			userdata := r.userdata[mt.SubjectType][mt.SubjectRelation]
			nextChecks = append(nextChecks, Check{
				Tuple: Tuple{
					ObjectType:      mt.SubjectType,
					ObjectID:        mt.SubjectID,
					ObjectRelation:  mt.SubjectRelation,
					SubjectType:     check.Tuple.SubjectType,
					SubjectID:       check.Tuple.SubjectID,
					SubjectRelation: check.Tuple.SubjectRelation,
				},
				Ruleset:  ruleset,
				Userdata: userdata,
			})
		case KindIndirect: // TODO: THIS CAN BE USERSET!?
			relations := rule.WithRelationToSubject
			for _, relation := range relations {
				ruleset, ok := r.rules[mt.SubjectType][relation]
				if !ok {
					return false, fmt.Errorf("failed to find %s > %s in query map", mt.SubjectType, relation)
				}
				userdata := r.userdata[mt.SubjectType][relation]
				nextChecks = append(nextChecks, Check{
					Tuple: Tuple{
						ObjectType:      mt.SubjectType,
						ObjectID:        mt.SubjectID,
						ObjectRelation:  relation,
						SubjectType:     check.Tuple.SubjectType,
						SubjectID:       check.Tuple.SubjectID,
						SubjectRelation: check.Tuple.SubjectRelation,
					},
					Ruleset:  ruleset,
					Userdata: userdata,
				})
			}
		default:
			panic("unreachable")
		}
	}

	return r.check(ctx, nextChecks, depth)
}

func (r *Resolver) RulesetFor(object, relation string) []InferredRule {
	return r.rules[object][relation]
}

func prepareUserdataForRules(storage Storage, inferredRules InferredRuleMap) (UserdataMap, error) {
	// Let's create the commands first
	userdata := UserdataMap{}
	for object, relations := range inferredRules {
		userdata[object] = map[string]Userdata{}
		for relation, rules := range relations {
			var err error
			userdata[object][relation], err = storage.PrepareRuleset(object, relation, rules)
			if err != nil {
				return nil, err
			}
		}
	}
	return userdata, nil
}
