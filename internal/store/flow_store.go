package store

import (
	"onlinejudge/internal/model"
	"sync"
)

type FlowStore struct {
	mu      sync.RWMutex
	tickets map[string]*model.FlowTicket
	events  []string
}

func NewFlowStore() *FlowStore { return &FlowStore{tickets: map[string]*model.FlowTicket{}} }
func (s *FlowStore) Save(t *model.FlowTicket) {
	if s == nil || t == nil || t.ID == "" {
		return
	}
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
	if s == nil || event == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
}
func (s *FlowStore) Events() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string(nil), s.events...)
}
