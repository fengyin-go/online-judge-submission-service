package service

import (
	"testing"
	"time"

	"onlinejudge/internal/model"
)

func TestAddContestProblem(t *testing.T) {
	s := newTestService()
	now := time.Now()
	c, _ := s.CreateContest(model.Contest{Title: "周赛", StartAt: now, EndAt: now.Add(time.Hour)})
	p1 := seedProblem(s, "A题")
	p2 := seedProblem(s, "B题")

	if _, err := s.AddContestProblem(c.ID, p1.ID); err != nil {
		t.Fatalf("add problem: %v", err)
	}
	if _, err := s.AddContestProblem(c.ID, p2.ID); err != nil {
		t.Fatalf("add second problem: %v", err)
	}
	// 重复添加
	if _, err := s.AddContestProblem(c.ID, p1.ID); err == nil {
		t.Fatal("want duplicate problem error")
	}
	// 不存在的题目
	if _, err := s.AddContestProblem(c.ID, "nope"); err == nil {
		t.Fatal("want missing problem error")
	}
	problems, err := s.ListContestProblems(c.ID)
	if err != nil || len(problems) != 2 {
		t.Fatalf("list problems: len=%d err=%v", len(problems), err)
	}
}

func TestRemoveContestProblem(t *testing.T) {
	s := newTestService()
	now := time.Now()
	c, _ := s.CreateContest(model.Contest{Title: "周赛", StartAt: now, EndAt: now.Add(time.Hour)})
	p1 := seedProblem(s, "A题")
	_, _ = s.AddContestProblem(c.ID, p1.ID)

	if _, err := s.RemoveContestProblem(c.ID, p1.ID); err != nil {
		t.Fatalf("remove: %v", err)
	}
	problems, _ := s.ListContestProblems(c.ID)
	if len(problems) != 0 {
		t.Fatalf("want 0 problems, got %d", len(problems))
	}
}

func TestAddContestProblemAfterStart(t *testing.T) {
	s := newTestService()
	now := time.Now()
	c, _ := s.CreateContest(model.Contest{Title: "周赛", StartAt: now, EndAt: now.Add(time.Hour)})
	p := seedProblem(s, "A题")
	_, _ = s.TransitionContest(c.ID, model.ContestRunning)
	if _, err := s.AddContestProblem(c.ID, p.ID); err == nil {
		t.Fatal("want error adding problem after start")
	}
}

func TestChangePassword(t *testing.T) {
	s := newTestService()
	u := seedUser(s, "alice")
	if err := s.ChangePassword(u.ID, "secret123", "newpass"); err != nil {
		t.Fatalf("change password: %v", err)
	}
	// 旧密码登录失败
	if _, err := s.Authenticate("alice", "secret123"); err == nil {
		t.Fatal("old password should fail")
	}
	// 新密码登录成功
	if _, err := s.Authenticate("alice", "newpass"); err != nil {
		t.Fatalf("new password should work: %v", err)
	}
	// 旧密码错误时改密失败
	if err := s.ChangePassword(u.ID, "wrong", "x"); err == nil {
		t.Fatal("want wrong old password error")
	}
}

func TestLanguageDistribution(t *testing.T) {
	s := newTestService()
	u := seedUser(s, "alice")
	p := seedProblem(s, "A题")
	seedSubmission(s, p.ID, u.ID, model.SubmissionAccepted)
	// 手动构造不同语言的提交
	sub, _ := s.CreateSubmission(model.Submission{ProblemID: p.ID, UserID: u.ID, Language: model.LanguageCpp, Code: "x"})
	sub.Status = model.SubmissionAccepted
	_ = s.store.UpdateSubmission(sub)

	dist, err := s.LanguageDistribution()
	if err != nil {
		t.Fatalf("language dist: %v", err)
	}
	total := 0
	for _, l := range dist {
		total += l.Count
	}
	if total != 2 {
		t.Fatalf("want 2 total, got %d", total)
	}
}

func TestUserSubmissionSummaries(t *testing.T) {
	s := newTestService()
	alice := seedUser(s, "alice")
	p := seedProblem(s, "A题")
	seedSubmission(s, p.ID, alice.ID, model.SubmissionAccepted)
	seedSubmission(s, p.ID, alice.ID, model.SubmissionWrongAnswer)

	summaries, err := s.UserSubmissionSummaries()
	if err != nil {
		t.Fatalf("summaries: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("want 1 summary, got %d", len(summaries))
	}
	if summaries[0].Total != 2 || summaries[0].Accepted != 1 {
		t.Fatalf("unexpected summary: %+v", summaries[0])
	}
}

func TestUserDisabledLogin(t *testing.T) {
	s := newTestService()
	u := seedUser(s, "alice")
	_, _ = s.UpdateUser(u.ID, model.User{Status: model.UserDisabled})
	if _, err := s.Authenticate("alice", "secret123"); err == nil {
		t.Fatal("disabled user should not login")
	}
}

func TestRejudgeSubmission(t *testing.T) {
	s := newTestService()
	u := seedUser(s, "alice")
	p := seedProblem(s, "A题")
	sub, _ := s.CreateSubmission(model.Submission{ProblemID: p.ID, UserID: u.ID, Code: "code"})
	_, _ = s.JudgeSubmission(sub.ID)

	// 重判终态提交
	result, err := s.RejudgeSubmission(sub.ID)
	if err != nil {
		t.Fatalf("rejudge: %v", err)
	}
	if !model.IsFinalVerdict(result.Verdict) {
		t.Fatalf("want final verdict after rejudge, got %s", result.Verdict)
	}
	// 判题明细只剩一条（旧的被删除）
	detail, err := s.GetJudgeResult(sub.ID)
	if err != nil || detail.ID != result.ID {
		t.Fatalf("judge result should be replaced: %v %v", detail, err)
	}
}

func TestRejudgePendingFails(t *testing.T) {
	s := newTestService()
	u := seedUser(s, "alice")
	p := seedProblem(s, "A题")
	sub, _ := s.CreateSubmission(model.Submission{ProblemID: p.ID, UserID: u.ID, Code: "code"})
	// pending 状态不可重判
	if _, err := s.RejudgeSubmission(sub.ID); err == nil {
		t.Fatal("want error rejudging pending submission")
	}
}

func TestListTags(t *testing.T) {
	s := newTestService()
	_, _ = s.CreateProblem(model.Problem{Title: "A题", Description: "x", Tags: []string{"dp", "math"}})
	_, _ = s.CreateProblem(model.Problem{Title: "B题", Description: "x", Tags: []string{"math", "graph"}})
	tags, err := s.ListTags()
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	if len(tags) != 3 {
		t.Fatalf("want 3 distinct tags, got %v", tags)
	}
	if tags[0] != "dp" || tags[1] != "graph" || tags[2] != "math" {
		t.Fatalf("tags should be sorted, got %v", tags)
	}
}

func TestListTagsEmpty(t *testing.T) {
	s := newTestService()
	tags, err := s.ListTags()
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	if len(tags) != 0 {
		t.Fatalf("want empty, got %v", tags)
	}
}
