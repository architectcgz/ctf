package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/authctx"
	authcmd "ctf-platform/internal/module/auth/application/commands"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

type stubInstanceQuery struct {
	items      []*instancecontracts.InstanceInfo
	lastUserID int64
	called     bool
}

func (s *stubInstanceQuery) GetUserInstances(ctx context.Context, userID int64) ([]*instancecontracts.InstanceInfo, error) {
	s.called = true
	s.lastUserID = userID
	return s.items, nil
}

type stubChallengeQuery struct {
	detail          *challengecontracts.ChallengeDetailResp
	lastUserID      int64
	lastChallengeID int64
	called          bool
}

func (s *stubChallengeQuery) GetPublishedChallenge(ctx context.Context, userID, challengeID int64) (*challengecontracts.ChallengeDetailResp, error) {
	s.called = true
	s.lastUserID = userID
	s.lastChallengeID = challengeID
	return s.detail, nil
}

type stubOAuthTokenResolver struct {
	user              *authctx.CurrentUser
	clientID          string
	scope             string
	err               error
	lastToken         string
	lastRequiredScope string
	calls             int
}

func (s *stubOAuthTokenResolver) ResolveOAuthAccessToken(_ context.Context, token string, requiredScope string) (*authcmd.OAuthAccessTokenResolution, error) {
	s.calls++
	s.lastToken = token
	s.lastRequiredScope = requiredScope
	if s.err != nil {
		return nil, s.err
	}
	if s.user == nil {
		return nil, authcontracts.NewOAuthInvalidGrant("access token is invalid or expired")
	}
	return &authcmd.OAuthAccessTokenResolution{
		User:     *s.user,
		ClientID: s.clientID,
		Scope:    s.scope,
	}, nil
}

type recordingAuditRecorder struct {
	entries []auditlog.Entry
}

func (r *recordingAuditRecorder) Record(_ context.Context, entry auditlog.Entry) error {
	r.entries = append(r.entries, entry)
	return nil
}

func TestHandlerListsCurrentChallengeTool(t *testing.T) {
	handler := NewHandler(Deps{})
	resp := postMCPRaw(t, handler, nil, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
	})

	if resp.Code != 0 {
		t.Fatalf("unexpected error response: %+v", resp)
	}
	tools := resp.Result["tools"].([]any)
	if len(tools) != 1 {
		t.Fatalf("tools length = %d, want 1", len(tools))
	}
	tool := tools[0].(map[string]any)
	if tool["name"] != "get_current_challenge" {
		t.Fatalf("tool name = %v, want get_current_challenge", tool["name"])
	}
}

func TestHandlerAcceptsInitializedNotification(t *testing.T) {
	handler := NewHandler(Deps{})
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/mcp", func(c *gin.Context) {
		authctx.SetCurrentUser(c, authctx.CurrentUser{UserID: 42, Username: "student"})
		c.Next()
	}, handler.ServeHTTP)

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(`{"jsonrpc":"2.0","method":"notifications/initialized"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d body = %s, want 204", rec.Code, rec.Body.String())
	}
}

func TestHandlerCallsCurrentChallengeToolWithCurrentUser(t *testing.T) {
	now := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	instances := &stubInstanceQuery{items: []*instancecontracts.InstanceInfo{
		{
			ID:             1001,
			ChallengeID:    7,
			ChallengeTitle: "SQL 注入基础",
			Status:         instancecontracts.InstanceStatusRunning,
			AccessURL:      "http://example.test/instances/1001/proxy",
			ExpiresAt:      now.Add(time.Hour),
			CreatedAt:      now,
		},
	}}
	challenges := &stubChallengeQuery{detail: &challengecontracts.ChallengeDetailResp{
		ID:          7,
		Title:       "SQL 注入基础",
		Description: "找到登录绕过点",
		Category:    "web",
		Difficulty:  "easy",
		Points:      100,
		NeedTarget:  true,
		IsSolved:    false,
	}}
	tokens := &stubOAuthTokenResolver{
		user:     &authctx.CurrentUser{UserID: 42, Username: "student", Role: "student"},
		clientID: "client_test",
		scope:    authcontracts.OAuthScopeMCPChallengeRead,
	}
	audit := &recordingAuditRecorder{}
	handler := NewHandler(Deps{Instances: instances, Challenges: challenges, Tokens: tokens, AuditRecorder: audit})

	resp := postMCPWithHeaders(t, handler, map[string]string{"Authorization": "Bearer access-token-1"}, map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-1",
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "get_current_challenge",
			"arguments": map[string]any{},
		},
	})

	if resp.Code != 0 {
		t.Fatalf("unexpected error response: %+v", resp)
	}
	if tokens.lastToken != "access-token-1" || tokens.lastRequiredScope != authcontracts.OAuthScopeMCPChallengeRead {
		t.Fatalf("oauth token=%q scope=%q", tokens.lastToken, tokens.lastRequiredScope)
	}
	if !instances.called || instances.lastUserID != 42 {
		t.Fatalf("instances called=%v user=%d, want called with 42", instances.called, instances.lastUserID)
	}
	if !challenges.called || challenges.lastUserID != 42 || challenges.lastChallengeID != 7 {
		t.Fatalf("challenge called=%v user=%d challenge=%d, want user=42 challenge=7", challenges.called, challenges.lastUserID, challenges.lastChallengeID)
	}

	content := resp.Result["structuredContent"].(map[string]any)
	if content["has_current_challenge"] != true {
		t.Fatalf("has_current_challenge = %v, want true", content["has_current_challenge"])
	}
	challenge := content["challenge"].(map[string]any)
	if challenge["title"] != "SQL 注入基础" {
		t.Fatalf("challenge title = %v", challenge["title"])
	}
	instance := content["instance"].(map[string]any)
	if instance["id"].(float64) != 1001 {
		t.Fatalf("instance id = %v, want 1001", instance["id"])
	}
	if len(audit.entries) != 1 {
		t.Fatalf("expected one MCP audit entry, got %+v", audit.entries)
	}
	entry := audit.entries[0]
	if entry.UserID == nil || *entry.UserID != 42 || entry.Action != auditlog.ActionRead || entry.ResourceType != "mcp_tool" {
		t.Fatalf("unexpected MCP audit entry: %+v", entry)
	}
	if entry.Detail["tool"] != toolGetCurrentChallenge || entry.Detail["result"] != "success" {
		t.Fatalf("unexpected MCP audit detail: %+v", entry.Detail)
	}
	if entry.Detail["client_id"] != "client_test" || entry.Detail["scope"] != authcontracts.OAuthScopeMCPChallengeRead {
		t.Fatalf("MCP audit should include oauth client and scope: %+v", entry.Detail)
	}
}

func TestHandlerReturnsOAuthChallengeWhenAuthRequired(t *testing.T) {
	handler := NewHandler(Deps{})

	rec := postMCPRecorder(t, handler, nil, map[string]any{
		"jsonrpc": "2.0",
		"id":      "auth-required",
		"method":  "tools/call",
		"params": map[string]any{
			"name": toolGetCurrentChallenge,
		},
	})

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body = %s, want 401", rec.Code, rec.Body.String())
	}
	challenge := rec.Header().Get("WWW-Authenticate")
	if !strings.Contains(challenge, `Bearer realm="ctf-mcp"`) || !strings.Contains(challenge, `resource_metadata=`) {
		t.Fatalf("unexpected WWW-Authenticate header: %q", challenge)
	}

	var resp mcpResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v\nbody: %s", err, rec.Body.String())
	}
	if resp.Error == nil {
		t.Fatalf("expected auth error, got %+v", resp)
	}
	errPayload := resp.Error.(map[string]any)
	if errPayload["code"].(float64) != -32001 {
		t.Fatalf("error code = %v, want -32001", errPayload["code"])
	}
	data := errPayload["data"].(map[string]any)
	if data["auth_method"] != "oauth" || data["resource_metadata"] == "" || data["required_scope"] != "mcp:challenge:read" {
		t.Fatalf("unexpected auth data: %+v", data)
	}
	if _, exists := data["token_url"]; exists {
		t.Fatalf("auth data must not expose legacy token_url: %+v", data)
	}
}

func TestHandlerRejectsInvalidOAuthTokenWithoutCallingTools(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{name: "invalid token", err: authcontracts.NewOAuthInvalidGrant("access token is invalid or expired")},
		{name: "missing scope", err: authcontracts.NewOAuthInvalidScope("access token does not include required scope")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			instances := &stubInstanceQuery{}
			challenges := &stubChallengeQuery{}
			tokens := &stubOAuthTokenResolver{err: tc.err}
			handler := NewHandler(Deps{Instances: instances, Challenges: challenges, Tokens: tokens})

			rec := postMCPRecorder(t, handler, func(c *gin.Context) {
				c.Request.Header.Set("Authorization", "Bearer bad-token")
			}, map[string]any{
				"jsonrpc": "2.0",
				"id":      "bad-token",
				"method":  "tools/call",
				"params": map[string]any{
					"name": toolGetCurrentChallenge,
				},
			})

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d body = %s, want 401", rec.Code, rec.Body.String())
			}
			if tokens.lastToken != "bad-token" || tokens.lastRequiredScope != authcontracts.OAuthScopeMCPChallengeRead {
				t.Fatalf("oauth token resolver token=%q scope=%q", tokens.lastToken, tokens.lastRequiredScope)
			}
			if instances.called || challenges.called {
				t.Fatalf("tool services must not be called for invalid OAuth token, instances=%v challenges=%v", instances.called, challenges.called)
			}
		})
	}
}

func TestHandlerRateLimitsAuthenticatedToolCalls(t *testing.T) {
	instances := &stubInstanceQuery{}
	handler := NewHandler(Deps{
		Instances: instances,
		Tokens: &stubOAuthTokenResolver{
			user:     &authctx.CurrentUser{UserID: 42, Username: "student", Role: "student"},
			clientID: "client_test",
			scope:    authcontracts.OAuthScopeMCPChallengeRead,
		},
		RateLimit: func(ctx context.Context, user authctx.CurrentUser) (RateLimitResult, error) {
			return RateLimitResult{
				Allowed:    false,
				Limit:      1,
				Remaining:  0,
				ResetAt:    time.Date(2026, 7, 1, 11, 0, 0, 0, time.UTC),
				RetryAfter: time.Minute,
			}, nil
		},
	})

	resp := postMCPWithHeaders(t, handler, map[string]string{"Authorization": "Bearer access-token-1"}, map[string]any{
		"jsonrpc": "2.0",
		"id":      "rate-limited",
		"method":  "tools/call",
		"params": map[string]any{
			"name": toolGetCurrentChallenge,
		},
	})

	if resp.Error == nil {
		t.Fatalf("expected rate limit error, got %+v", resp)
	}
	errPayload := resp.Error.(map[string]any)
	if errPayload["code"].(float64) != -32002 {
		t.Fatalf("error code = %v, want -32002", errPayload["code"])
	}
	if instances.called {
		t.Fatal("instances query should not be called after MCP rate limit is exceeded")
	}
}

func TestHandlerReturnsEmptyCurrentChallengeWhenUserHasNoActiveInstance(t *testing.T) {
	instances := &stubInstanceQuery{items: []*instancecontracts.InstanceInfo{
		{
			ID:          1001,
			ChallengeID: 7,
			Status:      instancecontracts.InstanceStatusExpired,
			CreatedAt:   time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC),
		},
	}}
	challenges := &stubChallengeQuery{}
	handler := NewHandler(Deps{Instances: instances, Challenges: challenges})

	resp := postMCP(t, handler, authctx.CurrentUser{UserID: 42, Username: "student"}, map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-empty",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "get_current_challenge",
		},
	})

	if resp.Code != 0 {
		t.Fatalf("unexpected error response: %+v", resp)
	}
	if challenges.called {
		t.Fatal("challenge query should not be called when no active instance exists")
	}
	content := resp.Result["structuredContent"].(map[string]any)
	if content["has_current_challenge"] != false {
		t.Fatalf("has_current_challenge = %v, want false", content["has_current_challenge"])
	}
}

type mcpResponse struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id"`
	Result  map[string]any `json:"result"`
	Error   any            `json:"error"`
	Code    int
}

func postMCP(t *testing.T, handler *Handler, user authctx.CurrentUser, body map[string]any) mcpResponse {
	t.Helper()
	return postMCPRaw(t, handler, func(c *gin.Context) {
		authctx.SetCurrentUser(c, user)
	}, body)
}

func postMCPWithHeaders(t *testing.T, handler *Handler, headers map[string]string, body map[string]any) mcpResponse {
	t.Helper()
	return postMCPRaw(t, handler, func(c *gin.Context) {
		for key, value := range headers {
			c.Request.Header.Set(key, value)
		}
	}, body)
}

func postMCPRaw(t *testing.T, handler *Handler, prepare func(*gin.Context), body map[string]any) mcpResponse {
	t.Helper()

	rec := postMCPRecorder(t, handler, prepare, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	var resp mcpResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v\nbody: %s", err, rec.Body.String())
	}
	return resp
}

func postMCPRecorder(t *testing.T, handler *Handler, prepare func(*gin.Context), body map[string]any) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/mcp", func(c *gin.Context) {
		if prepare != nil {
			prepare(c)
		}
		c.Next()
	}, handler.ServeHTTP)

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}
