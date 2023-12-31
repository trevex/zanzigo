package zanzigo

import (
	"context"
	"errors"
	"fmt"
)

// A map of object-types to relations to Userdata.
type UserdataMap map[string]map[string]Userdata

// A Resolver uses a [Model] and [Storage]-implementation to execute relationship checks.
type Resolver struct {
	storage  Storage
	userdata UserdataMap
	rules    InferredRuleMap
	maxDepth int
}

// NewResolver creates a new resolver for the particular [Model] using the designated [Storage]-implementation.
// The main purpose of a [Resolver] is to traverse the ReBAC-policies and check whether a [Tuple] is authorized or not.
// During creation the inferred rules of the [Model] are used to precompute storage-specific [Userdata] that can be used
// to speed up checks (when calling Storage.QueryChecks internally).
// When Check is called the [Userdata] is passed on to the [Storage]-implementation as part of the [Check].
//
// maxDepth limits the depth of the traversal of the authorization-model during checks.
func NewResolver(model *Model, storage Storage, maxDepth int) (*Resolver, error) {
	userdata, err := prepareUserdataForRules(storage, model.InferredRules)
	return &Resolver{
		storage, userdata, model.InferredRules, maxDepth,
	}, err
}

// Checks whether the relationship stated by [Tuple] t is true.
func (r *Resolver) Check(ctx context.Context, t Tuple) (bool, error) {
	// TODO: check if tuple is valid!? Does relation exist!
	ruleset, ok := r.rules[t.ObjectType][t.ObjectRelation]
	if !ok {
		return false, fmt.Errorf("failed to find %s > %s in query map", t.ObjectType, t.ObjectRelation)
	}
	// needs to exist, otherwise `NewResolver` would have failed
	userdata := r.userdata[t.ObjectType][t.ObjectRelation]
	depth := 0 // We start from zero and dive "upwards"
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

	// TODO: resolver should support caching:
	//       1. for each direct-rule of checks, check cache
	//       2. for marked tuples returned: if direct, store in cache
	// Alternatively, should rules with context be cached with results?

	nextChecks := []Check{}
	// Returned marked tuples are ordered by .RuleIndex and rules are ordered with directs first,
	// so we can exit early if we find a direct relationship before continuing to subsequent checks.
	for _, mt := range markedTuples {
		check := checks[mt.CheckIndex]
		rule := check.Ruleset[mt.RuleIndex]
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
		case KindIndirect:
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

// Returns an inferred ruleset for the given object-type and relation.
func (r *Resolver) RulesetFor(object, relation string) []InferredRule {
	return r.rules[object][relation]
}

func prepareUserdataForRules(storage Storage, inferredRules InferredRuleMap) (UserdataMap, error) {
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
