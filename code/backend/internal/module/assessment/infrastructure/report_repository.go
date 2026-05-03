package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	readmodelports "ctf-platform/internal/module/teaching_readmodel/ports"
	"ctf-platform/internal/teaching/evidence"
)

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func reportSolvedChallengesCTE() string {
	return `
		solved_challenges AS (
			SELECT DISTINCT s.user_id AS user_id, s.challenge_id AS challenge_id
			FROM submissions s
			WHERE s.is_correct = TRUE AND s.contest_id IS NULL
		)
	`
}

func reportAWDAttackDetailSQL(successExpr, victimTeamNameExpr, scoreExpr string) string {
	return fmt.Sprintf(
		"CASE WHEN %s THEN 'AWD 攻击命中 ' || COALESCE(%s, '目标队伍') || CASE WHEN %s > 0 THEN '，得分 ' || CAST(%s AS TEXT) ELSE '' END ELSE 'AWD 攻击未命中 ' || COALESCE(%s, '目标队伍') END",
		successExpr,
		victimTeamNameExpr,
		scoreExpr,
		scoreExpr,
		victimTeamNameExpr,
	)
}

func (r *ReportRepository) Create(ctx context.Context, report *model.Report) error {
	return r.db.WithContext(ctx).Create(report).Error
}

func (r *ReportRepository) FindByID(ctx context.Context, reportID int64) (*model.Report, error) {
	var report model.Report
	if err := r.db.WithContext(ctx).First(&report, reportID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcodeReportNotFound()
		}
		return nil, err
	}
	return &report, nil
}

func (r *ReportRepository) MarkReady(ctx context.Context, reportID int64, filePath string, expiresAt time.Time) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.Report{}).
		Where("id = ?", reportID).
		Updates(map[string]any{
			"status":       model.ReportStatusReady,
			"file_path":    filePath,
			"expires_at":   expiresAt,
			"error_msg":    nil,
			"completed_at": &now,
		}).Error
}

func (r *ReportRepository) MarkFailed(ctx context.Context, reportID int64, message string) error {
	return r.db.WithContext(ctx).Model(&model.Report{}).
		Where("id = ?", reportID).
		Updates(map[string]any{
			"status":    model.ReportStatusFailed,
			"error_msg": &message,
		}).Error
}

func (r *ReportRepository) FindUserByID(ctx context.Context, userID int64) (*assessmentdomain.ReportUser, error) {
	var user assessmentdomain.ReportUser
	err := r.db.WithContext(ctx).Model(&model.User{}).
		Select("id, username, COALESCE(name, '') AS name, class_name, role").
		Where("id = ? AND deleted_at IS NULL", userID).
		Scan(&user).Error
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}

func (r *ReportRepository) FindContestByID(ctx context.Context, contestID int64) (*model.Contest, error) {
	var contest model.Contest
	if err := r.db.WithContext(ctx).Where("id = ?", contestID).First(&contest).Error; err != nil {
		return nil, err
	}
	return &contest, nil
}

func (r *ReportRepository) GetPersonalStats(ctx context.Context, userID int64) (*assessmentdomain.PersonalReportStats, error) {
	var stats assessmentdomain.PersonalReportStats
	query := fmt.Sprintf(`
		WITH %s,
		user_solved AS (
			SELECT sc.challenge_id
			FROM solved_challenges sc
			WHERE sc.user_id = ?
		),
		user_scores AS (
			SELECT
				u.id AS user_id,
				COALESCE(SUM(c.points), 0) AS total_score
			FROM users u
			LEFT JOIN solved_challenges sc ON sc.user_id = u.id
			LEFT JOIN challenges c ON c.id = sc.challenge_id AND c.status = 'published'
			WHERE u.deleted_at IS NULL
			GROUP BY u.id
		),
		ranked AS (
			SELECT user_id, RANK() OVER (ORDER BY total_score DESC, user_id ASC) AS rank
			FROM user_scores
		)
		SELECT
			COALESCE((
				SELECT SUM(c.points)
				FROM user_solved us
				JOIN challenges c ON c.id = us.challenge_id AND c.status = 'published'
			), 0) AS total_score,
			COALESCE((
				SELECT COUNT(*)
				FROM user_solved us
				JOIN challenges c ON c.id = us.challenge_id AND c.status = 'published'
			), 0) AS total_solved,
			COALESCE((
				SELECT COUNT(*)
				FROM submissions
				WHERE user_id = ? AND contest_id IS NULL
			), 0) + COALESCE((
				SELECT COUNT(*)
				FROM awd_attack_logs aal
				WHERE aal.submitted_by_user_id = ?
					AND aal.source = '%s'
			), 0) AS total_attempts,
			COALESCE((SELECT rank FROM ranked WHERE user_id = ?), 1) AS rank
	`, reportSolvedChallengesCTE(), model.AWDAttackSourceSubmission)
	err := r.db.WithContext(ctx).Raw(query, userID, userID, userID, userID).Scan(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *ReportRepository) ListPersonalDimensionStats(ctx context.Context, userID int64) ([]assessmentdomain.ReportDimensionStat, error) {
	stats := make([]assessmentdomain.ReportDimensionStat, 0)
	query := fmt.Sprintf(`
		WITH %s,
		user_solved AS (
			SELECT sc.challenge_id
			FROM solved_challenges sc
			WHERE sc.user_id = ?
		)
		SELECT
			c.category AS dimension,
			COUNT(DISTINCT CASE WHEN us.challenge_id IS NOT NULL THEN c.id END) AS solved,
			COUNT(DISTINCT c.id) AS total
		FROM challenges c
		LEFT JOIN user_solved us ON us.challenge_id = c.id
		WHERE c.status = 'published'
		GROUP BY c.category
		ORDER BY c.category
	`, reportSolvedChallengesCTE())
	err := r.db.WithContext(ctx).Raw(query, userID).Scan(&stats).Error
	return stats, err
}

func (r *ReportRepository) CountClassStudents(ctx context.Context, className string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("class_name = ? AND role = ? AND deleted_at IS NULL", className, model.RoleStudent).
		Count(&count).Error
	return int(count), err
}

func (r *ReportRepository) GetClassAverageScore(ctx context.Context, className string) (float64, error) {
	var avgScore float64
	query := fmt.Sprintf(`
		WITH %s,
		user_scores AS (
			SELECT
				u.id AS user_id,
				COALESCE(SUM(c.points), 0) AS total_score
			FROM users u
			LEFT JOIN solved_challenges sc ON sc.user_id = u.id
			LEFT JOIN challenges c ON c.id = sc.challenge_id AND c.status = 'published'
			WHERE u.class_name = ? AND u.role = ? AND u.deleted_at IS NULL
			GROUP BY u.id
		)
		SELECT COALESCE(AVG(total_score), 0) AS avg_score
		FROM user_scores
	`, reportSolvedChallengesCTE())
	err := r.db.WithContext(ctx).Raw(query, className, model.RoleStudent).Scan(&avgScore).Error
	return avgScore, err
}

func (r *ReportRepository) ListClassDimensionAverages(ctx context.Context, className string) ([]assessmentdomain.ClassDimensionAverage, error) {
	rows := make([]assessmentdomain.ClassDimensionAverage, 0)
	err := r.db.WithContext(ctx).Raw(`
		SELECT sp.dimension, COALESCE(AVG(sp.score), 0) AS avg_score
		FROM users u
		JOIN skill_profiles sp ON sp.user_id = u.id
		WHERE u.class_name = ? AND u.role = ? AND u.deleted_at IS NULL
		GROUP BY sp.dimension
		ORDER BY sp.dimension
	`, className, model.RoleStudent).Scan(&rows).Error
	return rows, err
}

func (r *ReportRepository) ListClassTopStudents(ctx context.Context, className string, limit int) ([]assessmentdomain.ClassTopStudent, error) {
	rows := make([]assessmentdomain.ClassTopStudent, 0)
	query := fmt.Sprintf(`
		WITH %s,
		user_scores AS (
			SELECT
				u.id AS user_id,
				u.username,
				COALESCE(SUM(c.points), 0) AS total_score
			FROM users u
			LEFT JOIN solved_challenges sc ON sc.user_id = u.id
			LEFT JOIN challenges c ON c.id = sc.challenge_id AND c.status = 'published'
			WHERE u.class_name = ? AND u.role = ? AND u.deleted_at IS NULL
			GROUP BY u.id, u.username
		)
		SELECT
			user_id,
			username,
			total_score,
			RANK() OVER (ORDER BY total_score DESC, user_id ASC) AS rank
		FROM user_scores
		ORDER BY total_score DESC, user_id ASC
		LIMIT ?
	`, reportSolvedChallengesCTE())
	err := r.db.WithContext(ctx).Raw(query, className, model.RoleStudent, limit).Scan(&rows).Error
	return rows, err
}

type contestScoreboardRow struct {
	TeamID              int64  `gorm:"column:team_id"`
	TeamName            string `gorm:"column:team_name"`
	Score               int    `gorm:"column:score"`
	SolvedCount         int    `gorm:"column:solved_count"`
	LastSubmissionAtRaw string `gorm:"column:last_submission_at"`
}

func (r *ReportRepository) ListContestScoreboard(ctx context.Context, contestID int64) ([]assessmentdomain.ContestExportScoreboardItem, error) {
	rows := make([]contestScoreboardRow, 0)
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			t.id AS team_id,
			t.name AS team_name,
			t.total_score AS score,
			COUNT(DISTINCT CASE
				WHEN s.is_correct = TRUE THEN s.challenge_id
			END) AS solved_count,
			MAX(CASE WHEN s.is_correct = TRUE THEN s.submitted_at END) AS last_submission_at
		FROM teams t
		LEFT JOIN submissions s
			ON s.contest_id = t.contest_id
			AND s.team_id = t.id
		WHERE t.contest_id = ? AND t.deleted_at IS NULL
		GROUP BY t.id, t.name, t.total_score
		ORDER BY t.total_score DESC, MAX(CASE WHEN s.is_correct = TRUE THEN s.submitted_at END) ASC, t.id ASC
	`, contestID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]assessmentdomain.ContestExportScoreboardItem, 0, len(rows))
	for idx, row := range rows {
		items = append(items, assessmentdomain.ContestExportScoreboardItem{
			Rank:             idx + 1,
			TeamID:           row.TeamID,
			TeamName:         row.TeamName,
			Score:            row.Score,
			SolvedCount:      row.SolvedCount,
			LastSubmissionAt: parseAggregateTime(row.LastSubmissionAtRaw),
		})
	}
	return items, nil
}

func (r *ReportRepository) ListContestChallenges(ctx context.Context, contestID int64) ([]assessmentdomain.ContestExportChallengeItem, error) {
	rows := make([]assessmentdomain.ContestExportChallengeItem, 0)
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			cc.id AS contest_challenge_id,
			cc.challenge_id,
			c.title,
			c.category,
			c.difficulty,
			cc.points,
			cc."order",
			cc.is_visible,
			COUNT(DISTINCT CASE
				WHEN s.is_correct = TRUE THEN COALESCE(s.team_id, -s.user_id)
			END) AS solve_count,
			cc.first_blood_by,
			NULLIF(fb.name, '') AS first_blood_team_name
		FROM contest_challenges cc
		JOIN challenges c ON c.id = cc.challenge_id
		LEFT JOIN submissions s
			ON s.contest_id = cc.contest_id
			AND s.challenge_id = cc.challenge_id
		LEFT JOIN teams fb ON fb.id = cc.first_blood_by
		WHERE cc.contest_id = ? AND cc.deleted_at IS NULL
		GROUP BY
			cc.id, cc.challenge_id, c.title, c.category, c.difficulty,
			cc.points, cc."order", cc.is_visible, cc.first_blood_by, fb.name
		ORDER BY cc."order" ASC, cc.id ASC
	`, contestID).Scan(&rows).Error
	return rows, err
}

type contestTeamRow struct {
	TeamID          int64  `gorm:"column:team_id"`
	Name            string `gorm:"column:name"`
	CaptainID       int64  `gorm:"column:captain_id"`
	CaptainUsername string `gorm:"column:captain_username"`
	MaxMembers      int    `gorm:"column:max_members"`
	TotalScore      int    `gorm:"column:total_score"`
	LastSolveAtRaw  string `gorm:"column:last_solve_at"`
}

type contestTeamMemberRow struct {
	TeamID    int64      `gorm:"column:team_id"`
	UserID    int64      `gorm:"column:user_id"`
	Username  string     `gorm:"column:username"`
	Name      string     `gorm:"column:name"`
	ClassName string     `gorm:"column:class_name"`
	JoinedAt  *time.Time `gorm:"column:joined_at"`
}

func (r *ReportRepository) ListContestTeams(ctx context.Context, contestID int64) ([]assessmentdomain.ContestExportTeamItem, error) {
	teams := make([]contestTeamRow, 0)
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			t.id AS team_id,
			t.name,
			t.captain_id,
			cu.username AS captain_username,
			t.max_members,
			t.total_score,
			t.last_solve_at
		FROM teams t
		JOIN users cu ON cu.id = t.captain_id
		WHERE t.contest_id = ? AND t.deleted_at IS NULL
		ORDER BY t.total_score DESC, t.id ASC
	`, contestID).Scan(&teams).Error; err != nil {
		return nil, err
	}

	members := make([]contestTeamMemberRow, 0)
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			tm.team_id,
			u.id AS user_id,
			u.username,
			COALESCE(u.name, '') AS name,
			COALESCE(u.class_name, '') AS class_name,
			tm.joined_at
		FROM team_members tm
		JOIN users u ON u.id = tm.user_id
		WHERE tm.contest_id = ?
		ORDER BY tm.team_id ASC, tm.joined_at ASC, tm.id ASC
	`, contestID).Scan(&members).Error; err != nil {
		return nil, err
	}

	memberMap := make(map[int64][]assessmentdomain.ContestExportTeamMember, len(teams))
	for _, member := range members {
		memberMap[member.TeamID] = append(memberMap[member.TeamID], assessmentdomain.ContestExportTeamMember{
			UserID:    member.UserID,
			Username:  member.Username,
			Name:      member.Name,
			ClassName: member.ClassName,
			JoinedAt:  member.JoinedAt,
		})
	}

	items := make([]assessmentdomain.ContestExportTeamItem, 0, len(teams))
	for _, team := range teams {
		teamMembers := memberMap[team.TeamID]
		items = append(items, assessmentdomain.ContestExportTeamItem{
			TeamID:          team.TeamID,
			Name:            team.Name,
			CaptainID:       team.CaptainID,
			CaptainUsername: team.CaptainUsername,
			MaxMembers:      team.MaxMembers,
			TotalScore:      team.TotalScore,
			LastSolveAt:     parseAggregateTime(team.LastSolveAtRaw),
			MemberCount:     len(teamMembers),
			Members:         teamMembers,
		})
	}
	return items, nil
}

func (r *ReportRepository) CountPublishedChallenges(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Challenge{}).
		Where("status = ?", model.ChallengeStatusPublished).
		Count(&count).Error
	return count, err
}

func (r *ReportRepository) GetStudentTimeline(ctx context.Context, userID int64, limit, offset int) ([]assessmentdomain.ReviewArchiveTimelineEvent, error) {
	if limit <= 0 {
		limit = 200
	}

	rows := make([]assessmentdomain.ReviewArchiveTimelineEvent, 0)
	err := r.db.WithContext(ctx).Raw(fmt.Sprintf(`
		SELECT
			events.type,
			events.challenge_id,
			events.awd_challenge_id,
			events.awd_challenge_title,
			COALESCE(NULLIF(events.awd_challenge_title, ''), c.title, '') AS title,
			events.timestamp,
			events.is_correct,
			events.points,
			events.detail
		FROM (
			SELECT
				'instance_start' AS type,
				i.challenge_id,
				0 AS awd_challenge_id,
				'' AS awd_challenge_title,
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
				0 AS awd_challenge_id,
				'' AS awd_challenge_title,
				s.submitted_at AS timestamp,
				s.is_correct,
				CASE WHEN s.is_correct THEN c.points ELSE NULL END AS points,
				CASE WHEN s.is_correct THEN '提交命中 Flag' ELSE '提交未命中 Flag' END AS detail
			FROM submissions s
			LEFT JOIN challenges c ON c.id = s.challenge_id
			WHERE s.user_id = ? AND s.contest_id IS NULL
			UNION ALL
			SELECT
				'instance_destroy' AS type,
				i.challenge_id,
				0 AS awd_challenge_id,
				'' AS awd_challenge_title,
				i.destroyed_at AS timestamp,
				NULL AS is_correct,
				NULL AS points,
				'结束练习实例' AS detail
			FROM instances i
			WHERE i.user_id = ? AND i.status IN ('stopped', 'expired', 'destroyed') AND i.destroyed_at IS NOT NULL
			UNION ALL
			SELECT
				'awd_attack_submit' AS type,
				0 AS challenge_id,
				al.awd_challenge_id AS awd_challenge_id,
				COALESCE(ac.name, '') AS awd_challenge_title,
				al.created_at AS timestamp,
				al.is_success AS is_correct,
				CASE WHEN al.score_gained > 0 THEN al.score_gained ELSE NULL END AS points,
				%s AS detail
			FROM awd_attack_logs al
			LEFT JOIN awd_challenges ac ON ac.id = al.awd_challenge_id
			LEFT JOIN teams vt ON vt.id = al.victim_team_id
			WHERE al.submitted_by_user_id = ? AND al.source = '%s'
		) events
		LEFT JOIN challenges c ON c.id = events.challenge_id
		ORDER BY events.timestamp DESC
		LIMIT ? OFFSET ?
	`, reportAWDAttackDetailSQL("al.is_success", "vt.name", "al.score_gained"), model.AWDAttackSourceSubmission), userID, userID, userID, userID, limit, offset).Scan(&rows).Error
	return rows, err
}

func (r *ReportRepository) GetStudentEvidence(ctx context.Context, userID int64, query readmodelports.EvidenceQuery) ([]assessmentdomain.ReviewArchiveEvidenceEvent, error) {
	events := make([]assessmentdomain.ReviewArchiveEvidenceEvent, 0)

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
		return nil, err
	}
	for _, row := range accessRows {
		events = append(events, toReviewArchiveEvidenceEvent(evidence.NewInstanceAccessEvent(evidence.InstanceAccessInput{
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
		return nil, err
	}
	for _, row := range proxyRows {
		events = append(events, toReviewArchiveEvidenceEvent(evidence.NewProxyRequestEvent(evidence.ProxyRequestInput{
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
		return nil, err
	}
	for _, row := range submissionRows {
		events = append(events, toReviewArchiveEvidenceEvent(evidence.NewChallengeSubmissionEvent(evidence.ChallengeSubmissionInput{
			UserID:      userID,
			ChallengeID: row.ChallengeID,
			Title:       row.Title,
			Timestamp:   row.Timestamp,
			IsCorrect:   row.IsCorrect,
			Points:      row.Points,
		})))
	}

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
	awdWhere := "al.submitted_by_user_id = ?"
	awdArgs := []any{userID}
	if r.db.Migrator().HasTable("team_members") {
		awdWhere += " OR al.attacker_team_id IN (SELECT tm.team_id FROM team_members tm WHERE tm.user_id = ?)"
		awdArgs = append(awdArgs, userID)
	}
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
		Where(awdWhere, awdArgs...)
	if query.ChallengeID != nil {
		awdQuery = awdQuery.Where("al.awd_challenge_id = ?", *query.ChallengeID)
	}
	if err := awdQuery.Order("al.created_at ASC").Scan(&awdRows).Error; err != nil {
		return nil, err
	}
	for _, row := range awdRows {
		scope := "team"
		if row.SubmittedByUserID != nil && *row.SubmittedByUserID == userID {
			scope = "student"
		}
		events = append(events, toReviewArchiveEvidenceEvent(evidence.NewAWDAttackEvent(evidence.AWDAttackInput{
			UserID:            userID,
			TeamID:            &row.TeamID,
			ContestID:         &row.ContestID,
			RoundID:           &row.RoundID,
			ServiceID:         &row.ServiceID,
			VictimTeamID:      &row.VictimTeamID,
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

	if r.db.Migrator().HasTable("awd_traffic_events") && r.db.Migrator().HasTable("team_members") {
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
			return nil, err
		}
		for _, row := range trafficRows {
			events = append(events, toReviewArchiveEvidenceEvent(evidence.NewAWDTrafficEvent(evidence.AWDTrafficInput{
				UserID:            userID,
				TeamID:            &row.TeamID,
				ContestID:         &row.ContestID,
				RoundID:           &row.RoundID,
				ServiceID:         &row.ServiceID,
				VictimTeamID:      &row.VictimTeamID,
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

func (r *ReportRepository) ListStudentWriteups(ctx context.Context, userID int64) ([]assessmentdomain.ReviewArchiveWriteupItem, error) {
	rows := make([]assessmentdomain.ReviewArchiveWriteupItem, 0)
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			sw.id,
			sw.challenge_id,
			c.title AS challenge_title,
			sw.title,
			sw.submission_status,
			sw.visibility_status,
			sw.is_recommended,
			sw.published_at,
			sw.updated_at
		FROM submission_writeups sw
		JOIN challenges c ON c.id = sw.challenge_id
		WHERE sw.user_id = ?
		ORDER BY sw.updated_at DESC, sw.id DESC
	`, userID).Scan(&rows).Error
	return rows, err
}

func (r *ReportRepository) ListStudentManualReviews(ctx context.Context, userID int64) ([]assessmentdomain.ReviewArchiveManualReviewItem, error) {
	rows := make([]assessmentdomain.ReviewArchiveManualReviewItem, 0)
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			s.id,
			s.challenge_id,
			c.title AS challenge_title,
			s.flag AS answer,
			s.review_status,
			s.submitted_at,
			s.reviewed_at,
			s.review_comment,
			s.score,
			COALESCE(reviewer.name, reviewer.username, '') AS reviewer_name
		FROM submissions s
		JOIN challenges c ON c.id = s.challenge_id
		LEFT JOIN users reviewer ON reviewer.id = s.reviewed_by
		WHERE s.user_id = ? AND c.flag_type = ?
		ORDER BY s.updated_at DESC, s.id DESC
	`, userID, model.FlagTypeManualReview).Scan(&rows).Error
	return rows, err
}

func errcodeReportNotFound() error {
	return gorm.ErrRecordNotFound
}

func parseAggregateTime(raw string) *time.Time {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}

	layouts := []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, trimmed)
		if err == nil {
			return &parsed
		}
	}
	return nil
}

func toReviewArchiveEvidenceEvent(event evidence.Event) assessmentdomain.ReviewArchiveEvidenceEvent {
	return assessmentdomain.ReviewArchiveEvidenceEvent{
		Type:              event.Type,
		ChallengeID:       event.ChallengeID,
		AWDChallengeID:    event.AWDChallengeID,
		AWDChallengeTitle: event.AWDChallengeTitle,
		Title:             event.Title,
		Timestamp:         event.Timestamp,
		Detail:            event.Detail,
		Meta:              event.Meta,
	}
}
