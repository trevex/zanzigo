package zanzigo

import (
	"context"
	"errors"

	"github.com/gofrs/uuid/v5"
)

var (
	ErrNotFound = errors.New("not found")
)

type Query interface {
	Get() string
	NumArgs() int
}

type QueryContext struct {
	Root Tuple // TODO: simplify?
}

type MarkedTuple struct {
	CommandID int
	Tuple
}

type Storage interface {
	Write(ctx context.Context, t Tuple) error
	Read(ctx context.Context, t Tuple) (uuid.UUID, error)

	QueryForCommands(commands []Command) Query
	// Returns MarkedTuples ordered by CommandID!
	RunQuery(ctx context.Context, qc QueryContext, q Query, commands []Command) ([]MarkedTuple, error)

	Close() error
}
