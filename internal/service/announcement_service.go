package service

import (
	"sort"
	"time"

	"onlinejudge/internal/model"
	"onlinejudge/pkg/idgen"
)

func (s *Service) CreateAnnouncement(a model.Announcement) (*model.Announcement, error) {
	if err := a.Validate(); err != nil {
		return nil, err
	}
	a.ID = idgen.Hex()
	a.CreatedAt = time.Now()
	a.UpdatedAt = a.CreatedAt
	if err := s.store.CreateAnnouncement(&a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *Service) GetAnnouncement(id string) (*model.Announcement, error) {
	return s.store.GetAnnouncement(id)
}

func (s *Service) ListAnnouncements(filter model.AnnouncementFilter, page, size int) ([]*model.Announcement, int, error) {
	all := s.store.ListAnnouncements()
	matched := make([]*model.Announcement, 0, len(all))
	for _, a := range all {
		if filter.Match(a) {
			matched = append(matched, a)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		if matched[i].Pinned != matched[j].Pinned {
			return matched[i].Pinned
		}
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Announcement{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

func (s *Service) UpdateAnnouncement(id string, a model.Announcement) (*model.Announcement, error) {
	existing, err := s.store.GetAnnouncement(id)
	if err != nil {
		return nil, err
	}
	existing.Title = a.Title
	existing.Content = a.Content
	existing.Pinned = a.Pinned
	if err := existing.Validate(); err != nil {
		return nil, err
	}
	existing.UpdatedAt = time.Now()
	if err := s.store.UpdateAnnouncement(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteAnnouncement(id string) error {
	return s.store.DeleteAnnouncement(id)
}
