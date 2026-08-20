package store

import (
	"onlinejudge/internal/model"
)

func (s *MemoryStore) CreateProblem(p *model.Problem) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.problems[p.ID]; ok {
		return ErrConflict
	}
	for _, exist := range s.problems {
		if exist.Title == p.Title {
			return ErrConflict
		}
	}
	s.problems[p.ID] = p
	return nil
}

func (s *MemoryStore) GetProblem(id string) (*model.Problem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.problems[id]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *MemoryStore) ListProblems() []*model.Problem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Problem, 0, len(s.problems))
	for _, p := range s.problems {
		list = append(list, p)
	}
	return list
}

func (s *MemoryStore) UpdateProblem(p *model.Problem) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.problems[p.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.problems {
		if exist.ID != p.ID && exist.Title == p.Title {
			return ErrConflict
		}
	}
	s.problems[p.ID] = p
	return nil
}

func (s *MemoryStore) DeleteProblem(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.problems[id]; !ok {
		return ErrNotFound
	}
	delete(s.problems, id)
	return nil
}
