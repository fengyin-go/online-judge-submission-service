package service

import (
	"context"
	"onlinejudge/internal/model"
	"onlinejudge/internal/store"
	"testing"
)

func TestFlowChannelLifecycle(t *testing.T) {
	st := store.NewFlowStore()
	st.Save(&model.FlowTicket{ID: "t9", UserID: "u", ProblemID: "p"})
	svc := NewFlowService(st, model.DefaultFlowValidator{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := svc.Retry(ctx, "t9"); err == nil {
		t.Fatal("cancelled retry completed")
	}
	if len(st.Events()) != 0 {
		t.Fatal("retry emitted event after cancellation")
	}
}
