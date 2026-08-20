package service

import (
	"sort"
	"time"

	"onlinejudge/internal/model"
	"onlinejudge/pkg/idgen"
)

func (s *Service) Register(contestID, userID string) (*model.Registration, error) {
	contest, err := s.store.GetContest(contestID)
	if err != nil {
		return nil, model.NewValidationError("contest_id", "比赛不存在")
	}
	if contest.Status != model.ContestRegistration {
		return nil, model.NewValidationError("status", "当前阶段不可报名")
	}
	if _, err := s.store.GetUser(userID); err != nil {
		return nil, model.NewValidationError("user_id", "用户不存在")
	}
	r := &model.Registration{
		ID:           idgen.Hex(),
		ContestID:    contestID,
		UserID:       userID,
		RegisteredAt: time.Now(),
		CreatedAt:    time.Now(),
	}
	if err := s.store.CreateRegistration(r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Service) GetRegistration(id string) (*model.Registration, error) {
	return s.store.GetRegistration(id)
}

func (s *Service) ListRegistrations(filter model.RegistrationFilter, page, size int) ([]*model.Registration, int, error) {
	all := s.store.ListRegistrations()
	matched := make([]*model.Registration, 0, len(all))
	for _, r := range all {
		if filter.Match(r) {
			matched = append(matched, r)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].RegisteredAt.Before(matched[j].RegisteredAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Registration{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

func (s *Service) DeleteRegistration(id string) error {
	return s.store.DeleteRegistration(id)
}
