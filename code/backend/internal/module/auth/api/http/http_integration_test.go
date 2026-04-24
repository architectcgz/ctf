package http_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	authhttp "ctf-platform/internal/module/auth/api/http"
	authcmd "ctf-platform/internal/module/auth/application/commands"
	authqry "ctf-platform/internal/module/auth/application/queries"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	identitycmd "ctf-platform/internal/module/identity/application/commands"
	identityqry "ctf-platform/internal/module/identity/application/queries"
	identityinfra "ctf-platform/internal/module/identity/infrastructure"
	opshttp "ctf-platform/internal/module/ops/api/http"
	opscmd "ctf-platform/internal/module/ops/application/commands"
	opsqry "ctf-platform/internal/module/ops/application/queries"
	opsinfra "ctf-platform/internal/module/ops/infrastructure"
	"ctf-platform/internal/validation"
	"ctf-platform/pkg/errcode"
	"ctf-platform/pkg/response"
)

type testEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type testLoginResponse struct {
	User struct {
		ID        int64   `json:"id"`
		Username  string  `json:"username"`
		Role      string  `json:"role"`
		ClassName *string `json:"class_name"`
	} `json:"user"`
}

type testCASStatusResponse struct {
	Provider      string `json:"provider"`
	Enabled       bool   `json:"enabled"`
	Configured    bool   `json:"configured"`
	AutoProvision bool   `json:"auto_provision"`
	LoginPath     string `json:"login_path"`
	CallbackPath  string `json:"callback_path"`
}

type testCASLoginResponse struct {
	Provider    string `json:"provider"`
	RedirectURL string `json:"redirect_url"`
	CallbackURL string `json:"callback_url"`
}

type testProfileResponse struct {
	ID        int64   `json:"id"`
	Username  string  `json:"username"`
	Role      string  `json:"role"`
	ClassName *string `json:"class_name"`
}

type testPageResponse[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

type testAuditLogItem struct {
	ID            int64                  `json:"id"`
	Action        string                 `json:"action"`
	ResourceType  string                 `json:"resource_type"`
	ActorUserID   *int64                 `json:"actor_user_id"`
	ActorUsername string                 `json:"actor_username"`
	Detail        map[string]interface{} `json:"detail"`
}

type integrationTestEnv struct {
	router *gin.Engine
	db     *gorm.DB
}

type memoryTokenService struct {
	config   config.AuthConfig
	wsConfig config.WebSocketConfig

	mu       sync.Mutex
	sessions map[string]authcontracts.Session
	tickets  map[string]authctx.CurrentUser
}

var fallbackRequestIDCounter atomic.Uint64

func TestHTTP_RegisterLoginAndProfileFlow(t *testing.T) {
	env := newIntegrationTestEnv(t)

	registerResp := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/auth/register",
		map[string]any{
			"username":   "student_one",
			"password":   "Password123",
			"email":      "student_one@example.com",
			"class_name": "CTF-1",
		},
		nil,
		nil,
	)
	if registerResp.Code != http.StatusOK {
		t.Fatalf("unexpected register status: %d body=%s", registerResp.Code, registerResp.Body.String())
	}
	registerBody := decodeEnvelope(t, registerResp)
	if registerBody.Code != 0 {
		t.Fatalf("unexpected register code: %d body=%s", registerBody.Code, registerResp.Body.String())
	}
	registerData := decodeJSON[testLoginResponse](t, registerBody.Data)
	if registerData.User.Role != model.RoleStudent {
		t.Fatalf("expected student role, got %q", registerData.User.Role)
	}
	if cookieValue(registerResp.Result().Cookies(), "ctf_session") == "" {
		t.Fatalf("expected session cookie to be set")
	}

	loginResp := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/auth/login",
		map[string]any{
			"username": "student_one",
			"password": "Password123",
		},
		nil,
		nil,
	)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("unexpected login status: %d body=%s", loginResp.Code, loginResp.Body.String())
	}
	loginBody := decodeEnvelope(t, loginResp)
	loginData := decodeJSON[testLoginResponse](t, loginBody.Data)
	if cookieValue(loginResp.Result().Cookies(), "ctf_session") == "" {
		t.Fatalf("expected session cookie in login response")
	}

	profileResp := performJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/auth/profile",
		nil,
		nil,
		[]*http.Cookie{cloneCookie(loginResp.Result().Cookies(), "ctf_session")},
	)
	if profileResp.Code != http.StatusOK {
		t.Fatalf("unexpected profile status: %d body=%s", profileResp.Code, profileResp.Body.String())
	}
	profileBody := decodeEnvelope(t, profileResp)
	profileData := decodeJSON[testProfileResponse](t, profileBody.Data)
	if profileData.ID != loginData.User.ID {
		t.Fatalf("expected profile id %d, got %d", loginData.User.ID, profileData.ID)
	}
	if profileData.Username != "student_one" {
		t.Fatalf("expected profile username student_one, got %q", profileData.Username)
	}
	if profileData.Role != model.RoleStudent {
		t.Fatalf("expected profile role student, got %q", profileData.Role)
	}
}

func TestHTTP_LoginResponseDoesNotExposeAccessToken(t *testing.T) {
	env := newIntegrationTestEnv(t)

	createUser(t, env.db, "session_login_user", "Password123", model.RoleStudent, "")

	loginResp := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/auth/login",
		map[string]any{
			"username": "session_login_user",
			"password": "Password123",
		},
		nil,
		nil,
	)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("unexpected login status: %d body=%s", loginResp.Code, loginResp.Body.String())
	}

	loginBody := decodeEnvelope(t, loginResp)
	var payload map[string]any
	if err := json.Unmarshal(loginBody.Data, &payload); err != nil {
		t.Fatalf("decode login payload: %v body=%s", err, string(loginBody.Data))
	}
	if _, exists := payload["access_token"]; exists {
		t.Fatalf("expected login payload to omit access_token, got %s", string(loginBody.Data))
	}
}

func TestHTTP_ChangePasswordFlow(t *testing.T) {
	env := newIntegrationTestEnv(t)

	registerResp := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/auth/register",
		map[string]any{
			"username": "change_pwd_user",
			"password": "Password123",
		},
		nil,
		nil,
	)
	if registerResp.Code != http.StatusOK {
		t.Fatalf("unexpected register status: %d body=%s", registerResp.Code, registerResp.Body.String())
	}
	sessionCookie := cloneCookie(registerResp.Result().Cookies(), "ctf_session")
	if sessionCookie == nil {
		t.Fatalf("expected session cookie from register response")
	}

	changeResp := performJSONRequest(
		t,
		env.router,
		http.MethodPut,
		"/api/v1/auth/password",
		map[string]any{
			"old_password": "Password123",
			"new_password": "Password456",
		},
		nil,
		[]*http.Cookie{sessionCookie},
	)
	if changeResp.Code != http.StatusOK {
		t.Fatalf("unexpected change password status: %d body=%s", changeResp.Code, changeResp.Body.String())
	}
	changeBody := decodeEnvelope(t, changeResp)
	if changeBody.Code != 0 {
		t.Fatalf("unexpected change password code: %d body=%s", changeBody.Code, changeResp.Body.String())
	}

	oldLoginResp := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/auth/login",
		map[string]any{
			"username": "change_pwd_user",
			"password": "Password123",
		},
		nil,
		nil,
	)
	if oldLoginResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected old password login to fail, got %d body=%s", oldLoginResp.Code, oldLoginResp.Body.String())
	}
	oldLoginBody := decodeEnvelope(t, oldLoginResp)
	if oldLoginBody.Code != errcode.ErrInvalidCredentials.Code {
		t.Fatalf("expected invalid credentials code %d, got %d", errcode.ErrInvalidCredentials.Code, oldLoginBody.Code)
	}

	newLoginResp := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/auth/login",
		map[string]any{
			"username": "change_pwd_user",
			"password": "Password456",
		},
		nil,
		nil,
	)
	if newLoginResp.Code != http.StatusOK {
		t.Fatalf("expected new password login to succeed, got %d body=%s", newLoginResp.Code, newLoginResp.Body.String())
	}
	if cookieValue(newLoginResp.Result().Cookies(), "ctf_session") == "" {
		t.Fatalf("expected new session cookie after password change")
	}
}

func TestHTTP_LogoutRevokesSessionAndAdminCanQueryAuditLogs(t *testing.T) {
	env := newIntegrationTestEnv(t)

	createUser(t, env.db, "audit_admin", "Password123", model.RoleAdmin, "")

	registerResp := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/auth/register",
		map[string]any{
			"username": "logout_user",
			"password": "Password123",
		},
		nil,
		nil,
	)
	registerBody := decodeEnvelope(t, registerResp)
	registerData := decodeJSON[testLoginResponse](t, registerBody.Data)
	sessionCookie := cloneCookie(registerResp.Result().Cookies(), "ctf_session")
	if sessionCookie == nil {
		t.Fatalf("expected session cookie for logout flow")
	}

	logoutResp := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/auth/logout",
		nil,
		nil,
		[]*http.Cookie{sessionCookie},
	)
	if logoutResp.Code != http.StatusOK {
		t.Fatalf("unexpected logout status: %d body=%s", logoutResp.Code, logoutResp.Body.String())
	}

	revokedResp := performJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/auth/profile",
		nil,
		nil,
		[]*http.Cookie{sessionCookie},
	)
	if revokedResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected revoked session to return 401, got %d body=%s", revokedResp.Code, revokedResp.Body.String())
	}
	revokedBody := decodeEnvelope(t, revokedResp)
	if revokedBody.Code != errcode.ErrUnauthorized.Code {
		t.Fatalf("expected unauthorized code %d, got %d", errcode.ErrUnauthorized.Code, revokedBody.Code)
	}

	adminLoginResp := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/auth/login",
		map[string]any{
			"username": "audit_admin",
			"password": "Password123",
		},
		nil,
		nil,
	)
	adminLoginBody := decodeEnvelope(t, adminLoginResp)
	_ = decodeJSON[testLoginResponse](t, adminLoginBody.Data)
	adminSessionCookie := cloneCookie(adminLoginResp.Result().Cookies(), "ctf_session")
	if adminSessionCookie == nil {
		t.Fatalf("expected admin session cookie")
	}

	auditResp := performJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/admin/audit-logs?action=logout&user_id="+strconv.FormatInt(registerData.User.ID, 10),
		nil,
		nil,
		[]*http.Cookie{adminSessionCookie},
	)
	if auditResp.Code != http.StatusOK {
		t.Fatalf("unexpected audit status: %d body=%s", auditResp.Code, auditResp.Body.String())
	}
	auditBody := decodeEnvelope(t, auditResp)
	auditData := decodeJSON[testPageResponse[testAuditLogItem]](t, auditBody.Data)
	if len(auditData.List) == 0 {
		t.Fatalf("expected at least one logout audit log")
	}
	found := false
	for _, item := range auditData.List {
		if item.Action == model.AuditActionLogout && item.ActorUserID != nil && *item.ActorUserID == registerData.User.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected logout audit log for user %d, got %+v", registerData.User.ID, auditData.List)
	}
}

func TestHTTP_FailedLoginIsRecordedInAuditLog(t *testing.T) {
	env := newIntegrationTestEnv(t)

	createUser(t, env.db, "audit_admin", "Password123", model.RoleAdmin, "")

	failedResp := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/auth/login",
		map[string]any{
			"username": "ghost_user",
			"password": "Password123",
		},
		nil,
		nil,
	)
	if failedResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected failed login to return 401, got %d body=%s", failedResp.Code, failedResp.Body.String())
	}

	adminLoginResp := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/auth/login",
		map[string]any{
			"username": "audit_admin",
			"password": "Password123",
		},
		nil,
		nil,
	)
	adminLoginBody := decodeEnvelope(t, adminLoginResp)
	_ = decodeJSON[testLoginResponse](t, adminLoginBody.Data)
	adminSessionCookie := cloneCookie(adminLoginResp.Result().Cookies(), "ctf_session")
	if adminSessionCookie == nil {
		t.Fatalf("expected admin session cookie")
	}

	auditResp := performJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/admin/audit-logs?action=login&resource_type=auth",
		nil,
		nil,
		[]*http.Cookie{adminSessionCookie},
	)
	if auditResp.Code != http.StatusOK {
		t.Fatalf("unexpected audit status: %d body=%s", auditResp.Code, auditResp.Body.String())
	}
	auditBody := decodeEnvelope(t, auditResp)
	auditData := decodeJSON[testPageResponse[testAuditLogItem]](t, auditBody.Data)

	found := false
	for _, item := range auditData.List {
		if item.Action != model.AuditActionLogin || item.ResourceType != "auth" {
			continue
		}
		if item.Detail["username"] == "ghost_user" && item.Detail["result"] == "failed" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected failed login audit log for ghost_user, got %+v", auditData.List)
	}
}

func TestHTTP_LoginIsTemporarilyLockedAfterRepeatedFailures(t *testing.T) {
	env := newIntegrationTestEnv(t)

	createUser(t, env.db, "locked_user", "Password123", model.RoleStudent, "CTF-1")

	for attempt := 1; attempt <= 3; attempt++ {
		failedResp := performJSONRequest(
			t,
			env.router,
			http.MethodPost,
			"/api/v1/auth/login",
			map[string]any{
				"username": "locked_user",
				"password": "wrong-password",
			},
			nil,
			nil,
		)
		expectedStatus := http.StatusUnauthorized
		expectedCode := errcode.ErrInvalidCredentials.Code
		if attempt == 3 {
			expectedStatus = http.StatusTooManyRequests
			expectedCode = errcode.ErrLoginTooFrequent.Code
		}
		if failedResp.Code != expectedStatus {
			t.Fatalf("attempt %d expected status %d, got %d body=%s", attempt, expectedStatus, failedResp.Code, failedResp.Body.String())
		}
		failedBody := decodeEnvelope(t, failedResp)
		if failedBody.Code != expectedCode {
			t.Fatalf("attempt %d expected code %d, got %d", attempt, expectedCode, failedBody.Code)
		}
	}

	lockedResp := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/auth/login",
		map[string]any{
			"username": "locked_user",
			"password": "Password123",
		},
		nil,
		nil,
	)
	if lockedResp.Code != http.StatusForbidden {
		t.Fatalf("expected locked login status 403, got %d body=%s", lockedResp.Code, lockedResp.Body.String())
	}
	lockedBody := decodeEnvelope(t, lockedResp)
	if lockedBody.Code != errcode.ErrAccountLocked.Code {
		t.Fatalf("expected account locked code %d, got %d", errcode.ErrAccountLocked.Code, lockedBody.Code)
	}
}

func TestHTTP_CASStatusDisabledByDefault(t *testing.T) {
	env := newIntegrationTestEnv(t)

	resp := performJSONRequest(t, env.router, http.MethodGet, "/api/v1/auth/cas/status", nil, nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected cas status code: %d body=%s", resp.Code, resp.Body.String())
	}
	body := decodeEnvelope(t, resp)
	data := decodeJSON[testCASStatusResponse](t, body.Data)
	if data.Provider != "cas" {
		t.Fatalf("expected provider cas, got %+v", data)
	}
	if data.Enabled || data.Configured {
		t.Fatalf("expected disabled and unconfigured cas, got %+v", data)
	}
}

func TestHTTP_CASLoginReturnsConfiguredRedirectURL(t *testing.T) {
	env := newIntegrationTestEnvWithAuthConfig(t, func(cfg *config.AuthConfig) {
		cfg.CAS.Enabled = true
		cfg.CAS.BaseURL = "https://cas.example.edu/cas"
		cfg.CAS.ServiceURL = "https://ctf.example.edu/api/v1/auth/cas/callback"
		cfg.CAS.AutoProvision = true
	})

	resp := performJSONRequest(t, env.router, http.MethodGet, "/api/v1/auth/cas/login", nil, nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected cas login code: %d body=%s", resp.Code, resp.Body.String())
	}
	body := decodeEnvelope(t, resp)
	data := decodeJSON[testCASLoginResponse](t, body.Data)
	expectedRedirect := "https://cas.example.edu/cas/login?service=https%3A%2F%2Fctf.example.edu%2Fapi%2Fv1%2Fauth%2Fcas%2Fcallback"
	if data.RedirectURL != expectedRedirect {
		t.Fatalf("unexpected redirect url: %s", data.RedirectURL)
	}
	if data.CallbackURL != "https://ctf.example.edu/api/v1/auth/cas/callback" {
		t.Fatalf("unexpected callback url: %s", data.CallbackURL)
	}
}

func TestHTTP_CASCallbackAutoProvisionsUserAndIssuesCookie(t *testing.T) {
	casServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<serviceResponse>
  <authenticationSuccess>
    <user>cas_http_user</user>
    <attributes>
      <displayName>HTTP CAS User</displayName>
      <mail>cas_http_user@example.edu</mail>
      <className>CTF-HTTP</className>
      <studentNo>20269999</studentNo>
    </attributes>
  </authenticationSuccess>
</serviceResponse>`)
	}))
	defer casServer.Close()

	env := newIntegrationTestEnvWithAuthConfig(t, func(cfg *config.AuthConfig) {
		cfg.CAS.Enabled = true
		cfg.CAS.BaseURL = casServer.URL
		cfg.CAS.ServiceURL = "https://ctf.example.edu/api/v1/auth/cas/callback"
		cfg.CAS.AutoProvision = true
	})

	resp := performJSONRequest(t, env.router, http.MethodGet, "/api/v1/auth/cas/callback?ticket=ST-123", nil, nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected cas callback status: %d body=%s", resp.Code, resp.Body.String())
	}
	body := decodeEnvelope(t, resp)
	if body.Code != 0 {
		t.Fatalf("unexpected cas callback code %d body=%s", body.Code, resp.Body.String())
	}
	data := decodeJSON[testLoginResponse](t, body.Data)
	if data.User.Username != "cas_http_user" || data.User.Role != model.RoleStudent {
		t.Fatalf("unexpected cas callback user: %+v", data.User)
	}
	if cookieValue(resp.Result().Cookies(), "ctf_session") == "" {
		t.Fatalf("expected session cookie to be set")
	}

	var user model.User
	if err := env.db.Where("username = ?", "cas_http_user").First(&user).Error; err != nil {
		t.Fatalf("query cas user: %v", err)
	}
	if user.Email != "cas_http_user@example.edu" || user.ClassName != "CTF-HTTP" || user.StudentNo != "20269999" {
		t.Fatalf("unexpected provisioned cas user: %+v", user)
	}
}

func TestHTTP_CASCallbackRejectsUserWhenAutoProvisionDisabled(t *testing.T) {
	casServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<serviceResponse>
  <authenticationSuccess>
    <user>cas_http_user</user>
  </authenticationSuccess>
</serviceResponse>`)
	}))
	defer casServer.Close()

	env := newIntegrationTestEnvWithAuthConfig(t, func(cfg *config.AuthConfig) {
		cfg.CAS.Enabled = true
		cfg.CAS.BaseURL = casServer.URL
		cfg.CAS.ServiceURL = "https://ctf.example.edu/api/v1/auth/cas/callback"
		cfg.CAS.AutoProvision = false
	})

	resp := performJSONRequest(t, env.router, http.MethodGet, "/api/v1/auth/cas/callback?ticket=ST-123", nil, nil, nil)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("unexpected cas callback status: %d body=%s", resp.Code, resp.Body.String())
	}
	body := decodeEnvelope(t, resp)
	if body.Code != errcode.ErrCASUserNotProvisioned.Code {
		t.Fatalf("expected cas user not provisioned code %d, got %d", errcode.ErrCASUserNotProvisioned.Code, body.Code)
	}
}

func newIntegrationTestEnv(t *testing.T) *integrationTestEnv {
	return newIntegrationTestEnvWithAuthConfig(t, nil)
}

func newIntegrationTestEnvWithAuthConfig(t *testing.T, mutate func(*config.AuthConfig)) *integrationTestEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)

	if err := validation.Register(); err != nil {
		t.Fatalf("register validator: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), "auth-http-integration.sqlite")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Role{}, &model.User{}, &model.UserRole{}, &model.AuditLog{}); err != nil {
		t.Fatalf("auto migrate test schema: %v", err)
	}
	seedRoles(t, db)

	authCfg := newTestAuthConfig(t)
	if mutate != nil {
		mutate(&authCfg)
	}
	tokenService := newMemoryTokenService(authCfg, config.WebSocketConfig{
		TicketTTL:       30 * time.Second,
		TicketKeyPrefix: "test:ws:ticket",
	})
	authRepo := identityinfra.NewRepository(db)
	authService := authcmd.NewService(authRepo, tokenService, config.RateLimitPolicyConfig{
		Enabled:      true,
		Limit:        3,
		Window:       time.Minute,
		LockDuration: 15 * time.Minute,
	}, zap.NewNop())
	profileCommandService := identitycmd.NewProfileService(authRepo, zap.NewNop())
	profileQueryService := identityqry.NewProfileService(authRepo)
	casCommandService := authcmd.NewCASService(authCfg.CAS, authRepo, tokenService, zap.NewNop(), nil)
	casQueryService := authqry.NewCASService(authCfg.CAS)
	auditRepo := opsinfra.NewAuditRepository(db)
	auditCommandService := opscmd.NewAuditService(auditRepo, zap.NewNop())
	auditQueryService := opsqry.NewAuditService(auditRepo, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, zap.NewNop())
	authHandler := authhttp.NewHandler(authService, profileCommandService, profileQueryService, tokenService, casCommandService, casQueryService, authhttp.CookieConfig{
		Name:     authCfg.SessionCookieName,
		Path:     authCfg.SessionCookiePath,
		HTTPOnly: authCfg.SessionCookieHTTPOnly,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   authCfg.SessionTTL,
	}, zap.NewNop(), auditCommandService)
	auditHandler := opshttp.NewAuditHandler(auditQueryService)

	router := gin.New()
	router.Use(testRequestID())

	apiV1 := router.Group("/api/v1")
	authGroup := apiV1.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.GET("/cas/status", authHandler.CASStatus)
	authGroup.GET("/cas/login", authHandler.CASLogin)
	authGroup.GET("/cas/callback", authHandler.CASCallback)

	protected := apiV1.Group("")
	protected.Use(testAuthMiddleware(tokenService, authCfg.SessionCookieName))
	protected.POST("/auth/logout", authHandler.Logout)
	protected.GET("/auth/profile", authHandler.Profile)
	protected.PUT("/auth/password", authHandler.ChangePassword)

	adminOnly := protected.Group("/admin")
	adminOnly.Use(testRequireRole(model.RoleAdmin))
	adminOnly.GET("/audit-logs", auditHandler.ListAuditLogs)

	t.Cleanup(func() {
		if sqlDB, sqlErr := db.DB(); sqlErr == nil {
			_ = sqlDB.Close()
		}
	})

	return &integrationTestEnv{
		router: router,
		db:     db,
	}
}

func newTestAuthConfig(t *testing.T) config.AuthConfig {
	t.Helper()

	return config.AuthConfig{
		SessionTTL:            24 * time.Hour,
		SessionCookieName:     "ctf_session",
		SessionCookiePath:     "/",
		SessionCookieHTTPOnly: true,
		SessionCookieSameSite: "lax",
		SessionKeyPrefix:      "test:session",
	}
}

func newMemoryTokenService(cfg config.AuthConfig, wsCfg config.WebSocketConfig) authcontracts.TokenService {
	return &memoryTokenService{
		config:   cfg,
		wsConfig: wsCfg,
		sessions: make(map[string]authcontracts.Session),
		tickets:  make(map[string]authctx.CurrentUser),
	}
}

func (s *memoryTokenService) CreateSession(_ context.Context, userID int64, username, role string) (*authcontracts.Session, error) {
	sessionID, err := generateRandomHex(16)
	if err != nil {
		return nil, err
	}
	session := authcontracts.Session{
		ID:        sessionID,
		UserID:    userID,
		Username:  username,
		Role:      role,
		ExpiresAt: time.Now().Add(s.config.SessionTTL),
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sessionID] = session
	return &session, nil
}

func (s *memoryTokenService) GetSession(_ context.Context, sessionID string) (*authcontracts.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, errcode.ErrUnauthorized
	}
	return &session, nil
}

func (s *memoryTokenService) DeleteSession(_ context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
	return nil
}

func (s *memoryTokenService) IssueWSTicket(ctx context.Context, user authctx.CurrentUser) (*authcontracts.WSTicket, error) {
	ticket := fmt.Sprintf("ticket_%s", randomHex(12))
	expiresAt := time.Now().Add(s.wsConfig.TicketTTL)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.tickets[ticket] = authctx.CurrentUser{
		UserID:   user.UserID,
		Username: user.Username,
		Role:     user.Role,
	}

	return &authcontracts.WSTicket{
		Ticket:    ticket,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *memoryTokenService) ConsumeWSTicket(ctx context.Context, ticket string) (*authctx.CurrentUser, error) {
	if ticket == "" {
		return nil, errcode.ErrWSTicketInvalid
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.tickets[ticket]
	if !ok {
		return nil, errcode.ErrWSTicketInvalid
	}
	delete(s.tickets, ticket)
	return &user, nil
}

func seedRoles(t *testing.T, db *gorm.DB) {
	t.Helper()

	roles := []model.Role{
		{ID: 1, Code: model.RoleStudent, Name: "Student"},
		{ID: 2, Code: model.RoleTeacher, Name: "Teacher"},
		{ID: 3, Code: model.RoleAdmin, Name: "Admin"},
	}
	for _, role := range roles {
		if err := db.Create(&role).Error; err != nil {
			t.Fatalf("seed role %s: %v", role.Code, err)
		}
	}
}

func createUser(t *testing.T, db *gorm.DB, username, password, role, className string) *model.User {
	t.Helper()

	user := &model.User{
		Username:  username,
		Email:     fmt.Sprintf("%s@example.com", username),
		Role:      role,
		ClassName: className,
		Status:    model.UserStatusActive,
	}
	if err := user.SetPassword(password); err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func performJSONRequest(
	t *testing.T,
	router http.Handler,
	method string,
	target string,
	payload any,
	headers map[string]string,
	cookies []*http.Cookie,
) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encode request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, target, &body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func decodeEnvelope(t *testing.T, recorder *httptest.ResponseRecorder) testEnvelope {
	t.Helper()

	var envelope testEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, recorder.Body.String())
	}
	return envelope
}

func decodeJSON[T any](t *testing.T, data []byte) T {
	t.Helper()

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("decode payload: %v payload=%s", err, string(data))
	}
	return value
}

func cookieValue(cookies []*http.Cookie, name string) string {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

func cloneCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			cloned := *cookie
			return &cloned
		}
	}
	return nil
}

func generateRandomHex(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

func testRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = newTestRequestID()
		}

		c.Set("request_id", requestID)
		c.Request.Header.Set("X-Request-ID", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

func testAuthMiddleware(tokenService authcontracts.TokenService, cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(cookieName)
		if err != nil || sessionID == "" {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		session, err := tokenService.GetSession(c.Request.Context(), sessionID)
		if err != nil {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		authctx.SetCurrentUser(c, authctx.CurrentUser{
			UserID:    session.UserID,
			Username:  session.Username,
			Role:      session.Role,
			SessionID: session.ID,
			ExpiresAt: session.ExpiresAt,
		})
		c.Next()
	}
}

func testRequireRole(minRole string) gin.HandlerFunc {
	roleLevels := map[string]int{
		model.RoleStudent: 1,
		model.RoleTeacher: 2,
		model.RoleAdmin:   3,
	}

	return func(c *gin.Context) {
		currentUser := authctx.MustCurrentUser(c)
		if roleLevels[currentUser.Role] < roleLevels[minRole] {
			response.Error(c, errcode.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

func newTestRequestID() string {
	buffer := make([]byte, 8)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("req_fallback_%d_%d", time.Now().UnixNano(), fallbackRequestIDCounter.Add(1))
	}
	return "req_" + hex.EncodeToString(buffer)
}

func randomHex(size int) string {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("fallback_%d", fallbackRequestIDCounter.Add(1))
	}
	return hex.EncodeToString(buffer)
}
