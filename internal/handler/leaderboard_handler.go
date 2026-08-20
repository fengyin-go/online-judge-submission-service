package handler

import (
	"net/http"
	"strconv"

	"onlinejudge/pkg/httpx"
)

func (s *Server) registerLeaderboardRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/leaderboard", s.leaderboard)
	mux.HandleFunc("GET /api/stats/verdicts", s.verdictDistribution)
	mux.HandleFunc("GET /api/stats/pass-rates", s.problemPassRates)
	mux.HandleFunc("GET /api/stats/languages", s.languageDistribution)
	mux.HandleFunc("GET /api/stats/users", s.userSubmissionSummaries)
	mux.HandleFunc("GET /api/contests/{id}/ranking", s.contestRanking)
}

func (s *Server) leaderboard(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	entries, err := s.svc.Leaderboard(limit)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, entries)
}

func (s *Server) verdictDistribution(w http.ResponseWriter, r *http.Request) {
	dist, err := s.svc.VerdictDistribution()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, dist)
}

func (s *Server) problemPassRates(w http.ResponseWriter, r *http.Request) {
	rates, err := s.svc.ProblemPassRates()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, rates)
}

func (s *Server) contestRanking(w http.ResponseWriter, r *http.Request) {
	ranking, err := s.svc.ContestRanking(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, ranking)
}

func (s *Server) languageDistribution(w http.ResponseWriter, r *http.Request) {
	dist, err := s.svc.LanguageDistribution()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, dist)
}

func (s *Server) userSubmissionSummaries(w http.ResponseWriter, r *http.Request) {
	summaries, err := s.svc.UserSubmissionSummaries()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, summaries)
}
