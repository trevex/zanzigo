package zanzigo

import "context"

type Storage interface {
	Write(ctx context.Context, t Tuple) error
	Close() error
}
