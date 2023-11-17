package zanzigo

import (
	"context"
	"errors"
	"fmt"
	"slices"
)

type CommandKind int

const (
	KindUnknown CommandKind = iota
	KindDirect
	KindDirectUserset
	KindIndirect
)

type CheckCommand interface {
	Kind() CommandKind
	Rule() MergedRule
}

type CheckDirectCommand struct {
	MergedRule
}

func (c *CheckDirectCommand) Kind() CommandKind { return KindDirect }
func (c *CheckDirectCommand) Rule() MergedRule  { return c.MergedRule }

type CheckDirectUsersetCommand struct {
	MergedRule
}

func (c *CheckDirectUsersetCommand) Kind() CommandKind { return KindDirectUserset }
func (c *CheckDirectUsersetCommand) Rule() MergedRule  { return c.MergedRule }

// CheckIndirectCommand will first try to find subjects for r.Object r.Relations
// and then traverse them independently. (e.g. recursive .Check)
type CheckIndirectCommand struct {
	MergedRule
}

func (c *CheckIndirectCommand) Kind() CommandKind { return KindIndirect }
func (c *CheckIndirectCommand) Rule() MergedRule  { return c.MergedRule }

type CheckQueryAndCommands struct {
	Query    CheckQuery
	Commands []CheckCommand
}

type QueryMap map[string]map[string]CheckQueryAndCommands

type Resolver struct {
	storage  Storage
	queryMap QueryMap
	maxDepth int
}

func NewResolver(model *Model, storage Storage, maxDepth int) (*Resolver, error) {
	return &Resolver{
		storage, computeQueryMapForModel(model, storage), maxDepth,
	}, nil
}

func (r *Resolver) Check(ctx context.Context, t Tuple) (bool, error) {
	qac, ok := r.queryMap[t.ObjectType][t.ObjectRelation]
	if !ok {
		return false, fmt.Errorf("failed to find %s > %s in query map", t.ObjectType, t.ObjectRelation)
	}
	depth := 0
	return r.check(ctx, []CheckPayload{{
		Tuple:    t,
		Query:    qac.Query,
		Commands: qac.Commands,
	}}, depth)
}

func (r *Resolver) check(ctx context.Context, checks []CheckPayload, depth int) (bool, error) {
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

	nextChecks := []CheckPayload{}
	// Returned marked tuples are ordered by .CommandID and commands are ordered with directs first,
	// so we can exit early if we find a direct relationship.
	for _, mt := range markedTuples {
		cp := checks[mt.CheckID]
		command := cp.Commands[mt.CommandID]
		switch command.Kind() {
		case KindDirect:
			return true, nil
		case KindDirectUserset:
			qac, ok := r.queryMap[mt.SubjectType][mt.SubjectRelation]
			if !ok {
				fmt.Println(cp)
				fmt.Println(command.Kind())
				fmt.Println(mt)
				return false, fmt.Errorf("failed to find %s > %s in query map", mt.SubjectType, mt.SubjectRelation)
			}
			nextChecks = append(nextChecks, CheckPayload{
				Tuple: Tuple{
					ObjectType:      mt.SubjectType,
					ObjectID:        mt.SubjectID,
					ObjectRelation:  mt.SubjectRelation,
					SubjectType:     cp.Tuple.SubjectType,
					SubjectID:       cp.Tuple.SubjectID,
					SubjectRelation: cp.Tuple.SubjectRelation,
				},
				Query:    qac.Query,
				Commands: qac.Commands,
			})
		case KindIndirect: // TODO: THIS CAN BE USERSET!?
			relations := command.Rule().WithRelationToSubject
			for _, relation := range relations {
				qac, ok := r.queryMap[mt.SubjectType][relation]
				if !ok {
					fmt.Println(cp)
					fmt.Println(mt)
					fmt.Println(relation)
					return false, fmt.Errorf("failed to find %s > %s in query map", mt.SubjectType, relation)
				}
				nextChecks = append(nextChecks, CheckPayload{
					Tuple: Tuple{
						ObjectType:      mt.SubjectType,
						ObjectID:        mt.SubjectID,
						ObjectRelation:  relation,
						SubjectType:     cp.Tuple.SubjectType,
						SubjectID:       cp.Tuple.SubjectID,
						SubjectRelation: cp.Tuple.SubjectRelation,
					},
					Query:    qac.Query,
					Commands: qac.Commands,
				})
			}
		default:
			panic("unreachable")
		}
	}

	return r.check(ctx, nextChecks, depth)
}

func (r *Resolver) CheckCommandsFor(object, relation string) []CheckCommand {
	return r.queryMap[object][relation].Commands
}

func computeQueryMapForModel(model *Model, storage Storage) QueryMap {
	// Let's create the commands first
	queryMap := QueryMap{}
	for object, relations := range model.MergedRules {
		queryMap[object] = map[string]CheckQueryAndCommands{}
		for relation, rules := range relations {
			commands := []CheckCommand{}
			for _, rule := range rules {
				if len(rule.WithRelationToSubject) > 0 { // INDIRECT
					commands = append(commands, &CheckIndirectCommand{rule})
				} else { // DIRECT
					commands = append(commands, &CheckDirectCommand{rule}, &CheckDirectUsersetCommand{rule})
				}
			}
			commands = sortCommands(commands)
			queryMap[object][relation] = CheckQueryAndCommands{
				Query:    storage.PrecomputeQueryForCheckCommands(commands),
				Commands: commands,
			}
		}
	}
	return queryMap
}

func sortCommands(commands []CheckCommand) []CheckCommand {
	slices.SortFunc(commands, func(a, b CheckCommand) int {
		return int(a.Kind()) - int(b.Kind())
	})
	return commands
}
