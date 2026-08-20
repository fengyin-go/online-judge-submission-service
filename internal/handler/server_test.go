package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"onlinejudge/internal/config"
	"onlinejudge/internal/service"
	"onlinejudge/internal/store"
	"onlinejudge/pkg/logger"
)

func newTestServer() *Server {
	cfg := &config.Config{MaxPageSize: 100}
	log := logger.NewLevel(logger.LevelError)
	st := store.NewMemoryStore()
	svc := service.New(st, log, cfg)
	return NewServer(svc, log, cfg)
}

func doJSON(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	var rd io.Reader
	if body != "" {
		rd = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, rd)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func extractID(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v body=%s", err, rec.Body.String())
	}
	if resp.Data.ID == "" {
		t.Fatalf("empty id, body=%s", rec.Body.String())
	}
	return resp.Data.ID
}

func TestHandlerUserLoginFlow(t *testing.T) {
	h := newTestServer().Routes()
	rec := doJSON(h, http.MethodPost, "/api/users", `{"username":"alice","password":"secret123","email":"a@x.com"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create user: code=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = doJSON(h, http.MethodPost, "/api/auth/login", `{"username":"alice","password":"secret123"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("login: code=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = doJSON(h, http.MethodPost, "/api/auth/login", `{"username":"alice","password":"wrong"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("wrong password: code=%d", rec.Code)
	}
}

func TestHandlerProblemImport(t *testing.T) {
	h := newTestServer().Routes()
	body := `{"problems":[{"title":"A+B","description":"求和"},{"title":"排序","description":"排序"},{"title":"","description":"bad"}]}`
	rec := doJSON(h, http.MethodPost, "/api/problems/import", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("import: code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandlerSubmissionJudgeFlow(t *testing.T) {
	h := newTestServer().Routes()
	rec := doJSON(h, http.MethodPost, "/api/users", `{"username":"alice","password":"secret123"}`)
	userID := extractID(t, rec)
	rec = doJSON(h, http.MethodPost, "/api/problems", `{"title":"A+B","description":"求和"}`)
	problemID := extractID(t, rec)

	rec = doJSON(h, http.MethodPost, "/api/submissions", `{"problem_id":"`+problemID+`","user_id":"`+userID+`","language":"go","code":"package main"}`)
	subID := extractID(t, rec)

	rec = doJSON(h, http.MethodPost, "/api/submissions/"+subID+"/judge", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("judge: code=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = doJSON(h, http.MethodGet, "/api/submissions/"+subID+"/result", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("result: code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandlerContestRegisterAndRanking(t *testing.T) {
	h := newTestServer().Routes()
	rec := doJSON(h, http.MethodPost, "/api/users", `{"username":"alice","password":"secret123"}`)
	userID := extractID(t, rec)

	now := time.Now()
	start := now.Add(-time.Hour).Format(time.RFC3339)
	end := now.Add(time.Hour).Format(time.RFC3339)
	rec = doJSON(h, http.MethodPost, "/api/contests", `{"title":"周赛","start_at":"`+start+`","end_at":"`+end+`"}`)
	contestID := extractID(t, rec)

	rec = doJSON(h, http.MethodPost, "/api/registrations", `{"contest_id":"`+contestID+`","user_id":"`+userID+`"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register: code=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doJSON(h, http.MethodGet, "/api/contests/"+contestID+"/ranking", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("ranking: code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandlerLeaderboard(t *testing.T) {
	h := newTestServer().Routes()
	rec := doJSON(h, http.MethodGet, "/api/leaderboard?limit=5", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("leaderboard: code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandlerVerdictDistribution(t *testing.T) {
	h := newTestServer().Routes()
	rec := doJSON(h, http.MethodGet, "/api/stats/verdicts", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("verdicts: code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandlerAnnouncementCRUD(t *testing.T) {
	h := newTestServer().Routes()
	rec := doJSON(h, http.MethodPost, "/api/announcements", `{"title":"通知","content":"内容","pinned":true}`)
	id := extractID(t, rec)
	rec = doJSON(h, http.MethodGet, "/api/announcements/"+id, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("get: code=%d", rec.Code)
	}
	rec = doJSON(h, http.MethodPut, "/api/announcements/"+id, `{"title":"新","content":"新内容","pinned":false}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("update: code=%d", rec.Code)
	}
}

func TestHandlerNotFound(t *testing.T) {
	h := newTestServer().Routes()
	rec := doJSON(h, http.MethodGet, "/api/users/nope", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, code=%d", rec.Code)
	}
}

func TestHandlerValidationError(t *testing.T) {
	h := newTestServer().Routes()
	rec := doJSON(h, http.MethodPost, "/api/problems", `{"title":"","description":"x"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, code=%d", rec.Code)
	}
}

func TestHandlerAuthMiddleware(t *testing.T) {
	cfg := &config.Config{MaxPageSize: 100, AuthToken: "secret"}
	log := logger.NewLevel(logger.LevelError)
	st := store.NewMemoryStore()
	svc := service.New(st, log, cfg)
	srv := NewServer(svc, log, cfg)
	h := srv.Routes()

	rec := doJSON(h, http.MethodGet, "/api/leaderboard", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, code=%d", rec.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/leaderboard", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusOK {
		t.Fatalf("want 200 with token, code=%d", rec2.Code)
	}
}

func TestHandlerListProblemsFilter(t *testing.T) {
	h := newTestServer().Routes()
	_ = doJSON(h, http.MethodPost, "/api/problems", `{"title":"A+B","description":"求和","difficulty":"easy"}`)
	_ = doJSON(h, http.MethodPost, "/api/problems", `{"title":"背包","description":"DP","difficulty":"hard"}`)
	rec := doJSON(h, http.MethodGet, "/api/problems?difficulty=hard", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list: code=%d", rec.Code)
	}
	var resp struct {
		Data struct {
			Pagination struct {
				Total int `json:"total"`
			} `json:"pagination"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Data.Pagination.Total != 1 {
		t.Fatalf("want 1 hard problem, got %d", resp.Data.Pagination.Total)
	}
}

func TestHandlerContestProblems(t *testing.T) {
	h := newTestServer().Routes()
	rec := doJSON(h, http.MethodPost, "/api/problems", `{"title":"A+B","description":"求和"}`)
	problemID := extractID(t, rec)

	now := time.Now()
	start := now.Add(-time.Hour).Format(time.RFC3339)
	end := now.Add(time.Hour).Format(time.RFC3339)
	rec = doJSON(h, http.MethodPost, "/api/contests", `{"title":"周赛","start_at":"`+start+`","end_at":"`+end+`"}`)
	contestID := extractID(t, rec)

	rec = doJSON(h, http.MethodPost, "/api/contests/"+contestID+"/problems", `{"problem_id":"`+problemID+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("add problem: code=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = doJSON(h, http.MethodGet, "/api/contests/"+contestID+"/problems", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list problems: code=%d", rec.Code)
	}
	rec = doJSON(h, http.MethodDelete, "/api/contests/"+contestID+"/problems/"+problemID, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("remove problem: code=%d", rec.Code)
	}
}

func TestHandlerChangePassword(t *testing.T) {
	h := newTestServer().Routes()
	rec := doJSON(h, http.MethodPost, "/api/users", `{"username":"alice","password":"secret123"}`)
	userID := extractID(t, rec)

	rec = doJSON(h, http.MethodPost, "/api/users/"+userID+"/change-password", `{"old_password":"secret123","new_password":"newpass"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("change password: code=%d body=%s", rec.Code, rec.Body.String())
	}
	// 新密码可登录
	rec = doJSON(h, http.MethodPost, "/api/auth/login", `{"username":"alice","password":"newpass"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("login with new password: code=%d", rec.Code)
	}
}

func TestHandlerLanguageDistribution(t *testing.T) {
	h := newTestServer().Routes()
	rec := doJSON(h, http.MethodGet, "/api/stats/languages", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("languages: code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandlerUserSummaries(t *testing.T) {
	h := newTestServer().Routes()
	rec := doJSON(h, http.MethodGet, "/api/stats/users", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("user summaries: code=%d", rec.Code)
	}
}

func TestHandlerDuplicateUsername(t *testing.T) {
	h := newTestServer().Routes()
	_ = doJSON(h, http.MethodPost, "/api/users", `{"username":"alice","password":"secret123"}`)
	rec := doJSON(h, http.MethodPost, "/api/users", `{"username":"alice","password":"other"}`)
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusConflict {
		t.Fatalf("want 400/409, code=%d", rec.Code)
	}
}

func TestHandlerRejudge(t *testing.T) {
	h := newTestServer().Routes()
	rec := doJSON(h, http.MethodPost, "/api/users", `{"username":"alice","password":"secret123"}`)
	userID := extractID(t, rec)
	rec = doJSON(h, http.MethodPost, "/api/problems", `{"title":"A+B","description":"求和"}`)
	problemID := extractID(t, rec)
	rec = doJSON(h, http.MethodPost, "/api/submissions", `{"problem_id":"`+problemID+`","user_id":"`+userID+`","language":"go","code":"package main"}`)
	subID := extractID(t, rec)
	_ = doJSON(h, http.MethodPost, "/api/submissions/"+subID+"/judge", "")

	rec = doJSON(h, http.MethodPost, "/api/submissions/"+subID+"/rejudge", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("rejudge: code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandlerListTags(t *testing.T) {
	h := newTestServer().Routes()
	_ = doJSON(h, http.MethodPost, "/api/problems", `{"title":"A题","description":"x","tags":["dp","math"]}`)
	_ = doJSON(h, http.MethodPost, "/api/problems", `{"title":"B题","description":"x","tags":["math","graph"]}`)
	rec := doJSON(h, http.MethodGet, "/api/problems/tags", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("tags: code=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data []string `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Data) != 3 {
		t.Fatalf("want 3 tags, got %v", resp.Data)
	}
}

func TestHandlerRateLimitMiddleware(t *testing.T) {
	cfg := &config.Config{MaxPageSize: 100, RateLimitPerIP: 2}
	log := logger.NewLevel(logger.LevelError)
	st := store.NewMemoryStore()
	svc := service.New(st, log, cfg)
	srv := NewServer(svc, log, cfg)
	h := srv.Routes()

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/leaderboard", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: want 200, got %d", i, rec.Code)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/api/leaderboard", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("want 429, got %d", rec.Code)
	}
}
