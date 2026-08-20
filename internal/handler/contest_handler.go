package handler

import (
	"net/http"
	"time"

	"onlinejudge/internal/model"
	"onlinejudge/pkg/httpx"
)

func (s *Server) registerContestRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/contests", s.createContest)
	mux.HandleFunc("GET /api/contests", s.listContests)
	mux.HandleFunc("GET /api/contests/{id}", s.getContest)
	mux.HandleFunc("POST /api/contests/{id}/transition", s.transitionContest)
	mux.HandleFunc("POST /api/contests/{id}/problems", s.addContestProblem)
	mux.HandleFunc("DELETE /api/contests/{id}/problems/{problemID}", s.removeContestProblem)
	mux.HandleFunc("GET /api/contests/{id}/problems", s.listContestProblems)
	mux.HandleFunc("DELETE /api/contests/{id}", s.deleteContest)
}

type contestRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	StartAt     string `json:"start_at"` // RFC3339
	EndAt       string `json:"end_at"`
}

func (s *Server) createContest(w http.ResponseWriter, r *http.Request) {
	var req contestRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		httpx.BadRequest(w, "start_at 格式错误: "+err.Error())
		return
	}
	endAt, err := time.Parse(time.RFC3339, req.EndAt)
	if err != nil {
		httpx.BadRequest(w, "end_at 格式错误: "+err.Error())
		return
	}
	c, err := s.svc.CreateContest(model.Contest{
		Title: req.Title, Description: req.Description, StartAt: startAt, EndAt: endAt,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, c)
}

func (s *Server) listContests(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ContestFilter{
		Status:  r.URL.Query().Get("status"),
		Keyword: r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListContests(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getContest(w http.ResponseWriter, r *http.Request) {
	c, err := s.svc.GetContest(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, c)
}

type transitionRequest struct {
	Status string `json:"status"`
}

func (s *Server) transitionContest(w http.ResponseWriter, r *http.Request) {
	var req transitionRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	c, err := s.svc.TransitionContest(r.PathValue("id"), req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, c)
}

func (s *Server) deleteContest(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteContest(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

type contestProblemRequest struct {
	ProblemID string `json:"problem_id"`
}

func (s *Server) addContestProblem(w http.ResponseWriter, r *http.Request) {
	var req contestProblemRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	c, err := s.svc.AddContestProblem(r.PathValue("id"), req.ProblemID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, c)
}

func (s *Server) removeContestProblem(w http.ResponseWriter, r *http.Request) {
	c, err := s.svc.RemoveContestProblem(r.PathValue("id"), r.PathValue("problemID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, c)
}

func (s *Server) listContestProblems(w http.ResponseWriter, r *http.Request) {
	problems, err := s.svc.ListContestProblems(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, problems)
}
