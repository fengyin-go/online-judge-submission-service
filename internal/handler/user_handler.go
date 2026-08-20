package handler

import (
	"net/http"

	"onlinejudge/internal/model"
	"onlinejudge/pkg/httpx"
)

func (s *Server) registerUserRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/users", s.createUser)
	mux.HandleFunc("POST /api/auth/login", s.login)
	mux.HandleFunc("POST /api/users/{id}/change-password", s.changePassword)
	mux.HandleFunc("GET /api/users", s.listUsers)
	mux.HandleFunc("GET /api/users/{id}", s.getUser)
	mux.HandleFunc("PUT /api/users/{id}", s.updateUser)
	mux.HandleFunc("DELETE /api/users/{id}", s.deleteUser)
}

type createUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	u, err := s.svc.CreateUser(model.User{
		Username: req.Username, Email: req.Email, Role: req.Role,
	}, req.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, u)
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	u, err := s.svc.Authenticate(req.Username, req.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, u)
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.UserFilter{
		Role:    r.URL.Query().Get("role"),
		Status:  r.URL.Query().Get("status"),
		Keyword: r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListUsers(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	u, err := s.svc.GetUser(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, u)
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email  string `json:"email"`
		Role   string `json:"role"`
		Status string `json:"status"`
	}
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	u, err := s.svc.UpdateUser(r.PathValue("id"), model.User{Email: req.Email, Role: req.Role, Status: req.Status})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, u)
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteUser(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (s *Server) changePassword(w http.ResponseWriter, r *http.Request) {
	var req changePasswordRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	if err := s.svc.ChangePassword(r.PathValue("id"), req.OldPassword, req.NewPassword); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"message": "密码已修改"})
}
