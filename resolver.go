package zanzigo

import "context"

// A map of object-types to relations to Userdata.
type UserdataMap map[string]map[string]Userdata

// The main purpose of a [Resolver] is to traverse the ReBAC-policies and check whether a [Tuple] is authorized or not.
// See [SequentialResolver] for an implementation.
type Resolver interface {
	Check(ctx context.Context, t Tuple) (bool, error)
}
