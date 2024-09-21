package sqlite3

import (
	"context"
	"fmt"
	"strings"

	"github.com/trevex/zanzigo"

	"github.com/cockroachdb/pebble"
	"github.com/gofrs/uuid/v5"
)

type PebbleStorage struct {
	db *pebble.DB
}

func NewPebbleStorage(dirname string) (*PebbleStorage, error) {
	db, err := pebble.Open(dirname, &pebble.Options{})
	return &PebbleStorage{db}, err
}

func (s *PebbleStorage) Close() error {
	return s.db.Close()
}

func (s *PebbleStorage) Write(ctx context.Context, t zanzigo.Tuple) error {
	return s.db.Set(toKey(t), nil, pebble.Sync)
}

func (s *PebbleStorage) Read(ctx context.Context, t zanzigo.Tuple) (uuid.UUID, error) {
	// TODO: the uuid is pointless for kv-store...
	id := uuid.UUID{}
	_, closer, err := s.db.Get(toKey(t))
	if err == pebble.ErrNotFound {
		return id, zanzigo.ErrNotFound
	} else if err != nil {
		return id, err
	}
	closer.Close()
	return id, nil
}

func (s *PebbleStorage) CursorStart() zanzigo.Cursor {
	return []byte("")
}

func (s *PebbleStorage) List(ctx context.Context, t zanzigo.Tuple, p zanzigo.Pagination) ([]zanzigo.Tuple, zanzigo.Cursor, error) {
	prefix := ""
	fieldErr := fmt.Errorf("Pebble List implementation only allows listing for valid prefixes in tuple string notation: %v", t)

	if t.ObjectType != "" {
		prefix += t.ObjectType + ":"
	} else if t.ObjectID != "" || t.ObjectRelation != "" || t.SubjectType != "" || t.SubjectID != "" || t.SubjectRelation != "" {
		return nil, nil, fieldErr
	}
	if t.ObjectType != "" {
		prefix += t.ObjectID + "#"
	} else if t.ObjectRelation != "" || t.SubjectType != "" || t.SubjectID != "" || t.SubjectRelation != "" {
		return nil, nil, fieldErr
	}
	if t.ObjectRelation != "" {
		prefix += t.ObjectRelation + "@"
	} else if t.SubjectType != "" || t.SubjectID != "" || t.SubjectRelation != "" {
		return nil, nil, fieldErr
	}
	if t.SubjectType != "" {
		prefix += t.SubjectType + ":"
	} else if t.SubjectID != "" || t.SubjectRelation != "" {
		return nil, nil, fieldErr
	}
	if t.SubjectID != "" {
		prefix += t.SubjectID // TODO: might find too much, we should add an "end" character
	} else if t.SubjectRelation != "" {
		return nil, nil, fieldErr
	}
	if t.SubjectRelation != "" {
		prefix += "#" + t.SubjectRelation
	}

	iterOpts := prefixIterOptions([]byte(prefix))
	if len(p.Cursor) > 0 {
		iterOpts.LowerBound = p.Cursor
	}
	iter, err := s.db.NewIter(iterOpts)
	if err != nil {
		return nil, nil, err
	}
	i := p.Limit
	cursor := p.Cursor
	tuples := make([]zanzigo.Tuple, 0, p.Limit)
	for iter.First(); iter.Valid(); iter.Next() {
		cursor = iter.Key()
		t := zanzigo.TupleString(string(cursor))
		tuples = append(tuples, t)
		i -= 1
		if i == 0 {
			break
		}
	}
	if err := iter.Close(); err != nil {
		return nil, nil, err
	}

	return tuples, keyUpperBound(cursor), nil
}

func (s *PebbleStorage) PrepareRuleset(object, relation string, ruleset []zanzigo.InferredRule) (zanzigo.Userdata, error) {
	return nil, nil
}

func (s *PebbleStorage) QueryChecks(ctx context.Context, checks []zanzigo.Check) ([]zanzigo.MarkedTuple, error) {
	tuples := []zanzigo.MarkedTuple{}

	// We iterate over all check and combine all the results
	for i, check := range checks {
		for j, rule := range check.Ruleset {
			switch rule.Kind {
			case zanzigo.KindDirect:
				for _, relation := range rule.Relations {
					t := zanzigo.Tuple{
						ObjectType:      rule.Object,
						ObjectID:        check.Tuple.ObjectID,
						ObjectRelation:  relation,
						SubjectType:     check.Tuple.SubjectType,
						SubjectID:       check.Tuple.SubjectID,
						SubjectRelation: check.Tuple.SubjectRelation,
					}
					_, closer, err := s.db.Get(toKey(t))
					if err == nil {
						closer.Close()
						tuples = append(tuples, zanzigo.MarkedTuple{Tuple: t, CheckIndex: i, RuleIndex: j})
					}
				}
			case zanzigo.KindDirectUserset:
				for _, relation := range rule.Relations {
					prefix := toDirectUsersetPrefix(rule.Object, check.Tuple.ObjectID, relation)
					iter, err := s.db.NewIter(prefixIterOptions(prefix))
					if err != nil {
						return nil, err
					}
					for iter.First(); iter.Valid(); iter.Next() {
						st := strings.ReplaceAll(string(iter.Key()), "!", "")
						t := zanzigo.TupleString(st)
						tuples = append(tuples, zanzigo.MarkedTuple{Tuple: t, CheckIndex: i, RuleIndex: j})
					}
					if err := iter.Close(); err != nil {
						return nil, err
					}
				}
			case zanzigo.KindIndirect:
				// TODO: do we need to check usersets as well!? assumption: no
				for _, relation := range rule.Relations {
					prefix := toIndirectPrefix(rule.Object, check.Tuple.ObjectID, relation, rule.Subject)
					iter, err := s.db.NewIter(prefixIterOptions(prefix))
					if err != nil {
						return nil, err
					}
					for iter.First(); iter.Valid(); iter.Next() {
						t := zanzigo.TupleString(string(iter.Key()))
						tuples = append(tuples, zanzigo.MarkedTuple{Tuple: t, CheckIndex: i, RuleIndex: j})
					}
					if err := iter.Close(); err != nil {
						return nil, err
					}
				}
			default:
				panic("unreachable")
			}
		}
	}

	return tuples, nil
}

func keyUpperBound(b []byte) []byte {
	end := make([]byte, len(b))
	copy(end, b)
	for i := len(end) - 1; i >= 0; i-- {
		end[i] = end[i] + 1
		if end[i] != 0 {
			return end[:i+1]
		}
	}
	return nil // no upper-bound
}

func prefixIterOptions(prefix []byte) *pebble.IterOptions {
	return &pebble.IterOptions{
		LowerBound: prefix,
		UpperBound: keyUpperBound(prefix),
	}
}

func toKey(t zanzigo.Tuple) []byte {
	var s string
	if t.SubjectRelation != "" {
		s = fmt.Sprintf("%s:%s#%s@!%s:%s#%s", t.ObjectType, t.ObjectID, t.ObjectRelation, t.SubjectType, t.SubjectID, t.SubjectRelation)
	} else {
		s = fmt.Sprintf("%s:%s#%s@%s:%s", t.ObjectType, t.ObjectID, t.ObjectRelation, t.SubjectType, t.SubjectID)
	}
	return []byte(s)
}

func toDirectUsersetPrefix(objectType, objectID, objectRelation string) []byte {
	return []byte(fmt.Sprintf("%s:%s#%s@!", objectType, objectID, objectRelation))
}

func toIndirectPrefix(objectType, objectID, objectRelation, subjectType string) []byte {
	return []byte(fmt.Sprintf("%s:%s#%s@%s:", objectType, objectID, objectRelation, subjectType))
}
