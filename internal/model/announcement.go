package model

import (
	"strings"
	"time"
)

// Announcement 公告。
type Announcement struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Pinned    bool      `json:"pinned"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (a *Announcement) Validate() error {
	a.Title = strings.TrimSpace(a.Title)
	a.Content = strings.TrimSpace(a.Content)
	if a.Title == "" {
		return NewValidationError("title", "公告标题不能为空")
	}
	if a.Content == "" {
		return NewValidationError("content", "公告内容不能为空")
	}
	return nil
}

// AnnouncementFilter 公告筛选条件。
type AnnouncementFilter struct {
	Pinned  *bool
	Keyword string
}

func (f AnnouncementFilter) Match(a *Announcement) bool {
	if f.Pinned != nil && a.Pinned != *f.Pinned {
		return false
	}
	if k := strings.ToLower(strings.TrimSpace(f.Keyword)); k != "" {
		if !strings.Contains(strings.ToLower(a.Title), k) &&
			!strings.Contains(strings.ToLower(a.Content), k) {
			return false
		}
	}
	return true
}
