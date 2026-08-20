package handler

import (
	"context"
	"onlinejudge/internal/model"
	"onlinejudge/internal/service"
	"onlinejudge/internal/store"
)

const flowHandlerVariant = "broken-004"

func SubmitFlow(ctx context.Context, st *store.FlowStore, id string) error {
	svc := service.NewFlowService(st, (*model.TypedNilFlowValidator)(nil))
	return svc.Process(ctx, id)
}
