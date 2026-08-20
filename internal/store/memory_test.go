package store

import (
	"testing"
	"time"

	"onlinejudge/internal/model"
)

func TestUserCRUDAndUnique(t *testing.T) {
	s := NewMemoryStore()
	u := &model.User{ID: "u1", Username: "alice", Role: model.RoleUser}
	if err := s.CreateUser(u); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.CreateUser(&model.User{ID: "u2", Username: "alice"}); err != ErrConflict {
		t.Fatalf("want conflict, got %v", err)
	}
	got, err := s.GetUserByUsername("alice")
	if err != nil || got.ID != "u1" {
		t.Fatalf("get by username: %v %v", got, err)
	}
	if _, err := s.GetUserByUsername("bob"); err != ErrNotFound {
		t.Fatalf("want not found, got %v", err)
	}
	if err := s.DeleteUser("u1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(s.ListUsers()) != 0 {
		t.Fatal("want 0 users")
	}
}

func TestProblemCRUDAndUnique(t *testing.T) {
	s := NewMemoryStore()
	p := &model.Problem{ID: "p1", Title: "A+B", Description: "求和"}
	if err := s.CreateProblem(p); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.CreateProblem(&model.Problem{ID: "p2", Title: "A+B", Description: "x"}); err != ErrConflict {
		t.Fatalf("want conflict, got %v", err)
	}
	p.Description = "更新"
	if err := s.UpdateProblem(p); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := s.DeleteProblem("p1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestSubmissionCRUD(t *testing.T) {
	s := NewMemoryStore()
	sub := &model.Submission{ID: "s1", ProblemID: "p1", UserID: "u1", Status: model.SubmissionPending}
	if err := s.CreateSubmission(sub); err != nil {
		t.Fatalf("create: %v", err)
	}
	sub.Status = model.SubmissionAccepted
	if err := s.UpdateSubmission(sub); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := s.GetSubmission("s1"); err != nil {
		t.Fatalf("get: %v", err)
	}
	if err := s.DeleteSubmission("s1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetSubmission("s1"); err != ErrNotFound {
		t.Fatalf("want not found, got %v", err)
	}
}

func TestJudgeResultBySubmission(t *testing.T) {
	s := NewMemoryStore()
	j := &model.JudgeResult{ID: "j1", SubmissionID: "s1", Verdict: model.SubmissionAccepted}
	if err := s.CreateJudgeResult(j); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := s.GetJudgeResultBySubmission("s1")
	if err != nil || got.ID != "j1" {
		t.Fatalf("get by submission: %v %v", got, err)
	}
	if _, err := s.GetJudgeResultBySubmission("s2"); err != ErrNotFound {
		t.Fatalf("want not found, got %v", err)
	}
	if len(s.ListJudgeResults()) != 1 {
		t.Fatal("want 1 judge result")
	}
}

func TestContestCRUDAndUnique(t *testing.T) {
	s := NewMemoryStore()
	now := time.Now()
	c := &model.Contest{ID: "c1", Title: "周赛", StartAt: now, EndAt: now.Add(time.Hour)}
	if err := s.CreateContest(c); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.CreateContest(&model.Contest{ID: "c2", Title: "周赛", StartAt: now, EndAt: now.Add(time.Hour)}); err != ErrConflict {
		t.Fatalf("want conflict, got %v", err)
	}
	if err := s.DeleteContest("c1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestRegistrationUnique(t *testing.T) {
	s := NewMemoryStore()
	r := &model.Registration{ID: "r1", ContestID: "c1", UserID: "u1"}
	if err := s.CreateRegistration(r); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.CreateRegistration(&model.Registration{ID: "r2", ContestID: "c1", UserID: "u1"}); err != ErrConflict {
		t.Fatalf("want conflict, got %v", err)
	}
	if err := s.DeleteRegistration("r1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestAnnouncementCRUD(t *testing.T) {
	s := NewMemoryStore()
	a := &model.Announcement{ID: "a1", Title: "通知", Content: "内容", Pinned: true}
	if err := s.CreateAnnouncement(a); err != nil {
		t.Fatalf("create: %v", err)
	}
	a.Content = "更新"
	if err := s.UpdateAnnouncement(a); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := s.DeleteAnnouncement("a1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestUpdateNotFound(t *testing.T) {
	s := NewMemoryStore()
	if err := s.UpdateUser(&model.User{ID: "nope", Username: "x"}); err != ErrNotFound {
		t.Fatalf("user: want not found, got %v", err)
	}
	if err := s.UpdateProblem(&model.Problem{ID: "nope", Title: "x"}); err != ErrNotFound {
		t.Fatalf("problem: want not found, got %v", err)
	}
	if err := s.UpdateSubmission(&model.Submission{ID: "nope"}); err != ErrNotFound {
		t.Fatalf("submission: want not found, got %v", err)
	}
	if err := s.UpdateContest(&model.Contest{ID: "nope", Title: "x"}); err != ErrNotFound {
		t.Fatalf("contest: want not found, got %v", err)
	}
	if err := s.UpdateAnnouncement(&model.Announcement{ID: "nope", Title: "x"}); err != ErrNotFound {
		t.Fatalf("announcement: want not found, got %v", err)
	}
}

func TestDeleteNotFound(t *testing.T) {
	s := NewMemoryStore()
	if err := s.DeleteUser("nope"); err != ErrNotFound {
		t.Fatalf("user: %v", err)
	}
	if err := s.DeleteProblem("nope"); err != ErrNotFound {
		t.Fatalf("problem: %v", err)
	}
	if err := s.DeleteJudgeResult("nope"); err != ErrNotFound {
		t.Fatalf("judge: %v", err)
	}
	if err := s.DeleteRegistration("nope"); err != ErrNotFound {
		t.Fatalf("registration: %v", err)
	}
}
