package handler

import (
	"context"
	"onlinejudge/internal/model"
	"onlinejudge/internal/service"
	"onlinejudge/internal/store"
)

func SubmitFlow(ctx context.Context, st *store.FlowStore, id string) error {
	if st == nil {
		return context.Canceled
	}
	svc := service.NewFlowService(st, model.DefaultFlowValidator{})
	return svc.Process(ctx, id)
}
