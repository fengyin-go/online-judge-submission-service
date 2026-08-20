package store

import (
	"onlinejudge/internal/model"
)

func (s *MemoryStore) CreateContest(c *model.Contest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.contests[c.ID]; ok {
		return ErrConflict
	}
	for _, exist := range s.contests {
		if exist.Title == c.Title {
			return ErrConflict
		}
	}
	s.contests[c.ID] = c
	return nil
}

func (s *MemoryStore) GetContest(id string) (*model.Contest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.contests[id]
	if !ok {
		return nil, ErrNotFound
	}
	return c, nil
}

func (s *MemoryStore) ListContests() []*model.Contest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Contest, 0, len(s.contests))
	for _, c := range s.contests {
		list = append(list, c)
	}
	return list
}

func (s *MemoryStore) UpdateContest(c *model.Contest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.contests[c.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.contests {
		if exist.ID != c.ID && exist.Title == c.Title {
			return ErrConflict
		}
	}
	s.contests[c.ID] = c
	return nil
}

func (s *MemoryStore) DeleteContest(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.contests[id]; !ok {
		return ErrNotFound
	}
	delete(s.contests, id)
	return nil
}
