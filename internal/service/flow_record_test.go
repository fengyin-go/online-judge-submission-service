package service

import (
	"context"
	"onlinejudge/internal/model"
	"onlinejudge/internal/store"
	"testing"
)

func TestFlowItemsRemainOwned(t *testing.T) {
	items := []string{"first", "second"}
	st := store.NewFlowStore()
	st.Save(&model.FlowTicket{ID: "t7", UserID: "u", ProblemID: "p", Items: items})
	items[0] = "changed"
	got := st.Get("t7")
	if got.Items[0] != "first" {
		t.Fatalf("store aliased caller slice: %#v", got.Items)
	}
	svc := NewFlowService(st, model.DefaultFlowValidator{})
	_ = svc.Process(context.Background(), "t7")
	if st.Get("t7").Items[0] != "first" {
		t.Fatal("async boundary lost ownership")
	}
}
