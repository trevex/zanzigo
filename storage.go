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
	Commands []CheckCommand
}

type MarkedTuple struct {
	CheckID   int
	CommandID int
	Tuple
}

type Storage interface {
	Write(ctx context.Context, t Tuple) error
	Read(ctx context.Context, t Tuple) (uuid.UUID, error)

	PrepareForChecks(object, relation string, commands []CheckCommand) (Userdata, error)
	// Returns MarkedTuples ordered by CommandID!
	QueryChecks(ctx context.Context, checks []CheckRequest) ([]MarkedTuple, error)

	Close() error
}
