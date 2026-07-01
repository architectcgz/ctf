package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/authctx"
	response "ctf-platform/internal/httpresponse"
	"ctf-platform/internal/middleware"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	identityhttp "ctf-platform/internal/module/identity/api/http"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

type sessionTestTokenService struct {
	mu       sync.Mutex
	sessions map[string]authcontracts.Session
	version  map[int64]int64
}

func newSessionTestTokenService() *sessionTestTokenService {
	return &sessionTestTokenService{
		sessions: make(map[string]authcontracts.Session),
		version:  make(map[int64]int64),
	}
}

func (s *sessionTestTokenService) CreateSession(_ context.Context, userID int64, username, role string) (*authcontracts.Session, error) {
	session := authcontracts.Session{
		ID:        "session-" + username,
		UserID:    userID,
		Username:  username,
		Role:      role,
		ExpiresAt: time.Now().Add(24 * time.Hour).UTC(),
	}
	s.mu.Lock()
	s.sessions[session.ID] = session
	s.mu.Unlock()
	return &session, nil
}

func (s *sessionTestTokenService) GetSession(_ context.Context, sessionID string) (*authcontracts.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, authcontracts.ErrAccessTokenExpired
	}
	if !session.ExpiresAt.After(time.Now().UTC()) {
		return nil, authcontracts.ErrAccessTokenExpired
	}
	version := s.version[session.UserID]
	if version > 0 {
		return nil, authcontracts.ErrAccessTokenExpired
	}
	return &session, nil
}

func (s *sessionTestTokenService) DeleteSession(_ context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
	return nil
}

func (s *sessionTestTokenService) RevokeAllUserSessions(_ context.Context, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.version[userID]++
	for id, session := range s.sessions {
		if session.UserID == userID {
			delete(s.sessions, id)
		}
	}
	return nil
}

func (s *sessionTestTokenService) CurrentSessionVersion(_ context.Context, userID int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.version[userID], nil
}

func (s *sessionTestTokenService) ListUserSessions(_ context.Context, userID int64) ([]authcontracts.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]authcontracts.Session, 0)
	for _, session := range s.sessions {
		if session.UserID == userID {
			result = append(result, session)
		}
	}
	return result, nil
}

func (s *sessionTestTokenService) IssueWSTicket(_ context.Context, _ authctx.CurrentUser) (*authcontracts.WSTicket, error) {
	return nil, nil
}

func (s *sessionTestTokenService) ConsumeWSTicket(_ context.Context, _ string) (*authctx.CurrentUser, error) {
	return nil, nil
}

func (s *sessionTestTokenService) IssueMCPToken(_ context.Context, _ authctx.CurrentUser) (*authcontracts.MCPToken, error) {
	return nil, nil
}

func (s *sessionTestTokenService) ResolveMCPToken(_ context.Context, _ string) (*authctx.CurrentUser, error) {
	return nil, nil
}

func TestAdminSessionRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokenService := newSessionTestTokenService()

	// 种子：admin 用户
	tokenService.sessions["session-admin"] = authcontracts.Session{
		ID:        "session-admin",
		UserID:    1,
		Username:  "admin",
		Role:      identitycontracts.RoleAdmin,
		ExpiresAt: time.Now().Add(24 * time.Hour).UTC(),
	}

	engine := gin.New()
	adminOnly := engine.Group("/admin")
	adminOnly.Use(func(c *gin.Context) {
		c.Set(authctx.CurrentUserKey, authctx.CurrentUser{
			UserID:    1,
			Username:  "admin",
			Role:      identitycontracts.RoleAdmin,
			SessionID: "session-admin",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		})
		c.Next()
	})

	// 注册会话路由（与 router_routes.go 一致）
	deps := adminRouteDeps{
		tokenService: tokenService,
	}

	adminOnly.GET("/users/:id/sessions",
		middleware.ParseInt64Param("id"),
		func(c *gin.Context) {
			userID := c.GetInt64("id")
			sessions, err := deps.tokenService.ListUserSessions(c.Request.Context(), userID)
			if err != nil {
				response.FromError(c, err)
				return
			}
			resps := make([]identityhttp.UserSessionResp, 0, len(sessions))
			for _, s := range sessions {
				resps = append(resps, identityhttp.UserSessionResp{
					ID:        s.ID,
					Username:  s.Username,
					Role:      s.Role,
					ExpiresAt: s.ExpiresAt,
				})
			}
			response.Success(c, gin.H{"sessions": resps})
		},
	)
	adminOnly.DELETE("/users/:id/sessions",
		middleware.ParseInt64Param("id"),
		func(c *gin.Context) {
			userID := c.GetInt64("id")
			if err := deps.tokenService.RevokeAllUserSessions(c.Request.Context(), userID); err != nil {
				response.FromError(c, err)
				return
			}
			response.Success(c, gin.H{"message": "已撤销该用户所有会话"})
		},
	)
	adminOnly.DELETE("/users/:id/sessions/:sid",
		middleware.ParseInt64Param("id"),
		func(c *gin.Context) {
			sessionID := c.Param("sid")
			if sessionID == "" {
				response.InvalidParams(c, "缺少会话ID")
				return
			}
			session, err := deps.tokenService.GetSession(c.Request.Context(), sessionID)
			if err != nil {
				response.FromError(c, err)
				return
			}
			userID := c.GetInt64("id")
			if session.UserID != userID {
				response.Error(c, apperror.ErrForbidden)
				return
			}
			if err := deps.tokenService.DeleteSession(c.Request.Context(), sessionID); err != nil {
				response.FromError(c, err)
				return
			}
			response.Success(c, gin.H{"message": "会话已撤销"})
		},
	)

	t.Run("list sessions returns empty for user with no sessions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/users/999/sessions", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
		}

		var envelope struct {
			Code int `json:"code"`
			Data struct {
				Sessions []any `json:"sessions"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if envelope.Code != 0 {
			t.Fatalf("expected code 0, got %d", envelope.Code)
		}
		if len(envelope.Data.Sessions) != 0 {
			t.Fatalf("expected 0 sessions, got %d", len(envelope.Data.Sessions))
		}
	})

	t.Run("list sessions returns active sessions", func(t *testing.T) {
		tokenService.sessions["session-alice"] = authcontracts.Session{
			ID:        "session-alice",
			UserID:    42,
			Username:  "alice",
			Role:      identitycontracts.RoleStudent,
			ExpiresAt: time.Now().Add(1 * time.Hour).UTC(),
		}

		req := httptest.NewRequest(http.MethodGet, "/admin/users/42/sessions", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
		}

		var envelope struct {
			Code int `json:"code"`
			Data struct {
				Sessions []struct {
					ID       string `json:"id"`
					Username string `json:"username"`
				} `json:"sessions"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if len(envelope.Data.Sessions) != 1 {
			t.Fatalf("expected 1 session, got %d", len(envelope.Data.Sessions))
		}
		if envelope.Data.Sessions[0].Username != "alice" {
			t.Fatalf("expected alice, got %s", envelope.Data.Sessions[0].Username)
		}
	})

	t.Run("revoke single session", func(t *testing.T) {
		tokenService.sessions["session-bob"] = authcontracts.Session{
			ID:        "session-bob",
			UserID:    99,
			Username:  "bob",
			Role:      identitycontracts.RoleStudent,
			ExpiresAt: time.Now().Add(1 * time.Hour).UTC(),
		}

		req := httptest.NewRequest(http.MethodDelete, "/admin/users/99/sessions/session-bob", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
		}

		// 验证会话已被删除
		if _, ok := tokenService.sessions["session-bob"]; ok {
			t.Fatal("expected session to be deleted")
		}
	})

	t.Run("revoke single session returns 403 for wrong user", func(t *testing.T) {
		tokenService.sessions["session-charlie"] = authcontracts.Session{
			ID:        "session-charlie",
			UserID:    77,
			Username:  "charlie",
			Role:      identitycontracts.RoleStudent,
			ExpiresAt: time.Now().Add(1 * time.Hour).UTC(),
		}

		// 尝试用 admin (userID=1) 删除 charlie (userID=77) 的会话，但 URL 中 user ID 是 88（不匹配）
		req := httptest.NewRequest(http.MethodDelete, "/admin/users/88/sessions/session-charlie", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		var envelope struct {
			Code int `json:"code"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if envelope.Code != apperror.ErrForbidden.Code {
			t.Fatalf("expected forbidden code %d, got %d body=%s", apperror.ErrForbidden.Code, envelope.Code, w.Body.String())
		}
	})

	t.Run("revoke all sessions for user", func(t *testing.T) {
		tokenService.sessions["session-d1"] = authcontracts.Session{
			ID: "session-d1", UserID: 55, Username: "dave",
			Role: identitycontracts.RoleStudent, ExpiresAt: time.Now().Add(1 * time.Hour).UTC(),
		}
		tokenService.sessions["session-d2"] = authcontracts.Session{
			ID: "session-d2", UserID: 55, Username: "dave",
			Role: identitycontracts.RoleStudent, ExpiresAt: time.Now().Add(1 * time.Hour).UTC(),
		}

		req := httptest.NewRequest(http.MethodDelete, "/admin/users/55/sessions", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
		}

		// 验证所有会话已被删除
		if _, ok := tokenService.sessions["session-d1"]; ok {
			t.Fatal("expected session-d1 to be deleted")
		}
		if _, ok := tokenService.sessions["session-d2"]; ok {
			t.Fatal("expected session-d2 to be deleted")
		}

		// 版本号应已递增
		if tokenService.version[55] != 1 {
			t.Fatalf("expected version 1, got %d", tokenService.version[55])
		}
	})
}
