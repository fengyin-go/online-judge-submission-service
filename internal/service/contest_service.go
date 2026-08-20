package service

import (
	"sort"
	"time"

	"onlinejudge/internal/model"
	"onlinejudge/pkg/idgen"
)

func (s *Service) CreateContest(c model.Contest) (*model.Contest, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	c.ID = idgen.Hex()
	c.Status = model.ContestRegistration
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	if err := s.store.CreateContest(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Service) GetContest(id string) (*model.Contest, error) {
	return s.store.GetContest(id)
}

func (s *Service) ListContests(filter model.ContestFilter, page, size int) ([]*model.Contest, int, error) {
	all := s.store.ListContests()
	matched := make([]*model.Contest, 0, len(all))
	for _, c := range all {
		if filter.Match(c) {
			matched = append(matched, c)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].StartAt.Before(matched[j].StartAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Contest{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// TransitionContest 推进比赛状态。
func (s *Service) TransitionContest(id, to string) (*model.Contest, error) {
	c, err := s.store.GetContest(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionContest(c.Status, to) {
		return nil, model.NewValidationError("status", "不允许的状态流转: "+c.Status+" -> "+to)
	}
	c.Status = to
	c.UpdatedAt = time.Now()
	if err := s.store.UpdateContest(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) DeleteContest(id string) error {
	return s.store.DeleteContest(id)
}

// AddContestProblem 向比赛添加题目，仅报名阶段允许。
func (s *Service) AddContestProblem(contestID, problemID string) (*model.Contest, error) {
	c, err := s.store.GetContest(contestID)
	if err != nil {
		return nil, err
	}
	if c.Status != model.ContestRegistration {
		return nil, model.NewValidationError("status", "仅报名阶段可调整题目")
	}
	if _, err := s.store.GetProblem(problemID); err != nil {
		return nil, model.NewValidationError("problem_id", "题目不存在")
	}
	for _, id := range c.Problems {
		if id == problemID {
			return nil, model.NewValidationError("problem_id", "题目已在比赛中")
		}
	}
	c.Problems = append(c.Problems, problemID)
	c.UpdatedAt = time.Now()
	if err := s.store.UpdateContest(c); err != nil {
		return nil, err
	}
	return c, nil
}

// RemoveContestProblem 从比赛移除题目。
func (s *Service) RemoveContestProblem(contestID, problemID string) (*model.Contest, error) {
	c, err := s.store.GetContest(contestID)
	if err != nil {
		return nil, err
	}
	if c.Status != model.ContestRegistration {
		return nil, model.NewValidationError("status", "仅报名阶段可调整题目")
	}
	kept := make([]string, 0, len(c.Problems))
	for _, id := range c.Problems {
		if id != problemID {
			kept = append(kept, id)
		}
	}
	c.Problems = kept
	c.UpdatedAt = time.Now()
	if err := s.store.UpdateContest(c); err != nil {
		return nil, err
	}
	return c, nil
}

// ListContestProblems 返回比赛的题目列表。
func (s *Service) ListContestProblems(contestID string) ([]*model.Problem, error) {
	c, err := s.store.GetContest(contestID)
	if err != nil {
		return nil, err
	}
	problems := make([]*model.Problem, 0, len(c.Problems))
	for _, id := range c.Problems {
		if p, err := s.store.GetProblem(id); err == nil {
			problems = append(problems, p)
		}
	}
	sort.Slice(problems, func(i, j int) bool {
		return problems[i].ID < problems[j].ID
	})
	return problems, nil
}
