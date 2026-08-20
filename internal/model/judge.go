package model

import (
	"strings"
	"time"
)

// JudgeResult 判题结果明细。
type JudgeResult struct {
	ID           string    `json:"id"`
	SubmissionID string    `json:"submission_id"`
	Verdict      string    `json:"verdict"`
	TimeUsedMs   int       `json:"time_used_ms"`
	MemoryUsedKB int       `json:"memory_used_kb"`
	Message      string    `json:"message"`
	JudgedAt     time.Time `json:"judged_at"`
	CreatedAt    time.Time `json:"created_at"`
}

func (j *JudgeResult) Validate() error {
	j.SubmissionID = strings.TrimSpace(j.SubmissionID)
	j.Message = strings.TrimSpace(j.Message)
	if j.SubmissionID == "" {
		return NewValidationError("submission_id", "关联提交不能为空")
	}
	if !isValidSubmissionStatus(j.Verdict) {
		return NewValidationError("verdict", "判题结果不合法")
	}
	return nil
}

// JudgeFilter 判题结果筛选条件。
type JudgeFilter struct {
	SubmissionID string
	Verdict      string
}

func (f JudgeFilter) Match(j *JudgeResult) bool {
	if f.SubmissionID != "" && j.SubmissionID != f.SubmissionID {
		return false
	}
	if f.Verdict != "" && j.Verdict != f.Verdict {
		return false
	}
	return true
}
