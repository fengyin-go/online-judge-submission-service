package service

import (
	"context"
	"onlinejudge/internal/model"
	"onlinejudge/internal/store"
	"testing"
)

func TestFlowContextCancellation(t *testing.T) {
	st := store.NewFlowStore()
	st.Save(&model.FlowTicket{ID: "t2", UserID: "u", ProblemID: "p"})
	svc := NewFlowService(st, model.DefaultFlowValidator{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := svc.Process(ctx, "t2"); err == nil {
		t.Fatal("cancelled flow completed")
	}
	if got := st.Get("t2"); got.Status != "" || len(st.Events()) != 0 {
		t.Fatalf("cancel leaked side effects: %#v %#v", got, st.Events())
	}
}
