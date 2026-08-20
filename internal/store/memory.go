package store

import (
	"sync"

	"onlinejudge/internal/model"
)

// MemoryStore 内存实现，用 map 保存各实体。
type MemoryStore struct {
	mu            sync.RWMutex
	users         map[string]*model.User
	problems      map[string]*model.Problem
	submissions   map[string]*model.Submission
	judges        map[string]*model.JudgeResult
	contests      map[string]*model.Contest
	registrations map[string]*model.Registration
	announcements map[string]*model.Announcement
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:         make(map[string]*model.User),
		problems:      make(map[string]*model.Problem),
		submissions:   make(map[string]*model.Submission),
		judges:        make(map[string]*model.JudgeResult),
		contests:      make(map[string]*model.Contest),
		registrations: make(map[string]*model.Registration),
		announcements: make(map[string]*model.Announcement),
	}
}

var _ Store = (*MemoryStore)(nil)
