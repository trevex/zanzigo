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
// Primary purpose is to prepare for a given inferred ruleset with [Storage.PrepareRuleset]
// and supply additional data to [Storage.QueryChecks].
type Userdata any

// A request to check a given [Tuple] also providing the relevant inferred ruleset and
// the associated prepared userdata.
type Check struct {
	Tuple    Tuple
	Userdata Userdata
	Ruleset  []InferredRule
}

// A tuple with additional information which [Check] and which [InferredRule] from the [Check] resulted in this tuple.
// This is used by the [Resolver] to connect resulting tuples to the original instructions to deduct subsquent actions.
type MarkedTuple struct {
	CheckIndex int
	RuleIndex  int
	Tuple
}

type Storage interface {
	Write(ctx context.Context, t Tuple) error
	Read(ctx context.Context, t Tuple) (uuid.UUID, error)

	PrepareRuleset(object, relation string, ruleset []InferredRule) (Userdata, error)
	// Returns MarkedTuples ordered by CommandID!
	QueryChecks(ctx context.Context, checks []Check) ([]MarkedTuple, error)

	Close() error
}
