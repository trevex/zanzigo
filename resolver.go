package zanzigo

import (
	"context"
	"errors"
	"fmt"
	"slices"
)

type Kind int

const (
	KindUnknown Kind = iota
	KindDirect
	KindDirectUserset
	KindIndirect
)

type Check interface {
	Kind() Kind
	Rule() MergedRule
}

type CheckDirect struct {
	MergedRule
}

func (c *CheckDirect) Kind() Kind       { return KindDirect }
func (c *CheckDirect) Rule() MergedRule { return c.MergedRule }

type CheckDirectUserset struct {
	MergedRule
}

func (c *CheckDirectUserset) Kind() Kind       { return KindDirectUserset }
func (c *CheckDirectUserset) Rule() MergedRule { return c.MergedRule }

// CheckIndirect will first try to find subjects for r.Object r.Relations
// and then traverse them independently. (e.g. recursive .Check)
type CheckIndirect struct {
	MergedRule
}

func (c *CheckIndirect) Kind() Kind       { return KindIndirect }
func (c *CheckIndirect) Rule() MergedRule { return c.MergedRule }

type PreparedChecks struct {
	Userdata Userdata
	Checks   []Check
}

type CheckMap map[string]map[string]PreparedChecks

// During creation a set of static commands are precomputed which will also be passed on to the [Storage]-backend via [Storage.PrepareForCheckCommands].
type Resolver struct {
	storage  Storage
	checkMap CheckMap
	maxDepth int
}

// NewResolver creates a new resolver for the particular [Model] using the designated [Storage]-implementation.
// The main purpose of the [Resolver] is to traverse the ReBAC-policies and check whether a [Tuple] is authorized or not.
// During creation a set of static commands are precomputed which will also be passed on to the [Storage]-backend via [Storage.PrepareForCheckCommands].
func NewResolver(model *Model, storage Storage, maxDepth int) (*Resolver, error) {
	checkMap, err := prepareCheckMapForModel(model, storage)
	return &Resolver{
		storage, checkMap, maxDepth,
	}, err
}

func (r *Resolver) Check(ctx context.Context, t Tuple) (bool, error) {
	pc, ok := r.checkMap[t.ObjectType][t.ObjectRelation]
	if !ok {
		return false, fmt.Errorf("failed to find %s > %s in query map", t.ObjectType, t.ObjectRelation)
	}
	depth := 0
	return r.check(ctx, []CheckRequest{{
		Tuple:    t,
		Userdata: pc.Userdata,
		Checks:   pc.Checks,
	}}, depth)
}

func (r *Resolver) check(ctx context.Context, crs []CheckRequest, depth int) (bool, error) {
	if len(crs) == 0 {
		return false, nil
	}
	if depth > r.maxDepth {
		return false, errors.New("max depth exceeded")
	}
	depth += 1

	markedTuples, err := r.storage.QueryChecks(ctx, crs)
	if err != nil {
		return false, err
	}

	nextChecks := []CheckRequest{}
	// Returned marked tuples are ordered by .CommandID and commands are ordered with directs first,
	// so we can exit early if we find a direct relationship.
	for _, mt := range markedTuples {
		cp := crs[mt.RequestID]
		check := cp.Checks[mt.CheckID]
		switch check.Kind() {
		case KindDirect:
			return true, nil
		case KindDirectUserset:
			pc, ok := r.checkMap[mt.SubjectType][mt.SubjectRelation]
			if !ok {
				return false, fmt.Errorf("failed to find %s > %s in query map", mt.SubjectType, mt.SubjectRelation)
			}
			nextChecks = append(nextChecks, CheckRequest{
				Tuple: Tuple{
					ObjectType:      mt.SubjectType,
					ObjectID:        mt.SubjectID,
					ObjectRelation:  mt.SubjectRelation,
					SubjectType:     cp.Tuple.SubjectType,
					SubjectID:       cp.Tuple.SubjectID,
					SubjectRelation: cp.Tuple.SubjectRelation,
				},
				Userdata: pc.Userdata,
				Checks:   pc.Checks,
			})
		case KindIndirect: // TODO: THIS CAN BE USERSET!?
			relations := check.Rule().WithRelationToSubject
			for _, relation := range relations {
				pc, ok := r.checkMap[mt.SubjectType][relation]
				if !ok {
					return false, fmt.Errorf("failed to find %s > %s in query map", mt.SubjectType, relation)
				}
				nextChecks = append(nextChecks, CheckRequest{
					Tuple: Tuple{
						ObjectType:      mt.SubjectType,
						ObjectID:        mt.SubjectID,
						ObjectRelation:  relation,
						SubjectType:     cp.Tuple.SubjectType,
						SubjectID:       cp.Tuple.SubjectID,
						SubjectRelation: cp.Tuple.SubjectRelation,
					},
					Userdata: pc.Userdata,
					Checks:   pc.Checks,
				})
			}
		default:
			panic("unreachable")
		}
	}

	return r.check(ctx, nextChecks, depth)
}

func (r *Resolver) ChecksFor(object, relation string) []Check {
	return r.checkMap[object][relation].Checks
}

func prepareCheckMapForModel(model *Model, storage Storage) (CheckMap, error) {
	// Let's create the commands first
	checkMap := CheckMap{}
	for object, relations := range model.MergedRules {
		checkMap[object] = map[string]PreparedChecks{}
		for relation, rules := range relations {
			checks := []Check{}
			for _, rule := range rules {
				if len(rule.WithRelationToSubject) > 0 { // INDIRECT
					checks = append(checks, &CheckIndirect{rule})
				} else { // DIRECT
					checks = append(checks, &CheckDirect{rule}, &CheckDirectUserset{rule})
				}
			}
			checks = sortChecks(checks)
			userdata, err := storage.PrepareForChecks(object, relation, checks)
			if err != nil {
				return nil, err
			}
			checkMap[object][relation] = PreparedChecks{
				Userdata: userdata,
				Checks:   checks,
			}
		}
	}
	return checkMap, nil
}

func sortChecks(checks []Check) []Check {
	slices.SortFunc(checks, func(a, b Check) int {
		return int(a.Kind()) - int(b.Kind())
	})
	return checks
}
