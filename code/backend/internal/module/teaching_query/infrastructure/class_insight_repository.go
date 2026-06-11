package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	queryports "ctf-platform/internal/module/teaching_query/ports"
	"ctf-platform/internal/shared/taxonomy"
	teachingadvice "ctf-platform/internal/teaching/advice"
)

type ClassInsightRepository struct {
	db *gorm.DB
}

func NewClassInsightRepository(db *gorm.DB) *ClassInsightRepository {
	return &ClassInsightRepository{db: db}
}

func (r *ClassInsightRepository) ListClassTeachingFactSnapshots(
	ctx context.Context,
	className string,
	since time.Time,
) ([]teachingadvice.StudentFactSnapshot, error) {
	studentRows := make([]struct {
		ID       int64   `gorm:"column:id"`
		Username string  `gorm:"column:username"`
		Name     *string `gorm:"column:name"`
	}, 0)
	if err := r.db.WithContext(ctx).Table("users AS u").
		Select("u.id, u.username, NULLIF(u.name, '') AS name").
		Where("u.class_name = ? AND u.role = ? AND u.deleted_at IS NULL", className, identitycontracts.RoleStudent).
		Order("u.username ASC").
		Scan(&studentRows).Error; err != nil {
		return nil, fmt.Errorf("list class teaching fact students: %w", err)
	}
	if len(studentRows) == 0 {
		return []teachingadvice.StudentFactSnapshot{}, nil
	}

	userIDs := make([]int64, 0, len(studentRows))
	snapshotByID := make(map[int64]*teachingadvice.StudentFactSnapshot, len(studentRows))
	for _, row := range studentRows {
		snapshot := &teachingadvice.StudentFactSnapshot{
			UserID:   row.ID,
			Username: row.Username,
			Name:     row.Name,
		}
		userIDs = append(userIDs, row.ID)
		snapshotByID[row.ID] = snapshot
	}

	if err := r.fillClassRecentActivity(ctx, userIDs, since, snapshotByID); err != nil {
		return nil, err
	}
	if err := r.fillClassSubmissionStats(ctx, userIDs, snapshotByID); err != nil {
		return nil, err
	}
	if err := r.fillClassWriteupAndReviewStats(ctx, userIDs, snapshotByID); err != nil {
		return nil, err
	}
	if err := r.fillClassHandsOnStats(ctx, userIDs, snapshotByID); err != nil {
		return nil, err
	}
	if err := r.fillClassDimensionFacts(ctx, userIDs, snapshotByID); err != nil {
		return nil, err
	}

	snapshots := make([]teachingadvice.StudentFactSnapshot, 0, len(studentRows))
	for _, row := range studentRows {
		snapshot := snapshotByID[row.ID]
		if snapshot == nil {
			continue
		}
		snapshots = append(snapshots, *snapshot)
	}
	return snapshots, nil
}

func (r *ClassInsightRepository) fillClassRecentActivity(
	ctx context.Context,
	userIDs []int64,
	since time.Time,
	snapshotByID map[int64]*teachingadvice.StudentFactSnapshot,
) error {
	type eventRow struct {
		UserID    int64     `gorm:"column:user_id"`
		Timestamp time.Time `gorm:"column:timestamp"`
	}

	classActivityDays := make(map[int64]map[string]struct{}, len(userIDs))

	recordEvent := func(row eventRow) {
		snapshot := snapshotByID[row.UserID]
		if snapshot == nil {
			return
		}
		snapshot.RecentEventCount7d++
		if snapshot.LastActivityAt == nil || row.Timestamp.After(*snapshot.LastActivityAt) {
			timestamp := row.Timestamp
			snapshot.LastActivityAt = &timestamp
		}
		dateKey := row.Timestamp.UTC().Format("2006-01-02")
		if snapshotDimensions, ok := classActivityDays[row.UserID]; ok {
			snapshotDimensions[dateKey] = struct{}{}
			return
		}
		classActivityDays[row.UserID] = map[string]struct{}{dateKey: {}}
	}

	rows := make([]eventRow, 0)
	if err := r.db.WithContext(ctx).Table("submissions AS s").
		Select("s.user_id, s.submitted_at AS timestamp").
		Where("s.user_id IN ? AND s.submitted_at >= ?", userIDs, since).
		Scan(&rows).Error; err != nil {
		return fmt.Errorf("get class submission activity rows: %w", err)
	}
	for _, row := range rows {
		recordEvent(row)
	}

	if r.db.Migrator().HasTable("audit_logs") {
		rows = rows[:0]
		if err := r.db.WithContext(ctx).Table("audit_logs AS a").
			Select("a.user_id, a.created_at AS timestamp").
			Where("a.user_id IN ? AND a.resource_type IN (?, ?) AND a.created_at >= ?", userIDs, "instance_access", "instance_proxy_request", since).
			Scan(&rows).Error; err != nil {
			return fmt.Errorf("get class audit activity rows: %w", err)
		}
		for _, row := range rows {
			recordEvent(row)
		}
	}

	if r.db.Migrator().HasTable("submission_writeups") {
		rows = rows[:0]
		if err := r.db.WithContext(ctx).Table("submission_writeups AS sw").
			Select("sw.user_id, sw.updated_at AS timestamp").
			Where("sw.user_id IN ? AND sw.updated_at >= ?", userIDs, since).
			Scan(&rows).Error; err != nil {
			return fmt.Errorf("get class writeup activity rows: %w", err)
		}
		for _, row := range rows {
			recordEvent(row)
		}
	}

	if r.db.Migrator().HasTable("awd_attack_logs") {
		rows = rows[:0]
		if err := r.db.WithContext(ctx).Table("awd_attack_logs AS al").
			Select("al.submitted_by_user_id AS user_id, al.created_at AS timestamp").
			Where("al.submitted_by_user_id IN ? AND al.created_at >= ?", userIDs, since).
			Scan(&rows).Error; err != nil {
			return fmt.Errorf("get class awd activity rows: %w", err)
		}
		for _, row := range rows {
			recordEvent(row)
		}
	}

	for userID, days := range classActivityDays {
		snapshot := snapshotByID[userID]
		if snapshot == nil {
			continue
		}
		snapshot.ActiveDays7d = len(days)
	}
	return nil
}

func (r *ClassInsightRepository) fillClassSubmissionStats(
	ctx context.Context,
	userIDs []int64,
	snapshotByID map[int64]*teachingadvice.StudentFactSnapshot,
) error {
	type summaryRow struct {
		UserID                int64 `gorm:"column:user_id"`
		ChallengeSuccessCount int   `gorm:"column:challenge_success_count"`
		ChallengeFailureCount int   `gorm:"column:challenge_failure_count"`
	}
	rows := make([]summaryRow, 0)
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			s.user_id,
			COALESCE(SUM(CASE WHEN s.is_correct THEN 1 ELSE 0 END), 0) AS challenge_success_count,
			COALESCE(SUM(CASE WHEN s.is_correct THEN 0 ELSE 1 END), 0) AS challenge_failure_count
		FROM submissions s
		WHERE s.user_id IN ?
		GROUP BY s.user_id
	`, userIDs).Scan(&rows).Error; err != nil {
		return fmt.Errorf("get class submission summary rows: %w", err)
	}
	for _, row := range rows {
		snapshot := snapshotByID[row.UserID]
		if snapshot == nil {
			continue
		}
		snapshot.ChallengeSuccessCount = row.ChallengeSuccessCount
		snapshot.SubmissionSuccessCount = row.ChallengeSuccessCount
		snapshot.SubmissionFailureCount = row.ChallengeFailureCount
		snapshot.CorrectSubmissionCount = row.ChallengeSuccessCount
		snapshot.WrongSubmissionCount = row.ChallengeFailureCount
	}

	type awdSummaryRow struct {
		UserID          int64 `gorm:"column:user_id"`
		AWDAttemptCount int   `gorm:"column:awd_attempt_count"`
		AWDSuccessCount int   `gorm:"column:awd_success_count"`
	}
	awdRows := make([]awdSummaryRow, 0)
	if r.db.Migrator().HasTable("awd_attack_logs") {
		if err := r.db.WithContext(ctx).Raw(`
			SELECT
				al.submitted_by_user_id AS user_id,
				COUNT(*) AS awd_attempt_count,
				COALESCE(SUM(CASE WHEN al.is_success = TRUE AND al.score_gained > 0 THEN 1 ELSE 0 END), 0) AS awd_success_count
			FROM awd_attack_logs al
			WHERE al.submitted_by_user_id IN ?
				AND al.source = ?
			GROUP BY al.submitted_by_user_id
		`, userIDs, contestcontracts.AWDAttackSourceSubmission).Scan(&awdRows).Error; err != nil {
			return fmt.Errorf("get class awd submission summary rows: %w", err)
		}
	}
	for _, row := range awdRows {
		snapshot := snapshotByID[row.UserID]
		if snapshot == nil {
			continue
		}
		snapshot.AWDAttemptCount = row.AWDAttemptCount
		snapshot.AWDSuccessCount = row.AWDSuccessCount
		snapshot.SubmissionSuccessCount += row.AWDSuccessCount
		snapshot.SubmissionFailureCount += maxInt(row.AWDAttemptCount-row.AWDSuccessCount, 0)
		snapshot.CorrectSubmissionCount = snapshot.SubmissionSuccessCount
		snapshot.WrongSubmissionCount = snapshot.SubmissionFailureCount
	}

	type resultRow struct {
		UserID     int64     `gorm:"column:user_id"`
		IsCorrect  bool      `gorm:"column:is_correct"`
		OccurredAt time.Time `gorm:"column:occurred_at"`
		EventID    int64     `gorm:"column:event_id"`
	}
	results := make([]resultRow, 0)
	resultQueryParts := []string{
		`SELECT
			s.user_id AS user_id,
			s.is_correct AS is_correct,
			s.submitted_at AS occurred_at,
			s.id AS event_id
		FROM submissions s
		WHERE s.user_id IN ?`,
	}
	resultArgs := []any{userIDs}
	if r.db.Migrator().HasTable("awd_attack_logs") {
		resultQueryParts = append(resultQueryParts, `SELECT
			al.submitted_by_user_id AS user_id,
			CASE WHEN al.is_success = TRUE AND al.score_gained > 0 THEN TRUE ELSE FALSE END AS is_correct,
			al.created_at AS occurred_at,
			al.id AS event_id
		FROM awd_attack_logs al
		WHERE al.submitted_by_user_id IN ?
			AND al.source = ?`)
		resultArgs = append(resultArgs, userIDs, contestcontracts.AWDAttackSourceSubmission)
	}
	if err := r.db.WithContext(ctx).Raw(fmt.Sprintf(`
		SELECT user_id, is_correct, occurred_at, event_id
		FROM (
			%s
		) ordered_results
		ORDER BY user_id ASC, occurred_at ASC, event_id ASC
	`, strings.Join(resultQueryParts, " UNION ALL ")), resultArgs...).Scan(&results).Error; err != nil {
		return fmt.Errorf("list class submission results: %w", err)
	}

	currentUserID := int64(0)
	currentWrongStreak := 0
	maxWrongStreak := 0
	flush := func() {
		if currentUserID == 0 {
			return
		}
		snapshot := snapshotByID[currentUserID]
		if snapshot != nil {
			snapshot.MaxWrongStreak = maxWrongStreak
		}
	}
	for _, result := range results {
		if result.UserID != currentUserID {
			flush()
			currentUserID = result.UserID
			currentWrongStreak = 0
			maxWrongStreak = 0
		}
		if result.IsCorrect {
			currentWrongStreak = 0
			continue
		}
		currentWrongStreak++
		if currentWrongStreak > maxWrongStreak {
			maxWrongStreak = currentWrongStreak
		}
	}
	flush()
	return nil
}

func (r *ClassInsightRepository) fillClassWriteupAndReviewStats(
	ctx context.Context,
	userIDs []int64,
	snapshotByID map[int64]*teachingadvice.StudentFactSnapshot,
) error {
	if r.db.Migrator().HasTable("submission_writeups") {
		rows := make([]struct {
			UserID int64 `gorm:"column:user_id"`
			Count  int   `gorm:"column:count"`
		}, 0)
		if err := r.db.WithContext(ctx).Table("submission_writeups AS sw").
			Select("sw.user_id, COUNT(*) AS count").
			Where("sw.user_id IN ?", userIDs).
			Group("sw.user_id").
			Scan(&rows).Error; err != nil {
			return fmt.Errorf("count class writeups: %w", err)
		}
		for _, row := range rows {
			snapshot := snapshotByID[row.UserID]
			if snapshot != nil {
				snapshot.WriteupCount = row.Count
			}
		}
	}

	rows := make([]struct {
		UserID int64 `gorm:"column:user_id"`
		Count  int   `gorm:"column:count"`
	}, 0)
	if err := r.db.WithContext(ctx).Table("submissions AS s").
		Select("s.user_id, COUNT(*) AS count").
		Where("s.user_id IN ? AND s.review_status = ?", userIDs, contestcontracts.SubmissionReviewStatusApproved).
		Group("s.user_id").
		Scan(&rows).Error; err != nil {
		return fmt.Errorf("count class approved reviews: %w", err)
	}
	for _, row := range rows {
		snapshot := snapshotByID[row.UserID]
		if snapshot != nil {
			snapshot.ApprovedReviewCount = row.Count
		}
	}
	return nil
}

func (r *ClassInsightRepository) fillClassHandsOnStats(
	ctx context.Context,
	userIDs []int64,
	snapshotByID map[int64]*teachingadvice.StudentFactSnapshot,
) error {
	if r.db.Migrator().HasTable("audit_logs") {
		rows := make([]struct {
			UserID int64 `gorm:"column:user_id"`
			Count  int   `gorm:"column:count"`
		}, 0)
		if err := r.db.WithContext(ctx).Table("audit_logs AS a").
			Select("a.user_id, COUNT(*) AS count").
			Where("a.user_id IN ? AND a.resource_type IN (?, ?)", userIDs, "instance_access", "instance_proxy_request").
			Group("a.user_id").
			Scan(&rows).Error; err != nil {
			return fmt.Errorf("count class hands-on events: %w", err)
		}
		for _, row := range rows {
			snapshot := snapshotByID[row.UserID]
			if snapshot != nil {
				snapshot.HandsOnEventCount = row.Count
			}
		}
	}

	if r.db.Migrator().HasTable("awd_attack_logs") {
		rows := make([]struct {
			UserID int64 `gorm:"column:user_id"`
			Count  int   `gorm:"column:count"`
		}, 0)
		if err := r.db.WithContext(ctx).Table("awd_attack_logs AS al").
			Select("al.submitted_by_user_id AS user_id, COUNT(*) AS count").
			Where(
				"al.submitted_by_user_id IN ? AND al.is_success = ? AND al.source = ? AND al.score_gained > 0",
				userIDs,
				true,
				contestcontracts.AWDAttackSourceSubmission,
			).
			Group("al.submitted_by_user_id").
			Scan(&rows).Error; err != nil {
			return fmt.Errorf("count class awd success events: %w", err)
		}
		for _, row := range rows {
			snapshot := snapshotByID[row.UserID]
			if snapshot != nil {
				snapshot.AWDSuccessCount = row.Count
			}
		}
	}
	return nil
}

func (r *ClassInsightRepository) fillClassDimensionFacts(
	ctx context.Context,
	userIDs []int64,
	snapshotByID map[int64]*teachingadvice.StudentFactSnapshot,
) error {
	dimensionFactsByUser := make(map[int64]map[string]*teachingadvice.DimensionFact, len(userIDs))
	for _, userID := range userIDs {
		dimensionFactsByUser[userID] = make(map[string]*teachingadvice.DimensionFact, len(taxonomy.AllDimensions))
		for _, dimension := range taxonomy.AllDimensions {
			dimensionCopy := dimension
			dimensionFactsByUser[userID][dimension] = &teachingadvice.DimensionFact{Dimension: dimensionCopy}
		}
	}

	profiles := make([]struct {
		UserID    int64   `gorm:"column:user_id"`
		Dimension string  `gorm:"column:dimension"`
		Score     float64 `gorm:"column:score"`
	}, 0)
	if err := r.db.WithContext(ctx).Table("skill_profiles AS sp").
		Select("sp.user_id, sp.dimension, sp.score").
		Where("sp.user_id IN ?", userIDs).
		Scan(&profiles).Error; err != nil {
		return fmt.Errorf("get class skill profile facts: %w", err)
	}
	for _, profile := range profiles {
		fact := ensureClassDimensionFact(dimensionFactsByUser, profile.UserID, profile.Dimension)
		fact.ProfileScore = profile.Score
	}

	if r.db.Migrator().HasTable("challenges") {
		attemptRows := make([]struct {
			UserID       int64  `gorm:"column:user_id"`
			Dimension    string `gorm:"column:dimension"`
			AttemptCount int    `gorm:"column:attempt_count"`
			SuccessCount int    `gorm:"column:success_count"`
		}, 0)
		if err := r.db.WithContext(ctx).Raw(`
			SELECT
				s.user_id,
				c.category AS dimension,
				COUNT(*) AS attempt_count,
				COALESCE(SUM(CASE WHEN s.is_correct THEN 1 ELSE 0 END), 0) AS success_count
			FROM submissions s
			JOIN challenges c ON c.id = s.challenge_id
			WHERE s.user_id IN ? AND c.status = ?
			GROUP BY s.user_id, c.category
		`, userIDs, challengecontracts.ChallengeStatusPublished).Scan(&attemptRows).Error; err != nil {
			return fmt.Errorf("get class dimension attempt facts: %w", err)
		}
		for _, row := range attemptRows {
			fact := ensureClassDimensionFact(dimensionFactsByUser, row.UserID, row.Dimension)
			fact.AttemptCount = row.AttemptCount
			fact.SuccessCount = row.SuccessCount
			fact.EvidenceCount += row.AttemptCount
		}

		solvedDifficultyRows := make([]struct {
			UserID      int64  `gorm:"column:user_id"`
			Dimension   string `gorm:"column:dimension"`
			Difficulty  string `gorm:"column:difficulty"`
			SolvedCount int    `gorm:"column:solved_count"`
		}, 0)
		if err := r.db.WithContext(ctx).Raw(`
			SELECT
				s.user_id,
				c.category AS dimension,
				c.difficulty AS difficulty,
				COUNT(DISTINCT s.challenge_id) AS solved_count
			FROM submissions s
			JOIN challenges c ON c.id = s.challenge_id
			WHERE s.user_id IN ?
				AND s.is_correct = TRUE
				AND c.status = ?
			GROUP BY s.user_id, c.category, c.difficulty
		`, userIDs, challengecontracts.ChallengeStatusPublished).Scan(&solvedDifficultyRows).Error; err != nil {
			return fmt.Errorf("get class solved difficulty facts: %w", err)
		}
		for _, row := range solvedDifficultyRows {
			fact := ensureClassDimensionFact(dimensionFactsByUser, row.UserID, row.Dimension)
			if fact.SolvedDifficultyCounts == nil {
				fact.SolvedDifficultyCounts = make(map[string]int)
			}
			fact.SolvedDifficultyCounts[row.Difficulty] = row.SolvedCount
		}

		if r.db.Migrator().HasTable("audit_logs") && r.db.Migrator().HasTable("instances") {
			auditRows := make([]struct {
				UserID    int64  `gorm:"column:user_id"`
				Dimension string `gorm:"column:dimension"`
				Count     int    `gorm:"column:count"`
			}, 0)
			if err := r.db.WithContext(ctx).Raw(`
				SELECT
					a.user_id,
					c.category AS dimension,
					COUNT(*) AS count
				FROM audit_logs a
				JOIN instances i ON i.id = a.resource_id
				JOIN challenges c ON c.id = i.challenge_id
				WHERE a.user_id IN ?
					AND a.resource_type IN (?, ?)
					AND c.status = ?
				GROUP BY a.user_id, c.category
			`, userIDs, "instance_access", "instance_proxy_request", challengecontracts.ChallengeStatusPublished).Scan(&auditRows).Error; err != nil {
				return fmt.Errorf("get class audit dimension facts: %w", err)
			}
			for _, row := range auditRows {
				fact := ensureClassDimensionFact(dimensionFactsByUser, row.UserID, row.Dimension)
				fact.EvidenceCount += row.Count
			}
		}

		if r.db.Migrator().HasTable("submission_writeups") {
			writeupRows := make([]struct {
				UserID    int64  `gorm:"column:user_id"`
				Dimension string `gorm:"column:dimension"`
				Count     int    `gorm:"column:count"`
			}, 0)
			if err := r.db.WithContext(ctx).Raw(`
				SELECT
					sw.user_id,
					c.category AS dimension,
					COUNT(*) AS count
				FROM submission_writeups sw
				JOIN challenges c ON c.id = sw.challenge_id
				WHERE sw.user_id IN ? AND c.status = ?
				GROUP BY sw.user_id, c.category
			`, userIDs, challengecontracts.ChallengeStatusPublished).Scan(&writeupRows).Error; err != nil {
				return fmt.Errorf("get class writeup dimension facts: %w", err)
			}
			for _, row := range writeupRows {
				fact := ensureClassDimensionFact(dimensionFactsByUser, row.UserID, row.Dimension)
				fact.EvidenceCount += row.Count
			}
		}

		reviewRows := make([]struct {
			UserID    int64  `gorm:"column:user_id"`
			Dimension string `gorm:"column:dimension"`
			Count     int    `gorm:"column:count"`
		}, 0)
		if err := r.db.WithContext(ctx).Raw(`
			SELECT
				s.user_id,
				c.category AS dimension,
				COUNT(*) AS count
			FROM submissions s
			JOIN challenges c ON c.id = s.challenge_id
			WHERE s.user_id IN ?
				AND s.review_status = ?
				AND c.status = ?
			GROUP BY s.user_id, c.category
		`, userIDs, contestcontracts.SubmissionReviewStatusApproved, challengecontracts.ChallengeStatusPublished).Scan(&reviewRows).Error; err != nil {
			return fmt.Errorf("get class review dimension facts: %w", err)
		}
		for _, row := range reviewRows {
			fact := ensureClassDimensionFact(dimensionFactsByUser, row.UserID, row.Dimension)
			fact.EvidenceCount += row.Count
		}
	}

	if r.db.Migrator().HasTable("awd_attack_logs") && r.db.Migrator().HasTable("awd_challenges") {
		awdAttemptRows := make([]struct {
			UserID       int64  `gorm:"column:user_id"`
			Dimension    string `gorm:"column:dimension"`
			AttemptCount int    `gorm:"column:attempt_count"`
			SuccessCount int    `gorm:"column:success_count"`
		}, 0)
		if err := r.db.WithContext(ctx).Raw(`
			SELECT
				al.submitted_by_user_id AS user_id,
				ac.category AS dimension,
				COUNT(*) AS attempt_count,
				COALESCE(SUM(CASE WHEN al.is_success = TRUE AND al.score_gained > 0 THEN 1 ELSE 0 END), 0) AS success_count
			FROM awd_attack_logs al
			JOIN awd_challenges ac ON ac.id = al.awd_challenge_id
			WHERE al.submitted_by_user_id IN ?
				AND al.source = ?
				AND ac.status = ?
			GROUP BY al.submitted_by_user_id, ac.category
		`, userIDs, contestcontracts.AWDAttackSourceSubmission, challengecontracts.AWDChallengeStatusPublished).Scan(&awdAttemptRows).Error; err != nil {
			return fmt.Errorf("get class awd attempt dimension facts: %w", err)
		}
		for _, row := range awdAttemptRows {
			fact := ensureClassDimensionFact(dimensionFactsByUser, row.UserID, row.Dimension)
			fact.AttemptCount += row.AttemptCount
			fact.SuccessCount += row.SuccessCount
			fact.EvidenceCount += row.AttemptCount
		}

		publishedRows := make([]struct {
			Dimension  string `gorm:"column:dimension"`
			TotalCount int    `gorm:"column:total_count"`
		}, 0)
		if err := r.db.WithContext(ctx).Raw(`
			SELECT
				ac.category AS dimension,
				COUNT(DISTINCT ac.id) AS total_count
			FROM awd_challenges ac
			WHERE ac.status = ?
			GROUP BY ac.category
		`, challengecontracts.AWDChallengeStatusPublished).Scan(&publishedRows).Error; err != nil {
			return fmt.Errorf("get class awd published dimension totals: %w", err)
		}
		awdTotals := make(map[string]int, len(publishedRows))
		for _, row := range publishedRows {
			awdTotals[row.Dimension] = row.TotalCount
		}

		successRows := make([]struct {
			UserID      int64  `gorm:"column:user_id"`
			Dimension   string `gorm:"column:dimension"`
			SolvedCount int    `gorm:"column:solved_count"`
		}, 0)
		if err := r.db.WithContext(ctx).Raw(`
			SELECT
				al.submitted_by_user_id AS user_id,
				ac.category AS dimension,
				COUNT(DISTINCT al.awd_challenge_id) AS solved_count
			FROM awd_attack_logs al
			JOIN awd_challenges ac ON ac.id = al.awd_challenge_id
			WHERE al.submitted_by_user_id IN ?
				AND al.source = ?
				AND al.is_success = TRUE
				AND al.score_gained > 0
				AND ac.status = ?
			GROUP BY al.submitted_by_user_id, ac.category
		`, userIDs, contestcontracts.AWDAttackSourceSubmission, challengecontracts.AWDChallengeStatusPublished).Scan(&successRows).Error; err != nil {
			return fmt.Errorf("get class awd success dimension facts: %w", err)
		}
		for _, row := range successRows {
			fact := ensureClassDimensionFact(dimensionFactsByUser, row.UserID, row.Dimension)
			if total := awdTotals[row.Dimension]; total > 0 {
				coverage := float64(row.SolvedCount) / float64(total)
				if coverage > fact.ProfileScore {
					fact.ProfileScore = coverage
				}
			}
		}

		difficultyRows := make([]struct {
			UserID      int64  `gorm:"column:user_id"`
			Dimension   string `gorm:"column:dimension"`
			Difficulty  string `gorm:"column:difficulty"`
			SolvedCount int    `gorm:"column:solved_count"`
		}, 0)
		if err := r.db.WithContext(ctx).Raw(`
			SELECT
				al.submitted_by_user_id AS user_id,
				ac.category AS dimension,
				ac.difficulty AS difficulty,
				COUNT(DISTINCT al.awd_challenge_id) AS solved_count
			FROM awd_attack_logs al
			JOIN awd_challenges ac ON ac.id = al.awd_challenge_id
			WHERE al.submitted_by_user_id IN ?
				AND al.source = ?
				AND al.is_success = TRUE
				AND al.score_gained > 0
				AND ac.status = ?
			GROUP BY al.submitted_by_user_id, ac.category, ac.difficulty
		`, userIDs, contestcontracts.AWDAttackSourceSubmission, challengecontracts.AWDChallengeStatusPublished).Scan(&difficultyRows).Error; err != nil {
			return fmt.Errorf("get class awd solved difficulty facts: %w", err)
		}
		for _, row := range difficultyRows {
			fact := ensureClassDimensionFact(dimensionFactsByUser, row.UserID, row.Dimension)
			if fact.SolvedDifficultyCounts == nil {
				fact.SolvedDifficultyCounts = make(map[string]int)
			}
			fact.SolvedDifficultyCounts[row.Difficulty] += row.SolvedCount
		}
	}

	for userID, dimensions := range dimensionFactsByUser {
		snapshot := snapshotByID[userID]
		if snapshot == nil {
			continue
		}
		items := make([]teachingadvice.DimensionFact, 0, len(taxonomy.AllDimensions))
		for _, dimension := range taxonomy.AllDimensions {
			fact := dimensions[dimension]
			if fact == nil {
				continue
			}
			items = append(items, *fact)
		}
		snapshot.Dimensions = items
	}
	return nil
}

func ensureClassDimensionFact(
	facts map[int64]map[string]*teachingadvice.DimensionFact,
	userID int64,
	dimension string,
) *teachingadvice.DimensionFact {
	perUser, ok := facts[userID]
	if !ok {
		perUser = make(map[string]*teachingadvice.DimensionFact)
		facts[userID] = perUser
	}
	if fact, exists := perUser[dimension]; exists {
		return fact
	}
	fact := &teachingadvice.DimensionFact{Dimension: dimension}
	perUser[dimension] = fact
	return fact
}

func (r *ClassInsightRepository) GetClassSummary(ctx context.Context, className string, since time.Time) (*queryports.ClassSummary, error) {
	studentCount, err := NewClassQueryRepository(r.db).CountStudentsByClass(ctx, className)
	if err != nil {
		return nil, err
	}

	summary := &queryports.ClassSummary{
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

func (r *ClassInsightRepository) GetClassTrend(ctx context.Context, className string, since time.Time, days int) (*queryports.ClassTrend, error) {
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
	`, identitycontracts.RoleStudent, className, since, identitycontracts.RoleStudent, className, since, identitycontracts.RoleStudent, className, since).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("get class trend: %w", err)
	}

	points := make([]queryports.ClassTrendPoint, days)
	indexByDate := make(map[string]int, days)
	for i := 0; i < days; i++ {
		date := since.AddDate(0, 0, i).Format("2006-01-02")
		points[i] = queryports.ClassTrendPoint{Date: date}
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

	return &queryports.ClassTrend{
		ClassName: className,
		Points:    points,
	}, nil
}

func (r *ClassInsightRepository) getAverageSolvedByClass(ctx context.Context, className string) (float64, error) {
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
	`, challengecontracts.ChallengeStatusPublished, identitycontracts.RoleStudent, className).Scan(&result).Error; err != nil {
		return 0, fmt.Errorf("get average solved by class: %w", err)
	}
	return result.AverageSolved, nil
}

func (r *ClassInsightRepository) getActiveStudentCountByClass(ctx context.Context, className string, since time.Time) (int64, error) {
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
			SELECT i.user_id
			FROM instances i
			JOIN users u ON u.id = i.user_id
			WHERE u.role = ? AND u.class_name = ? AND u.deleted_at IS NULL
				AND (i.created_at >= ? OR i.updated_at >= ?)
		) active
	`, identitycontracts.RoleStudent, className, since, identitycontracts.RoleStudent, className, since, since).Scan(&result).Error; err != nil {
		return 0, fmt.Errorf("get active student count by class: %w", err)
	}
	return result.Count, nil
}

func (r *ClassInsightRepository) getRecentEventCountByClass(ctx context.Context, className string, since time.Time) (int64, error) {
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
	`, identitycontracts.RoleStudent, className, since, identitycontracts.RoleStudent, className, since, identitycontracts.RoleStudent, className, since).Scan(&result).Error; err != nil {
		return 0, fmt.Errorf("get recent event count by class: %w", err)
	}
	return result.Count, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
