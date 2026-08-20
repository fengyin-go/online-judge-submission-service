package handler

import (
	"net/http"

	"onlinejudge/internal/model"
	"onlinejudge/pkg/httpx"
)

func (s *Server) registerProblemRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/problems", s.createProblem)
	mux.HandleFunc("POST /api/problems/import", s.importProblems)
	mux.HandleFunc("GET /api/problems", s.listProblems)
	mux.HandleFunc("GET /api/problems/tags", s.listTags)
	mux.HandleFunc("GET /api/problems/{id}", s.getProblem)
	mux.HandleFunc("PUT /api/problems/{id}", s.updateProblem)
	mux.HandleFunc("DELETE /api/problems/{id}", s.deleteProblem)
}

type problemRequest struct {
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Difficulty    string   `json:"difficulty"`
	TimeLimitMs   int      `json:"time_limit_ms"`
	MemoryLimitMB int      `json:"memory_limit_mb"`
	Tags          []string `json:"tags"`
	Status        string   `json:"status"`
}

func (s *Server) createProblem(w http.ResponseWriter, r *http.Request) {
	var req problemRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	p, err := s.svc.CreateProblem(model.Problem{
		Title: req.Title, Description: req.Description, Difficulty: req.Difficulty,
		TimeLimitMs: req.TimeLimitMs, MemoryLimitMB: req.MemoryLimitMB,
		Tags: req.Tags, Status: req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, p)
}

type importProblemsRequest struct {
	Problems []problemRequest `json:"problems"`
}

func (s *Server) importProblems(w http.ResponseWriter, r *http.Request) {
	var req importProblemsRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	problems := make([]model.Problem, 0, len(req.Problems))
	for _, p := range req.Problems {
		problems = append(problems, model.Problem{
			Title: p.Title, Description: p.Description, Difficulty: p.Difficulty,
			TimeLimitMs: p.TimeLimitMs, MemoryLimitMB: p.MemoryLimitMB,
			Tags: p.Tags, Status: p.Status,
		})
	}
	result := s.svc.BatchImportProblems(problems)
	httpx.OK(w, result)
}

func (s *Server) listProblems(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ProblemFilter{
		Difficulty: r.URL.Query().Get("difficulty"),
		Status:     r.URL.Query().Get("status"),
		Tag:        r.URL.Query().Get("tag"),
		Keyword:    r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListProblems(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getProblem(w http.ResponseWriter, r *http.Request) {
	p, err := s.svc.GetProblem(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, p)
}

func (s *Server) updateProblem(w http.ResponseWriter, r *http.Request) {
	var req problemRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	p, err := s.svc.UpdateProblem(r.PathValue("id"), model.Problem{
		Title: req.Title, Description: req.Description, Difficulty: req.Difficulty,
		TimeLimitMs: req.TimeLimitMs, MemoryLimitMB: req.MemoryLimitMB,
		Tags: req.Tags, Status: req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, p)
}

func (s *Server) deleteProblem(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteProblem(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

func (s *Server) listTags(w http.ResponseWriter, r *http.Request) {
	tags, err := s.svc.ListTags()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, tags)
}
