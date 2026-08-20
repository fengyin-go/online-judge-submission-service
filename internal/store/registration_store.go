package store

import (
	"onlinejudge/internal/model"
)

func (s *MemoryStore) CreateRegistration(r *model.Registration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.registrations[r.ID]; ok {
		return ErrConflict
	}
	for _, exist := range s.registrations {
		if exist.ContestID == r.ContestID && exist.UserID == r.UserID {
			return ErrConflict
		}
	}
	s.registrations[r.ID] = r
	return nil
}

func (s *MemoryStore) GetRegistration(id string) (*model.Registration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.registrations[id]
	if !ok {
		return nil, ErrNotFound
	}
	return r, nil
}

func (s *MemoryStore) ListRegistrations() []*model.Registration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Registration, 0, len(s.registrations))
	for _, r := range s.registrations {
		list = append(list, r)
	}
	return list
}

func (s *MemoryStore) DeleteRegistration(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.registrations[id]; !ok {
		return ErrNotFound
	}
	delete(s.registrations, id)
	return nil
}
