package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type timelineEventRow struct {
	Type        string
	ChallengeID int64
	Title       string
	Timestamp   time.Time
	IsCorrect   *bool
	Points      *int
	Detail      string
}

func (r *Repository) GetUserTimeline(ctx context.Context, userID int64, limit, offset int) ([]practiceports.TimelineEventRecord, error) {
	if limit <= 0 {
		limit = 100
	}

	events := make([]timelineEventRow, 0)
	if err := r.dbWithContext(ctx).Raw(`
		SELECT
			events.type,
			events.challenge_id,
			c.title,
			events.timestamp,
			events.is_correct,
			events.points,
			events.detail
		FROM (
			SELECT
				'instance_start' AS type,
				i.challenge_id,
				i.created_at AS timestamp,
				NULL AS is_correct,
				NULL AS points,
				'启动练习实例' AS detail
			FROM instances i
			WHERE i.user_id = ?
			UNION ALL
			SELECT
				'flag_submit' AS type,
				s.challenge_id,
				s.submitted_at AS timestamp,
				s.is_correct,
				CASE WHEN s.is_correct THEN c.points ELSE NULL END AS points,
				CASE
					WHEN s.is_correct THEN '第 ' || CAST(s.attempt_no AS TEXT) || ' 次提交命中 Flag，获得 ' || CAST(COALESCE(c.points, 0) AS TEXT) || ' 分'
					ELSE '第 ' || CAST(s.attempt_no AS TEXT) || ' 次提交未命中 Flag'
				END AS detail
			FROM (
				SELECT
					submissions.*,
					ROW_NUMBER() OVER (PARTITION BY submissions.user_id, submissions.challenge_id ORDER BY submissions.submitted_at ASC, submissions.id ASC) AS attempt_no
				FROM submissions
				WHERE submissions.user_id = ?
			) s
			LEFT JOIN challenges c ON s.challenge_id = c.id
			UNION ALL
			SELECT
				'instance_destroy' AS type,
				i.challenge_id,
				i.destroyed_at AS timestamp,
				NULL AS is_correct,
				NULL AS points,
				'结束练习实例' AS detail
			FROM instances i
			WHERE i.user_id = ? AND i.status IN ('stopped', 'expired') AND i.destroyed_at IS NOT NULL
		) events
		LEFT JOIN challenges c ON events.challenge_id = c.id
		ORDER BY events.timestamp DESC
	`, userID, userID, userID).Scan(&events).Error; err != nil {
		return nil, fmt.Errorf("get user timeline: %w", err)
	}

	auditEvents, err := r.listUserAuditTimelineEvents(ctx, userID)
	if err != nil {
		return nil, err
	}
	events = append(events, auditEvents...)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})

	if offset >= len(events) {
		return []practiceports.TimelineEventRecord{}, nil
	}

	end := offset + limit
	if end > len(events) {
		end = len(events)
	}

	items := make([]practiceports.TimelineEventRecord, 0, end-offset)
	for _, event := range events[offset:end] {
		items = append(items, practiceports.TimelineEventRecord{
			Type:        event.Type,
			ChallengeID: event.ChallengeID,
			Title:       event.Title,
			Timestamp:   event.Timestamp,
			IsCorrect:   event.IsCorrect,
			Points:      event.Points,
			Detail:      event.Detail,
		})
	}

	return items, nil
}

func (r *Repository) listUserAuditTimelineEvents(ctx context.Context, userID int64) ([]timelineEventRow, error) {
	rows := make([]timelineEventRow, 0)
	if err := r.dbWithContext(ctx).Raw(`
		SELECT
			'challenge_detail_view' AS type,
			a.resource_id AS challenge_id,
			COALESCE(c.title, '') AS title,
			a.created_at AS timestamp,
			NULL AS is_correct,
			NULL AS points,
			'查看题目详情，开始分析题面与环境线索' AS detail
		FROM audit_logs a
		LEFT JOIN challenges c ON c.id = a.resource_id
		WHERE a.user_id = ? AND a.action = ? AND a.resource_type = ?
		UNION ALL
		SELECT
			'instance_extend' AS type,
			i.challenge_id,
			COALESCE(c.title, '') AS title,
			a.created_at AS timestamp,
			NULL AS is_correct,
			NULL AS points,
			'延长实例有效期，继续当前利用过程' AS detail
		FROM audit_logs a
		JOIN instances i ON i.id = a.resource_id
		LEFT JOIN challenges c ON c.id = i.challenge_id
		WHERE a.user_id = ? AND a.action = ? AND a.resource_type = ?
		UNION ALL
		SELECT
			'instance_access' AS type,
			i.challenge_id,
			COALESCE(c.title, '') AS title,
			a.created_at AS timestamp,
			NULL AS is_correct,
			NULL AS points,
			'访问攻击目标，开始与靶机进行实际交互' AS detail
		FROM audit_logs a
		JOIN instances i ON i.id = a.resource_id
		LEFT JOIN challenges c ON c.id = i.challenge_id
		WHERE a.user_id = ? AND a.action = ? AND a.resource_type = ?
	`,
		userID,
		model.AuditActionRead,
		"challenge_detail",
		userID,
		model.AuditActionUpdate,
		"instance",
		userID,
		model.AuditActionRead,
		"instance_access",
	).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list audit timeline events: %w", err)
	}

	proxyRows := make([]struct {
		ChallengeID int64     `gorm:"column:challenge_id"`
		Title       string    `gorm:"column:title"`
		Timestamp   time.Time `gorm:"column:timestamp"`
		Detail      string    `gorm:"column:detail"`
	}, 0)
	if err := r.dbWithContext(ctx).Table("audit_logs AS a").
		Select(strings.Join([]string{
			"i.challenge_id AS challenge_id",
			"COALESCE(c.title, '') AS title",
			"a.created_at AS timestamp",
			"a.detail AS detail",
		}, ", ")).
		Joins("JOIN instances i ON i.id = a.resource_id").
		Joins("LEFT JOIN challenges c ON c.id = i.challenge_id").
		Where("a.user_id = ? AND a.resource_type = ?", userID, "instance_proxy_request").
		Order("a.created_at DESC").
		Scan(&proxyRows).Error; err != nil {
		return nil, fmt.Errorf("list proxy audit timeline events: %w", err)
	}
	for _, row := range proxyRows {
		rows = append(rows, timelineEventRow{
			Type:        "instance_proxy_request",
			ChallengeID: row.ChallengeID,
			Title:       row.Title,
			Timestamp:   row.Timestamp,
			Detail:      buildProxyTimelineDetail(row.Detail),
		})
	}

	return rows, nil
}

func buildProxyTimelineDetail(rawDetail string) string {
	if strings.TrimSpace(rawDetail) == "" {
		return "经平台代理向靶机发起了一次请求"
	}

	var detail struct {
		Method         string `json:"method"`
		TargetPath     string `json:"target_path"`
		TargetQuery    string `json:"target_query"`
		Status         int    `json:"status"`
		PayloadPreview string `json:"payload_preview"`
	}
	if err := json.Unmarshal([]byte(rawDetail), &detail); err != nil {
		return "经平台代理向靶机发起了一次请求"
	}

	method := strings.ToUpper(strings.TrimSpace(detail.Method))
	if method == "" {
		method = "REQUEST"
	}
	target := detail.TargetPath
	if target == "" {
		target = "/"
	}
	if strings.TrimSpace(detail.TargetQuery) != "" {
		target += "?" + detail.TargetQuery
	}

	summary := fmt.Sprintf("经平台代理发起 %s %s，请求返回 %d", method, target, detail.Status)
	if strings.TrimSpace(detail.PayloadPreview) != "" {
		summary += "，携带请求摘要"
	}
	return summary
}
