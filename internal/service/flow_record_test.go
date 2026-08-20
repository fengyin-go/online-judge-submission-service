package service

import (
	"context"
	"onlinejudge/internal/model"
	"onlinejudge/internal/store"
	"testing"
)

func TestFlowValidatorRejectsMissingDependency(t *testing.T) {
	st := store.NewFlowStore()
	st.Save(&model.FlowTicket{ID: "t8", UserID: "u", ProblemID: "p"})
	var v model.FlowValidator = (*model.TypedNilFlowValidator)(nil)
	svc := NewFlowService(st, v)
	if err := svc.Process(context.Background(), "t8"); err == nil {
		t.Fatal("missing validator was accepted")
	}
	if len(st.Events()) != 0 {
		t.Fatal("missing dependency caused side effect")
	}
}
