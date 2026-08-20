// Package store 定义数据访问接口与内存实现。
package store

import (
	"errors"

	"onlinejudge/internal/model"
)

var (
	ErrNotFound = errors.New("记录不存在")
	ErrConflict = errors.New("记录已存在或状态冲突")
)

// Store 聚合全部实体的数据访问方法。
type Store interface {
	CreateUser(u *model.User) error
	GetUser(id string) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
	ListUsers() []*model.User
	UpdateUser(u *model.User) error
	DeleteUser(id string) error

	CreateProblem(p *model.Problem) error
	GetProblem(id string) (*model.Problem, error)
	ListProblems() []*model.Problem
	UpdateProblem(p *model.Problem) error
	DeleteProblem(id string) error

	CreateSubmission(s *model.Submission) error
	GetSubmission(id string) (*model.Submission, error)
	ListSubmissions() []*model.Submission
	UpdateSubmission(s *model.Submission) error
	DeleteSubmission(id string) error

	CreateJudgeResult(j *model.JudgeResult) error
	GetJudgeResult(id string) (*model.JudgeResult, error)
	GetJudgeResultBySubmission(submissionID string) (*model.JudgeResult, error)
	ListJudgeResults() []*model.JudgeResult
	DeleteJudgeResult(id string) error

	CreateContest(c *model.Contest) error
	GetContest(id string) (*model.Contest, error)
	ListContests() []*model.Contest
	UpdateContest(c *model.Contest) error
	DeleteContest(id string) error

	CreateRegistration(r *model.Registration) error
	GetRegistration(id string) (*model.Registration, error)
	ListRegistrations() []*model.Registration
	DeleteRegistration(id string) error

	CreateAnnouncement(a *model.Announcement) error
	GetAnnouncement(id string) (*model.Announcement, error)
	ListAnnouncements() []*model.Announcement
	UpdateAnnouncement(a *model.Announcement) error
	DeleteAnnouncement(id string) error
}
