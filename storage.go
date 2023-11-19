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

type CheckRequest struct {
	Tuple    Tuple
	Userdata Userdata
	Checks   []Check
}

type MarkedTuple struct {
	RequestID int
	CheckID   int
	Tuple
}

type Storage interface {
	Write(ctx context.Context, t Tuple) error
	Read(ctx context.Context, t Tuple) (uuid.UUID, error)

	PrepareForChecks(object, relation string, checks []Check) (Userdata, error)
	// Returns MarkedTuples ordered by CommandID!
	QueryChecks(ctx context.Context, crs []CheckRequest) ([]MarkedTuple, error)

	Close() error
}
