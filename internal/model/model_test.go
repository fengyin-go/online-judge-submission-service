package model

import (
	"testing"
	"time"
)

func TestUserValidate(t *testing.T) {
	cases := []struct {
		name    string
		u       User
		wantErr bool
	}{
		{"ok", User{Username: "alice", Role: RoleUser}, false},
		{"empty username", User{Username: " "}, true},
		{"short username", User{Username: "ab"}, true},
		{"bad role", User{Username: "alice", Role: "boss"}, true},
		{"bad status", User{Username: "alice", Status: "x"}, true},
		{"default role", User{Username: "alice"}, false},
	}
	for _, c := range cases {
		if err := c.u.Validate(); (err != nil) != c.wantErr {
			t.Errorf("%s: got err=%v wantErr=%v", c.name, err, c.wantErr)
		}
	}
	u := User{Username: "alice"}
	_ = u.Validate()
	if u.Role != RoleUser || u.Status != UserActive {
		t.Errorf("defaults not filled: role=%s status=%s", u.Role, u.Status)
	}
}

func TestUserFilterMatch(t *testing.T) {
	u := &User{Username: "alice", Email: "a@x.com", Role: RoleUser, Status: UserActive}
	cases := []struct {
		f    UserFilter
		want bool
	}{
		{UserFilter{}, true},
		{UserFilter{Role: RoleUser}, true},
		{UserFilter{Role: RoleAdmin}, false},
		{UserFilter{Keyword: "ali"}, true},
		{UserFilter{Keyword: "bob"}, false},
		{UserFilter{Status: UserDisabled}, false},
	}
	for _, c := range cases {
		if got := c.f.Match(u); got != c.want {
			t.Errorf("filter %+v: got %v want %v", c.f, got, c.want)
		}
	}
}

func TestProblemValidate(t *testing.T) {
	cases := []struct {
		name    string
		p       Problem
		wantErr bool
	}{
		{"ok", Problem{Title: "A+B", Description: "求和", Difficulty: DifficultyEasy}, false},
		{"empty title", Problem{Description: "x"}, true},
		{"empty desc", Problem{Title: "A+B"}, true},
		{"bad difficulty", Problem{Title: "A+B", Description: "x", Difficulty: "x"}, true},
		{"default difficulty", Problem{Title: "A+B", Description: "x"}, false},
		{"default limits", Problem{Title: "A+B", Description: "x", TimeLimitMs: 0}, false},
	}
	for _, c := range cases {
		if err := c.p.Validate(); (err != nil) != c.wantErr {
			t.Errorf("%s: got err=%v wantErr=%v", c.name, err, c.wantErr)
		}
	}
	p := Problem{Title: "A+B", Description: "x"}
	_ = p.Validate()
	if p.TimeLimitMs != 1000 || p.MemoryLimitMB != 256 {
		t.Errorf("limits not defaulted: %d %d", p.TimeLimitMs, p.MemoryLimitMB)
	}
}

func TestProblemFilterTag(t *testing.T) {
	p := &Problem{Title: "A+B", Difficulty: DifficultyEasy, Tags: []string{"math", "beginner"}}
	if !(ProblemFilter{Tag: "math"}).Match(p) {
		t.Error("tag math should match")
	}
	if (ProblemFilter{Tag: "dp"}).Match(p) {
		t.Error("tag dp should not match")
	}
	if !(ProblemFilter{Difficulty: DifficultyEasy}).Match(p) {
		t.Error("difficulty should match")
	}
}

func TestSubmissionValidate(t *testing.T) {
	cases := []struct {
		name    string
		s       Submission
		wantErr bool
	}{
		{"ok", Submission{ProblemID: "p1", UserID: "u1", Language: LanguageGo, Code: "package main"}, false},
		{"empty problem", Submission{UserID: "u1", Code: "x"}, true},
		{"empty user", Submission{ProblemID: "p1", Code: "x"}, true},
		{"empty code", Submission{ProblemID: "p1", UserID: "u1", Code: " "}, true},
		{"bad language", Submission{ProblemID: "p1", UserID: "u1", Language: "rust", Code: "x"}, true},
		{"default language", Submission{ProblemID: "p1", UserID: "u1", Code: "x"}, false},
		{"bad status", Submission{ProblemID: "p1", UserID: "u1", Code: "x", Status: "x"}, true},
	}
	for _, c := range cases {
		if err := c.s.Validate(); (err != nil) != c.wantErr {
			t.Errorf("%s: got err=%v wantErr=%v", c.name, err, c.wantErr)
		}
	}
}

func TestSubmissionStateMachine(t *testing.T) {
	cases := []struct {
		from, to string
		want     bool
	}{
		{SubmissionPending, SubmissionJudging, true},
		{SubmissionJudging, SubmissionAccepted, true},
		{SubmissionJudging, SubmissionWrongAnswer, true},
		{SubmissionPending, SubmissionAccepted, false},
		{SubmissionAccepted, SubmissionJudging, false},
	}
	for _, c := range cases {
		if got := CanTransitionSubmission(c.from, c.to); got != c.want {
			t.Errorf("transition %s->%s: got %v want %v", c.from, c.to, got, c.want)
		}
	}
}

func TestIsFinalVerdict(t *testing.T) {
	for _, v := range []string{SubmissionAccepted, SubmissionWrongAnswer, SubmissionRuntimeError, SubmissionTimeLimit, SubmissionMemoryLimit, SubmissionCompileError} {
		if !IsFinalVerdict(v) {
			t.Errorf("%s should be final", v)
		}
	}
	if IsFinalVerdict(SubmissionPending) || IsFinalVerdict(SubmissionJudging) {
		t.Error("pending/judging should not be final")
	}
}

func TestContestValidate(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name    string
		c       Contest
		wantErr bool
	}{
		{"ok", Contest{Title: "周赛", StartAt: now, EndAt: now.Add(time.Hour)}, false},
		{"empty title", Contest{StartAt: now, EndAt: now.Add(time.Hour)}, true},
		{"end before start", Contest{Title: "周赛", StartAt: now, EndAt: now.Add(-time.Hour)}, true},
		{"bad status", Contest{Title: "周赛", StartAt: now, EndAt: now.Add(time.Hour), Status: "x"}, true},
	}
	for _, c := range cases {
		if err := c.c.Validate(); (err != nil) != c.wantErr {
			t.Errorf("%s: got err=%v wantErr=%v", c.name, err, c.wantErr)
		}
	}
}

func TestContestStateMachine(t *testing.T) {
	if !CanTransitionContest(ContestRegistration, ContestRunning) {
		t.Error("registration->running should be allowed")
	}
	if !CanTransitionContest(ContestRunning, ContestEnded) {
		t.Error("running->ended should be allowed")
	}
	if CanTransitionContest(ContestRegistration, ContestEnded) {
		t.Error("registration->ended should be denied")
	}
	if CanTransitionContest(ContestEnded, ContestRunning) {
		t.Error("ended->running should be denied")
	}
}

func TestRegistrationValidate(t *testing.T) {
	if err := (&Registration{ContestID: "c1", UserID: "u1"}).Validate(); err != nil {
		t.Errorf("valid registration: %v", err)
	}
	if err := (&Registration{UserID: "u1"}).Validate(); err == nil {
		t.Error("missing contest should fail")
	}
	if err := (&Registration{ContestID: "c1"}).Validate(); err == nil {
		t.Error("missing user should fail")
	}
}

func TestAnnouncementValidate(t *testing.T) {
	if err := (&Announcement{Title: "通知", Content: "内容"}).Validate(); err != nil {
		t.Errorf("valid announcement: %v", err)
	}
	if err := (&Announcement{Content: "内容"}).Validate(); err == nil {
		t.Error("missing title should fail")
	}
	if err := (&Announcement{Title: "通知"}).Validate(); err == nil {
		t.Error("missing content should fail")
	}
}

func TestAnnouncementFilterPinned(t *testing.T) {
	pinned := true
	a := &Announcement{Title: "通知", Content: "x", Pinned: true}
	if !(AnnouncementFilter{Pinned: &pinned}).Match(a) {
		t.Error("pinned filter should match")
	}
	notPinned := false
	if (AnnouncementFilter{Pinned: &notPinned}).Match(a) {
		t.Error("not-pinned filter should not match")
	}
}

func TestSubmissionFilterMatch(t *testing.T) {
	sub := &Submission{ProblemID: "p1", UserID: "u1", Status: SubmissionAccepted, Language: LanguageGo}
	cases := []struct {
		f    SubmissionFilter
		want bool
	}{
		{SubmissionFilter{}, true},
		{SubmissionFilter{ProblemID: "p1"}, true},
		{SubmissionFilter{ProblemID: "p2"}, false},
		{SubmissionFilter{UserID: "u1"}, true},
		{SubmissionFilter{Status: SubmissionAccepted}, true},
		{SubmissionFilter{Status: SubmissionWrongAnswer}, false},
		{SubmissionFilter{Language: LanguageGo}, true},
		{SubmissionFilter{Language: LanguageCpp}, false},
	}
	for _, c := range cases {
		if got := c.f.Match(sub); got != c.want {
			t.Errorf("filter %+v: got %v want %v", c.f, got, c.want)
		}
	}
}

func TestJudgeResultValidate(t *testing.T) {
	cases := []struct {
		name    string
		j       JudgeResult
		wantErr bool
	}{
		{"ok", JudgeResult{SubmissionID: "s1", Verdict: SubmissionAccepted}, false},
		{"empty submission", JudgeResult{Verdict: SubmissionAccepted}, true},
		{"bad verdict", JudgeResult{SubmissionID: "s1", Verdict: "x"}, true},
	}
	for _, c := range cases {
		if err := c.j.Validate(); (err != nil) != c.wantErr {
			t.Errorf("%s: got err=%v wantErr=%v", c.name, err, c.wantErr)
		}
	}
}

func TestJudgeFilterMatch(t *testing.T) {
	j := &JudgeResult{SubmissionID: "s1", Verdict: SubmissionAccepted}
	if !(JudgeFilter{SubmissionID: "s1"}).Match(j) {
		t.Error("submission filter should match")
	}
	if (JudgeFilter{SubmissionID: "s2"}).Match(j) {
		t.Error("wrong submission should not match")
	}
	if !(JudgeFilter{Verdict: SubmissionAccepted}).Match(j) {
		t.Error("verdict filter should match")
	}
}

func TestRegistrationFilterMatch(t *testing.T) {
	r := &Registration{ContestID: "c1", UserID: "u1"}
	if !(RegistrationFilter{ContestID: "c1"}).Match(r) {
		t.Error("contest filter should match")
	}
	if (RegistrationFilter{UserID: "u2"}).Match(r) {
		t.Error("wrong user should not match")
	}
}

func TestContestFilterMatch(t *testing.T) {
	now := time.Now()
	c := &Contest{Title: "周赛", StartAt: now, EndAt: now.Add(time.Hour), Status: ContestRegistration}
	if !(ContestFilter{Status: ContestRegistration}).Match(c) {
		t.Error("status filter should match")
	}
	if !(ContestFilter{Keyword: "周赛"}).Match(c) {
		t.Error("keyword filter should match")
	}
	if (ContestFilter{Keyword: "月赛"}).Match(c) {
		t.Error("wrong keyword should not match")
	}
}
