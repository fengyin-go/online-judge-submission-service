package handler

import (
	"net/http"

	"onlinejudge/internal/model"
	"onlinejudge/pkg/httpx"
)

func (s *Server) registerSubmissionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/submissions", s.createSubmission)
	mux.HandleFunc("GET /api/submissions", s.listSubmissions)
	mux.HandleFunc("GET /api/submissions/{id}", s.getSubmission)
	mux.HandleFunc("POST /api/submissions/{id}/judge", s.judgeSubmission)
	mux.HandleFunc("POST /api/submissions/{id}/rejudge", s.rejudgeSubmission)
	mux.HandleFunc("GET /api/submissions/{id}/result", s.submissionResult)
	mux.HandleFunc("DELETE /api/submissions/{id}", s.deleteSubmission)
}

type submissionRequest struct {
	ProblemID string `json:"problem_id"`
	UserID    string `json:"user_id"`
	Language  string `json:"language"`
	Code      string `json:"code"`
}

func (s *Server) createSubmission(w http.ResponseWriter, r *http.Request) {
	var req submissionRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	sub, err := s.svc.CreateSubmission(model.Submission{
		ProblemID: req.ProblemID, UserID: req.UserID, Language: req.Language, Code: req.Code,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, sub)
}

func (s *Server) listSubmissions(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.SubmissionFilter{
		ProblemID: r.URL.Query().Get("problem_id"),
		UserID:    r.URL.Query().Get("user_id"),
		Status:    r.URL.Query().Get("status"),
		Language:  r.URL.Query().Get("language"),
	}
	items, total, err := s.svc.ListSubmissions(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getSubmission(w http.ResponseWriter, r *http.Request) {
	sub, err := s.svc.GetSubmission(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, sub)
}

func (s *Server) judgeSubmission(w http.ResponseWriter, r *http.Request) {
	result, err := s.svc.JudgeSubmission(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, result)
}

func (s *Server) rejudgeSubmission(w http.ResponseWriter, r *http.Request) {
	result, err := s.svc.RejudgeSubmission(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, result)
}

func (s *Server) submissionResult(w http.ResponseWriter, r *http.Request) {
	sub, err := s.svc.GetSubmission(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	result, err := s.svc.GetJudgeResult(sub.ID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, result)
}

func (s *Server) deleteSubmission(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteSubmission(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
