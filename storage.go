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
type CheckQuery interface{}

type CheckPayload struct {
	Tuple    Tuple
	Query    CheckQuery
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

	PrecomputeQueryForCheckCommands(commands []CheckCommand) CheckQuery
	// Returns MarkedTuples ordered by CommandID!
	QueryChecks(ctx context.Context, checks []CheckPayload) ([]MarkedTuple, error)

	Close() error
}
