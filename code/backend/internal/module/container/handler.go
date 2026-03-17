package container

import (
	"bytes"
	"io"
	"net/http"
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
	"ctf-platform/pkg/errcode"
	"ctf-platform/pkg/response"
)

const proxyAccessCookieName = "ctf_instance_proxy_ticket"

type ProxyCookieConfig struct {
	Secure   bool
	SameSite http.SameSite
}

type Handler struct {
	service       *Service
	proxyTickets  *ProxyTicketService
	auditRecorder auditlog.Recorder
	cookieConfig  ProxyCookieConfig
}

func NewHandler(service *Service, proxyTickets *ProxyTicketService, auditRecorder auditlog.Recorder, cookieConfig ProxyCookieConfig) *Handler {
	return &Handler{
		service:       service,
		proxyTickets:  proxyTickets,
		auditRecorder: auditRecorder,
		cookieConfig:  cookieConfig,
	}
}

func (h *Handler) CreateInstance(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.CreateInstanceWithContext(c.Request.Context(), userID, challengeID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
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

	ticket, _, err := h.proxyTickets.IssueTicket(c.Request.Context(), currentUser, instanceID)
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
		c.Redirect(http.StatusFound, redirectURL)
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

	bodyPreview, bodyCaptured, bodyTruncated := captureProxyBodyPreview(c.Request, h.service.config.ProxyBodyPreviewSize)
	shouldAudit := shouldAuditProxyRequest(c.Request.Method, joinedPath)
	requestID := c.GetString("request_id")
	username := claims.Username

	proxy := httputil.NewSingleHostReverseProxy(parsedTarget)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = parsedTarget.Scheme
		req.URL.Host = parsedTarget.Host
		req.URL.Path = joinedPath
		req.URL.RawPath = joinedPath
		req.URL.RawQuery = rawQuery.Encode()
		req.Host = parsedTarget.Host
		req.Header.Del("Authorization")
		req.Header.Del("Cookie")
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		if shouldAudit {
			h.recordProxyAudit(c, claims, instanceID, username, requestID, joinedPath, resp.StatusCode, bodyPreview, bodyCaptured, bodyTruncated)
		}
		return nil
	}
	proxy.ErrorHandler = func(writer http.ResponseWriter, req *http.Request, proxyErr error) {
		if shouldAudit {
			h.recordProxyAudit(c, claims, instanceID, username, requestID, joinedPath, http.StatusBadGateway, bodyPreview, bodyCaptured, bodyTruncated)
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

func (h *Handler) resolveProxyClaims(c *gin.Context, instanceID int64) (*ProxyTicketClaims, string, error) {
	if h.proxyTickets == nil {
		return nil, "", errcode.ErrInternal.WithCause(errcode.ErrServiceUnavailable)
	}

	if ticket := strings.TrimSpace(c.Query("ticket")); ticket != "" {
		claims, err := h.proxyTickets.ResolveTicket(c.Request.Context(), ticket)
		if err != nil {
			return nil, "", err
		}
		if claims.InstanceID != instanceID {
			return nil, "", errcode.ErrForbidden
		}

		setProxyAccessCookie(c, ticket, instanceID, int(h.service.config.ProxyTicketTTL.Seconds()), h.cookieConfig)
		return claims, sanitizedProxyRedirectURL(c), nil
	}

	ticketCookie, err := c.Cookie(proxyAccessCookieName)
	if err != nil {
		return nil, "", errcode.ErrProxyTicketInvalid
	}
	claims, err := h.proxyTickets.ResolveTicket(c.Request.Context(), ticketCookie)
	if err != nil {
		return nil, "", err
	}
	if claims.InstanceID != instanceID {
		return nil, "", errcode.ErrForbidden
	}
	return claims, "", nil
}

func setProxyAccessCookie(c *gin.Context, ticket string, instanceID int64, maxAge int, cfg ProxyCookieConfig) {
	http.SetCookie(c.Writer, &http.Cookie{
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

func captureProxyBodyPreview(req *http.Request, limit int) (string, bool, bool) {
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
	if method != http.MethodGet && method != http.MethodHead {
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
	claims *ProxyTicketClaims,
	instanceID int64,
	username string,
	requestID string,
	targetPath string,
	status int,
	bodyPreview string,
	bodyCaptured bool,
	bodyTruncated bool,
) {
	if h.auditRecorder == nil || claims == nil {
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

func proxyAuditAction(method string) string {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodPost:
		return model.AuditActionSubmit
	case http.MethodPut, http.MethodPatch:
		return model.AuditActionUpdate
	case http.MethodDelete:
		return model.AuditActionDelete
	default:
		return model.AuditActionRead
	}
}

type proxyResponseWriter struct {
	gin.ResponseWriter
}

func newProxyResponseWriter(writer gin.ResponseWriter) http.ResponseWriter {
	return proxyResponseWriter{ResponseWriter: writer}
}

func (w proxyResponseWriter) CloseNotify() <-chan bool {
	return make(chan bool)
}
