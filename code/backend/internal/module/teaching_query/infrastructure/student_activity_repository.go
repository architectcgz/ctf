package infrastructure

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/auditlog"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	queryports "ctf-platform/internal/module/teaching_query/ports"
	"ctf-platform/internal/teaching/evidence"
)

type StudentActivityRepository struct {
	db *gorm.DB
}

type timelineEventRow struct {
	Type        string    `gorm:"column:type"`
	ChallengeID int64     `gorm:"column:challenge_id"`
	Title       string    `gorm:"column:title"`
	Timestamp   time.Time `gorm:"column:timestamp"`
	IsCorrect   *bool     `gorm:"column:is_correct"`
	Points      *int      `gorm:"column:points"`
	Detail      string    `gorm:"column:detail"`
}

func NewStudentActivityRepository(db *gorm.DB) *StudentActivityRepository {
	return &StudentActivityRepository{db: db}
}

func (r *StudentActivityRepository) GetStudentTimeline(ctx context.Context, userID int64, limit, offset int) ([]queryports.TimelineEventRecord, error) {
	if limit <= 0 {
		limit = 100
	}

	rows := make([]timelineEventRow, 0)
	if err := r.db.WithContext(ctx).Raw(`
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
		LIMIT ? OFFSET ?
	`, userID, userID, userID, limit, offset).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("get student timeline: %w", err)
	}

	auditRows, err := r.listStudentAuditTimelineRows(ctx, userID)
	if err != nil {
		return nil, err
	}
	rows = append(rows, auditRows...)
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Timestamp.After(rows[j].Timestamp)
	})

	if offset >= len(rows) {
		return []queryports.TimelineEventRecord{}, nil
	}
	end := offset + limit
	if end > len(rows) {
		end = len(rows)
	}
	rows = rows[offset:end]

	events := make([]queryports.TimelineEventRecord, len(rows))
	for i, row := range rows {
		events[i] = queryports.TimelineEventRecord{
			Type:        row.Type,
			ChallengeID: row.ChallengeID,
			Title:       row.Title,
			Timestamp:   row.Timestamp,
			IsCorrect:   row.IsCorrect,
			Points:      row.Points,
			Detail:      row.Detail,
		}
	}
	return events, nil
}

func (r *StudentActivityRepository) GetStudentEvidence(ctx context.Context, userID int64, query evidence.Query) ([]queryports.EvidenceEventRecord, error) {
	events := make([]queryports.EvidenceEventRecord, 0)

	accessRows := make([]struct {
		ChallengeID int64     `gorm:"column:challenge_id"`
		Title       string    `gorm:"column:title"`
		Timestamp   time.Time `gorm:"column:timestamp"`
	}, 0)
	accessQuery := r.db.WithContext(ctx).Table("audit_logs AS a").
		Select(strings.Join([]string{
			"i.challenge_id AS challenge_id",
			"COALESCE(c.title, '') AS title",
			"a.created_at AS timestamp",
		}, ", ")).
		Joins("JOIN instances i ON i.id = a.resource_id").
		Joins("LEFT JOIN challenges c ON c.id = i.challenge_id").
		Where("a.user_id = ? AND a.resource_type = ?", userID, "instance_access")
	if query.ChallengeID != nil {
		accessQuery = accessQuery.Where("i.challenge_id = ?", *query.ChallengeID)
	}
	if err := accessQuery.Order("a.created_at ASC").Scan(&accessRows).Error; err != nil {
		return nil, fmt.Errorf("get student evidence access rows: %w", err)
	}
	for _, row := range accessRows {
		events = append(events, toEvidenceRecord(evidence.NewInstanceAccessEvent(evidence.InstanceAccessInput{
			UserID:      userID,
			ChallengeID: row.ChallengeID,
			Title:       row.Title,
			Timestamp:   row.Timestamp,
		})))
	}

	proxyRows := make([]struct {
		ChallengeID int64     `gorm:"column:challenge_id"`
		Title       string    `gorm:"column:title"`
		Timestamp   time.Time `gorm:"column:timestamp"`
		Detail      string    `gorm:"column:detail"`
	}, 0)
	proxyQuery := r.db.WithContext(ctx).Table("audit_logs AS a").
		Select(strings.Join([]string{
			"i.challenge_id AS challenge_id",
			"COALESCE(c.title, '') AS title",
			"a.created_at AS timestamp",
			"a.detail AS detail",
		}, ", ")).
		Joins("JOIN instances i ON i.id = a.resource_id").
		Joins("LEFT JOIN challenges c ON c.id = i.challenge_id").
		Where("a.user_id = ? AND a.resource_type = ?", userID, "instance_proxy_request")
	if query.ChallengeID != nil {
		proxyQuery = proxyQuery.Where("i.challenge_id = ?", *query.ChallengeID)
	}
	if err := proxyQuery.Order("a.created_at ASC").Scan(&proxyRows).Error; err != nil {
		return nil, fmt.Errorf("get student evidence proxy rows: %w", err)
	}
	for _, row := range proxyRows {
		events = append(events, toEvidenceRecord(evidence.NewProxyRequestEvent(evidence.ProxyRequestInput{
			UserID:      userID,
			ChallengeID: row.ChallengeID,
			Title:       row.Title,
			Timestamp:   row.Timestamp,
			RawDetail:   row.Detail,
		})))
	}

	submissionRows := make([]struct {
		ChallengeID int64     `gorm:"column:challenge_id"`
		Title       string    `gorm:"column:title"`
		Timestamp   time.Time `gorm:"column:timestamp"`
		IsCorrect   bool      `gorm:"column:is_correct"`
		Points      int       `gorm:"column:points"`
		Detail      string    `gorm:"column:detail"`
	}, 0)
	submissionQuery := r.db.WithContext(ctx).Table("submissions AS s").
		Select(strings.Join([]string{
			"s.challenge_id AS challenge_id",
			"COALESCE(c.title, '') AS title",
			"s.submitted_at AS timestamp",
			"s.is_correct AS is_correct",
			"CASE WHEN s.is_correct THEN COALESCE(c.points, 0) ELSE 0 END AS points",
			"CASE WHEN s.is_correct THEN '提交命中 Flag' ELSE '提交未命中 Flag' END AS detail",
		}, ", ")).
		Joins("LEFT JOIN challenges c ON c.id = s.challenge_id").
		Where("s.user_id = ?", userID)
	if query.ChallengeID != nil {
		submissionQuery = submissionQuery.Where("s.challenge_id = ?", *query.ChallengeID)
	}
	if err := submissionQuery.Order("s.submitted_at ASC").Scan(&submissionRows).Error; err != nil {
		return nil, fmt.Errorf("get student evidence submission rows: %w", err)
	}
	for _, row := range submissionRows {
		events = append(events, toEvidenceRecord(evidence.NewChallengeSubmissionEvent(evidence.ChallengeSubmissionInput{
			UserID:      userID,
			ChallengeID: row.ChallengeID,
			Title:       row.Title,
			Timestamp:   row.Timestamp,
			IsCorrect:   row.IsCorrect,
			Points:      row.Points,
		})))
	}

	manualReviewRows := make([]struct {
		ChallengeID  int64     `gorm:"column:challenge_id"`
		Title        string    `gorm:"column:title"`
		Timestamp    time.Time `gorm:"column:timestamp"`
		ReviewStatus string    `gorm:"column:review_status"`
		Score        int       `gorm:"column:score"`
	}, 0)
	manualReviewQuery := r.db.WithContext(ctx).Table("submissions AS s").
		Select(strings.Join([]string{
			"s.challenge_id AS challenge_id",
			"COALESCE(c.title, '') AS title",
			"s.updated_at AS timestamp",
			"s.review_status AS review_status",
			"s.score AS score",
		}, ", ")).
		Joins("JOIN challenges c ON c.id = s.challenge_id").
		Where("s.user_id = ? AND c.flag_type = ?", userID, challengecontracts.FlagTypeManualReview)
	if query.ChallengeID != nil {
		manualReviewQuery = manualReviewQuery.Where("s.challenge_id = ?", *query.ChallengeID)
	}
	if err := manualReviewQuery.Order("s.updated_at ASC").Scan(&manualReviewRows).Error; err != nil {
		return nil, fmt.Errorf("get student evidence manual review rows: %w", err)
	}
	for _, row := range manualReviewRows {
		events = append(events, toEvidenceRecord(evidence.NewManualReviewEvent(evidence.ManualReviewInput{
			UserID:       userID,
			ChallengeID:  row.ChallengeID,
			Title:        row.Title,
			Timestamp:    row.Timestamp,
			ReviewStatus: row.ReviewStatus,
			Score:        row.Score,
		})))
	}

	if r.db.Migrator().HasTable("submission_writeups") {
		writeupRows := make([]struct {
			ChallengeID      int64     `gorm:"column:challenge_id"`
			Title            string    `gorm:"column:title"`
			WriteupTitle     string    `gorm:"column:writeup_title"`
			Timestamp        time.Time `gorm:"column:timestamp"`
			SubmissionStatus string    `gorm:"column:submission_status"`
			VisibilityStatus string    `gorm:"column:visibility_status"`
			IsRecommended    bool      `gorm:"column:is_recommended"`
		}, 0)
		writeupQuery := r.db.WithContext(ctx).Table("submission_writeups AS sw").
			Select(strings.Join([]string{
				"sw.challenge_id AS challenge_id",
				"COALESCE(c.title, '') AS title",
				"sw.title AS writeup_title",
				"sw.updated_at AS timestamp",
				"sw.submission_status AS submission_status",
				"sw.visibility_status AS visibility_status",
				"sw.is_recommended AS is_recommended",
			}, ", ")).
			Joins("LEFT JOIN challenges c ON c.id = sw.challenge_id").
			Where("sw.user_id = ?", userID)
		if query.ChallengeID != nil {
			writeupQuery = writeupQuery.Where("sw.challenge_id = ?", *query.ChallengeID)
		}
		if err := writeupQuery.Order("sw.updated_at ASC").Scan(&writeupRows).Error; err != nil {
			return nil, fmt.Errorf("get student evidence writeup rows: %w", err)
		}
		for _, row := range writeupRows {
			events = append(events, toEvidenceRecord(evidence.NewWriteupEvent(evidence.WriteupInput{
				UserID:           userID,
				ChallengeID:      row.ChallengeID,
				Title:            row.Title,
				Timestamp:        row.Timestamp,
				WriteupTitle:     row.WriteupTitle,
				SubmissionStatus: row.SubmissionStatus,
				VisibilityStatus: row.VisibilityStatus,
				IsRecommended:    row.IsRecommended,
			})))
		}
	}

	if r.db.Migrator().HasTable("awd_attack_logs") && r.db.Migrator().HasTable("awd_rounds") {
		awdRows := make([]struct {
			ContestID         int64     `gorm:"column:contest_id"`
			RoundID           int64     `gorm:"column:round_id"`
			TeamID            int64     `gorm:"column:team_id"`
			VictimTeamID      int64     `gorm:"column:victim_team_id"`
			VictimTeamName    string    `gorm:"column:victim_team_name"`
			ServiceID         int64     `gorm:"column:service_id"`
			AWDChallengeID    int64     `gorm:"column:awd_challenge_id"`
			AWDChallengeTitle string    `gorm:"column:awd_challenge_title"`
			Timestamp         time.Time `gorm:"column:timestamp"`
			IsSuccess         bool      `gorm:"column:is_success"`
			ScoreGained       int       `gorm:"column:score_gained"`
			SubmittedByUserID *int64    `gorm:"column:submitted_by_user_id"`
			Source            string    `gorm:"column:source"`
		}, 0)
		awdQuery := r.db.WithContext(ctx).Table("awd_attack_logs AS al").
			Select(strings.Join([]string{
				"ar.contest_id AS contest_id",
				"al.round_id AS round_id",
				"al.attacker_team_id AS team_id",
				"al.victim_team_id AS victim_team_id",
				"COALESCE(vt.name, '') AS victim_team_name",
				"al.service_id AS service_id",
				"al.awd_challenge_id AS awd_challenge_id",
				"COALESCE(ac.name, '') AS awd_challenge_title",
				"al.created_at AS timestamp",
				"al.is_success AS is_success",
				"al.score_gained AS score_gained",
				"al.submitted_by_user_id AS submitted_by_user_id",
				"al.source AS source",
			}, ", ")).
			Joins("JOIN awd_rounds ar ON ar.id = al.round_id").
			Joins("LEFT JOIN awd_challenges ac ON ac.id = al.awd_challenge_id").
			Joins("LEFT JOIN teams vt ON vt.id = al.victim_team_id").
			Where("al.submitted_by_user_id = ? OR al.attacker_team_id IN (SELECT tm.team_id FROM team_members tm WHERE tm.user_id = ?)", userID, userID)
		if query.ChallengeID != nil {
			awdQuery = awdQuery.Where("al.awd_challenge_id = ?", *query.ChallengeID)
		}
		if err := awdQuery.Order("al.created_at ASC").Scan(&awdRows).Error; err != nil {
			return nil, fmt.Errorf("get student evidence awd attack rows: %w", err)
		}
		for _, row := range awdRows {
			teamID := row.TeamID
			contestID := row.ContestID
			roundID := row.RoundID
			serviceID := row.ServiceID
			victimTeamID := row.VictimTeamID
			scope := evidence.EventScopeTeam
			if row.SubmittedByUserID != nil && *row.SubmittedByUserID == userID {
				scope = evidence.EventScopeStudent
			}
			events = append(events, toEvidenceRecord(evidence.NewAWDAttackEvent(evidence.AWDAttackInput{
				UserID:            userID,
				TeamID:            &teamID,
				ContestID:         &contestID,
				RoundID:           &roundID,
				ServiceID:         &serviceID,
				VictimTeamID:      &victimTeamID,
				AWDChallengeID:    row.AWDChallengeID,
				AWDChallengeTitle: row.AWDChallengeTitle,
				VictimTeamName:    row.VictimTeamName,
				Timestamp:         row.Timestamp,
				IsSuccess:         row.IsSuccess,
				ScoreGained:       row.ScoreGained,
				Scope:             scope,
				AttackSource:      row.Source,
			})))
		}
	}

	if r.db.Migrator().HasTable("awd_traffic_events") {
		trafficRows := make([]struct {
			ContestID         int64     `gorm:"column:contest_id"`
			RoundID           int64     `gorm:"column:round_id"`
			TeamID            int64     `gorm:"column:team_id"`
			VictimTeamID      int64     `gorm:"column:victim_team_id"`
			VictimTeamName    string    `gorm:"column:victim_team_name"`
			ServiceID         int64     `gorm:"column:service_id"`
			AWDChallengeID    int64     `gorm:"column:awd_challenge_id"`
			AWDChallengeTitle string    `gorm:"column:awd_challenge_title"`
			Method            string    `gorm:"column:method"`
			Path              string    `gorm:"column:path"`
			StatusCode        int       `gorm:"column:status_code"`
			Timestamp         time.Time `gorm:"column:timestamp"`
		}, 0)
		trafficQuery := r.db.WithContext(ctx).Table("awd_traffic_events AS te").
			Select(strings.Join([]string{
				"te.contest_id AS contest_id",
				"te.round_id AS round_id",
				"te.attacker_team_id AS team_id",
				"te.victim_team_id AS victim_team_id",
				"COALESCE(vt.name, '') AS victim_team_name",
				"te.service_id AS service_id",
				"te.awd_challenge_id AS awd_challenge_id",
				"COALESCE(ac.name, '') AS awd_challenge_title",
				"te.method AS method",
				"te.path AS path",
				"te.status_code AS status_code",
				"te.created_at AS timestamp",
			}, ", ")).
			Joins("JOIN team_members tm ON tm.contest_id = te.contest_id AND tm.team_id = te.attacker_team_id").
			Joins("LEFT JOIN awd_challenges ac ON ac.id = te.awd_challenge_id").
			Joins("LEFT JOIN teams vt ON vt.id = te.victim_team_id").
			Where("tm.user_id = ?", userID)
		if query.ChallengeID != nil {
			trafficQuery = trafficQuery.Where("te.awd_challenge_id = ?", *query.ChallengeID)
		}
		if err := trafficQuery.Order("te.created_at ASC").Limit(500).Scan(&trafficRows).Error; err != nil {
			return nil, fmt.Errorf("get student evidence awd traffic rows: %w", err)
		}
		for _, row := range trafficRows {
			teamID := row.TeamID
			contestID := row.ContestID
			roundID := row.RoundID
			serviceID := row.ServiceID
			victimTeamID := row.VictimTeamID
			events = append(events, toEvidenceRecord(evidence.NewAWDTrafficEvent(evidence.AWDTrafficInput{
				UserID:            userID,
				TeamID:            &teamID,
				ContestID:         &contestID,
				RoundID:           &roundID,
				ServiceID:         &serviceID,
				VictimTeamID:      &victimTeamID,
				AWDChallengeID:    row.AWDChallengeID,
				AWDChallengeTitle: row.AWDChallengeTitle,
				VictimTeamName:    row.VictimTeamName,
				Method:            row.Method,
				Path:              row.Path,
				StatusCode:        row.StatusCode,
				Timestamp:         row.Timestamp,
			})))
		}
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	return events, nil
}

func (r *StudentActivityRepository) listStudentAuditTimelineRows(ctx context.Context, userID int64) ([]timelineEventRow, error) {
	rows := make([]timelineEventRow, 0)
	if err := r.db.WithContext(ctx).Raw(`
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
		auditlog.ActionRead,
		"challenge_detail",
		userID,
		auditlog.ActionUpdate,
		"instance",
		userID,
		auditlog.ActionRead,
		"instance_access",
	).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list student audit timeline rows: %w", err)
	}

	proxyRows := make([]struct {
		ChallengeID int64     `gorm:"column:challenge_id"`
		Title       string    `gorm:"column:title"`
		Timestamp   time.Time `gorm:"column:timestamp"`
		Detail      string    `gorm:"column:detail"`
	}, 0)
	if err := r.db.WithContext(ctx).
		Table("audit_logs AS a").
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
		return nil, fmt.Errorf("list student proxy timeline rows: %w", err)
	}
	for _, row := range proxyRows {
		rows = append(rows, timelineEventRow{
			Type:        "instance_proxy_request",
			ChallengeID: row.ChallengeID,
			Title:       row.Title,
			Timestamp:   row.Timestamp,
			Detail:      evidence.BuildProxyRequestDetail(row.Detail),
		})
	}
	return rows, nil
}

func toEvidenceRecord(event evidence.Event) queryports.EvidenceEventRecord {
	return queryports.EvidenceEventRecord{
		Type:         event.Type,
		Source:       event.Source,
		Stage:        event.Stage,
		UserID:       event.UserID,
		TeamID:       event.TeamID,
		ChallengeID:  event.ChallengeID,
		ContestID:    event.ContestID,
		RoundID:      event.RoundID,
		ServiceID:    event.ServiceID,
		VictimTeamID: event.VictimTeamID,
		Title:        event.Title,
		Timestamp:    event.Timestamp,
		Detail:       event.Detail,
		Meta:         event.Meta,
	}
}
