package evidence

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type proxyDetail struct {
	Method         string `json:"method"`
	TargetPath     string `json:"target_path"`
	TargetQuery    string `json:"target_query"`
	Status         int    `json:"status"`
	PayloadPreview string `json:"payload_preview"`
}

type Event struct {
	Type              string
	Source            string
	Stage             string
	UserID            int64
	TeamID            *int64
	ChallengeID       int64
	ContestID         *int64
	RoundID           *int64
	ServiceID         *int64
	VictimTeamID      *int64
	AWDChallengeID    int64
	AWDChallengeTitle string
	Title             string
	Timestamp         time.Time
	Detail            string
	Meta              map[string]any
}

type InstanceAccessInput struct {
	UserID      int64
	ChallengeID int64
	Title       string
	Timestamp   time.Time
}

type ProxyRequestInput struct {
	UserID      int64
	ChallengeID int64
	Title       string
	Timestamp   time.Time
	RawDetail   string
}

type ChallengeSubmissionInput struct {
	UserID      int64
	ChallengeID int64
	Title       string
	Timestamp   time.Time
	IsCorrect   bool
	Points      int
}

type ManualReviewInput struct {
	UserID       int64
	ChallengeID  int64
	Title        string
	Timestamp    time.Time
	ReviewStatus string
	Score        int
}

type WriteupInput struct {
	UserID           int64
	ChallengeID      int64
	Title            string
	Timestamp        time.Time
	WriteupTitle     string
	SubmissionStatus string
	VisibilityStatus string
	IsRecommended    bool
}

type AWDAttackInput struct {
	UserID            int64
	TeamID            *int64
	ContestID         *int64
	RoundID           *int64
	ServiceID         *int64
	VictimTeamID      *int64
	AWDChallengeID    int64
	AWDChallengeTitle string
	VictimTeamName    string
	Timestamp         time.Time
	IsSuccess         bool
	ScoreGained       int
	Scope             string
	AttackSource      string
}

type AWDTrafficInput struct {
	UserID            int64
	TeamID            *int64
	ContestID         *int64
	RoundID           *int64
	ServiceID         *int64
	VictimTeamID      *int64
	AWDChallengeID    int64
	AWDChallengeTitle string
	VictimTeamName    string
	Method            string
	Path              string
	StatusCode        int
	Timestamp         time.Time
}

func NewInstanceAccessEvent(input InstanceAccessInput) Event {
	return Event{
		Type:        "instance_access",
		Source:      "audit_logs",
		Stage:       "access",
		UserID:      input.UserID,
		ChallengeID: input.ChallengeID,
		Title:       input.Title,
		Timestamp:   input.Timestamp,
		Detail:      "访问攻击目标，开始与靶机进行实际交互",
		Meta: map[string]any{
			"event_stage": "access",
		},
	}
}

func NewProxyRequestEvent(input ProxyRequestInput) Event {
	meta := BuildProxyRequestMeta(input.RawDetail)
	meta["event_stage"] = "exploit"
	return Event{
		Type:        "instance_proxy_request",
		Source:      "audit_logs",
		Stage:       "exploit",
		UserID:      input.UserID,
		ChallengeID: input.ChallengeID,
		Title:       input.Title,
		Timestamp:   input.Timestamp,
		Detail:      BuildProxyRequestDetail(input.RawDetail),
		Meta:        meta,
	}
}

func NewChallengeSubmissionEvent(input ChallengeSubmissionInput) Event {
	detail := "提交未命中 Flag"
	if input.IsCorrect {
		detail = "提交命中 Flag"
	}
	return Event{
		Type:        "challenge_submission",
		Source:      "submissions",
		Stage:       "submit",
		UserID:      input.UserID,
		ChallengeID: input.ChallengeID,
		Title:       input.Title,
		Timestamp:   input.Timestamp,
		Detail:      detail,
		Meta: map[string]any{
			"event_stage": "submit",
			"is_correct":  input.IsCorrect,
			"points":      input.Points,
		},
	}
}

func NewManualReviewEvent(input ManualReviewInput) Event {
	return Event{
		Type:        "manual_review",
		Source:      "submissions",
		Stage:       "review",
		UserID:      input.UserID,
		ChallengeID: input.ChallengeID,
		Title:       input.Title,
		Timestamp:   input.Timestamp,
		Detail:      BuildManualReviewDetail(input.ReviewStatus),
		Meta: map[string]any{
			"event_stage":   "review",
			"review_status": input.ReviewStatus,
			"score":         input.Score,
		},
	}
}

func NewWriteupEvent(input WriteupInput) Event {
	return Event{
		Type:        "writeup",
		Source:      "submission_writeups",
		Stage:       "review",
		UserID:      input.UserID,
		ChallengeID: input.ChallengeID,
		Title:       input.Title,
		Timestamp:   input.Timestamp,
		Detail:      "提交或更新了解题复盘材料",
		Meta: map[string]any{
			"event_stage":       "review",
			"writeup_title":     input.WriteupTitle,
			"submission_status": input.SubmissionStatus,
			"visibility_status": input.VisibilityStatus,
			"is_recommended":    input.IsRecommended,
		},
	}
}

func NewAWDAttackEvent(input AWDAttackInput) Event {
	return Event{
		Type:              "awd_attack_submission",
		Source:            "awd_attack_logs",
		Stage:             "exploit",
		UserID:            input.UserID,
		TeamID:            input.TeamID,
		ChallengeID:       input.AWDChallengeID,
		ContestID:         input.ContestID,
		RoundID:           input.RoundID,
		ServiceID:         input.ServiceID,
		VictimTeamID:      input.VictimTeamID,
		AWDChallengeID:    input.AWDChallengeID,
		AWDChallengeTitle: input.AWDChallengeTitle,
		Title:             input.AWDChallengeTitle,
		Timestamp:         input.Timestamp,
		Detail:            BuildAWDAttackDetail(input.IsSuccess),
		Meta: map[string]any{
			"event_stage":         "exploit",
			"is_success":          input.IsSuccess,
			"score_gained":        input.ScoreGained,
			"contest_id":          derefInt64(input.ContestID),
			"round_id":            derefInt64(input.RoundID),
			"team_id":             derefInt64(input.TeamID),
			"victim_team_id":      derefInt64(input.VictimTeamID),
			"victim_team_name":    input.VictimTeamName,
			"service_id":          derefInt64(input.ServiceID),
			"awd_challenge_id":    input.AWDChallengeID,
			"awd_challenge_title": input.AWDChallengeTitle,
			"scope":               input.Scope,
			"source":              input.AttackSource,
		},
	}
}

func NewAWDTrafficEvent(input AWDTrafficInput) Event {
	return Event{
		Type:              "awd_traffic",
		Source:            "awd_traffic_events",
		Stage:             "exploit",
		UserID:            input.UserID,
		TeamID:            input.TeamID,
		ChallengeID:       input.AWDChallengeID,
		ContestID:         input.ContestID,
		RoundID:           input.RoundID,
		ServiceID:         input.ServiceID,
		VictimTeamID:      input.VictimTeamID,
		AWDChallengeID:    input.AWDChallengeID,
		AWDChallengeTitle: input.AWDChallengeTitle,
		Title:             input.AWDChallengeTitle,
		Timestamp:         input.Timestamp,
		Detail:            BuildAWDTrafficDetail(input.Method, input.Path, input.StatusCode),
		Meta: map[string]any{
			"event_stage":         "exploit",
			"request_method":      strings.ToUpper(strings.TrimSpace(input.Method)),
			"target_path":         input.Path,
			"status_code":         input.StatusCode,
			"contest_id":          derefInt64(input.ContestID),
			"round_id":            derefInt64(input.RoundID),
			"team_id":             derefInt64(input.TeamID),
			"victim_team_id":      derefInt64(input.VictimTeamID),
			"victim_team_name":    input.VictimTeamName,
			"service_id":          derefInt64(input.ServiceID),
			"awd_challenge_id":    input.AWDChallengeID,
			"awd_challenge_title": input.AWDChallengeTitle,
		},
	}
}

func BuildProxyRequestDetail(rawDetail string) string {
	parsed, ok := parseProxyDetail(rawDetail)
	if !ok {
		return "经平台代理向靶机发起了一次请求"
	}

	method := strings.ToUpper(strings.TrimSpace(parsed.Method))
	if method == "" {
		method = "REQUEST"
	}
	target := strings.TrimSpace(parsed.TargetPath)
	if target == "" {
		target = "/"
	}
	if query := strings.TrimSpace(parsed.TargetQuery); query != "" {
		target += "?" + query
	}

	summary := fmt.Sprintf("经平台代理发起 %s %s，请求返回 %d", method, target, parsed.Status)
	if strings.TrimSpace(parsed.PayloadPreview) != "" {
		summary += "，携带请求摘要"
	}
	return summary
}

func BuildProxyRequestMeta(rawDetail string) map[string]any {
	parsed, ok := parseProxyDetail(rawDetail)
	if !ok {
		return map[string]any{}
	}

	meta := map[string]any{}
	if value := strings.ToUpper(strings.TrimSpace(parsed.Method)); value != "" {
		meta["request_method"] = value
	}
	if value := strings.TrimSpace(parsed.TargetPath); value != "" {
		meta["target_path"] = value
	}
	if value := strings.TrimSpace(parsed.TargetQuery); value != "" {
		meta["target_query"] = value
	}
	if parsed.Status > 0 {
		meta["status_code"] = parsed.Status
	}
	if value := strings.TrimSpace(parsed.PayloadPreview); value != "" {
		meta["payload_preview"] = value
	}
	return meta
}

func BuildManualReviewDetail(status string) string {
	switch strings.TrimSpace(status) {
	case "approved":
		return "人工评审已通过"
	case "rejected":
		return "人工评审已驳回"
	case "pending":
		return "提交了待人工评审的答案"
	default:
		return "提交了人工评审答案"
	}
}

func BuildAWDAttackDetail(isSuccess bool) string {
	if isSuccess {
		return "AWD 攻击提交成功"
	}
	return "AWD 攻击提交未命中"
}

func BuildAWDTrafficDetail(method, path string, statusCode int) string {
	return fmt.Sprintf("AWD 流量 %s %s 返回 %d", strings.ToUpper(strings.TrimSpace(method)), path, statusCode)
}

func parseProxyDetail(rawDetail string) (proxyDetail, bool) {
	var detail proxyDetail
	if strings.TrimSpace(rawDetail) == "" {
		return detail, false
	}
	if err := json.Unmarshal([]byte(rawDetail), &detail); err != nil {
		return detail, false
	}
	return detail, true
}

func derefInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}
