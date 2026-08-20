package service

import (
	"context"
	"onlinejudge/internal/model"
	"onlinejudge/internal/store"
	"testing"
)

func TestFlowTypedNilValidator(t *testing.T) {
	st := store.NewFlowStore()
	st.Save(&model.FlowTicket{ID: "t4", UserID: "u", ProblemID: "p"})
	var v *model.TypedNilFlowValidator
	svc := NewFlowService(st, v)
	if err := svc.Process(context.Background(), "t4"); err == nil {
		t.Fatal("nil validator bypassed")
	}
	if got := st.Get("t4"); got.Status != "" {
		t.Fatalf("invalid flow changed state: %#v", got)
	}
}
