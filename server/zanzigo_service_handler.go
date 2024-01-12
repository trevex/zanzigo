package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofrs/uuid/v5"
	"github.com/trevex/zanzigo"
	v1 "github.com/trevex/zanzigo/api/zanzigo/v1"
	v1connect "github.com/trevex/zanzigo/api/zanzigo/v1/zanzigov1connect"

	"connectrpc.com/connect"
)

type zanzigoServiceHandler struct {
	log      *slog.Logger
	model    *zanzigo.Model
	storage  zanzigo.Storage
	resolver *zanzigo.Resolver
}

func NewZanzigoServiceHandler(log *slog.Logger, model *zanzigo.Model, storage zanzigo.Storage, resolver *zanzigo.Resolver) v1connect.ZanzigoServiceHandler {
	return &zanzigoServiceHandler{log, model, storage, resolver}
}

func (h *zanzigoServiceHandler) Write(ctx context.Context, req *connect.Request[v1.WriteRequest]) (*connect.Response[v1.WriteResponse], error) {
	tuple, err := h.isTupleValid(req.Msg.Tuple)
	if err != nil {
		return nil, err
	}

	err = h.storage.Write(ctx, tuple)
	if err != nil {
		// TODO: what if already exists? let's properly catch that error!
		h.log.Error("failed to write tuple", slog.Any("tuple", tuple), slog.Any("error", err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed writing tuple"))
	}

	return connect.NewResponse(&v1.WriteResponse{}), nil
}

func (h *zanzigoServiceHandler) Read(ctx context.Context, req *connect.Request[v1.ReadRequest]) (*connect.Response[v1.ReadResponse], error) {
	tuple, err := h.isTupleValid(req.Msg.Tuple)
	if err != nil {
		return nil, err
	}

	uuid, err := h.storage.Read(ctx, tuple)
	if errors.Is(err, zanzigo.ErrNotFound) {
		return nil, connect.NewError(connect.CodeDataLoss, fmt.Errorf("tuple not found"))
	} else if err != nil {
		h.log.Error("failed to read tuple", slog.Any("tuple", tuple), slog.Any("error", err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed reading tuple"))
	}

	return connect.NewResponse(&v1.ReadResponse{
		Uuid: uuid.String(),
	}), nil
}

func (h *zanzigoServiceHandler) Check(ctx context.Context, req *connect.Request[v1.CheckRequest]) (*connect.Response[v1.CheckResponse], error) {
	tuple, err := h.isTupleValid(req.Msg.Tuple)
	if err != nil {
		return nil, err
	}

	result, err := h.resolver.Check(ctx, tuple)
	if err != nil {
		h.log.Error("failed to check tuple", slog.Any("tuple", tuple), slog.Any("error", err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to check tuple"))
	}

	return connect.NewResponse(&v1.CheckResponse{
		Result: result,
	}), nil
}

func (h *zanzigoServiceHandler) List(ctx context.Context, req *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error) {
	filter := toZanzigoTuple(req.Msg.Filter) // TODO: filter tuple can be partially validated
	pagination, err := toZanzigoPagination(req.Msg.Pagination)
	if err != nil {
		h.log.Debug("failed to parse cursor", slog.String("cursor", req.Msg.Pagination.Cursor))
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("malformed cursor"))
	}

	tuples, cursor, err := h.storage.List(ctx, filter, pagination)
	if err != nil {
		h.log.Error("failed to list tuples", slog.Any("error", err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list tuples"))
	}

	return connect.NewResponse(&v1.ListResponse{
		Tuples: toProtobufTuples(tuples),
		Cursor: cursor.String(),
	}), nil
}

func (h *zanzigoServiceHandler) isTupleValid(t *v1.Tuple) (zanzigo.Tuple, error) {
	if t == nil {
		return zanzigo.EmptyTuple, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("missing tuple"))
	}
	tuple := toZanzigoTuple(t)
	if !h.model.IsValid(tuple) {
		return zanzigo.EmptyTuple, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid tuple: %v", tuple))
	}
	return tuple, nil
}

func toZanzigoTuple(t *v1.Tuple) zanzigo.Tuple {
	return zanzigo.Tuple{
		ObjectType:      t.ObjectType,
		ObjectID:        t.ObjectId,
		ObjectRelation:  t.ObjectRelation,
		SubjectType:     t.SubjectType,
		SubjectID:       t.SubjectId,
		SubjectRelation: t.SubjectRelation,
	}
}

func toProtobufTuple(t *zanzigo.Tuple) v1.Tuple {
	return v1.Tuple{
		ObjectType:      t.ObjectType,
		ObjectId:        t.ObjectID,
		ObjectRelation:  t.ObjectRelation,
		SubjectType:     t.SubjectType,
		SubjectId:       t.SubjectID,
		SubjectRelation: t.SubjectRelation,
	}
}

func toProtobufTuples(ts []zanzigo.Tuple) []*v1.Tuple {
	ps := make([]*v1.Tuple, 0, len(ts))
	for _, t := range ts {
		p := toProtobufTuple(&t)
		ps = append(ps, &p)
	}
	return ps
}

func toZanzigoPagination(p *v1.Pagination) (zanzigo.Pagination, error) {
	cursor, err := uuid.FromString(p.Cursor)
	return zanzigo.Pagination{
		Cursor: cursor,
		Limit:  int(p.Limit),
	}, err
}
