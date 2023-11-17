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

type PreparedCommands struct {
	Userdata Userdata
	Commands []CheckCommand
}

type CommandMap map[string]map[string]PreparedCommands

// During creation a set of static commands are precomputed which will also be passed on to the [Storage]-backend via [Storage.PrepareForCheckCommands].
type Resolver struct {
	storage    Storage
	commandMap CommandMap
	maxDepth   int
}

// NewResolver creates a new resolver for the particular [Model] using the designated [Storage]-implementation.
// The main purpose of the [Resolver] is to traverse the ReBAC-policies and check whether a [Tuple] is authorized or not.
// During creation a set of static commands are precomputed which will also be passed on to the [Storage]-backend via [Storage.PrepareForCheckCommands].
func NewResolver(model *Model, storage Storage, maxDepth int) (*Resolver, error) {
	commandMap, err := prepareCommandMapForModel(model, storage)
	return &Resolver{
		storage, commandMap, maxDepth,
	}, err
}

func (r *Resolver) Check(ctx context.Context, t Tuple) (bool, error) {
	pc, ok := r.commandMap[t.ObjectType][t.ObjectRelation]
	if !ok {
		return false, fmt.Errorf("failed to find %s > %s in query map", t.ObjectType, t.ObjectRelation)
	}
	depth := 0
	return r.check(ctx, []CheckRequest{{
		Tuple:    t,
		Userdata: pc.Userdata,
		Commands: pc.Commands,
	}}, depth)
}

func (r *Resolver) check(ctx context.Context, checks []CheckRequest, depth int) (bool, error) {
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

	nextChecks := []CheckRequest{}
	// Returned marked tuples are ordered by .CommandID and commands are ordered with directs first,
	// so we can exit early if we find a direct relationship.
	for _, mt := range markedTuples {
		cp := checks[mt.CheckID]
		command := cp.Commands[mt.CommandID]
		switch command.Kind() {
		case KindDirect:
			return true, nil
		case KindDirectUserset:
			pc, ok := r.commandMap[mt.SubjectType][mt.SubjectRelation]
			if !ok {
				fmt.Println(cp)
				fmt.Println(command.Kind())
				fmt.Println(mt)
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
				Commands: pc.Commands,
			})
		case KindIndirect: // TODO: THIS CAN BE USERSET!?
			relations := command.Rule().WithRelationToSubject
			for _, relation := range relations {
				pc, ok := r.commandMap[mt.SubjectType][relation]
				if !ok {
					fmt.Println(cp)
					fmt.Println(mt)
					fmt.Println(relation)
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
					Commands: pc.Commands,
				})
			}
		default:
			panic("unreachable")
		}
	}

	return r.check(ctx, nextChecks, depth)
}

func (r *Resolver) CheckCommandsFor(object, relation string) []CheckCommand {
	return r.commandMap[object][relation].Commands
}

func prepareCommandMapForModel(model *Model, storage Storage) (CommandMap, error) {
	// Let's create the commands first
	commandMap := CommandMap{}
	for object, relations := range model.MergedRules {
		commandMap[object] = map[string]PreparedCommands{}
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
			userdata, err := storage.PrepareForCheckCommands(object, relation, commands)
			if err != nil {
				return nil, err
			}
			commandMap[object][relation] = PreparedCommands{
				Userdata: userdata,
				Commands: commands,
			}
		}
	}
	return commandMap, nil
}

func sortCommands(commands []CheckCommand) []CheckCommand {
	slices.SortFunc(commands, func(a, b CheckCommand) int {
		return int(a.Kind()) - int(b.Kind())
	})
	return commands
}
