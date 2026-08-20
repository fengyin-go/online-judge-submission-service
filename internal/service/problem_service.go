package service

import (
	"sort"
	"time"

	"onlinejudge/internal/model"
	"onlinejudge/pkg/idgen"
)

func (s *Service) CreateProblem(p model.Problem) (*model.Problem, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	p.ID = idgen.Hex()
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	if err := s.store.CreateProblem(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Service) GetProblem(id string) (*model.Problem, error) {
	return s.store.GetProblem(id)
}

func (s *Service) ListProblems(filter model.ProblemFilter, page, size int) ([]*model.Problem, int, error) {
	all := s.store.ListProblems()
	matched := make([]*model.Problem, 0, len(all))
	for _, p := range all {
		if filter.Match(p) {
			matched = append(matched, p)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.Before(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Problem{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

func (s *Service) UpdateProblem(id string, p model.Problem) (*model.Problem, error) {
	existing, err := s.store.GetProblem(id)
	if err != nil {
		return nil, err
	}
	existing.Title = p.Title
	existing.Description = p.Description
	existing.Difficulty = p.Difficulty
	existing.TimeLimitMs = p.TimeLimitMs
	existing.MemoryLimitMB = p.MemoryLimitMB
	existing.Tags = p.Tags
	existing.Status = p.Status
	if err := existing.Validate(); err != nil {
		return nil, err
	}
	existing.UpdatedAt = time.Now()
	if err := s.store.UpdateProblem(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteProblem(id string) error {
	return s.store.DeleteProblem(id)
}

// ListTags 返回全部题目出现过的去重标签，按字典序排序。
func (s *Service) ListTags() ([]string, error) {
	set := make(map[string]bool)
	for _, p := range s.store.ListProblems() {
		for _, t := range p.Tags {
			set[t] = true
		}
	}
	tags := make([]string, 0, len(set))
	for t := range set {
		tags = append(tags, t)
	}
	sort.Strings(tags)
	return tags, nil
}

// BatchImportProblems 批量导入题目。
type ImportProblemResult struct {
	Success []*model.Problem `json:"success"`
	Failed  []BatchFailure   `json:"failed"`
}

func (s *Service) BatchImportProblems(problems []model.Problem) *ImportProblemResult {
	result := &ImportProblemResult{Success: []*model.Problem{}, Failed: []BatchFailure{}}
	for i, p := range problems {
		created, err := s.CreateProblem(p)
		if err != nil {
			result.Failed = append(result.Failed, BatchFailure{Index: i, Error: err.Error()})
			continue
		}
		result.Success = append(result.Success, created)
	}
	return result
}
