package middleware

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	commonmapper "ctf-platform/internal/shared/mapperhelper"
)

const awdReadinessAuditPayloadKey = "awd_readiness_audit_payload"

type AWDReadinessAuditItem struct {
	AWDChallengeID  int64                `json:"awd_challenge_id"`
	Title           string               `json:"title"`
	CheckerType     model.AWDCheckerType `json:"checker_type,omitempty"`
	ValidationState string               `json:"validation_state"`
	BlockingReason  string               `json:"blocking_reason"`
}

type AWDReadinessAuditPayload struct {
	GateAction            string
	ForceOverride         bool
	GateAllowed           bool
	OverrideReason        string
	BlockingCount         int
	GlobalBlockingReasons []string
	BlockingItems         []AWDReadinessAuditItem
	ExecutionOutcome      string
	ExecutionError        string
}

func BuildAWDReadinessAuditPayload(gateAction string, overrideReason *string, snapshot *dto.AWDReadinessResp, executionErr error) *AWDReadinessAuditPayload {
	if snapshot == nil {
		return nil
	}

	items := make([]AWDReadinessAuditItem, 0, len(snapshot.Items))
	for _, item := range snapshot.Items {
		if item == nil || strings.TrimSpace(item.BlockingReason) == "" {
			continue
		}
		items = append(items, AWDReadinessAuditItem{
			AWDChallengeID:  item.AWDChallengeID,
			Title:           item.Title,
			CheckerType:     item.CheckerType,
			ValidationState: item.ValidationState,
			BlockingReason:  item.BlockingReason,
		})
	}

	payload := &AWDReadinessAuditPayload{
		GateAction:            gateAction,
		ForceOverride:         true,
		GateAllowed:           true,
		BlockingCount:         snapshot.BlockingCount,
		GlobalBlockingReasons: append([]string(nil), snapshot.GlobalBlockingReasons...),
		BlockingItems:         items,
		ExecutionOutcome:      "succeeded",
	}
	if overrideReason != nil {
		payload.OverrideReason = strings.TrimSpace(*overrideReason)
	}
	if executionErr != nil {
		payload.ExecutionOutcome = "failed"
		payload.ExecutionError = executionErr.Error()
	}
	return payload
}

func SetAWDReadinessAuditPayload(c *gin.Context, payload *AWDReadinessAuditPayload) {
	if c == nil || payload == nil {
		return
	}
	c.Set(awdReadinessAuditPayloadKey, payload)
}

func AWDReadinessAudit(recorder auditlog.Recorder, log *zap.Logger) gin.HandlerFunc {
	if log == nil {
		log = zap.NewNop()
	}

	return func(c *gin.Context) {
		c.Next()

		if recorder == nil {
			return
		}

		payload := getAWDReadinessAuditPayload(c)
		if payload == nil || !payload.ForceOverride || !payload.GateAllowed {
			return
		}

		currentUser := authctx.MustCurrentUser(c)
		var userID *int64
		if currentUser.UserID > 0 {
			userID = &currentUser.UserID
		}

		detail := map[string]any{
			"module":                  "awd_readiness_gate",
			"method":                  c.Request.Method,
			"path":                    c.FullPath(),
			"status":                  c.Writer.Status(),
			"request_id":              c.GetString(RequestIDKey),
			"gate_action":             payload.GateAction,
			"override_reason":         payload.OverrideReason,
			"blocking_count":          payload.BlockingCount,
			"global_blocking_reasons": append([]string(nil), payload.GlobalBlockingReasons...),
			"blocking_items":          awdReadinessAuditItemsToDetail(payload.BlockingItems),
			"execution_outcome":       payload.ExecutionOutcome,
			"execution_error":         payload.ExecutionError,
		}
		if detail["path"] == "" {
			detail["path"] = c.Request.URL.Path
		}
		if currentUser.Username != "" {
			detail["username"] = currentUser.Username
		}

		resourceID := awdReadinessAuditResourceID(c)
		if err := recorder.Record(c.Request.Context(), auditlog.Entry{
			UserID:       userID,
			Action:       model.AuditActionAdminOp,
			ResourceType: "contest",
			ResourceID:   resourceID,
			Detail:       detail,
			IPAddress:    c.ClientIP(),
			UserAgent:    commonmapper.NormalizeOptionalTrimmedString(c.Request.UserAgent()),
		}); err != nil {
			log.Error("awd_readiness_audit_record_failed", zap.Error(err))
		}
	}
}

func getAWDReadinessAuditPayload(c *gin.Context) *AWDReadinessAuditPayload {
	if c == nil {
		return nil
	}
	value, exists := c.Get(awdReadinessAuditPayloadKey)
	if !exists {
		return nil
	}
	payload, _ := value.(*AWDReadinessAuditPayload)
	return payload
}

func awdReadinessAuditItemsToDetail(items []AWDReadinessAuditItem) []map[string]any {
	if len(items) == 0 {
		return nil
	}
	detail := make([]map[string]any, 0, len(items))
	for _, item := range items {
		detail = append(detail, map[string]any{
			"awd_challenge_id": item.AWDChallengeID,
			"title":            item.Title,
			"checker_type":     item.CheckerType,
			"validation_state": item.ValidationState,
			"blocking_reason":  item.BlockingReason,
		})
	}
	return detail
}

func awdReadinessAuditResourceID(c *gin.Context) *int64 {
	if c == nil {
		return nil
	}
	if id := c.GetInt64("id"); id > 0 {
		return &id
	}
	parsed, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || parsed <= 0 {
		return nil
	}
	return &parsed
}
