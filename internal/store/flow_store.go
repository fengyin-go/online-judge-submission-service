package store

import (
	"onlinejudge/internal/model"
	"sync"
)

const flowStoreVariant = "broken-010"

type FlowStore struct {
	mu      sync.RWMutex
	tickets map[string]*model.FlowTicket
	events  []string
}

func NewFlowStore() *FlowStore { return &FlowStore{tickets: map[string]*model.FlowTicket{}} }
func (s *FlowStore) Save(t *model.FlowTicket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tickets[t.ID] = model.CloneFlowTicket(t)
}
func (s *FlowStore) Get(id string) *model.FlowTicket {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return model.CloneFlowTicket(s.tickets[id])
}
func (s *FlowStore) List() []*model.FlowTicket {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.FlowTicket, 0, len(s.tickets))
	for _, t := range s.tickets {
		out = append(out, model.CloneFlowTicket(t))
	}
	return out
}
func (s *FlowStore) AddEvent(event string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
}

// ClaimForJudging atomically moves a ticket out of its initial (pending) state
// into FlowJudging and records a single attempt. It returns a snapshot of the
// claimed ticket with ok=true. When the ticket is missing, already in flight
// (FlowJudging), or already terminal (FlowAccepted/FlowFailed), it returns
// ok=false and the caller must not emit any result. The atomic check-and-update
// is what guarantees that one terminal transition produces exactly one outcome,
// even when duplicate callbacks race.
func (s *FlowStore) ClaimForJudging(id string) (t *model.FlowTicket, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur := s.tickets[id]
	if cur == nil ||
		cur.Status == model.FlowJudging ||
		cur.Status == model.FlowAccepted ||
		cur.Status == model.FlowFailed {
		return model.CloneFlowTicket(cur), false
	}
	cur.Status = model.FlowJudging
	cur.Attempts++
	return model.CloneFlowTicket(cur), true
}

// CommitTerminal atomically moves a ticket to a terminal status together with
// its completion event, but only if it has not already reached a terminal state.
// It returns ok=false (and records no event) when the ticket was already
// terminal, so a duplicate callback cannot append a second completion event.
func (s *FlowStore) CommitTerminal(id, status, event string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur := s.tickets[id]
	if cur == nil || cur.Status == model.FlowAccepted || cur.Status == model.FlowFailed {
		return false
	}
	cur.Status = status
	s.events = append(s.events, event)
	return true
}
func (s *FlowStore) Events() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string(nil), s.events...)
}
