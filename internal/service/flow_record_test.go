package service

import (
	"context"
	"onlinejudge/internal/model"
	"onlinejudge/internal/store"
	"sync"
	"testing"
)

func TestFlowSnapshotIsolation(t *testing.T) {
	st := store.NewFlowStore()
	st.Save(&model.FlowTicket{ID: "t1", UserID: "u", ProblemID: "p", Items: []string{"a"}})
	got := st.Get("t1")
	got.Items[0] = "corrupt"
	if st.Get("t1").Items[0] != "a" {
		t.Fatal("store leaked mutable snapshot")
	}
	svc := NewFlowService(st, model.DefaultFlowValidator{})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _ = svc.Process(context.Background(), "t1") }()
	}
	wg.Wait()
	got = st.Get("t1")
	if got == nil || got.Items[0] != "a" {
		t.Fatalf("snapshot corrupted: %#v", got)
	}
	if len(st.Events()) == 0 {
		t.Fatal("missing completion event")
	}
}
