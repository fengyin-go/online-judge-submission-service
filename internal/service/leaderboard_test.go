package service

import (
	"testing"
	"time"

	"onlinejudge/internal/model"
)

// seedSubmission 创建提交并直接置为指定状态（用于测试聚合逻辑，隔离判题的不确定性）。
func seedSubmission(s *Service, problemID, userID, status string) {
	sub, _ := s.CreateSubmission(model.Submission{ProblemID: problemID, UserID: userID, Code: "code"})
	sub.Status = status
	_ = s.store.UpdateSubmission(sub)
}

func TestLeaderboard(t *testing.T) {
	s := newTestService()
	alice := seedUser(s, "alice")
	bob := seedUser(s, "bob")
	p1 := seedProblem(s, "A题")
	p2 := seedProblem(s, "B题")

	// alice 通过 2 题，bob 通过 1 题
	seedSubmission(s, p1.ID, alice.ID, model.SubmissionAccepted)
	seedSubmission(s, p2.ID, alice.ID, model.SubmissionAccepted)
	seedSubmission(s, p1.ID, bob.ID, model.SubmissionAccepted)

	board, err := s.Leaderboard(10)
	if err != nil {
		t.Fatalf("leaderboard: %v", err)
	}
	if len(board) != 2 {
		t.Fatalf("want 2 entries, got %d", len(board))
	}
	if board[0].Username != "alice" || board[0].AcceptedCount != 2 {
		t.Fatalf("want alice first with 2 accepted, got %+v", board[0])
	}
	if board[1].Username != "bob" || board[1].AcceptedCount != 1 {
		t.Fatalf("want bob second with 1 accepted, got %+v", board[1])
	}
}

func TestLeaderboardLimit(t *testing.T) {
	s := newTestService()
	p := seedProblem(s, "A题")
	for i := 0; i < 5; i++ {
		u := seedUser(s, "user"+string(rune('a'+i)))
		seedSubmission(s, p.ID, u.ID, model.SubmissionAccepted)
	}
	board, _ := s.Leaderboard(3)
	if len(board) != 3 {
		t.Fatalf("want 3 entries, got %d", len(board))
	}
}

func TestLeaderboardEmpty(t *testing.T) {
	s := newTestService()
	board, err := s.Leaderboard(10)
	if err != nil {
		t.Fatalf("leaderboard: %v", err)
	}
	if len(board) != 0 {
		t.Fatalf("want empty, got %d", len(board))
	}
}

func TestVerdictDistribution(t *testing.T) {
	s := newTestService()
	u := seedUser(s, "alice")
	p := seedProblem(s, "A题")
	seedSubmission(s, p.ID, u.ID, model.SubmissionAccepted)
	seedSubmission(s, p.ID, u.ID, model.SubmissionWrongAnswer)
	seedSubmission(s, p.ID, u.ID, model.SubmissionAccepted)

	dist, err := s.VerdictDistribution()
	if err != nil {
		t.Fatalf("verdict dist: %v", err)
	}
	total := 0
	for _, v := range dist {
		total += v.Count
		if v.Verdict == model.SubmissionAccepted && v.Count != 2 {
			t.Fatalf("want 2 accepted, got %d", v.Count)
		}
	}
	if total != 3 {
		t.Fatalf("want 3 total, got %d", total)
	}
}

func TestProblemPassRates(t *testing.T) {
	s := newTestService()
	u := seedUser(s, "alice")
	p := seedProblem(s, "A题")
	seedSubmission(s, p.ID, u.ID, model.SubmissionAccepted)
	seedSubmission(s, p.ID, u.ID, model.SubmissionWrongAnswer)
	seedSubmission(s, p.ID, u.ID, model.SubmissionAccepted)

	rates, err := s.ProblemPassRates()
	if err != nil {
		t.Fatalf("pass rates: %v", err)
	}
	if len(rates) != 1 || rates[0].Submissions != 3 || rates[0].Accepted != 2 {
		t.Fatalf("unexpected rates: %+v", rates)
	}
	if rates[0].PassRate < 66 || rates[0].PassRate > 67 {
		t.Fatalf("unexpected pass rate: %f", rates[0].PassRate)
	}
}

func TestContestRanking(t *testing.T) {
	s := newTestService()
	now := time.Now()
	contest, _ := s.CreateContest(model.Contest{Title: "周赛", StartAt: now.Add(-time.Hour), EndAt: now.Add(time.Hour)})

	alice := seedUser(s, "alice")
	bob := seedUser(s, "bob")
	p := seedProblem(s, "A题")

	_, _ = s.Register(contest.ID, alice.ID)
	_, _ = s.Register(contest.ID, bob.ID)

	seedSubmission(s, p.ID, alice.ID, model.SubmissionAccepted)
	seedSubmission(s, p.ID, bob.ID, model.SubmissionWrongAnswer)

	ranking, err := s.ContestRanking(contest.ID)
	if err != nil {
		t.Fatalf("ranking: %v", err)
	}
	if len(ranking) != 2 {
		t.Fatalf("want 2 entries, got %d", len(ranking))
	}
	if ranking[0].Username != "alice" || ranking[0].Solved != 1 {
		t.Fatalf("want alice first solved 1, got %+v", ranking[0])
	}
	if ranking[1].Username != "bob" || ranking[1].Solved != 0 {
		t.Fatalf("want bob second solved 0, got %+v", ranking[1])
	}
	if ranking[0].Rank != 1 || ranking[1].Rank != 2 {
		t.Fatalf("ranks should be 1 and 2, got %+v", ranking)
	}
}

func TestContestRankingRequiresContest(t *testing.T) {
	s := newTestService()
	if _, err := s.ContestRanking("nope"); err == nil {
		t.Fatal("want error for missing contest")
	}
}
