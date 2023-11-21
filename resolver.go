package zanzigo

import "context"

// A map of object-types to relations to Userdata.
type UserdataMap map[string]map[string]Userdata

type Resolver interface {
	Check(ctx context.Context, t Tuple) (bool, error)
}
