package store

import (
	"onlinejudge/internal/model"
)

func (s *MemoryStore) CreateJudgeResult(j *model.JudgeResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.judges[j.ID]; ok {
		return ErrConflict
	}
	s.judges[j.ID] = j
	return nil
}

func (s *MemoryStore) GetJudgeResult(id string) (*model.JudgeResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.judges[id]
	if !ok {
		return nil, ErrNotFound
	}
	return j, nil
}

func (s *MemoryStore) GetJudgeResultBySubmission(submissionID string) (*model.JudgeResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, j := range s.judges {
		if j.SubmissionID == submissionID {
			return j, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) ListJudgeResults() []*model.JudgeResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.JudgeResult, 0, len(s.judges))
	for _, j := range s.judges {
		list = append(list, j)
	}
	return list
}

func (s *MemoryStore) DeleteJudgeResult(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.judges[id]; !ok {
		return ErrNotFound
	}
	delete(s.judges, id)
	return nil
}
