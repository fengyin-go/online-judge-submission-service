package service

import (
	"context"
	"onlinejudge/internal/model"
	"onlinejudge/internal/store"
	"testing"
)

func TestFlowConcurrentTerminalTransition(t *testing.T) {
	st := store.NewFlowStore()
	st.Save(&model.FlowTicket{ID: "t10", UserID: "u", ProblemID: "p"})
	svc := NewFlowService(st, model.DefaultFlowValidator{})
	for i := 0; i < 3; i++ {
		_ = svc.Process(context.Background(), "t10")
	}
	got := st.Get("t10")
	if got.Attempts != 1 || got.Status != model.FlowAccepted {
		t.Fatalf("non-idempotent transition: %#v", got)
	}
	if len(st.Events()) != 1 {
		t.Fatalf("duplicate terminal events: %#v", st.Events())
	}
}
