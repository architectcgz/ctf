package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/authctx"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

const (
	methodInitialize  = "initialize"
	methodInitialized = "notifications/initialized"
	methodToolsList   = "tools/list"
	methodToolsCall   = "tools/call"

	toolGetCurrentChallenge = "get_current_challenge"

	defaultLoginURL = "/login"
	defaultTokenURL = "/api/v1/auth/mcp-token"

	rpcCodeAuthRequired = -32001
	rpcCodeRateLimited  = -32002
)

type instanceQueryService interface {
	GetUserInstances(ctx context.Context, userID int64) ([]*instancecontracts.InstanceInfo, error)
}

type challengeQueryService interface {
	GetPublishedChallenge(ctx context.Context, userID, challengeID int64) (*challengecontracts.ChallengeDetailResp, error)
}

type tokenResolver interface {
	ResolveMCPToken(ctx context.Context, token string) (*authctx.CurrentUser, error)
}

type RateLimitResult struct {
	Allowed    bool
	Limit      int
	Remaining  int
	ResetAt    time.Time
	RetryAfter time.Duration
}

type RateLimitFunc func(ctx context.Context, user authctx.CurrentUser) (RateLimitResult, error)

type Deps struct {
	Instances     instanceQueryService
	Challenges    challengeQueryService
	Tokens        tokenResolver
	RateLimit     RateLimitFunc
	AuditRecorder auditlog.Recorder
	LoginURL      string
	TokenURL      string
}

type Handler struct {
	instances     instanceQueryService
	challenges    challengeQueryService
	tokens        tokenResolver
	rateLimit     RateLimitFunc
	auditRecorder auditlog.Recorder
	loginURL      string
	tokenURL      string
}

func NewHandler(deps Deps) *Handler {
	loginURL := deps.LoginURL
	if loginURL == "" {
		loginURL = defaultLoginURL
	}
	tokenURL := deps.TokenURL
	if tokenURL == "" {
		tokenURL = defaultTokenURL
	}
	return &Handler{
		instances:     deps.Instances,
		challenges:    deps.Challenges,
		tokens:        deps.Tokens,
		rateLimit:     deps.RateLimit,
		auditRecorder: deps.AuditRecorder,
		loginURL:      loginURL,
		tokenURL:      tokenURL,
	}
}

func (h *Handler) ServeHTTP(c *gin.Context) {
	var req rpcRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, rpcErrorResponse(nil, -32700, "parse error"))
		return
	}

	switch req.Method {
	case methodInitialize:
		c.JSON(http.StatusOK, rpcSuccessResponse(req.ID, initializeResult()))
	case methodInitialized:
		c.Status(http.StatusNoContent)
	case methodToolsList:
		c.JSON(http.StatusOK, rpcSuccessResponse(req.ID, toolsListResult()))
	case methodToolsCall:
		user, err := h.currentUser(c)
		if err != nil {
			c.JSON(http.StatusOK, h.authRequiredResponse(req.ID))
			return
		}
		allowed, err := h.applyRateLimit(c, user)
		if err != nil {
			c.JSON(http.StatusOK, rpcErrorFromError(req.ID, err))
			return
		}
		if !allowed {
			c.JSON(http.StatusOK, h.rateLimitedResponse(req.ID))
			return
		}
		result, err := h.callTool(c.Request.Context(), user, req.Params)
		if err != nil {
			c.JSON(http.StatusOK, rpcErrorFromError(req.ID, err))
			return
		}
		h.recordToolAudit(c, user, toolGetCurrentChallenge, "success")
		c.JSON(http.StatusOK, rpcSuccessResponse(req.ID, result))
	default:
		c.JSON(http.StatusOK, rpcErrorResponse(req.ID, -32601, "method not found"))
	}
}

func (h *Handler) currentUser(c *gin.Context) (authctx.CurrentUser, error) {
	if value, ok := c.Get(authctx.CurrentUserKey); ok {
		if user, ok := value.(authctx.CurrentUser); ok && user.UserID > 0 {
			return user, nil
		}
	}

	token := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
	if token == "" || token == c.GetHeader("Authorization") {
		return authctx.CurrentUser{}, authcontracts.ErrMCPTokenInvalid
	}
	if h == nil || h.tokens == nil {
		return authctx.CurrentUser{}, authcontracts.ErrMCPTokenInvalid
	}
	user, err := h.tokens.ResolveMCPToken(c.Request.Context(), token)
	if err != nil {
		return authctx.CurrentUser{}, err
	}
	if user == nil || user.UserID <= 0 {
		return authctx.CurrentUser{}, authcontracts.ErrMCPTokenInvalid
	}
	authctx.SetCurrentUser(c, *user)
	return *user, nil
}

func (h *Handler) applyRateLimit(c *gin.Context, user authctx.CurrentUser) (bool, error) {
	if h == nil || h.rateLimit == nil {
		return true, nil
	}
	result, err := h.rateLimit(c.Request.Context(), user)
	if err != nil {
		return false, apperror.ErrInternal.WithCause(err)
	}
	c.Header("X-RateLimit-Limit", strconv.Itoa(result.Limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
	if !result.ResetAt.IsZero() {
		c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))
	}
	if result.RetryAfter > 0 {
		c.Header("Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))
	}
	return result.Allowed, nil
}

func (h *Handler) recordToolAudit(c *gin.Context, user authctx.CurrentUser, toolName, result string) {
	if h == nil || h.auditRecorder == nil {
		return
	}
	var userID *int64
	if user.UserID > 0 {
		userID = &user.UserID
	}
	_ = h.auditRecorder.Record(c.Request.Context(), auditlog.Entry{
		UserID:       userID,
		Action:       auditlog.ActionRead,
		ResourceType: "mcp_tool",
		Detail: map[string]any{
			"tool":       toolName,
			"result":     result,
			"request_id": c.GetString("request_id"),
		},
		IPAddress: c.ClientIP(),
		UserAgent: normalizeOptionalString(c.Request.UserAgent()),
	})
}

func (h *Handler) callTool(ctx context.Context, user authctx.CurrentUser, raw json.RawMessage) (toolCallResult, error) {
	var params toolCallParams
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &params); err != nil {
			return toolCallResult{}, apperror.ErrInvalidParams.WithCause(err)
		}
	}
	if params.Name != toolGetCurrentChallenge {
		return toolCallResult{}, apperror.ErrInvalidParams.WithMessage("未知的 MCP 工具")
	}
	return h.getCurrentChallenge(ctx, user.UserID)
}

func (h *Handler) getCurrentChallenge(ctx context.Context, userID int64) (toolCallResult, error) {
	if h == nil || h.instances == nil || h.challenges == nil {
		return toolCallResult{}, apperror.ErrInternal.WithCause(errors.New("mcp handler dependencies are not configured"))
	}

	instances, err := h.instances.GetUserInstances(ctx, userID)
	if err != nil {
		return toolCallResult{}, err
	}
	current := selectCurrentInstance(instances)
	if current == nil {
		return newToolResult(currentChallengeResult{HasCurrentChallenge: false})
	}

	detail, err := h.challenges.GetPublishedChallenge(ctx, userID, current.ChallengeID)
	if err != nil {
		return toolCallResult{}, err
	}
	return newToolResult(currentChallengeResult{
		HasCurrentChallenge: true,
		Instance:            current,
		Challenge:           detail,
	})
}

func selectCurrentInstance(items []*instancecontracts.InstanceInfo) *instancecontracts.InstanceInfo {
	var selected *instancecontracts.InstanceInfo
	for _, item := range items {
		if item == nil || !isCurrentInstanceStatus(item.Status) {
			continue
		}
		// 当前做题状态没有独立表，MCP 以用户最近的活动实例作为唯一可复用、权限受控的当前题目信号。
		if selected == nil ||
			instanceStatusPriority(item.Status) > instanceStatusPriority(selected.Status) ||
			(instanceStatusPriority(item.Status) == instanceStatusPriority(selected.Status) && item.CreatedAt.After(selected.CreatedAt)) ||
			(instanceStatusPriority(item.Status) == instanceStatusPriority(selected.Status) && item.CreatedAt.Equal(selected.CreatedAt) && item.ID > selected.ID) {
			selected = item
		}
	}
	return selected
}

func isCurrentInstanceStatus(status string) bool {
	switch status {
	case instancecontracts.InstanceStatusRunning, instancecontracts.InstanceStatusCreating, instancecontracts.InstanceStatusPending:
		return true
	default:
		return false
	}
}

func instanceStatusPriority(status string) int {
	switch status {
	case instancecontracts.InstanceStatusRunning:
		return 3
	case instancecontracts.InstanceStatusCreating:
		return 2
	case instancecontracts.InstanceStatusPending:
		return 1
	default:
		return 0
	}
}

func initializeResult() map[string]any {
	return map[string]any{
		"protocolVersion": "2025-06-18",
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    "ctf-platform-mcp",
			"version": "0.1.0",
		},
	}
}

func toolsListResult() map[string]any {
	return map[string]any{
		"tools": []toolDefinition{{
			Name:        toolGetCurrentChallenge,
			Description: "读取当前登录用户最近的活动实例及对应题目信息。",
			InputSchema: map[string]any{
				"type":                 "object",
				"properties":           map[string]any{},
				"additionalProperties": false,
			},
		}},
	}
}

func newToolResult(value currentChallengeResult) (toolCallResult, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return toolCallResult{}, apperror.ErrInternal.WithCause(err)
	}
	return toolCallResult{
		Content: []toolContent{{
			Type: "text",
			Text: string(payload),
		}},
		StructuredContent: value,
	}, nil
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func rpcSuccessResponse(id any, result any) rpcResponse {
	return rpcResponse{JSONRPC: "2.0", ID: id, Result: result}
}

func rpcErrorResponse(id any, code int, message string) rpcResponse {
	return rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: message}}
}

func rpcErrorResponseWithData(id any, code int, message string, data any) rpcResponse {
	return rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: message, Data: data}}
}

func (h *Handler) authRequiredResponse(id any) rpcResponse {
	return rpcErrorResponseWithData(id, rpcCodeAuthRequired, "请先登录 CTF 平台并签发 MCP Token", map[string]any{
		"login_url":   h.loginURL,
		"token_url":   h.tokenURL,
		"auth_method": "bearer_token",
	})
}

func (h *Handler) rateLimitedResponse(id any) rpcResponse {
	return rpcErrorResponseWithData(id, rpcCodeRateLimited, "MCP 调用过于频繁，请稍后再试", map[string]any{
		"retryable": true,
	})
}

func rpcErrorFromError(id any, err error) rpcResponse {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return rpcErrorResponse(id, -32000, appErr.Message)
	}
	return rpcErrorResponse(id, -32000, apperror.ErrInternal.Message)
}

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

type toolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type toolCallResult struct {
	Content           []toolContent          `json:"content"`
	StructuredContent currentChallengeResult `json:"structuredContent"`
}

type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type currentChallengeResult struct {
	HasCurrentChallenge bool                                    `json:"has_current_challenge"`
	Instance            *instancecontracts.InstanceInfo         `json:"instance,omitempty"`
	Challenge           *challengecontracts.ChallengeDetailResp `json:"challenge,omitempty"`
}

func normalizeOptionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}
