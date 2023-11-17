package zanzigo

import (
	"context"
	"errors"
	"slices"
)

type CommandKind int

const (
	KindUnknown CommandKind = iota
	KindDirect
	KindDirectUserset
	KindIndirect
)

type Command interface {
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

type QueryAndCommands struct {
	Query    Query
	Commands []Command
}

type QueryMap map[string]map[string]QueryAndCommands

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
	qc := QueryContext{t}
	depth := 0
	return r.check(ctx, qc, depth)
}

func (r *Resolver) check(ctx context.Context, qc QueryContext, depth int) (bool, error) {
	if depth > r.maxDepth {
		return false, errors.New("max depth exceeded")
	}
	depth += 1
	qac := r.queryMap[qc.Root.ObjectType][qc.Root.ObjectRelation]
	markedTuples, err := r.storage.RunQuery(ctx, qc, qac.Query, qac.Commands)
	if err != nil {
		return false, err
	}

	// Tuples are ordered by .CommandID and commands are ordered with directs first,
	// so we can exit early if we find a direct relationship
	for _, mt := range markedTuples {
		command := qac.Commands[mt.CommandID]
		if command.Kind() == KindDirect {
			return true, nil
		}
	}
	return true, nil
}

func (r *Resolver) CommandsFor(object, relation string) []Command {
	return r.queryMap[object][relation].Commands
}

func computeQueryMapForModel(model *Model, storage Storage) QueryMap {
	// Let's create the commands first
	queryMap := QueryMap{}
	for object, relations := range model.MergedRules {
		queryMap[object] = map[string]QueryAndCommands{}
		for relation, rules := range relations {
			commands := []Command{}
			for _, rule := range rules {
				if len(rule.WithRelationToSubject) > 0 { // INDIRECT
					commands = append(commands, &CheckIndirectCommand{rule})
				} else { // DIRECT
					commands = append(commands, &CheckDirectCommand{rule}, &CheckDirectUsersetCommand{rule})
				}
			}
			commands = sortCommands(commands)
			queryMap[object][relation] = QueryAndCommands{
				Query:    storage.QueryForCommands(commands),
				Commands: commands,
			}
		}
	}
	return queryMap
}

func sortCommands(commands []Command) []Command {
	slices.SortFunc(commands, func(a, b Command) int {
		return int(a.Kind()) - int(b.Kind())
	})
	return commands
}

// func (r *Resolver) getCommands(checkTuple Tuple, targetUser string) []Cmd {
// 	rule := r.model[checkTuple.Object][checkTuple.Relation]
// 	return append([]Cmd{Cmd{
// 		CheckTuple: checkTuple,
// 		TargetUser: targetUser,
// 	}}, r.getCommandsForRule(rule, targetUser)...)

// }

// func (r *Resolver) getCommandsForRule(rule Rule, targetUser string) []Cmd {
// 	if rule.InheritIf == AnyOfPlaceholder {

// 	}
// }

// func (r *Resolver) execsForTuple(ctx context.Context, t Tuple, rule Rule) []ExecFn {
// 	execs := []ExecFn{} // TODO: capacity?

// 	// Read the tuple itself
// 	execs := append(execs, func() (bool, []Tuple, error) {
// 		_, err := r.storage.Read(t)
// 		if err == ErrNotFound {
// 			return false, nil, nil
// 		} else if err != nil {
// 			return false, nil, err
// 		}
// 		return true, nil, nil
// 	})

// 	// Let's check the rule
// 	if rule.InheritIf == AnyOfPlaceholder { // Multiple rules
// 		tuples := []Tuple{} // TODO: reserve capacity
// 		for _, subrule := range rule.Rules {
// 			execs = append(execs, r.execsForTuple(ctx, Tuple{}, subrule))...
// 			tuples = append(tuples, r.tuplesForRule(t, subrule)...)
// 		}
// 		return tuples

// 	} else if rule.InheritIf != "" { // Single rule
// 		// Rule references other type
// 		if rule.OfType != "" && rule.WithRelation != "" {
// 			relations := r.model[rule.OfType]
// 			relrule := relations[]

// 		}

// 		// TODO: get indirect tuples
// 	}

// }

// func (r *Resolver) tuplesForRule(t Tuple, rule Rule) []Tuple {
// 	if rule.InheritIf == AnyOfPlaceholder { // Multiple rules
// 		tuples := []Tuple{} // TODO: reserve capacity
// 		for _, subrule := range rule.Rules {
// 			tuples = append(tuples, r.tuplesForRule(t, subrule)...)
// 		}
// 		return tuples

// 	} else if rule.InheritIf != "" { // Single rule
// 		// Rule references other type
// 		if rule.OfType != "" && rule.WithRelation != "" {
// 			relations := r.model[rule.OfType]
// 			relrule := relations[]

// 		}

// 		// TODO: get indirect tuples
// 	}
// 	return []Tuple{}
// }
