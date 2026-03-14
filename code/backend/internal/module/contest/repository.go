package contest

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

var (
	ErrContestNotFound = errors.New("contest not found")
)

type Repository interface {
	Create(ctx context.Context, contest *model.Contest) error
	FindByID(ctx context.Context, id int64) (*model.Contest, error)
	Update(ctx context.Context, contest *model.Contest) error
	List(ctx context.Context, status *string, offset, limit int) ([]*model.Contest, int64, error)
	ListByStatusesAndTimeRange(ctx context.Context, statuses []string, now time.Time, offset, limit int) ([]*model.Contest, int64, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	FindTeamsByIDs(ctx context.Context, ids []int64) ([]*model.Team, error)
	FindTeamsByContest(ctx context.Context, contestID int64) ([]*model.Team, error)
	FindScoreboardTeamStats(ctx context.Context, contestID int64, contestMode string, teamIDs []int64) (map[int64]scoreboardTeamStats, error)
}

type scoreboardTeamStats struct {
	SolvedCount      int
	LastSubmissionAt *time.Time
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, contest *model.Contest) error {
	return r.db.WithContext(ctx).Create(contest).Error
}

func (r *repository) FindByID(ctx context.Context, id int64) (*model.Contest, error) {
	var contest model.Contest
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&contest).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContestNotFound
		}
		return nil, err
	}
	return &contest, nil
}

func (r *repository) Update(ctx context.Context, contest *model.Contest) error {
	return r.db.WithContext(ctx).Save(contest).Error
}

func (r *repository) List(ctx context.Context, status *string, offset, limit int) ([]*model.Contest, int64, error) {
	var contests []*model.Contest
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Contest{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&contests).Error
	return contests, total, err
}

func (r *repository) ListByStatusesAndTimeRange(ctx context.Context, statuses []string, now time.Time, offset, limit int) ([]*model.Contest, int64, error) {
	var contests []*model.Contest
	var total int64

	conditions := make([]string, 0, len(statuses))
	args := make([]any, 0, len(statuses)*2)
	for _, status := range statuses {
		switch status {
		case model.ContestStatusRegistration:
			conditions = append(conditions, "(status = ? AND start_time <= ?)")
			args = append(args, status, now)
		case model.ContestStatusRunning:
			conditions = append(conditions, "(status = ? AND ((freeze_time IS NOT NULL AND freeze_time <= ?) OR end_time <= ?))")
			args = append(args, status, now, now)
		case model.ContestStatusFrozen:
			conditions = append(conditions, "(status = ? AND end_time <= ?)")
			args = append(args, status, now)
		}
	}
	if len(conditions) == 0 {
		return contests, 0, nil
	}

	query := r.db.WithContext(ctx).Model(&model.Contest{}).
		Where(strings.Join(conditions, " OR "), args...)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Find(&contests).Error
	return contests, total, err
}

func (r *repository) UpdateStatus(ctx context.Context, id int64, status string) error {
	result := r.db.WithContext(ctx).Model(&model.Contest{}).
		Where("id = ? AND status != ?", id, status).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}

	// RowsAffected == 0 可能是不存在或状态已相同，需要区分
	if result.RowsAffected == 0 {
		var exists bool
		err := r.db.WithContext(ctx).Model(&model.Contest{}).
			Select("1").Where("id = ?", id).Limit(1).Find(&exists).Error
		if err != nil {
			return err
		}
		if !exists {
			return ErrContestNotFound
		}
	}

	return nil
}

func (r *repository) FindTeamsByIDs(ctx context.Context, ids []int64) ([]*model.Team, error) {
	if len(ids) == 0 {
		return []*model.Team{}, nil
	}

	var teams []*model.Team
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&teams).Error
	return teams, err
}

func (r *repository) FindTeamsByContest(ctx context.Context, contestID int64) ([]*model.Team, error) {
	var teams []*model.Team
	err := r.db.WithContext(ctx).
		Where("contest_id = ?", contestID).
		Order("id ASC").
		Find(&teams).Error
	return teams, err
}

type scoreboardTeamStatsRow struct {
	TeamID              int64  `gorm:"column:team_id"`
	SolvedCount         int    `gorm:"column:solved_count"`
	LastSubmissionAtRaw string `gorm:"column:last_submission_at"`
}

func (r *repository) FindScoreboardTeamStats(ctx context.Context, contestID int64, contestMode string, teamIDs []int64) (map[int64]scoreboardTeamStats, error) {
	result := make(map[int64]scoreboardTeamStats, len(teamIDs))
	if len(teamIDs) == 0 {
		return result, nil
	}

	var rows []scoreboardTeamStatsRow
	switch contestMode {
	case model.ContestModeAWD:
		if err := r.db.WithContext(ctx).
			Table("awd_attack_logs AS aal").
			Select("aal.attacker_team_id AS team_id, COUNT(*) AS solved_count, MAX(aal.created_at) AS last_submission_at").
			Joins("JOIN awd_rounds AS ar ON ar.id = aal.round_id").
			Where("ar.contest_id = ? AND aal.is_success = ? AND aal.source = ?", contestID, true, model.AWDAttackSourceSubmission).
			Where("aal.attacker_team_id IN ?", teamIDs).
			Group("aal.attacker_team_id").
			Scan(&rows).Error; err != nil {
			return nil, err
		}
	default:
		if err := r.db.WithContext(ctx).
			Table("submissions").
			Select("team_id AS team_id, COUNT(*) AS solved_count, MAX(submitted_at) AS last_submission_at").
			Where("contest_id = ? AND is_correct = ? AND team_id IS NOT NULL", contestID, true).
			Where("team_id IN ?", teamIDs).
			Group("team_id").
			Scan(&rows).Error; err != nil {
			return nil, err
		}
	}

	for _, row := range rows {
		result[row.TeamID] = scoreboardTeamStats{
			SolvedCount:      row.SolvedCount,
			LastSubmissionAt: parseContestAggregateTime(row.LastSubmissionAtRaw),
		}
	}
	return result, nil
}

func parseContestAggregateTime(raw string) *time.Time {
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
