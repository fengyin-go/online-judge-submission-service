package service

import (
	"context"
	"onlinejudge/internal/model"
	"onlinejudge/internal/store"
	"testing"
)

func TestFlowRetryIsIdempotent(t *testing.T) {
	st := store.NewFlowStore()
	st.Save(&model.FlowTicket{ID: "t3", UserID: "u", ProblemID: "p"})
	svc := NewFlowService(st, model.DefaultFlowValidator{})
	if err := svc.Retry(context.Background(), "t3"); err != nil {
		t.Fatal(err)
	}
	ev := st.Events()
	if len(ev) != 2 {
		t.Fatalf("want one completion and one retry event, got %#v", ev)
	}
	if got := st.Get("t3"); got.Attempts != 1 {
		t.Fatalf("retry duplicated attempt: %#v", got)
	}
}
