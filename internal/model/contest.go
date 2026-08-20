package model

import (
	"strings"
	"time"
)

const (
	ContestRegistration = "registration"
	ContestRunning      = "running"
	ContestEnded        = "ended"
)

// Contest 比赛。
type Contest struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	Problems    []string  `json:"problems"` // 关联题目 ID 列表
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (c *Contest) Validate() error {
	c.Title = strings.TrimSpace(c.Title)
	c.Description = strings.TrimSpace(c.Description)
	if c.Title == "" {
		return NewValidationError("title", "比赛名称不能为空")
	}
	if !c.EndAt.After(c.StartAt) {
		return NewValidationError("end_at", "结束时间必须晚于开始时间")
	}
	if c.Status == "" {
		c.Status = ContestRegistration
	}
	if !isValidContestStatus(c.Status) {
		return NewValidationError("status", "比赛状态不合法")
	}
	return nil
}

func isValidContestStatus(s string) bool {
	switch s {
	case ContestRegistration, ContestRunning, ContestEnded:
		return true
	}
	return false
}

// contestTransitions 比赛状态机。
var contestTransitions = map[string][]string{
	ContestRegistration: {ContestRunning},
	ContestRunning:      {ContestEnded},
	ContestEnded:        {},
}

// CanTransitionContest 判断比赛状态 from 能否流转到 to。
func CanTransitionContest(from, to string) bool {
	for _, t := range contestTransitions[from] {
		if t == to {
			return true
		}
	}
	return false
}

// ContestFilter 比赛筛选条件。
type ContestFilter struct {
	Status  string
	Keyword string
}

func (f ContestFilter) Match(c *Contest) bool {
	if f.Status != "" && c.Status != f.Status {
		return false
	}
	if k := strings.ToLower(strings.TrimSpace(f.Keyword)); k != "" {
		if !strings.Contains(strings.ToLower(c.Title), k) {
			return false
		}
	}
	return true
}
