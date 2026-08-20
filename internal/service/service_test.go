package service

import (
	"testing"
	"time"

	"onlinejudge/internal/config"
	"onlinejudge/internal/model"
	"onlinejudge/internal/store"
	"onlinejudge/pkg/logger"
)

func newTestService() *Service {
	cfg := &config.Config{MaxPageSize: 100}
	log := logger.NewLevel(logger.LevelError)
	return New(store.NewMemoryStore(), log, cfg)
}

func seedUser(s *Service, username string) *model.User {
	u, _ := s.CreateUser(model.User{Username: username}, "secret123")
	return u
}

func seedProblem(s *Service, title string) *model.Problem {
	p, _ := s.CreateProblem(model.Problem{Title: title, Description: "描述", Difficulty: model.DifficultyEasy})
	return p
}

func TestCreateUserAndAuthenticate(t *testing.T) {
	s := newTestService()
	u := seedUser(s, "alice")
	if u.Role != model.RoleUser {
		t.Fatalf("want default role user, got %s", u.Role)
	}
	// 用户名重复
	if _, err := s.CreateUser(model.User{Username: "alice"}, "x"); err == nil {
		t.Fatal("want duplicate username error")
	}
	// 登录成功
	got, err := s.Authenticate("alice", "secret123")
	if err != nil || got.ID != u.ID {
		t.Fatalf("auth: %v %v", got, err)
	}
	// 密码错误
	if _, err := s.Authenticate("alice", "wrong"); err == nil {
		t.Fatal("want wrong password error")
	}
	// 用户不存在
	if _, err := s.Authenticate("nobody", "x"); err == nil {
		t.Fatal("want unknown user error")
	}
}

func TestHashPassword(t *testing.T) {
	if HashPassword("abc") == HashPassword("abd") {
		t.Fatal("different passwords should hash differently")
	}
	if HashPassword("abc") != HashPassword("abc") {
		t.Fatal("same password should hash consistently")
	}
}

func TestCreateUserRequiresPassword(t *testing.T) {
	s := newTestService()
	if _, err := s.CreateUser(model.User{Username: "bob"}, ""); err == nil {
		t.Fatal("want password required error")
	}
}

func TestBatchImportProblems(t *testing.T) {
	s := newTestService()
	result := s.BatchImportProblems([]model.Problem{
		{Title: "A+B", Description: "求和"},
		{Title: "排序", Description: "排序"},
		{Title: "", Description: "无标题"},
	})
	if len(result.Success) != 2 || len(result.Failed) != 1 {
		t.Fatalf("import: success=%d failed=%d", len(result.Success), len(result.Failed))
	}
}

func TestCreateSubmissionRequiresProblem(t *testing.T) {
	s := newTestService()
	u := seedUser(s, "alice")
	if _, err := s.CreateSubmission(model.Submission{ProblemID: "nope", UserID: u.ID, Code: "x"}); err == nil {
		t.Fatal("want missing problem error")
	}
	if _, err := s.CreateSubmission(model.Submission{ProblemID: "p1", UserID: "nope", Code: "x"}); err == nil {
		t.Fatal("want missing user error")
	}
}

func TestJudgeSubmissionFlow(t *testing.T) {
	s := newTestService()
	u := seedUser(s, "alice")
	p := seedProblem(s, "A+B")
	sub, err := s.CreateSubmission(model.Submission{ProblemID: p.ID, UserID: u.ID, Language: model.LanguageGo, Code: "package main\nfunc main(){}"})
	if err != nil {
		t.Fatalf("create submission: %v", err)
	}
	if sub.Status != model.SubmissionPending {
		t.Fatalf("want pending, got %s", sub.Status)
	}
	result, err := s.JudgeSubmission(sub.ID)
	if err != nil {
		t.Fatalf("judge: %v", err)
	}
	if !model.IsFinalVerdict(result.Verdict) {
		t.Fatalf("want final verdict, got %s", result.Verdict)
	}
	got, _ := s.GetSubmission(sub.ID)
	if got.Status != result.Verdict {
		t.Fatalf("submission status %s != verdict %s", got.Status, result.Verdict)
	}
	// 重复判题应报错
	if _, err := s.JudgeSubmission(sub.ID); err == nil {
		t.Fatal("want re-judge error")
	}
	// 判题明细可查
	detail, err := s.GetJudgeResult(sub.ID)
	if err != nil || detail.Verdict != result.Verdict {
		t.Fatalf("get judge result: %v %v", detail, err)
	}
}

func TestSimulateJudgeDeterministic(t *testing.T) {
	sub := &model.Submission{Code: "hello"}
	v1, t1, m1 := simulateJudge(sub, 1000)
	v2, t2, m2 := simulateJudge(sub, 1000)
	if v1 != v2 || t1 != t2 || m1 != m2 {
		t.Fatal("simulateJudge should be deterministic")
	}
}

func TestContestLifecycle(t *testing.T) {
	s := newTestService()
	now := time.Now()
	c, err := s.CreateContest(model.Contest{Title: "周赛", StartAt: now, EndAt: now.Add(time.Hour)})
	if err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if c.Status != model.ContestRegistration {
		t.Fatalf("want registration, got %s", c.Status)
	}
	if _, err := s.TransitionContest(c.ID, model.ContestEnded); err == nil {
		t.Fatal("registration->ended should fail")
	}
	if _, err := s.TransitionContest(c.ID, model.ContestRunning); err != nil {
		t.Fatalf("registration->running: %v", err)
	}
	if _, err := s.TransitionContest(c.ID, model.ContestEnded); err != nil {
		t.Fatalf("running->ended: %v", err)
	}
}

func TestRegisterFlow(t *testing.T) {
	s := newTestService()
	u := seedUser(s, "alice")
	now := time.Now()
	c, _ := s.CreateContest(model.Contest{Title: "周赛", StartAt: now, EndAt: now.Add(time.Hour)})

	if _, err := s.Register(c.ID, u.ID); err != nil {
		t.Fatalf("register: %v", err)
	}
	// 重复报名
	if _, err := s.Register(c.ID, u.ID); err == nil {
		t.Fatal("want duplicate registration error")
	}
	// 比赛开始后不可报名
	_, _ = s.TransitionContest(c.ID, model.ContestRunning)
	if _, err := s.Register(c.ID, u.ID); err == nil {
		t.Fatal("want register-after-start error")
	}
}

func TestAnnouncementCRUD(t *testing.T) {
	s := newTestService()
	a, err := s.CreateAnnouncement(model.Announcement{Title: "通知", Content: "内容", Pinned: true})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	updated, err := s.UpdateAnnouncement(a.ID, model.Announcement{Title: "新通知", Content: "新内容", Pinned: false})
	if err != nil || updated.Title != "新通知" {
		t.Fatalf("update: %v %v", updated, err)
	}
	if err := s.DeleteAnnouncement(a.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestListProblemsFilter(t *testing.T) {
	s := newTestService()
	seedProblem(s, "A+B")
	seedProblem(s, "背包问题")
	items, total, err := s.ListProblems(model.ProblemFilter{Keyword: "背包"}, 1, 10)
	if err != nil || total != 1 {
		t.Fatalf("filter: total=%d err=%v", total, err)
	}
	if items[0].Title != "背包问题" {
		t.Fatalf("want 背包问题, got %s", items[0].Title)
	}
}
