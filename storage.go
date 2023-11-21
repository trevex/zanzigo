package zanzigo

import (
	"context"
	"errors"

	"github.com/gofrs/uuid/v5"
)

var (
	// Returned by Storage-implementation for example if a given Read did not return a result.
	ErrNotFound = errors.New("not found")
)

// Marker interface for Userdata returned by a storage-implementation.
// Primary purpose is to prepare for a given inferred ruleset with Storage.PrepareRuleset
// and supply the Userdata to subsequent Storage.QueryChecks involving the associated ruleset.
type Userdata any

// A request to check a given [Tuple] with the inferred ruleset and
// the prepared [Userdata].
type Check struct {
	Tuple    Tuple
	Userdata Userdata
	Ruleset  []InferredRule
}

// A tuple with additional information which [Check] and which [InferredRule] from the [Check] resulted in this tuple.
// This is used by the [Resolver] to connect resulting tuples to the original instructions to deduct subsquent actions.
type MarkedTuple struct {
	Tuple
	CheckIndex int
	RuleIndex  int
}

// Storage provides simple CRUD operations for persistence as well as more complex methods
// required to permission checks as performant as possible.
type Storage interface {
	// Creates the [Tuple] t or errors, if creations fails.
	Write(ctx context.Context, t Tuple) error
	// Reads the specified [Tuple]. As all fields need to be known to read it, the UUID is returned.
	// If the tuple was not found, [ErrNotFound] is returned.
	Read(ctx context.Context, t Tuple) (uuid.UUID, error)

	// PrepareRuleset takes an object-type and relation with the inferred ruleset and prepares
	// the storage-implementation for subsequent checks by optionally returning [Userdata].
	PrepareRuleset(object, relation string, ruleset []InferredRule) (Userdata, error)
	// QueryChecks will retrieve matching tuples for all the checks if they exist.
	// The tuples are marked with the CheckIndex and RuleIndex to be able to identify precisely, the associated ruleset.
	// Returned marked tuples are sorted by RuleIndex as rulesets always begin with direct-relationships.
	// This allows returning as soon as possible by minimizing the rules to be checked for matches.
	QueryChecks(ctx context.Context, checks []Check) ([]MarkedTuple, error)

	Close() error
}
