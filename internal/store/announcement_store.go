package store

import (
	"onlinejudge/internal/model"
)

func (s *MemoryStore) CreateAnnouncement(a *model.Announcement) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.announcements[a.ID]; ok {
		return ErrConflict
	}
	s.announcements[a.ID] = a
	return nil
}

func (s *MemoryStore) GetAnnouncement(id string) (*model.Announcement, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.announcements[id]
	if !ok {
		return nil, ErrNotFound
	}
	return a, nil
}

func (s *MemoryStore) ListAnnouncements() []*model.Announcement {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Announcement, 0, len(s.announcements))
	for _, a := range s.announcements {
		list = append(list, a)
	}
	return list
}

func (s *MemoryStore) UpdateAnnouncement(a *model.Announcement) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.announcements[a.ID]; !ok {
		return ErrNotFound
	}
	s.announcements[a.ID] = a
	return nil
}

func (s *MemoryStore) DeleteAnnouncement(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.announcements[id]; !ok {
		return ErrNotFound
	}
	delete(s.announcements, id)
	return nil
}
