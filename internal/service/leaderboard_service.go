package service

import (
	"sort"

	"onlinejudge/internal/model"
)

// LeaderboardEntry 排行榜条目。
type LeaderboardEntry struct {
	UserID          string  `json:"user_id"`
	Username        string  `json:"username"`
	AcceptedCount   int     `json:"accepted_count"`
	SubmissionCount int     `json:"submission_count"`
	PassRate        float64 `json:"pass_rate"`
}

// Leaderboard 返回按通过题数排名的用户榜（通过题数降序，提交数升序）。
func (s *Service) Leaderboard(limit int) ([]LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 10
	}
	users := s.store.ListUsers()
	submissions := s.store.ListSubmissions()

	nameMap := make(map[string]string)
	for _, u := range users {
		nameMap[u.ID] = u.Username
	}

	type agg struct {
		accepted    map[string]bool
		submissions int
	}
	stats := make(map[string]*agg)
	for _, sub := range submissions {
		st, ok := stats[sub.UserID]
		if !ok {
			st = &agg{accepted: make(map[string]bool)}
			stats[sub.UserID] = st
		}
		st.submissions++
		if sub.Status == model.SubmissionAccepted {
			st.accepted[sub.ProblemID] = true
		}
	}

	entries := make([]LeaderboardEntry, 0, len(stats))
	for userID, st := range stats {
		e := LeaderboardEntry{
			UserID:          userID,
			Username:        nameMap[userID],
			AcceptedCount:   len(st.accepted),
			SubmissionCount: st.submissions,
		}
		if st.submissions > 0 {
			e.PassRate = float64(len(st.accepted)) / float64(st.submissions) * 100
		}
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].AcceptedCount != entries[j].AcceptedCount {
			return entries[i].AcceptedCount > entries[j].AcceptedCount
		}
		return entries[i].SubmissionCount < entries[j].SubmissionCount
	})
	if len(entries) > limit {
		entries = entries[:limit]
	}
	return entries, nil
}

// VerdictCount 某个判题结果的数量。
type VerdictCount struct {
	Verdict string `json:"verdict"`
	Count   int    `json:"count"`
}

// VerdictDistribution 提交判题结果分布统计。
func (s *Service) VerdictDistribution() ([]VerdictCount, error) {
	order := []string{
		model.SubmissionAccepted,
		model.SubmissionWrongAnswer,
		model.SubmissionTimeLimit,
		model.SubmissionMemoryLimit,
		model.SubmissionRuntimeError,
		model.SubmissionCompileError,
		model.SubmissionPending,
		model.SubmissionJudging,
	}
	counts := make(map[string]int)
	for _, sub := range s.store.ListSubmissions() {
		counts[sub.Status]++
	}
	result := make([]VerdictCount, 0, len(order))
	for _, v := range order {
		if c, ok := counts[v]; ok {
			result = append(result, VerdictCount{Verdict: v, Count: c})
		}
	}
	return result, nil
}

// ProblemPassRate 单题通过率。
type ProblemPassRate struct {
	ProblemID   string  `json:"problem_id"`
	Title       string  `json:"title"`
	Difficulty  string  `json:"difficulty"`
	Submissions int     `json:"submissions"`
	Accepted    int     `json:"accepted"`
	PassRate    float64 `json:"pass_rate"`
}

// ProblemPassRates 返回各题目通过率（按通过率降序）。
func (s *Service) ProblemPassRates() ([]ProblemPassRate, error) {
	problems := s.store.ListProblems()
	submissions := s.store.ListSubmissions()

	titleMap := make(map[string]string)
	diffMap := make(map[string]string)
	for _, p := range problems {
		titleMap[p.ID] = p.Title
		diffMap[p.ID] = p.Difficulty
	}

	type agg struct {
		submissions int
		accepted    int
	}
	stats := make(map[string]*agg)
	for _, sub := range submissions {
		st, ok := stats[sub.ProblemID]
		if !ok {
			st = &agg{}
			stats[sub.ProblemID] = st
		}
		st.submissions++
		if sub.Status == model.SubmissionAccepted {
			st.accepted++
		}
	}

	result := make([]ProblemPassRate, 0, len(stats))
	for problemID, st := range stats {
		e := ProblemPassRate{
			ProblemID:   problemID,
			Title:       titleMap[problemID],
			Difficulty:  diffMap[problemID],
			Submissions: st.submissions,
			Accepted:    st.accepted,
		}
		if st.submissions > 0 {
			e.PassRate = float64(st.accepted) / float64(st.submissions) * 100
		}
		result = append(result, e)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].PassRate > result[j].PassRate
	})
	return result, nil
}

// ContestRankEntry 比赛排名条目。
type ContestRankEntry struct {
	Rank     int    `json:"rank"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Solved   int    `json:"solved"`
	Attempts int    `json:"attempts"`
	Penalty  int    `json:"penalty"`
}

// ContestRanking 计算指定比赛的排名（按通过题数降序、罚时升序）。
func (s *Service) ContestRanking(contestID string) ([]ContestRankEntry, error) {
	contest, err := s.store.GetContest(contestID)
	if err != nil {
		return nil, err
	}
	users := s.store.ListUsers()
	submissions := s.store.ListSubmissions()
	registrations := s.store.ListRegistrations()

	nameMap := make(map[string]string)
	for _, u := range users {
		nameMap[u.ID] = u.Username
	}

	// 参与比赛的用户集合。
	participants := make(map[string]bool)
	for _, r := range registrations {
		if r.ContestID == contestID {
			participants[r.UserID] = true
		}
	}

	type agg struct {
		solved   map[string]bool
		attempts int
		penalty  int
	}
	stats := make(map[string]*agg)
	for _, sub := range submissions {
		if !sub.CreatedAt.After(contest.StartAt) || !sub.CreatedAt.Before(contest.EndAt) {
			continue
		}
		if !participants[sub.UserID] {
			continue
		}
		st, ok := stats[sub.UserID]
		if !ok {
			st = &agg{solved: make(map[string]bool)}
			stats[sub.UserID] = st
		}
		if st.solved[sub.ProblemID] {
			continue
		}
		st.attempts++
		if sub.Status == model.SubmissionAccepted {
			st.solved[sub.ProblemID] = true
			st.penalty += st.attempts * 20
		}
	}

	entries := make([]ContestRankEntry, 0, len(stats))
	for userID, st := range stats {
		entries = append(entries, ContestRankEntry{
			UserID:   userID,
			Username: nameMap[userID],
			Solved:   len(st.solved),
			Attempts: st.attempts,
			Penalty:  st.penalty,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Solved != entries[j].Solved {
			return entries[i].Solved > entries[j].Solved
		}
		return entries[i].Penalty < entries[j].Penalty
	})
	for i := range entries {
		entries[i].Rank = i + 1
	}
	return entries, nil
}

// LanguageCount 某语言的提交数量。
type LanguageCount struct {
	Language string `json:"language"`
	Count    int    `json:"count"`
}

// LanguageDistribution 提交语言分布统计。
func (s *Service) LanguageDistribution() ([]LanguageCount, error) {
	counts := make(map[string]int)
	for _, sub := range s.store.ListSubmissions() {
		counts[sub.Language]++
	}
	result := make([]LanguageCount, 0, len(counts))
	for lang, c := range counts {
		result = append(result, LanguageCount{Language: lang, Count: c})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	return result, nil
}

// UserSubmissionSummary 用户提交汇总。
type UserSubmissionSummary struct {
	UserID    string   `json:"user_id"`
	Username  string   `json:"username"`
	Total     int      `json:"total"`
	Accepted  int      `json:"accepted"`
	PassRate  float64  `json:"pass_rate"`
	Languages []string `json:"languages"`
}

// UserSubmissionSummaries 返回各用户的提交汇总。
func (s *Service) UserSubmissionSummaries() ([]UserSubmissionSummary, error) {
	users := s.store.ListUsers()
	submissions := s.store.ListSubmissions()

	nameMap := make(map[string]string)
	for _, u := range users {
		nameMap[u.ID] = u.Username
	}

	type agg struct {
		total     int
		accepted  int
		languages map[string]bool
	}
	stats := make(map[string]*agg)
	for _, sub := range submissions {
		st, ok := stats[sub.UserID]
		if !ok {
			st = &agg{languages: make(map[string]bool)}
			stats[sub.UserID] = st
		}
		st.total++
		st.languages[sub.Language] = true
		if sub.Status == model.SubmissionAccepted {
			st.accepted++
		}
	}

	result := make([]UserSubmissionSummary, 0, len(stats))
	for userID, st := range stats {
		langs := make([]string, 0, len(st.languages))
		for l := range st.languages {
			langs = append(langs, l)
		}
		sort.Strings(langs)
		summary := UserSubmissionSummary{
			UserID:    userID,
			Username:  nameMap[userID],
			Total:     st.total,
			Accepted:  st.accepted,
			Languages: langs,
		}
		if st.total > 0 {
			summary.PassRate = float64(st.accepted) / float64(st.total) * 100
		}
		result = append(result, summary)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Accepted > result[j].Accepted
	})
	return result, nil
}
