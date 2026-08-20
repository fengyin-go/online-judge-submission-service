package model

import (
	"strings"
	"time"
)

const (
	SubmissionPending      = "pending"
	SubmissionJudging      = "judging"
	SubmissionAccepted     = "accepted"
	SubmissionWrongAnswer  = "wrong_answer"
	SubmissionRuntimeError = "runtime_error"
	SubmissionTimeLimit    = "time_limit_exceeded"
	SubmissionMemoryLimit  = "memory_limit_exceeded"
	SubmissionCompileError = "compile_error"
)

const (
	LanguageGo   = "go"
	LanguageCpp  = "cpp"
	LanguageJava = "java"
	LanguagePy   = "python"
)

// Submission 提交记录。
type Submission struct {
	ID           string    `json:"id"`
	ProblemID    string    `json:"problem_id"`
	UserID       string    `json:"user_id"`
	Language     string    `json:"language"`
	Code         string    `json:"code"`
	Status       string    `json:"status"`
	TimeUsedMs   int       `json:"time_used_ms"`
	MemoryUsedKB int       `json:"memory_used_kb"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (s *Submission) Validate() error {
	s.ProblemID = strings.TrimSpace(s.ProblemID)
	s.UserID = strings.TrimSpace(s.UserID)
	s.Language = strings.TrimSpace(s.Language)
	if s.ProblemID == "" {
		return NewValidationError("problem_id", "题目不能为空")
	}
	if s.UserID == "" {
		return NewValidationError("user_id", "提交用户不能为空")
	}
	if s.Language == "" {
		s.Language = LanguageGo
	}
	if !isValidLanguage(s.Language) {
		return NewValidationError("language", "编程语言不合法")
	}
	if strings.TrimSpace(s.Code) == "" {
		return NewValidationError("code", "代码不能为空")
	}
	if s.Status == "" {
		s.Status = SubmissionPending
	}
	if !isValidSubmissionStatus(s.Status) {
		return NewValidationError("status", "提交状态不合法")
	}
	return nil
}

func isValidLanguage(s string) bool {
	switch s {
	case LanguageGo, LanguageCpp, LanguageJava, LanguagePy:
		return true
	}
	return false
}

func isValidSubmissionStatus(s string) bool {
	switch s {
	case SubmissionPending, SubmissionJudging, SubmissionAccepted, SubmissionWrongAnswer,
		SubmissionRuntimeError, SubmissionTimeLimit, SubmissionMemoryLimit, SubmissionCompileError:
		return true
	}
	return false
}

// IsFinalVerdict 判断是否为终态判题结果。
func IsFinalVerdict(s string) bool {
	switch s {
	case SubmissionAccepted, SubmissionWrongAnswer, SubmissionRuntimeError,
		SubmissionTimeLimit, SubmissionMemoryLimit, SubmissionCompileError:
		return true
	}
	return false
}

// submissionTransitions 提交状态机。
var submissionTransitions = map[string][]string{
	SubmissionPending:      {SubmissionJudging},
	SubmissionJudging:      {SubmissionAccepted, SubmissionWrongAnswer, SubmissionRuntimeError, SubmissionTimeLimit, SubmissionMemoryLimit, SubmissionCompileError},
	SubmissionAccepted:     {},
	SubmissionWrongAnswer:  {},
	SubmissionRuntimeError: {},
	SubmissionTimeLimit:    {},
	SubmissionMemoryLimit:  {},
	SubmissionCompileError: {},
}

// CanTransitionSubmission 判断提交状态 from 能否流转到 to。
func CanTransitionSubmission(from, to string) bool {
	for _, t := range submissionTransitions[from] {
		if t == to {
			return true
		}
	}
	return false
}

// SubmissionFilter 提交筛选条件。
type SubmissionFilter struct {
	ProblemID string
	UserID    string
	Status    string
	Language  string
}

func (f SubmissionFilter) Match(s *Submission) bool {
	if f.ProblemID != "" && s.ProblemID != f.ProblemID {
		return false
	}
	if f.UserID != "" && s.UserID != f.UserID {
		return false
	}
	if f.Status != "" && s.Status != f.Status {
		return false
	}
	if f.Language != "" && s.Language != f.Language {
		return false
	}
	return true
}
