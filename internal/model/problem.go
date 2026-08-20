package model

import (
	"strings"
	"time"
)

const (
	DifficultyEasy   = "easy"
	DifficultyMedium = "medium"
	DifficultyHard   = "hard"

	ProblemVisible = "visible"
	ProblemHidden  = "hidden"
)

// Problem 编程题目。
type Problem struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Difficulty    string    `json:"difficulty"`
	TimeLimitMs   int       `json:"time_limit_ms"`
	MemoryLimitMB int       `json:"memory_limit_mb"`
	Tags          []string  `json:"tags"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (p *Problem) Validate() error {
	p.Title = strings.TrimSpace(p.Title)
	p.Description = strings.TrimSpace(p.Description)
	if p.Title == "" {
		return NewValidationError("title", "题目标题不能为空")
	}
	if p.Description == "" {
		return NewValidationError("description", "题目描述不能为空")
	}
	if p.Difficulty == "" {
		p.Difficulty = DifficultyEasy
	}
	if !isValidDifficulty(p.Difficulty) {
		return NewValidationError("difficulty", "难度不合法")
	}
	if p.TimeLimitMs <= 0 {
		p.TimeLimitMs = 1000
	}
	if p.MemoryLimitMB <= 0 {
		p.MemoryLimitMB = 256
	}
	if p.Status == "" {
		p.Status = ProblemVisible
	}
	if p.Status != ProblemVisible && p.Status != ProblemHidden {
		return NewValidationError("status", "题目状态不合法")
	}
	return nil
}

func isValidDifficulty(s string) bool {
	switch s {
	case DifficultyEasy, DifficultyMedium, DifficultyHard:
		return true
	}
	return false
}

// ProblemFilter 题目筛选条件。
type ProblemFilter struct {
	Difficulty string
	Status     string
	Tag        string
	Keyword    string
}

func (f ProblemFilter) Match(p *Problem) bool {
	if f.Difficulty != "" && p.Difficulty != f.Difficulty {
		return false
	}
	if f.Status != "" && p.Status != f.Status {
		return false
	}
	if f.Tag != "" && !containsString(p.Tags, f.Tag) {
		return false
	}
	if k := strings.ToLower(strings.TrimSpace(f.Keyword)); k != "" {
		if !strings.Contains(strings.ToLower(p.Title), k) {
			return false
		}
	}
	return true
}

func containsString(list []string, target string) bool {
	for _, s := range list {
		if s == target {
			return true
		}
	}
	return false
}
