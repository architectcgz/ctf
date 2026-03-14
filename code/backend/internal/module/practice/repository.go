package practice

import (
	"ctf-platform/internal/model"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
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

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateSubmission 创建提交记录
func (r *Repository) CreateSubmission(submission *model.Submission) error {
	return r.db.Create(submission).Error
}

// FindCorrectSubmission 查找用户是否已正确提交过该题
func (r *Repository) FindCorrectSubmission(userID, challengeID int64) (*model.Submission, error) {
	var submission model.Submission
	err := r.db.Where("user_id = ? AND challenge_id = ? AND is_correct = ?", userID, challengeID, true).
		First(&submission).Error
	return &submission, err
}

// CountRecentSubmissions 统计时间窗口内的提交次数
func (r *Repository) CountRecentSubmissions(userID, challengeID int64, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&model.Submission{}).
		Where("user_id = ? AND challenge_id = ? AND submitted_at >= ?", userID, challengeID, since).
		Count(&count).Error
	return count, err
}

// IsUniqueViolation 检测是否为唯一约束冲突错误
func (r *Repository) IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// 23505 是 PostgreSQL 唯一约束冲突错误码
		return pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, "idx_submissions_user_challenge_correct")
	}
	return false
}

// GetUserProgress 获取用户解题进度统计
func (r *Repository) GetUserProgress(userID int64) (totalScore int, totalSolved int, err error) {
	var result struct {
		TotalScore  int `gorm:"column:total_score"`
		TotalSolved int `gorm:"column:total_solved"`
	}
	err = r.db.Table("submissions s").
		Select("COALESCE(SUM(c.points), 0) as total_score, COUNT(DISTINCT s.challenge_id) as total_solved").
		Joins("JOIN challenges c ON s.challenge_id = c.id").
		Where("s.user_id = ? AND s.is_correct = ? AND c.status = ?", userID, true, "published").
		Scan(&result).Error
	return result.TotalScore, result.TotalSolved, err
}

// GetCategoryStats 获取分类统计
func (r *Repository) GetCategoryStats(userID int64) ([]struct {
	Category string
	Solved   int
	Total    int
}, error) {
	var stats []struct {
		Category string
		Solved   int
		Total    int
	}
	err := r.db.Raw(`
		SELECT
			c.category,
			COUNT(DISTINCT CASE WHEN s.is_correct THEN c.id END) as solved,
			COUNT(DISTINCT c.id) as total
		FROM challenges c
		LEFT JOIN submissions s ON c.id = s.challenge_id AND s.user_id = ? AND s.is_correct = true
		WHERE c.status = ?
		GROUP BY c.category
		ORDER BY c.category
	`, userID, "published").Scan(&stats).Error
	return stats, err
}

// GetDifficultyStats 获取难度统计
func (r *Repository) GetDifficultyStats(userID int64) ([]struct {
	Difficulty string
	Solved     int
	Total      int
}, error) {
	var stats []struct {
		Difficulty string
		Solved     int
		Total      int
	}
	err := r.db.Raw(`
		SELECT
			c.difficulty,
			COUNT(DISTINCT CASE WHEN s.is_correct THEN c.id END) as solved,
			COUNT(DISTINCT c.id) as total
		FROM challenges c
		LEFT JOIN submissions s ON c.id = s.challenge_id AND s.user_id = ? AND s.is_correct = true
		WHERE c.status = ?
		GROUP BY c.difficulty
		ORDER BY
			CASE c.difficulty
				WHEN 'beginner' THEN 1
				WHEN 'easy' THEN 2
				WHEN 'medium' THEN 3
				WHEN 'hard' THEN 4
				WHEN 'insane' THEN 5
			END
	`, userID, "published").Scan(&stats).Error
	return stats, err
}

// GetUserRank 获取用户排名
func (r *Repository) GetUserRank(userID int64) (int, error) {
	var rank int
	err := r.db.Raw(`
		WITH ranked_users AS (
			SELECT
				s.user_id,
				RANK() OVER (ORDER BY SUM(c.points) DESC) as rank
			FROM submissions s
			JOIN challenges c ON s.challenge_id = c.id
			WHERE s.is_correct = true AND c.status = ?
			GROUP BY s.user_id
		)
		SELECT COALESCE(rank, (SELECT COUNT(DISTINCT user_id) + 1 FROM ranked_users))
		FROM ranked_users WHERE user_id = ?
	`, "published", userID).Scan(&rank).Error
	if err != nil {
		return 0, err
	}
	return rank, nil
}

// GetUserTimeline 获取用户时间线
func (r *Repository) GetUserTimeline(userID int64, limit, offset int) ([]struct {
	Type        string
	ChallengeID int64
	Title       string
	Timestamp   time.Time
	IsCorrect   *bool
	Points      *int
	Detail      string
}, error) {
	events := make([]timelineEventRow, 0)

	if limit <= 0 {
		limit = 100
	}

	err := r.db.Raw(`
		SELECT events.*, c.title FROM (
			SELECT 'instance_start' as type, i.challenge_id, i.created_at as timestamp,
				NULL as is_correct, NULL as points, '启动练习实例' AS detail
			FROM instances i
			WHERE i.user_id = ?
			UNION ALL
			SELECT
				'flag_submit' as type,
				s.challenge_id,
				s.submitted_at as timestamp,
				s.is_correct,
				CASE WHEN s.is_correct THEN c.points ELSE NULL END as points,
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
			SELECT 'instance_destroy' as type, i.challenge_id, i.updated_at as timestamp,
				NULL as is_correct, NULL as points, '结束练习实例' AS detail
			FROM instances i
			WHERE i.user_id = ? AND i.status IN ('stopped', 'expired')
		) events
		LEFT JOIN challenges c ON events.challenge_id = c.id
		ORDER BY events.timestamp DESC
	`, userID, userID, userID, userID).Scan(&events).Error
	if err != nil {
		return nil, err
	}

	auditEvents, err := r.listUserAuditTimelineEvents(userID)
	if err != nil {
		return nil, err
	}
	events = append(events, auditEvents...)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})

	if offset >= len(events) {
		return []struct {
			Type        string
			ChallengeID int64
			Title       string
			Timestamp   time.Time
			IsCorrect   *bool
			Points      *int
			Detail      string
		}{}, nil
	}

	end := offset + limit
	if end > len(events) {
		end = len(events)
	}

	result := make([]struct {
		Type        string
		ChallengeID int64
		Title       string
		Timestamp   time.Time
		IsCorrect   *bool
		Points      *int
		Detail      string
	}, 0, end-offset)
	for _, event := range events[offset:end] {
		result = append(result, struct {
			Type        string
			ChallengeID int64
			Title       string
			Timestamp   time.Time
			IsCorrect   *bool
			Points      *int
			Detail      string
		}{
			Type:        event.Type,
			ChallengeID: event.ChallengeID,
			Title:       event.Title,
			Timestamp:   event.Timestamp,
			IsCorrect:   event.IsCorrect,
			Points:      event.Points,
			Detail:      event.Detail,
		})
	}
	return result, nil
}

func (r *Repository) listUserAuditTimelineEvents(userID int64) ([]timelineEventRow, error) {
	rows := make([]timelineEventRow, 0)

	err := r.db.Raw(`
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
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	proxyRows := make([]struct {
		ChallengeID int64     `gorm:"column:challenge_id"`
		Title       string    `gorm:"column:title"`
		Timestamp   time.Time `gorm:"column:timestamp"`
		Detail      string    `gorm:"column:detail"`
	}, 0)
	if err := r.db.Table("audit_logs AS a").
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
		return nil, err
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

	return rows, err
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
