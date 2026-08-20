package handler

import (
	"net/http"
	"strconv"

	"onlinejudge/internal/model"
	"onlinejudge/pkg/httpx"
)

func (s *Server) registerAnnouncementRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/announcements", s.createAnnouncement)
	mux.HandleFunc("GET /api/announcements", s.listAnnouncements)
	mux.HandleFunc("GET /api/announcements/{id}", s.getAnnouncement)
	mux.HandleFunc("PUT /api/announcements/{id}", s.updateAnnouncement)
	mux.HandleFunc("DELETE /api/announcements/{id}", s.deleteAnnouncement)
}

type announcementRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Pinned  bool   `json:"pinned"`
}

func (s *Server) createAnnouncement(w http.ResponseWriter, r *http.Request) {
	var req announcementRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.CreateAnnouncement(model.Announcement{
		Title: req.Title, Content: req.Content, Pinned: req.Pinned,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, a)
}

func (s *Server) listAnnouncements(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.AnnouncementFilter{Keyword: r.URL.Query().Get("keyword")}
	if v := r.URL.Query().Get("pinned"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			filter.Pinned = &b
		}
	}
	items, total, err := s.svc.ListAnnouncements(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getAnnouncement(w http.ResponseWriter, r *http.Request) {
	a, err := s.svc.GetAnnouncement(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

func (s *Server) updateAnnouncement(w http.ResponseWriter, r *http.Request) {
	var req announcementRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.UpdateAnnouncement(r.PathValue("id"), model.Announcement{
		Title: req.Title, Content: req.Content, Pinned: req.Pinned,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

func (s *Server) deleteAnnouncement(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteAnnouncement(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
