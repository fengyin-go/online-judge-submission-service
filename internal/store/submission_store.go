package store

import (
	"onlinejudge/internal/model"
)

func (s *MemoryStore) CreateSubmission(sub *model.Submission) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.submissions[sub.ID]; ok {
		return ErrConflict
	}
	s.submissions[sub.ID] = sub
	return nil
}

func (s *MemoryStore) GetSubmission(id string) (*model.Submission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sub, ok := s.submissions[id]
	if !ok {
		return nil, ErrNotFound
	}
	return sub, nil
}

func (s *MemoryStore) ListSubmissions() []*model.Submission {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Submission, 0, len(s.submissions))
	for _, sub := range s.submissions {
		list = append(list, sub)
	}
	return list
}

func (s *MemoryStore) UpdateSubmission(sub *model.Submission) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.submissions[sub.ID]; !ok {
		return ErrNotFound
	}
	s.submissions[sub.ID] = sub
	return nil
}

func (s *MemoryStore) DeleteSubmission(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.submissions[id]; !ok {
		return ErrNotFound
	}
	delete(s.submissions, id)
	return nil
}
