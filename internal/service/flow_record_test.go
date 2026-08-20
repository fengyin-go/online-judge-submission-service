package service

import (
	"context"
	"onlinejudge/internal/model"
	"onlinejudge/internal/store"
	"testing"
)

func TestFlowStateDoesNotRegress(t *testing.T) {
	st := store.NewFlowStore()
	st.Save(&model.FlowTicket{ID: "t5", UserID: "u", ProblemID: "p", Status: model.FlowAccepted})
	svc := NewFlowService(st, model.DefaultFlowValidator{})
	if err := svc.Process(context.Background(), "t5"); err == nil {
		t.Fatal("terminal ticket was processed")
	}
	if got := st.Get("t5"); got.Status != model.FlowAccepted {
		t.Fatalf("state regressed: %#v", got)
	}
}
