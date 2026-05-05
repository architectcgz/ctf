package middleware

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/authctx"
	commonmapper "ctf-platform/internal/shared/mapperhelper"
)

type AuditOptions struct {
	Action          string
	ResourceType    string
	ResourceIDParam string
	DetailBuilder   func(*gin.Context) map[string]any
}

const SkipAuditKey = "skip_audit"

func SetSkipAudit(c *gin.Context) {
	if c != nil {
		c.Set(SkipAuditKey, true)
	}
}

func ShouldSkipAudit(c *gin.Context) bool {
	if c == nil {
		return false
	}
	skip, _ := c.Get(SkipAuditKey)
	shouldSkip, _ := skip.(bool)
	return shouldSkip
}

func Audit(recorder auditlog.Recorder, options AuditOptions, log *zap.Logger) gin.HandlerFunc {
	if log == nil {
		log = zap.NewNop()
	}

	return func(c *gin.Context) {
		c.Next()

		if recorder == nil || c.Writer.Status() < 200 || c.Writer.Status() >= 300 || ShouldSkipAudit(c) {
			return
		}

		currentUser := authctx.MustCurrentUser(c)
		var userID *int64
		if currentUser.UserID > 0 {
			userID = &currentUser.UserID
		}

		detail := map[string]any{
			"method":     c.Request.Method,
			"path":       c.FullPath(),
			"status":     c.Writer.Status(),
			"request_id": c.GetString(RequestIDKey),
		}
		if detail["path"] == "" {
			detail["path"] = c.Request.URL.Path
		}
		if currentUser.Username != "" {
			detail["username"] = currentUser.Username
		}

		if options.DetailBuilder != nil {
			for key, value := range options.DetailBuilder(c) {
				detail[key] = value
			}
		}

		var resourceID *int64
		if options.ResourceIDParam != "" {
			if parsed, err := strconv.ParseInt(c.Param(options.ResourceIDParam), 10, 64); err == nil && parsed > 0 {
				resourceID = &parsed
			}
		}

		if err := recorder.Record(c.Request.Context(), auditlog.Entry{
			UserID:       userID,
			Action:       options.Action,
			ResourceType: options.ResourceType,
			ResourceID:   resourceID,
			Detail:       detail,
			IPAddress:    c.ClientIP(),
			UserAgent:    commonmapper.NormalizeOptionalTrimmedString(c.Request.UserAgent()),
		}); err != nil {
			log.Error("audit_log_record_failed",
				zap.String("action", options.Action),
				zap.String("resource_type", options.ResourceType),
				zap.Error(err),
			)
		}
	}
}

func DetailFromParams(params ...string) func(*gin.Context) map[string]any {
	return func(c *gin.Context) map[string]any {
		detail := make(map[string]any, len(params))
		for _, param := range params {
			value := strings.TrimSpace(c.Param(param))
			if value != "" {
				detail[param] = value
			}
		}
		if len(detail) == 0 {
			return nil
		}
		return detail
	}
}
