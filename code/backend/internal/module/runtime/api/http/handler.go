package http

import (
	"bytes"
	"context"
	"io"
	stdhttp "net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/pkg/errcode"
	"ctf-platform/pkg/response"
)

const proxyAccessCookieName = "ctf_instance_proxy_ticket"
const sharedProofTicketHeader = "X-CTF-Proxy-Ticket"

type CookieConfig struct {
	Secure   bool
	SameSite stdhttp.SameSite
}

type runtimeService interface {
	DestroyInstanceWithContext(ctx context.Context, instanceID, userID int64) error
	ExtendInstanceWithContext(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error)
	GetAccessURLWithContext(ctx context.Context, instanceID, userID int64) (string, error)
	GetUserInstancesWithContext(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error)
	ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error)
	DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error
	IssueProxyTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, error)
	IssueSharedProof(ctx context.Context, proxyTicket string) (*dto.SharedProofIssueResp, error)
	ResolveProxyTicket(ctx context.Context, ticket string) (*runtimeports.ProxyTicketClaims, error)
	ProxyTicketMaxAge() int
	ProxyBodyPreviewSize() int
}

type runtimeProxyTrafficRecorder interface {
	RecordRuntimeProxyTrafficEvent(ctx context.Context, instanceID, userID int64, method, requestPath string, statusCode int) error
}

type Handler struct {
	service              runtimeService
	auditRecorder        auditlog.Recorder
	cookieConfig         CookieConfig
	proxyTrafficRecorder runtimeProxyTrafficRecorder
}

func NewHandler(
	service runtimeService,
	auditRecorder auditlog.Recorder,
	cookieConfig CookieConfig,
	proxyTrafficRecorder runtimeProxyTrafficRecorder,
) *Handler {
	return &Handler{
		service:              service,
		auditRecorder:        auditRecorder,
		cookieConfig:         cookieConfig,
		proxyTrafficRecorder: proxyTrafficRecorder,
	}
}

func (h *Handler) DestroyInstance(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	instanceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.service.DestroyInstanceWithContext(c.Request.Context(), instanceID, userID); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *Handler) ExtendInstance(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	instanceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.ExtendInstanceWithContext(c.Request.Context(), instanceID, userID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *Handler) AccessInstance(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	instanceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ValidationError(c, err)
		return
	}

	_, err = h.service.GetAccessURLWithContext(c.Request.Context(), instanceID, currentUser.UserID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	ticket, err := h.service.IssueProxyTicket(c.Request.Context(), currentUser, instanceID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	if h.auditRecorder != nil {
		_ = h.auditRecorder.Record(c.Request.Context(), auditlog.Entry{
			UserID:       &currentUser.UserID,
			Action:       model.AuditActionRead,
			ResourceType: "instance_access",
			ResourceID:   &instanceID,
			Detail: map[string]any{
				"method":     c.Request.Method,
				"path":       c.FullPath(),
				"status":     200,
				"request_id": c.GetString("request_id"),
				"username":   currentUser.Username,
			},
			IPAddress: c.ClientIP(),
			UserAgent: stringPtr(c.Request.UserAgent()),
		})
	}

	response.Success(c, &dto.InstanceAccessResp{
		AccessURL: buildProxyAccessURL(instanceID, ticket),
	})
}

func (h *Handler) IssueSharedProof(c *gin.Context) {
	resp, err := h.service.IssueSharedProof(c.Request.Context(), strings.TrimSpace(c.GetHeader(sharedProofTicketHeader)))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *Handler) ProxyInstance(c *gin.Context) {
	instanceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ValidationError(c, err)
		return
	}

	claims, redirectURL, err := h.resolveProxyClaims(c, instanceID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	if redirectURL != "" {
		c.Redirect(stdhttp.StatusFound, redirectURL)
		return
	}

	targetURL, err := h.service.GetAccessURLWithContext(c.Request.Context(), instanceID, claims.UserID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	parsedTarget, err := url.Parse(targetURL)
	if err != nil || parsedTarget.Scheme == "" || parsedTarget.Host == "" {
		if err == nil {
			err = errcode.ErrInternal.WithCause(io.ErrUnexpectedEOF)
		}
		response.FromError(c, errcode.ErrInternal.WithCause(err))
		return
	}

	proxyPath := c.Param("proxyPath")
	joinedPath := joinProxyPath(parsedTarget.Path, proxyPath)
	rawQuery := c.Request.URL.Query()
	rawQuery.Del("ticket")

	bodyPreview, bodyCaptured, bodyTruncated := captureProxyBodyPreview(c.Request, h.service.ProxyBodyPreviewSize())
	shouldAudit := shouldAuditProxyRequest(c.Request.Method, joinedPath)
	requestID := c.GetString("request_id")
	username := claims.Username

	proxy := httputil.NewSingleHostReverseProxy(parsedTarget)
	proxy.Director = func(req *stdhttp.Request) {
		req.URL.Scheme = parsedTarget.Scheme
		req.URL.Host = parsedTarget.Host
		req.URL.Path = joinedPath
		req.URL.RawPath = joinedPath
		req.URL.RawQuery = rawQuery.Encode()
		req.Host = parsedTarget.Host
		req.Header.Del("Authorization")
		req.Header.Del("Cookie")
	}
	proxy.ModifyResponse = func(resp *stdhttp.Response) error {
		if shouldAudit {
			h.recordProxyAudit(c, claims, instanceID, username, requestID, joinedPath, resp.StatusCode, bodyPreview, bodyCaptured, bodyTruncated)
		}
		return nil
	}
	proxy.ErrorHandler = func(writer stdhttp.ResponseWriter, req *stdhttp.Request, proxyErr error) {
		if shouldAudit {
			h.recordProxyAudit(c, claims, instanceID, username, requestID, joinedPath, stdhttp.StatusBadGateway, bodyPreview, bodyCaptured, bodyTruncated)
		}
		response.FromError(c, errcode.ErrServiceUnavailable.WithCause(proxyErr))
	}

	proxy.ServeHTTP(newProxyResponseWriter(c.Writer), c.Request)
}

func (h *Handler) ListInstances(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID

	instances, err := h.service.GetUserInstancesWithContext(c.Request.Context(), userID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, instances)
}

func (h *Handler) ListTeacherInstances(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	var query dto.TeacherInstanceQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	items, err := h.service.ListTeacherInstances(c.Request.Context(), currentUser.UserID, currentUser.Role, &query)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, items)
}

func (h *Handler) DestroyTeacherInstance(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	instanceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.service.DestroyTeacherInstance(c.Request.Context(), instanceID, currentUser.UserID, currentUser.Role); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func buildProxyAccessURL(instanceID int64, ticket string) string {
	return "/api/v1/instances/" + strconv.FormatInt(instanceID, 10) + "/proxy/?ticket=" + url.QueryEscape(ticket)
}

func (h *Handler) resolveProxyClaims(c *gin.Context, instanceID int64) (*runtimeports.ProxyTicketClaims, string, error) {
	if h.service == nil {
		return nil, "", errcode.ErrInternal.WithCause(errcode.ErrServiceUnavailable)
	}

	if ticket := strings.TrimSpace(c.Query("ticket")); ticket != "" {
		claims, err := h.service.ResolveProxyTicket(c.Request.Context(), ticket)
		if err != nil {
			return nil, "", err
		}
		if claims.InstanceID != instanceID {
			return nil, "", errcode.ErrForbidden
		}

		setProxyAccessCookie(c, ticket, instanceID, h.service.ProxyTicketMaxAge(), h.cookieConfig)
		return claims, sanitizedProxyRedirectURL(c), nil
	}

	ticketCookie, err := c.Cookie(proxyAccessCookieName)
	if err != nil {
		return nil, "", errcode.ErrProxyTicketInvalid
	}
	claims, err := h.service.ResolveProxyTicket(c.Request.Context(), ticketCookie)
	if err != nil {
		return nil, "", err
	}
	if claims.InstanceID != instanceID {
		return nil, "", errcode.ErrForbidden
	}
	return claims, "", nil
}

func setProxyAccessCookie(c *gin.Context, ticket string, instanceID int64, maxAge int, cfg CookieConfig) {
	stdhttp.SetCookie(c.Writer, &stdhttp.Cookie{
		Name:     proxyAccessCookieName,
		Value:    ticket,
		Path:     "/api/v1/instances/" + strconv.FormatInt(instanceID, 10) + "/proxy",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	})
}

func sanitizedProxyRedirectURL(c *gin.Context) string {
	query := c.Request.URL.Query()
	query.Del("ticket")
	encoded := query.Encode()
	if encoded == "" {
		return c.Request.URL.Path
	}
	return c.Request.URL.Path + "?" + encoded
}

func joinProxyPath(basePath, proxyPath string) string {
	if proxyPath == "" {
		if basePath == "" {
			return "/"
		}
		return basePath
	}

	if basePath == "" || basePath == "/" {
		if strings.HasPrefix(proxyPath, "/") {
			return proxyPath
		}
		return "/" + proxyPath
	}
	return path.Clean(strings.TrimRight(basePath, "/") + "/" + strings.TrimLeft(proxyPath, "/"))
}

func captureProxyBodyPreview(req *stdhttp.Request, limit int) (string, bool, bool) {
	if req == nil || req.Body == nil || limit <= 0 {
		return "", false, false
	}
	if req.ContentLength < 0 || req.ContentLength > int64(limit) || !isTextualContentType(req.Header.Get("Content-Type")) {
		return "", false, req.ContentLength > int64(limit)
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", false, false
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	return strings.TrimSpace(string(body)), true, false
}

func isTextualContentType(contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	return contentType == "" ||
		strings.HasPrefix(contentType, "text/") ||
		strings.Contains(contentType, "json") ||
		strings.Contains(contentType, "xml") ||
		strings.Contains(contentType, "x-www-form-urlencoded")
}

func shouldAuditProxyRequest(method, requestPath string) bool {
	if method != stdhttp.MethodGet && method != stdhttp.MethodHead {
		return true
	}

	lowerPath := strings.ToLower(requestPath)
	for _, ext := range []string{".css", ".js", ".map", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf"} {
		if strings.HasSuffix(lowerPath, ext) {
			return false
		}
	}
	return true
}

func (h *Handler) recordProxyAudit(
	c *gin.Context,
	claims *runtimeports.ProxyTicketClaims,
	instanceID int64,
	username string,
	requestID string,
	targetPath string,
	status int,
	bodyPreview string,
	bodyCaptured bool,
	bodyTruncated bool,
) {
	if claims == nil {
		return
	}

	detail := map[string]any{
		"method":         c.Request.Method,
		"target_path":    targetPath,
		"target_query":   c.Request.URL.RawQuery,
		"status":         status,
		"request_id":     requestID,
		"username":       username,
		"content_type":   c.Request.Header.Get("Content-Type"),
		"content_length": c.Request.ContentLength,
	}
	if bodyCaptured && bodyPreview != "" {
		detail["payload_preview"] = bodyPreview
	}
	if bodyTruncated {
		detail["payload_truncated"] = true
	}

	if h.auditRecorder != nil {
		_ = h.auditRecorder.Record(c.Request.Context(), auditlog.Entry{
			UserID:       &claims.UserID,
			Action:       proxyAuditAction(c.Request.Method),
			ResourceType: "instance_proxy_request",
			ResourceID:   &instanceID,
			Detail:       detail,
			IPAddress:    c.ClientIP(),
			UserAgent:    stringPtr(c.Request.UserAgent()),
		})
	}
	if h.proxyTrafficRecorder != nil {
		_ = h.proxyTrafficRecorder.RecordRuntimeProxyTrafficEvent(
			c.Request.Context(),
			instanceID,
			claims.UserID,
			c.Request.Method,
			targetPath,
			status,
		)
	}
}

func proxyAuditAction(method string) string {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case stdhttp.MethodPost:
		return model.AuditActionSubmit
	case stdhttp.MethodPut, stdhttp.MethodPatch:
		return model.AuditActionUpdate
	case stdhttp.MethodDelete:
		return model.AuditActionDelete
	default:
		return model.AuditActionRead
	}
}

type proxyResponseWriter struct {
	gin.ResponseWriter
}

func newProxyResponseWriter(writer gin.ResponseWriter) stdhttp.ResponseWriter {
	return proxyResponseWriter{ResponseWriter: writer}
}

func (w proxyResponseWriter) CloseNotify() <-chan bool {
	return make(chan bool)
}
