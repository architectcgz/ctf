package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/model"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	jwtpkg "ctf-platform/pkg/jwt"
)

type stubTokenService struct {
	claims  *jwtpkg.Claims
	revoked bool
}

func (s *stubTokenService) IssueTokens(int64, string, string) (*authcontracts.TokenPair, error) {
	panic("unexpected call to IssueTokens")
}

func (s *stubTokenService) IssueTokensWithContext(context.Context, int64, string, string) (*authcontracts.TokenPair, error) {
	panic("unexpected call to IssueTokensWithContext")
}

func (s *stubTokenService) RefreshAccessToken(context.Context, string) (*authcontracts.RefreshAccessPayload, error) {
	panic("unexpected call to RefreshAccessToken")
}

func (s *stubTokenService) RevokeToken(context.Context, string, time.Duration) error {
	panic("unexpected call to RevokeToken")
}

func (s *stubTokenService) ClearRefreshSession(context.Context, int64, string) error {
	panic("unexpected call to ClearRefreshSession")
}

func (s *stubTokenService) IsRevoked(context.Context, string) (bool, error) {
	return s.revoked, nil
}

func (s *stubTokenService) ParseToken(string) (*jwtpkg.Claims, error) {
	return s.claims, nil
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
		claims: &jwtpkg.Claims{
			UserID:    42,
			Username:  "teacher-token",
			Role:      model.RoleTeacher,
			TokenType: jwtpkg.TokenTypeAccess,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
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
	router.Use(Auth(tokenService, users))
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
	req.Header.Set("Authorization", "Bearer test-token")
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
		claims: &jwtpkg.Claims{
			UserID:    99,
			Username:  "admin-token",
			Role:      model.RoleAdmin,
			TokenType: jwtpkg.TokenTypeAccess,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
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
	router.Use(Auth(tokenService, users))
	adminOnly := router.Group("/admin")
	adminOnly.Use(RequireRole(model.RoleAdmin))
	adminOnly.PUT("/contests/1", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPut, "/admin/contests/1", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d with body %s", resp.Code, resp.Body.String())
	}
}
