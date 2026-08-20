package model

import (
	"strings"
	"time"
)

// Registration 比赛报名记录，同一用户对同一比赛唯一。
type Registration struct {
	ID           string    `json:"id"`
	ContestID    string    `json:"contest_id"`
	UserID       string    `json:"user_id"`
	RegisteredAt time.Time `json:"registered_at"`
	CreatedAt    time.Time `json:"created_at"`
}

func (r *Registration) Validate() error {
	r.ContestID = strings.TrimSpace(r.ContestID)
	r.UserID = strings.TrimSpace(r.UserID)
	if r.ContestID == "" {
		return NewValidationError("contest_id", "比赛不能为空")
	}
	if r.UserID == "" {
		return NewValidationError("user_id", "报名用户不能为空")
	}
	return nil
}

// RegistrationFilter 报名筛选条件。
type RegistrationFilter struct {
	ContestID string
	UserID    string
}

func (f RegistrationFilter) Match(r *Registration) bool {
	if f.ContestID != "" && r.ContestID != f.ContestID {
		return false
	}
	if f.UserID != "" && r.UserID != f.UserID {
		return false
	}
	return true
}
