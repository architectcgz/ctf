package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ctf-platform/internal/model"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	teachingadvice "ctf-platform/internal/teaching/advice"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *Repository) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	var user model.User
	if err := r.dbWithContext(ctx).Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &user, nil
}

// Upsert 插入或更新能力画像
func (r *Repository) Upsert(ctx context.Context, profile *model.SkillProfile) error {
	return r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "dimension"}},
		DoUpdates: clause.AssignmentColumns([]string{"score", "updated_at"}),
	}).Create(profile).Error
}

// FindByUserID 查询用户所有维度画像
func (r *Repository) FindByUserID(ctx context.Context, userID int64) ([]*model.SkillProfile, error) {
	var profiles []*model.SkillProfile
	err := r.dbWithContext(ctx).Where("user_id = ?", userID).Find(&profiles).Error
	return profiles, err
}

func (r *Repository) ListSolvedChallengeIDs(ctx context.Context, userID int64) ([]int64, error) {
	var ids []int64
	err := r.dbWithContext(ctx).Raw(`
		SELECT DISTINCT s.challenge_id AS challenge_id
		FROM submissions s
		WHERE s.user_id = ?
			AND s.is_correct = TRUE
			AND s.contest_id IS NULL
		ORDER BY challenge_id ASC
	`, userID).Scan(&ids).Error
	return ids, err
}

func (r *Repository) GetStudentTeachingFactSnapshot(ctx context.Context, userID int64) (*teachingadvice.StudentFactSnapshot, error) {
	if userID <= 0 {
		return nil, nil
	}

	snapshot := &teachingadvice.StudentFactSnapshot{UserID: userID}
	if user, err := r.FindUserByID(ctx, userID); err != nil {
		return nil, err
	} else if user != nil {
		snapshot.Username = user.Username
		if trimmed := user.Name; trimmed != "" {
			name := trimmed
			snapshot.Name = &name
		}
	}

	since := time.Now().AddDate(0, 0, -7)
	if err := r.fillStudentRecentActivity(ctx, userID, since, snapshot); err != nil {
		return nil, err
	}
	if err := r.fillStudentSubmissionStats(ctx, userID, snapshot); err != nil {
		return nil, err
	}
	if err := r.fillStudentWriteupAndReviewStats(ctx, userID, snapshot); err != nil {
		return nil, err
	}
	if err := r.fillStudentHandsOnStats(ctx, userID, snapshot); err != nil {
		return nil, err
	}
	if err := r.fillStudentDimensionFacts(ctx, userID, snapshot); err != nil {
		return nil, err
	}

	return snapshot, nil
}

// BatchUpsert 批量插入或更新
func (r *Repository) BatchUpsert(ctx context.Context, profiles []*model.SkillProfile) error {
	if len(profiles) == 0 {
		return nil
	}
	return r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "dimension"}},
		DoUpdates: clause.AssignmentColumns([]string{"score", "updated_at"}),
	}).Create(profiles).Error
}

func (r *Repository) ListStudentIDs(ctx context.Context) ([]int64, error) {
	var ids []int64
	err := r.dbWithContext(ctx).Model(&model.User{}).
		Where("role = ? AND deleted_at IS NULL", model.RoleStudent).
		Pluck("id", &ids).Error
	return ids, err
}

// GetDimensionScores 查询用户各维度得分统计
func (r *Repository) GetDimensionScores(ctx context.Context, userID int64) ([]assessmentdomain.DimensionScore, error) {
	var scores []assessmentdomain.DimensionScore
	err := r.dbWithContext(ctx).Raw(`
		SELECT
			c.category AS dimension,
			COALESCE(SUM(c.points), 0) AS total_score,
			COALESCE(SUM(
				CASE WHEN EXISTS (
					SELECT 1
					FROM submissions s
					WHERE s.challenge_id = c.id
						AND s.user_id = ?
						AND s.is_correct = TRUE
						AND s.contest_id IS NULL
				) THEN c.points ELSE 0 END
			), 0) AS user_score
		FROM challenges c
		WHERE c.status = 'published'
		GROUP BY c.category
		ORDER BY c.category
	`, userID).Scan(&scores).Error
	return scores, err
}

// GetDimensionScore 查询用户单个维度得分统计（增量更新用）
func (r *Repository) GetDimensionScore(ctx context.Context, userID int64, dimension string) (*assessmentdomain.DimensionScore, error) {
	var score assessmentdomain.DimensionScore
	err := r.dbWithContext(ctx).Raw(`
		SELECT
			c.category AS dimension,
			COALESCE(SUM(c.points), 0) AS total_score,
			COALESCE(SUM(
				CASE WHEN EXISTS (
					SELECT 1
					FROM submissions s
					WHERE s.challenge_id = c.id
						AND s.user_id = ?
						AND s.is_correct = TRUE
						AND s.contest_id IS NULL
				) THEN c.points ELSE 0 END
			), 0) AS user_score
		FROM challenges c
		WHERE c.status = 'published' AND c.category = ?
		GROUP BY c.category
	`, userID, dimension).Scan(&score).Error
	if err != nil {
		return nil, err
	}
	return &score, nil
}

func (r *Repository) fillStudentRecentActivity(
	ctx context.Context,
	userID int64,
	since time.Time,
	snapshot *teachingadvice.StudentFactSnapshot,
) error {
	if snapshot == nil {
		return nil
	}

	type activityRow struct {
		RecentEventCount int            `gorm:"column:recent_event_count"`
		ActiveDays       int            `gorm:"column:active_days"`
		LastActivityAt   sql.NullString `gorm:"column:last_activity_at"`
	}

	row := activityRow{}
	unionParts := []string{
		"SELECT s.submitted_at AS event_at FROM submissions s WHERE s.user_id = ? AND s.submitted_at >= ?",
	}
	args := []any{userID, since}

	if r.db.Migrator().HasTable("audit_logs") {
		unionParts = append(unionParts, "SELECT a.created_at AS event_at FROM audit_logs a WHERE a.user_id = ? AND a.resource_type IN (?, ?) AND a.created_at >= ?")
		args = append(args, userID, "instance_access", "instance_proxy_request", since)
	}
	if r.db.Migrator().HasTable("submission_writeups") {
		unionParts = append(unionParts, "SELECT sw.updated_at AS event_at FROM submission_writeups sw WHERE sw.user_id = ? AND sw.updated_at >= ?")
		args = append(args, userID, since)
	}
	if r.db.Migrator().HasTable("awd_attack_logs") {
		unionParts = append(unionParts, "SELECT al.created_at AS event_at FROM awd_attack_logs al WHERE al.submitted_by_user_id = ? AND al.created_at >= ?")
		args = append(args, userID, since)
	}

	query := fmt.Sprintf(`
		SELECT
			COUNT(*) AS recent_event_count,
			COUNT(DISTINCT DATE(event_at)) AS active_days,
			MAX(event_at) AS last_activity_at
		FROM (
			%s
		) events
	`, stringsJoin(unionParts, " UNION ALL "))

	if err := r.dbWithContext(ctx).Raw(query, args...).Scan(&row).Error; err != nil {
		return fmt.Errorf("get student recent activity: %w", err)
	}

	snapshot.RecentEventCount7d = row.RecentEventCount
	snapshot.ActiveDays7d = row.ActiveDays
	if row.LastActivityAt.Valid {
		snapshot.LastActivityAt = parseAggregateTime(row.LastActivityAt.String)
	}
	return nil
}

func (r *Repository) fillStudentSubmissionStats(
	ctx context.Context,
	userID int64,
	snapshot *teachingadvice.StudentFactSnapshot,
) error {
	if snapshot == nil {
		return nil
	}

	type summaryRow struct {
		CorrectSubmissionCount int `gorm:"column:correct_submission_count"`
		WrongSubmissionCount   int `gorm:"column:wrong_submission_count"`
	}

	row := summaryRow{}
	if err := r.dbWithContext(ctx).Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN s.is_correct THEN 1 ELSE 0 END), 0) AS correct_submission_count,
			COALESCE(SUM(CASE WHEN s.is_correct THEN 0 ELSE 1 END), 0) AS wrong_submission_count
		FROM submissions s
		WHERE s.user_id = ? AND s.contest_id IS NULL
	`, userID).Scan(&row).Error; err != nil {
		return fmt.Errorf("get student submission stats: %w", err)
	}

	type resultRow struct {
		ID        int64 `gorm:"column:id"`
		IsCorrect bool  `gorm:"column:is_correct"`
	}
	results := make([]resultRow, 0)
	if err := r.dbWithContext(ctx).Raw(`
		SELECT s.id, s.is_correct
		FROM submissions s
		WHERE s.user_id = ? AND s.contest_id IS NULL
		ORDER BY s.submitted_at ASC, s.id ASC
	`, userID).Scan(&results).Error; err != nil {
		return fmt.Errorf("list student submission results: %w", err)
	}

	maxWrongStreak := 0
	currentWrongStreak := 0
	for _, result := range results {
		if result.IsCorrect {
			currentWrongStreak = 0
			continue
		}
		currentWrongStreak++
		if currentWrongStreak > maxWrongStreak {
			maxWrongStreak = currentWrongStreak
		}
	}

	snapshot.CorrectSubmissionCount = row.CorrectSubmissionCount
	snapshot.WrongSubmissionCount = row.WrongSubmissionCount
	snapshot.MaxWrongStreak = maxWrongStreak
	return nil
}

func (r *Repository) fillStudentWriteupAndReviewStats(
	ctx context.Context,
	userID int64,
	snapshot *teachingadvice.StudentFactSnapshot,
) error {
	if snapshot == nil {
		return nil
	}

	if r.db.Migrator().HasTable("submission_writeups") {
		var writeupCount int64
		if err := r.dbWithContext(ctx).Table("submission_writeups").
			Where("user_id = ?", userID).
			Count(&writeupCount).Error; err != nil {
			return fmt.Errorf("count student writeups: %w", err)
		}
		snapshot.WriteupCount = int(writeupCount)
	}

	var approvedReviewCount int64
	if err := r.dbWithContext(ctx).Table("submissions").
		Where("user_id = ? AND review_status = ?", userID, model.SubmissionReviewStatusApproved).
		Count(&approvedReviewCount).Error; err != nil {
		return fmt.Errorf("count approved manual reviews: %w", err)
	}
	snapshot.ApprovedReviewCount = int(approvedReviewCount)
	return nil
}

func (r *Repository) fillStudentHandsOnStats(
	ctx context.Context,
	userID int64,
	snapshot *teachingadvice.StudentFactSnapshot,
) error {
	if snapshot == nil {
		return nil
	}

	if r.db.Migrator().HasTable("audit_logs") {
		var handsOnCount int64
		if err := r.dbWithContext(ctx).Table("audit_logs").
			Where("user_id = ? AND resource_type IN (?, ?)", userID, "instance_access", "instance_proxy_request").
			Count(&handsOnCount).Error; err != nil {
			return fmt.Errorf("count hands-on events: %w", err)
		}
		snapshot.HandsOnEventCount = int(handsOnCount)
	}

	if r.db.Migrator().HasTable("awd_attack_logs") {
		var awdSuccessCount int64
		if err := r.dbWithContext(ctx).Table("awd_attack_logs").
			Where("submitted_by_user_id = ? AND is_success = ?", userID, true).
			Count(&awdSuccessCount).Error; err != nil {
			return fmt.Errorf("count awd success events: %w", err)
		}
		snapshot.AWDSuccessCount = int(awdSuccessCount)
	}
	return nil
}

func (r *Repository) fillStudentDimensionFacts(
	ctx context.Context,
	userID int64,
	snapshot *teachingadvice.StudentFactSnapshot,
) error {
	if snapshot == nil {
		return nil
	}

	factMap := make(map[string]*teachingadvice.DimensionFact, len(model.AllDimensions))
	for _, dimension := range model.AllDimensions {
		dimensionCopy := dimension
		factMap[dimension] = &teachingadvice.DimensionFact{Dimension: dimensionCopy}
	}

	profiles, err := r.FindByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find student skill profiles: %w", err)
	}
	for _, profile := range profiles {
		if profile == nil {
			continue
		}
		fact := ensureDimensionFact(factMap, profile.Dimension)
		fact.ProfileScore = profile.Score
	}

	if r.db.Migrator().HasTable("challenges") {
		type attemptRow struct {
			Dimension    string `gorm:"column:dimension"`
			AttemptCount int    `gorm:"column:attempt_count"`
			SuccessCount int    `gorm:"column:success_count"`
		}
		attemptRows := make([]attemptRow, 0)
		if err := r.dbWithContext(ctx).Raw(`
			SELECT
				c.category AS dimension,
				COUNT(*) AS attempt_count,
				COALESCE(SUM(CASE WHEN s.is_correct THEN 1 ELSE 0 END), 0) AS success_count
			FROM submissions s
			JOIN challenges c ON c.id = s.challenge_id
			WHERE s.user_id = ? AND s.contest_id IS NULL AND c.status = ?
			GROUP BY c.category
		`, userID, model.ChallengeStatusPublished).Scan(&attemptRows).Error; err != nil {
			return fmt.Errorf("get student dimension attempt facts: %w", err)
		}
		for _, row := range attemptRows {
			fact := ensureDimensionFact(factMap, row.Dimension)
			fact.AttemptCount = row.AttemptCount
			fact.SuccessCount = row.SuccessCount
			fact.EvidenceCount += row.AttemptCount
		}

		if r.db.Migrator().HasTable("audit_logs") && r.db.Migrator().HasTable("instances") {
			type evidenceRow struct {
				Dimension string `gorm:"column:dimension"`
				Count     int    `gorm:"column:count"`
			}
			auditRows := make([]evidenceRow, 0)
			if err := r.dbWithContext(ctx).Raw(`
				SELECT
					c.category AS dimension,
					COUNT(*) AS count
				FROM audit_logs a
				JOIN instances i ON i.id = a.resource_id
				JOIN challenges c ON c.id = i.challenge_id
				WHERE a.user_id = ?
					AND a.resource_type IN (?, ?)
					AND c.status = ?
				GROUP BY c.category
			`, userID, "instance_access", "instance_proxy_request", model.ChallengeStatusPublished).Scan(&auditRows).Error; err != nil {
				return fmt.Errorf("get student audit evidence facts: %w", err)
			}
			for _, row := range auditRows {
				fact := ensureDimensionFact(factMap, row.Dimension)
				fact.EvidenceCount += row.Count
			}
		}

		if r.db.Migrator().HasTable("submission_writeups") {
			type evidenceRow struct {
				Dimension string `gorm:"column:dimension"`
				Count     int    `gorm:"column:count"`
			}
			writeupRows := make([]evidenceRow, 0)
			if err := r.dbWithContext(ctx).Raw(`
				SELECT
					c.category AS dimension,
					COUNT(*) AS count
				FROM submission_writeups sw
				JOIN challenges c ON c.id = sw.challenge_id
				WHERE sw.user_id = ? AND c.status = ?
				GROUP BY c.category
			`, userID, model.ChallengeStatusPublished).Scan(&writeupRows).Error; err != nil {
				return fmt.Errorf("get student writeup evidence facts: %w", err)
			}
			for _, row := range writeupRows {
				fact := ensureDimensionFact(factMap, row.Dimension)
				fact.EvidenceCount += row.Count
			}
		}

		type evidenceRow struct {
			Dimension string `gorm:"column:dimension"`
			Count     int    `gorm:"column:count"`
		}
		reviewRows := make([]evidenceRow, 0)
		if err := r.dbWithContext(ctx).Raw(`
			SELECT
				c.category AS dimension,
				COUNT(*) AS count
			FROM submissions s
			JOIN challenges c ON c.id = s.challenge_id
			WHERE s.user_id = ?
				AND s.contest_id IS NULL
				AND s.review_status = ?
				AND c.status = ?
			GROUP BY c.category
		`, userID, model.SubmissionReviewStatusApproved, model.ChallengeStatusPublished).Scan(&reviewRows).Error; err != nil {
			return fmt.Errorf("get student review evidence facts: %w", err)
		}
		for _, row := range reviewRows {
			fact := ensureDimensionFact(factMap, row.Dimension)
			fact.EvidenceCount += row.Count
		}
	}

	dimensions := make([]teachingadvice.DimensionFact, 0, len(factMap))
	for _, dimension := range model.AllDimensions {
		fact := ensureDimensionFact(factMap, dimension)
		dimensions = append(dimensions, *fact)
	}
	snapshot.Dimensions = dimensions
	return nil
}

func ensureDimensionFact(
	facts map[string]*teachingadvice.DimensionFact,
	dimension string,
) *teachingadvice.DimensionFact {
	if fact, ok := facts[dimension]; ok {
		return fact
	}
	fact := &teachingadvice.DimensionFact{Dimension: dimension}
	facts[dimension] = fact
	return fact
}

func stringsJoin(items []string, sep string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	default:
		result := items[0]
		for _, item := range items[1:] {
			result += sep + item
		}
		return result
	}
}
