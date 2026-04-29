package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
)

type teacherAWDReviewContestCardRow struct {
	ID           int64
	Title        string
	Mode         string
	Status       string
	CurrentRound *int
	RoundCount   int
	TeamCount    int
	ExportReady  bool
}

type teacherAWDReviewContestMetaRow struct {
	ID           int64
	Title        string
	Mode         string
	Status       string
	CurrentRound *int
	RoundCount   int
	TeamCount    int
	ExportReady  bool
}

type TeacherAWDReviewRepository struct {
	db *gorm.DB
}

func NewTeacherAWDReviewRepository(db *gorm.DB) *TeacherAWDReviewRepository {
	return &TeacherAWDReviewRepository{db: db}
}

func (r *TeacherAWDReviewRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *TeacherAWDReviewRepository) ListTeacherAWDReviewContests(ctx context.Context) ([]assessmentdomain.TeacherAWDReviewContestCard, error) {
	rows := make([]teacherAWDReviewContestCardRow, 0)
	err := r.dbWithContext(ctx).Raw(`
		SELECT
			c.id,
			c.title,
			c.mode,
			c.status,
			(
				SELECT ar.round_number
				FROM awd_rounds ar
				WHERE ar.contest_id = c.id AND ar.status = ?
				ORDER BY ar.round_number DESC
				LIMIT 1
			) AS current_round,
			(
				SELECT COUNT(*)
				FROM awd_rounds ar
				WHERE ar.contest_id = c.id
			) AS round_count,
			(
				SELECT COUNT(*)
				FROM teams t
				WHERE t.contest_id = c.id AND t.deleted_at IS NULL
			) AS team_count,
			(
				SELECT MAX(created_at)
				FROM (
					SELECT te.created_at AS created_at
					FROM awd_traffic_events te
					WHERE te.contest_id = c.id
					UNION ALL
					SELECT al.created_at AS created_at
					FROM awd_attack_logs al
					JOIN awd_rounds ar ON ar.id = al.round_id
					WHERE ar.contest_id = c.id
				) AS evidence_events
			) AS latest_evidence_at,
			CASE WHEN c.status = ? THEN 1 ELSE 0 END AS export_ready
		FROM contests c
		WHERE c.mode = ? AND c.deleted_at IS NULL
		ORDER BY c.id DESC
	`, model.AWDRoundStatusRunning, model.ContestStatusEnded, model.ContestModeAWD).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list teacher awd review contests: %w", err)
	}
	items := make([]assessmentdomain.TeacherAWDReviewContestCard, 0, len(rows))
	for _, row := range rows {
		latestEvidenceAt, err := r.findLatestEvidenceAt(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, assessmentdomain.TeacherAWDReviewContestCard{
			ID:               row.ID,
			Title:            row.Title,
			Mode:             row.Mode,
			Status:           row.Status,
			CurrentRound:     row.CurrentRound,
			RoundCount:       row.RoundCount,
			TeamCount:        row.TeamCount,
			LatestEvidenceAt: latestEvidenceAt,
			ExportReady:      row.ExportReady,
		})
	}
	return items, nil
}

func (r *TeacherAWDReviewRepository) FindTeacherAWDReviewContest(ctx context.Context, contestID int64) (*assessmentdomain.TeacherAWDReviewContestMeta, error) {
	var row teacherAWDReviewContestMetaRow
	err := r.dbWithContext(ctx).Raw(`
		SELECT
			c.id,
			c.title,
			c.mode,
			c.status,
			(
				SELECT ar.round_number
				FROM awd_rounds ar
				WHERE ar.contest_id = c.id AND ar.status = ?
				ORDER BY ar.round_number DESC
				LIMIT 1
			) AS current_round,
			(
				SELECT COUNT(*)
				FROM awd_rounds ar
				WHERE ar.contest_id = c.id
			) AS round_count,
			(
				SELECT COUNT(*)
				FROM teams t
				WHERE t.contest_id = c.id AND t.deleted_at IS NULL
			) AS team_count,
			(
				SELECT MAX(created_at)
				FROM (
					SELECT te.created_at AS created_at
					FROM awd_traffic_events te
					WHERE te.contest_id = c.id
					UNION ALL
					SELECT al.created_at AS created_at
					FROM awd_attack_logs al
					JOIN awd_rounds ar ON ar.id = al.round_id
					WHERE ar.contest_id = c.id
				) AS evidence_events
			) AS latest_evidence_at,
			CASE WHEN c.status = ? THEN 1 ELSE 0 END AS export_ready
		FROM contests c
		WHERE c.id = ? AND c.mode = ? AND c.deleted_at IS NULL
	`, model.AWDRoundStatusRunning, model.ContestStatusEnded, contestID, model.ContestModeAWD).Scan(&row).Error
	if err != nil {
		return nil, fmt.Errorf("find teacher awd review contest %d: %w", contestID, err)
	}
	if row.ID == 0 {
		return nil, nil
	}
	latestEvidenceAt, err := r.findLatestEvidenceAt(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	return &assessmentdomain.TeacherAWDReviewContestMeta{
		ID:               row.ID,
		Title:            row.Title,
		Mode:             row.Mode,
		Status:           row.Status,
		CurrentRound:     row.CurrentRound,
		RoundCount:       row.RoundCount,
		TeamCount:        row.TeamCount,
		LatestEvidenceAt: latestEvidenceAt,
		ExportReady:      row.ExportReady,
	}, nil
}

func (r *TeacherAWDReviewRepository) ListTeacherAWDReviewRounds(ctx context.Context, contestID int64) ([]assessmentdomain.TeacherAWDReviewRoundSummary, error) {
	rows := make([]assessmentdomain.TeacherAWDReviewRoundSummary, 0)
	err := r.dbWithContext(ctx).Model(&model.AWDRound{}).
		Select("id, contest_id, round_number, status, started_at, ended_at, attack_score, defense_score").
		Where("contest_id = ?", contestID).
		Order("round_number ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list teacher awd review rounds: %w", err)
	}
	return rows, nil
}

func (r *TeacherAWDReviewRepository) ListTeacherAWDReviewTeams(ctx context.Context, contestID int64) ([]assessmentdomain.TeacherAWDReviewTeamSummary, error) {
	rows := make([]assessmentdomain.TeacherAWDReviewTeamSummary, 0)
	err := r.dbWithContext(ctx).Raw(`
		SELECT
			t.id AS team_id,
			t.name AS team_name,
			t.captain_id,
			t.total_score,
			COUNT(DISTINCT tm.user_id) AS member_count,
			t.last_solve_at
		FROM teams t
		LEFT JOIN team_members tm
			ON tm.team_id = t.id AND tm.contest_id = t.contest_id
		WHERE t.contest_id = ? AND t.deleted_at IS NULL
		GROUP BY t.id, t.name, t.captain_id, t.total_score, t.last_solve_at
		ORDER BY t.total_score DESC, t.id ASC
	`, contestID).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list teacher awd review teams: %w", err)
	}
	return rows, nil
}

func (r *TeacherAWDReviewRepository) ListTeacherAWDReviewRoundServices(ctx context.Context, roundID int64) ([]assessmentdomain.TeacherAWDReviewServiceRecord, error) {
	rows := make([]assessmentdomain.TeacherAWDReviewServiceRecord, 0)
	err := r.dbWithContext(ctx).Raw(`
		SELECT
			ts.id,
			ts.round_id,
			ts.team_id,
			t.name AS team_name,
			ts.service_id,
			ts.awd_challenge_id,
			COALESCE(ch.name, '') AS awd_challenge_title,
			ts.service_status,
			ts.attack_received,
			ts.sla_score,
			ts.defense_score,
			ts.attack_score,
			ts.updated_at
		FROM awd_team_services ts
		JOIN teams t ON t.id = ts.team_id
		LEFT JOIN awd_challenges ch ON ch.id = ts.awd_challenge_id
		WHERE ts.round_id = ? AND t.deleted_at IS NULL
		ORDER BY ts.team_id ASC, ts.service_id ASC, ts.awd_challenge_id ASC
	`, roundID).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list teacher awd review round services: %w", err)
	}
	return rows, nil
}

func (r *TeacherAWDReviewRepository) ListTeacherAWDReviewRoundAttacks(ctx context.Context, roundID int64) ([]assessmentdomain.TeacherAWDReviewAttackRecord, error) {
	rows := make([]assessmentdomain.TeacherAWDReviewAttackRecord, 0)
	err := r.dbWithContext(ctx).Raw(`
		SELECT
			al.id,
			al.round_id,
			al.attacker_team_id,
			COALESCE(attacker.name, '') AS attacker_team_name,
			al.victim_team_id,
			COALESCE(victim.name, '') AS victim_team_name,
			al.service_id,
			al.awd_challenge_id,
			COALESCE(ch.name, '') AS awd_challenge_title,
			al.attack_type,
			al.source,
			COALESCE(al.submitted_flag, '') AS submitted_flag,
			al.is_success,
			al.score_gained,
			al.created_at
		FROM awd_attack_logs al
		LEFT JOIN teams attacker ON attacker.id = al.attacker_team_id
		LEFT JOIN teams victim ON victim.id = al.victim_team_id
		LEFT JOIN awd_challenges ch ON ch.id = al.awd_challenge_id
		WHERE al.round_id = ?
		ORDER BY al.created_at DESC, al.id DESC
	`, roundID).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list teacher awd review round attacks: %w", err)
	}
	return rows, nil
}

func (r *TeacherAWDReviewRepository) ListTeacherAWDReviewRoundTraffic(ctx context.Context, contestID, roundID int64) ([]assessmentdomain.TeacherAWDReviewTrafficRecord, error) {
	rows := make([]assessmentdomain.TeacherAWDReviewTrafficRecord, 0)
	err := r.dbWithContext(ctx).Raw(`
		SELECT
			te.id,
			te.contest_id,
			te.round_id,
			te.attacker_team_id,
			COALESCE(attacker.name, '') AS attacker_team_name,
			te.victim_team_id,
			COALESCE(victim.name, '') AS victim_team_name,
			te.service_id,
			te.awd_challenge_id,
			COALESCE(ch.name, '') AS awd_challenge_title,
			te.method,
			te.path,
			te.status_code,
			te.source,
			te.created_at
		FROM awd_traffic_events te
		LEFT JOIN teams attacker ON attacker.id = te.attacker_team_id
		LEFT JOIN teams victim ON victim.id = te.victim_team_id
		LEFT JOIN awd_challenges ch ON ch.id = te.awd_challenge_id
		WHERE te.contest_id = ? AND te.round_id = ?
		ORDER BY te.created_at DESC, te.id DESC
	`, contestID, roundID).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list teacher awd review round traffic: %w", err)
	}
	return rows, nil
}

func (r *TeacherAWDReviewRepository) findLatestEvidenceAt(ctx context.Context, contestID int64) (*time.Time, error) {
	trafficAt, err := r.findLatestTrafficAt(ctx, contestID)
	if err != nil {
		return nil, err
	}
	attackAt, err := r.findLatestAttackAt(ctx, contestID)
	if err != nil {
		return nil, err
	}

	switch {
	case trafficAt == nil:
		return attackAt, nil
	case attackAt == nil:
		return trafficAt, nil
	case trafficAt.After(*attackAt):
		return trafficAt, nil
	default:
		return attackAt, nil
	}
}

func (r *TeacherAWDReviewRepository) findLatestTrafficAt(ctx context.Context, contestID int64) (*time.Time, error) {
	var row struct {
		CreatedAt time.Time
	}
	err := r.dbWithContext(ctx).Model(&model.AWDTrafficEvent{}).
		Select("created_at").
		Where("contest_id = ?", contestID).
		Order("created_at DESC").
		Limit(1).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find latest traffic evidence for contest %d: %w", contestID, err)
	}
	return &row.CreatedAt, nil
}

func (r *TeacherAWDReviewRepository) findLatestAttackAt(ctx context.Context, contestID int64) (*time.Time, error) {
	var row struct {
		CreatedAt time.Time
	}
	err := r.dbWithContext(ctx).Model(&model.AWDAttackLog{}).
		Select("awd_attack_logs.created_at").
		Joins("JOIN awd_rounds ON awd_rounds.id = awd_attack_logs.round_id").
		Where("awd_rounds.contest_id = ?", contestID).
		Order("awd_attack_logs.created_at DESC").
		Limit(1).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find latest attack evidence for contest %d: %w", contestID, err)
	}
	return &row.CreatedAt, nil
}
