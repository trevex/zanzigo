package zanzigo

import (
	"context"
	"errors"
	"fmt"
)

// A SequentialResolver uses a [Model] and [Storage]-implementation to execute relationship checks.
type SequentialResolver struct {
	storage  Storage
	userdata UserdataMap
	rules    InferredRuleMap
	maxDepth int
}

// NewSequentialResolver creates a new resolver for the particular [Model] using the designated [Storage]-implementation.
// During creation the inferred rules of the [Model] are used to precompute storage-specific [Userdata] that can be used
// to speed up checks (when calling Storage.QueryChecks internally).
// When Check is called the [Userdata] is passed on to the [Storage]-implementation as part of the [Check].
//
// maxDepth limits the depth of the traversal of the authorization-model during checks.
func NewSequentialResolver(model *Model, storage Storage, maxDepth int) (*SequentialResolver, error) {
	userdata, err := prepareUserdataForRules(storage, model.InferredRules)
	return &SequentialResolver{
		storage, userdata, model.InferredRules, maxDepth,
	}, err
}

// Checks whether the relationship stated by [Tuple] t is true.
func (r *SequentialResolver) Check(ctx context.Context, t Tuple) (bool, error) {
	ruleset, ok := r.rules[t.ObjectType][t.ObjectRelation]
	if !ok {
		return false, fmt.Errorf("failed to find %s > %s in query map", t.ObjectType, t.ObjectRelation)
	}
	// needs to exist, otherwise `NewSequentialResolver` would have failed
	userdata := r.userdata[t.ObjectType][t.ObjectRelation]
	depth := 0 // We start from zero and dive "upwards"
	return r.check(ctx, []Check{{
		Tuple:    t,
		Ruleset:  ruleset,
		Userdata: userdata,
	}}, depth)
}

func (r *SequentialResolver) check(ctx context.Context, checks []Check, depth int) (bool, error) {
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
func (r *SequentialResolver) RulesetFor(object, relation string) []InferredRule {
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
