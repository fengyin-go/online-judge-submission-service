package handler

import (
	"net/http"

	"onlinejudge/internal/model"
	"onlinejudge/pkg/httpx"
)

func (s *Server) registerRegistrationRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/registrations", s.register)
	mux.HandleFunc("GET /api/registrations", s.listRegistrations)
	mux.HandleFunc("GET /api/registrations/{id}", s.getRegistration)
	mux.HandleFunc("DELETE /api/registrations/{id}", s.deleteRegistration)
}

type registrationRequest struct {
	ContestID string `json:"contest_id"`
	UserID    string `json:"user_id"`
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var req registrationRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	reg, err := s.svc.Register(req.ContestID, req.UserID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, reg)
}

func (s *Server) listRegistrations(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.RegistrationFilter{
		ContestID: r.URL.Query().Get("contest_id"),
		UserID:    r.URL.Query().Get("user_id"),
	}
	items, total, err := s.svc.ListRegistrations(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getRegistration(w http.ResponseWriter, r *http.Request) {
	reg, err := s.svc.GetRegistration(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, reg)
}

func (s *Server) deleteRegistration(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteRegistration(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
