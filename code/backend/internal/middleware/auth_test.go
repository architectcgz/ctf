package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/model"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

type stubTokenService struct {
	session *authcontracts.Session
}

func (s *stubTokenService) CreateSession(context.Context, int64, string, string) (*authcontracts.Session, error) {
	panic("unexpected call to CreateSession")
}

func (s *stubTokenService) GetSession(context.Context, string) (*authcontracts.Session, error) {
	return s.session, nil
}

func (s *stubTokenService) DeleteSession(context.Context, string) error {
	panic("unexpected call to DeleteSession")
}

func (s *stubTokenService) IssueWSTicket(context.Context, authctx.CurrentUser) (*authcontracts.WSTicket, error) {
	panic("unexpected call to IssueWSTicket")
}

func (s *stubTokenService) ConsumeWSTicket(context.Context, string) (*authctx.CurrentUser, error) {
	panic("unexpected call to ConsumeWSTicket")
}

type stubUserRepository struct {
	user *model.User
}

func (s *stubUserRepository) List(context.Context, identitycontracts.UserListFilter) ([]*model.User, int64, error) {
	panic("unexpected call to List")
}

func (s *stubUserRepository) FindByID(context.Context, int64) (*model.User, error) {
	return s.user, nil
}

func (s *stubUserRepository) FindByUsername(context.Context, string) (*model.User, error) {
	panic("unexpected call to FindByUsername")
}

func (s *stubUserRepository) Create(context.Context, *model.User) error {
	panic("unexpected call to Create")
}

func (s *stubUserRepository) Update(context.Context, *model.User) error {
	panic("unexpected call to Update")
}

func (s *stubUserRepository) Delete(context.Context, int64) error {
	panic("unexpected call to Delete")
}

func (s *stubUserRepository) UpdatePassword(context.Context, int64, string) error {
	panic("unexpected call to UpdatePassword")
}

func (s *stubUserRepository) UpdateLoginState(context.Context, int64, int, *time.Time, *time.Time, string) error {
	panic("unexpected call to UpdateLoginState")
}

func (s *stubUserRepository) UpdateProfile(context.Context, *model.User) error {
	panic("unexpected call to UpdateProfile")
}

func TestAuthUsesCurrentPersistedRoleForRBAC(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenService := &stubTokenService{
		session: &authcontracts.Session{
			ID:        "sess-1",
			UserID:    42,
			Username:  "teacher-token",
			Role:      model.RoleTeacher,
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}
	users := &stubUserRepository{
		user: &model.User{
			ID:       42,
			Username: "admin-db",
			Role:     model.RoleAdmin,
		},
	}

	router := gin.New()
	router.Use(Auth(tokenService, "ctf_session", users))
	adminOnly := router.Group("/admin")
	adminOnly.Use(RequireRole(model.RoleAdmin))
	adminOnly.PUT("/contests/1", func(c *gin.Context) {
		currentUser := MustCurrentUser(c)
		c.JSON(http.StatusOK, gin.H{
			"role":     currentUser.Role,
			"username": currentUser.Username,
		})
	})

	req := httptest.NewRequest(http.MethodPut, "/admin/contests/1", nil)
	req.AddCookie(&http.Cookie{Name: "ctf_session", Value: "sess-1"})
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", resp.Code, resp.Body.String())
	}

	var payload map[string]string
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload["role"] != model.RoleAdmin {
		t.Fatalf("expected current role %q, got %q", model.RoleAdmin, payload["role"])
	}
	if payload["username"] != "admin-db" {
		t.Fatalf("expected current username %q, got %q", "admin-db", payload["username"])
	}
}

func TestAuthRejectsPrivilegesRemovedAfterTokenIssued(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenService := &stubTokenService{
		session: &authcontracts.Session{
			ID:        "sess-2",
			UserID:    99,
			Username:  "admin-token",
			Role:      model.RoleAdmin,
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}
	users := &stubUserRepository{
		user: &model.User{
			ID:       99,
			Username: "teacher-db",
			Role:     model.RoleTeacher,
		},
	}

	router := gin.New()
	router.Use(Auth(tokenService, "ctf_session", users))
	adminOnly := router.Group("/admin")
	adminOnly.Use(RequireRole(model.RoleAdmin))
	adminOnly.PUT("/contests/1", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPut, "/admin/contests/1", nil)
	req.AddCookie(&http.Cookie{Name: "ctf_session", Value: "sess-2"})
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d with body %s", resp.Code, resp.Body.String())
	}
}
