package service

import (
	"context"
	"onlinejudge/internal/model"
	"onlinejudge/internal/store"
	"testing"
)

func TestFlowFailureClosesLifecycle(t *testing.T) {
	st := store.NewFlowStore()
	st.Save(&model.FlowTicket{ID: "t6", UserID: "u", ProblemID: "p"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	svc := NewFlowService(st, model.DefaultFlowValidator{})
	if err := svc.Process(ctx, "t6"); err == nil {
		t.Fatal("cancelled process succeeded")
	}
	if len(st.Events()) != 0 {
		t.Fatal("lifecycle emitted event after failure")
	}
}
