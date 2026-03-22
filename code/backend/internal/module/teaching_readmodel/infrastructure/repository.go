package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type ProgressRow struct {
	Key    string
	Total  int
	Solved int
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &user, nil
}

func (r *Repository) CountStudentsByClass(ctx context.Context, className string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("class_name = ? AND role = ? AND deleted_at IS NULL", className, model.RoleStudent).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count students by class: %w", err)
	}
	return count, nil
}

func (r *Repository) ListClasses(ctx context.Context) ([]dto.TeacherClassItem, error) {
	items := make([]dto.TeacherClassItem, 0)
	if err := r.db.WithContext(ctx).Model(&model.User{}).
		Select("class_name AS name, COUNT(*) AS student_count").
		Where("role = ? AND class_name <> '' AND deleted_at IS NULL", model.RoleStudent).
		Group("class_name").
		Order("class_name ASC").
		Scan(&items).Error; err != nil {
		return nil, fmt.Errorf("list classes: %w", err)
	}
	return items, nil
}

func (r *Repository) ListStudentsByClass(ctx context.Context, className, keyword, studentNo string, since time.Time) ([]dto.TeacherStudentItem, error) {
	items := make([]dto.TeacherStudentItem, 0)
	query := r.db.WithContext(ctx).Table("users AS u").
		Select(`
			u.id,
			u.username,
			NULLIF(u.name, '') AS name,
			NULLIF(u.student_no, '') AS student_no,
			COALESCE((
				SELECT COUNT(DISTINCT s.challenge_id)
				FROM submissions s
				JOIN challenges c ON c.id = s.challenge_id
				WHERE s.user_id = u.id AND s.is_correct = TRUE AND c.status = ?
			), 0) AS solved_count,
			COALESCE((
				SELECT SUM(c.points)
				FROM submissions s
				JOIN challenges c ON c.id = s.challenge_id
				WHERE s.user_id = u.id AND s.is_correct = TRUE AND c.status = ?
			), 0) AS total_score,
			COALESCE((
				SELECT COUNT(*)
				FROM (
					SELECT s.id
					FROM submissions s
					WHERE s.user_id = u.id AND s.submitted_at >= ?
					UNION ALL
					SELECT hu.id
					FROM challenge_hint_unlocks hu
					WHERE hu.user_id = u.id AND hu.unlocked_at >= ?
					UNION ALL
					SELECT i.id
					FROM instances i
					WHERE i.user_id = u.id AND i.created_at >= ?
					UNION ALL
					SELECT i.id
					FROM instances i
					WHERE i.user_id = u.id AND i.status IN ('stopped', 'expired') AND i.updated_at >= ?
				) recent_events
			), 0) AS recent_event_count,
			(
				SELECT sp.dimension
				FROM skill_profiles sp
				WHERE sp.user_id = u.id
				ORDER BY sp.score ASC, sp.updated_at DESC
				LIMIT 1
			) AS weak_dimension
		`, model.ChallengeStatusPublished, model.ChallengeStatusPublished, since, since, since, since).
		Where("u.role = ? AND u.class_name = ? AND u.deleted_at IS NULL", model.RoleStudent, className)
	if keyword != "" {
		likeKeyword := "%" + strings.ToLower(keyword) + "%"
		query = query.Where("(LOWER(u.username) LIKE ? OR LOWER(u.name) LIKE ?)", likeKeyword, likeKeyword)
	}
	if studentNo != "" {
		query = query.Where("u.student_no = ?", studentNo)
	}
	if err := query.Order("solved_count DESC, total_score DESC, COALESCE(NULLIF(u.student_no, ''), u.username) ASC, u.username ASC").
		Scan(&items).Error; err != nil {
		return nil, fmt.Errorf("list students by class: %w", err)
	}
	return items, nil
}

func (r *Repository) CountPublishedChallenges(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Challenge{}).
		Where("status = ?", model.ChallengeStatusPublished).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count published challenges: %w", err)
	}
	return count, nil
}

func (r *Repository) CountSolvedChallenges(ctx context.Context, userID int64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Table("submissions AS s").
		Joins("JOIN challenges c ON c.id = s.challenge_id").
		Where("s.user_id = ? AND s.is_correct = ? AND c.status = ?", userID, true, model.ChallengeStatusPublished).
		Distinct("s.challenge_id").
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count solved challenges: %w", err)
	}
	return count, nil
}

func (r *Repository) GetCategoryProgress(ctx context.Context, userID int64) ([]ProgressRow, error) {
	rows := make([]ProgressRow, 0)
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			c.category AS key,
			COUNT(DISTINCT c.id) AS total,
			COUNT(DISTINCT CASE WHEN s.is_correct THEN c.id END) AS solved
		FROM challenges c
		LEFT JOIN submissions s
			ON s.challenge_id = c.id
			AND s.user_id = ?
			AND s.is_correct = TRUE
		WHERE c.status = ?
		GROUP BY c.category
		ORDER BY c.category
	`, userID, model.ChallengeStatusPublished).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("get category progress: %w", err)
	}
	return rows, nil
}

func (r *Repository) GetDifficultyProgress(ctx context.Context, userID int64) ([]ProgressRow, error) {
	rows := make([]ProgressRow, 0)
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			c.difficulty AS key,
			COUNT(DISTINCT c.id) AS total,
			COUNT(DISTINCT CASE WHEN s.is_correct THEN c.id END) AS solved
		FROM challenges c
		LEFT JOIN submissions s
			ON s.challenge_id = c.id
			AND s.user_id = ?
			AND s.is_correct = TRUE
		WHERE c.status = ?
		GROUP BY c.difficulty
		ORDER BY
			CASE c.difficulty
				WHEN 'beginner' THEN 1
				WHEN 'easy' THEN 2
				WHEN 'medium' THEN 3
				WHEN 'hard' THEN 4
				WHEN 'hell' THEN 5
				ELSE 99
			END
	`, userID, model.ChallengeStatusPublished).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("get difficulty progress: %w", err)
	}
	return rows, nil
}

func (r *Repository) GetStudentTimeline(ctx context.Context, userID int64, limit, offset int) ([]dto.TimelineEvent, error) {
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
				'hint_unlock' AS type,
				hu.challenge_id,
				hu.unlocked_at AS timestamp,
				NULL AS is_correct,
				NULL AS points,
				CASE
					WHEN COALESCE(NULLIF(h.title, ''), '') <> '' THEN '解锁第 ' || CAST(h.level AS TEXT) || ' 级提示：' || h.title
					ELSE '解锁第 ' || CAST(h.level AS TEXT) || ' 级提示'
				END AS detail
			FROM challenge_hint_unlocks hu
			JOIN challenge_hints h ON h.id = hu.challenge_hint_id
			WHERE hu.user_id = ?
			UNION ALL
			SELECT
				'instance_destroy' AS type,
				i.challenge_id,
				i.updated_at AS timestamp,
				NULL AS is_correct,
				NULL AS points,
				'结束练习实例' AS detail
			FROM instances i
			WHERE i.user_id = ? AND i.status IN ('stopped', 'expired')
		) events
		LEFT JOIN challenges c ON events.challenge_id = c.id
		ORDER BY events.timestamp DESC
		LIMIT ? OFFSET ?
	`, userID, userID, userID, userID, limit, offset).Scan(&rows).Error; err != nil {
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
		return []dto.TimelineEvent{}, nil
	}
	end := offset + limit
	if end > len(rows) {
		end = len(rows)
	}
	rows = rows[offset:end]

	events := make([]dto.TimelineEvent, len(rows))
	for i, row := range rows {
		events[i] = dto.TimelineEvent{
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

func (r *Repository) listStudentAuditTimelineRows(ctx context.Context, userID int64) ([]timelineEventRow, error) {
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
		model.AuditActionRead,
		"challenge_detail",
		userID,
		model.AuditActionUpdate,
		"instance",
		userID,
		model.AuditActionRead,
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
			Detail:      buildTeacherProxyTimelineDetail(row.Detail),
		})
	}
	return rows, nil
}

func (r *Repository) GetClassSummary(ctx context.Context, className string, since time.Time) (*dto.TeacherClassSummaryResp, error) {
	studentCount, err := r.CountStudentsByClass(ctx, className)
	if err != nil {
		return nil, err
	}

	summary := &dto.TeacherClassSummaryResp{
		ClassName:    className,
		StudentCount: studentCount,
	}
	if studentCount == 0 {
		return summary, nil
	}

	averageSolved, err := r.getAverageSolvedByClass(ctx, className)
	if err != nil {
		return nil, err
	}
	activeStudentCount, err := r.getActiveStudentCountByClass(ctx, className, since)
	if err != nil {
		return nil, err
	}
	recentEventCount, err := r.getRecentEventCountByClass(ctx, className, since)
	if err != nil {
		return nil, err
	}

	summary.AverageSolved = averageSolved
	summary.ActiveStudentCount = activeStudentCount
	summary.ActiveRate = float64(activeStudentCount) * 100 / float64(studentCount)
	summary.RecentEventCount = recentEventCount
	return summary, nil
}

func (r *Repository) GetClassTrend(ctx context.Context, className string, since time.Time, days int) (*dto.TeacherClassTrendResp, error) {
	if days <= 0 {
		days = 7
	}

	type eventRow struct {
		UserID     int64     `gorm:"column:user_id"`
		OccurredAt time.Time `gorm:"column:occurred_at"`
		IsSolve    bool      `gorm:"column:is_solve"`
	}

	rows := make([]eventRow, 0)
	if err := r.db.WithContext(ctx).Raw(`
		SELECT s.user_id, s.submitted_at AS occurred_at, s.is_correct AS is_solve
		FROM submissions s
		JOIN users u ON u.id = s.user_id
		WHERE u.role = ? AND u.class_name = ? AND u.deleted_at IS NULL AND s.submitted_at >= ?
		UNION ALL
		SELECT hu.user_id, hu.unlocked_at AS occurred_at, FALSE AS is_solve
		FROM challenge_hint_unlocks hu
		JOIN users u ON u.id = hu.user_id
		WHERE u.role = ? AND u.class_name = ? AND u.deleted_at IS NULL AND hu.unlocked_at >= ?
		UNION ALL
		SELECT i.user_id, i.created_at AS occurred_at, FALSE AS is_solve
		FROM instances i
		JOIN users u ON u.id = i.user_id
		WHERE u.role = ? AND u.class_name = ? AND u.deleted_at IS NULL AND i.created_at >= ?
		UNION ALL
		SELECT i.user_id, i.updated_at AS occurred_at, FALSE AS is_solve
		FROM instances i
		JOIN users u ON u.id = i.user_id
		WHERE u.role = ? AND u.class_name = ? AND u.deleted_at IS NULL
			AND i.status IN ('stopped', 'expired') AND i.updated_at >= ?
	`, model.RoleStudent, className, since, model.RoleStudent, className, since, model.RoleStudent, className, since, model.RoleStudent, className, since).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("get class trend: %w", err)
	}

	points := make([]dto.TeacherClassTrendPoint, days)
	indexByDate := make(map[string]int, days)
	for i := 0; i < days; i++ {
		date := since.AddDate(0, 0, i).Format("2006-01-02")
		points[i] = dto.TeacherClassTrendPoint{Date: date}
		indexByDate[date] = i
	}

	activeUsersByDate := make(map[string]map[int64]struct{}, days)
	for _, row := range rows {
		date := row.OccurredAt.Format("2006-01-02")
		idx, ok := indexByDate[date]
		if !ok {
			continue
		}
		points[idx].EventCount++
		if row.IsSolve {
			points[idx].SolveCount++
		}
		users := activeUsersByDate[date]
		if users == nil {
			users = make(map[int64]struct{})
			activeUsersByDate[date] = users
		}
		users[row.UserID] = struct{}{}
	}

	for i := range points {
		points[i].ActiveStudentCount = int64(len(activeUsersByDate[points[i].Date]))
	}

	return &dto.TeacherClassTrendResp{
		ClassName: className,
		Points:    points,
	}, nil
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

func (r *Repository) getAverageSolvedByClass(ctx context.Context, className string) (float64, error) {
	var result struct {
		AverageSolved float64 `gorm:"column:average_solved"`
	}
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(AVG(student_solved.solved_count), 0) AS average_solved
		FROM (
			SELECT
				u.id,
				COUNT(DISTINCT CASE WHEN s.is_correct = TRUE AND c.status = ? THEN s.challenge_id END) AS solved_count
			FROM users u
			LEFT JOIN submissions s ON s.user_id = u.id
			LEFT JOIN challenges c ON c.id = s.challenge_id
			WHERE u.role = ? AND u.class_name = ? AND u.deleted_at IS NULL
			GROUP BY u.id
		) student_solved
	`, model.ChallengeStatusPublished, model.RoleStudent, className).Scan(&result).Error; err != nil {
		return 0, fmt.Errorf("get average solved by class: %w", err)
	}
	return result.AverageSolved, nil
}

func (r *Repository) getActiveStudentCountByClass(ctx context.Context, className string, since time.Time) (int64, error) {
	var result struct {
		Count int64 `gorm:"column:count"`
	}
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT active.user_id) AS count
		FROM (
			SELECT s.user_id
			FROM submissions s
			JOIN users u ON u.id = s.user_id
			WHERE u.role = ? AND u.class_name = ? AND u.deleted_at IS NULL AND s.submitted_at >= ?
			UNION
			SELECT hu.user_id
			FROM challenge_hint_unlocks hu
			JOIN users u ON u.id = hu.user_id
			WHERE u.role = ? AND u.class_name = ? AND u.deleted_at IS NULL AND hu.unlocked_at >= ?
			UNION
			SELECT i.user_id
			FROM instances i
			JOIN users u ON u.id = i.user_id
			WHERE u.role = ? AND u.class_name = ? AND u.deleted_at IS NULL
				AND (i.created_at >= ? OR i.updated_at >= ?)
		) active
	`, model.RoleStudent, className, since, model.RoleStudent, className, since, model.RoleStudent, className, since, since).Scan(&result).Error; err != nil {
		return 0, fmt.Errorf("get active student count by class: %w", err)
	}
	return result.Count, nil
}

func (r *Repository) getRecentEventCountByClass(ctx context.Context, className string, since time.Time) (int64, error) {
	var result struct {
		Count int64 `gorm:"column:count"`
	}
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) AS count
		FROM (
			SELECT s.id
			FROM submissions s
			JOIN users u ON u.id = s.user_id
			WHERE u.role = ? AND u.class_name = ? AND u.deleted_at IS NULL AND s.submitted_at >= ?
			UNION ALL
			SELECT hu.id
			FROM challenge_hint_unlocks hu
			JOIN users u ON u.id = hu.user_id
			WHERE u.role = ? AND u.class_name = ? AND u.deleted_at IS NULL AND hu.unlocked_at >= ?
			UNION ALL
			SELECT i.id
			FROM instances i
			JOIN users u ON u.id = i.user_id
			WHERE u.role = ? AND u.class_name = ? AND u.deleted_at IS NULL AND i.created_at >= ?
			UNION ALL
			SELECT i.id
			FROM instances i
			JOIN users u ON u.id = i.user_id
			WHERE u.role = ? AND u.class_name = ? AND u.deleted_at IS NULL
				AND i.status IN ('stopped', 'expired') AND i.updated_at >= ?
		) recent_events
	`, model.RoleStudent, className, since, model.RoleStudent, className, since, model.RoleStudent, className, since, model.RoleStudent, className, since).Scan(&result).Error; err != nil {
		return 0, fmt.Errorf("get recent event count by class: %w", err)
	}
	return result.Count, nil
}

func buildTeacherProxyTimelineDetail(rawDetail string) string {
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
