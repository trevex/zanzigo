package zanzigo

import (
	"context"
	"errors"

	"github.com/gofrs/uuid/v5"
)

var (
	ErrNotFound = errors.New("not found")
)

// Marker interface for
type Userdata any

type Check struct {
	Tuple    Tuple
	Userdata Userdata
	Ruleset  []InferredRule
}

type MarkedTuple struct {
	CheckID int
	RuleID  int
	Tuple
}

type Storage interface {
	Write(ctx context.Context, t Tuple) error
	Read(ctx context.Context, t Tuple) (uuid.UUID, error)

	PrepareRuleset(object, relation string, ruleset []InferredRule) (Userdata, error)
	// Returns MarkedTuples ordered by CommandID!
	QueryChecks(ctx context.Context, check []Check) ([]MarkedTuple, error)

	Close() error
}
