package infrastructure

import (
	"context"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ctf-platform/internal/model"
	contestapp "ctf-platform/internal/module/contest/application"
)

type AWDRepository struct {
	db *gorm.DB
}

func NewAWDRepository(db *gorm.DB) *AWDRepository {
	return &AWDRepository{db: db}
}

func (r *AWDRepository) WithDB(db *gorm.DB) *AWDRepository {
	return &AWDRepository{db: db}
}

func (r *AWDRepository) dbWithContext(ctx context.Context) *gorm.DB {
	if ctx == nil {
		ctx = context.Background()
	}
	return r.db.WithContext(ctx)
}

func (r *AWDRepository) WithinTransaction(ctx context.Context, fn func(txRepo contestapp.AWDRepository) error) error {
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(r.WithDB(tx))
	})
}

func (r *AWDRepository) CreateRound(ctx context.Context, round *model.AWDRound) error {
	return r.dbWithContext(ctx).Create(round).Error
}

func (r *AWDRepository) UpsertRound(ctx context.Context, round *model.AWDRound) error {
	return r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "contest_id"},
			{Name: "round_number"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"status":     round.Status,
			"started_at": round.StartedAt,
			"ended_at":   round.EndedAt,
			"updated_at": time.Now(),
		}),
	}).Create(round).Error
}

func (r *AWDRepository) ListRoundsByContest(ctx context.Context, contestID int64) ([]model.AWDRound, error) {
	var rounds []model.AWDRound
	err := r.dbWithContext(ctx).
		Where("contest_id = ?", contestID).
		Order("round_number ASC, id ASC").
		Find(&rounds).Error
	return rounds, err
}

func (r *AWDRepository) FindRoundByContestAndID(ctx context.Context, contestID, roundID int64) (*model.AWDRound, error) {
	var round model.AWDRound
	if err := r.dbWithContext(ctx).
		Where("id = ? AND contest_id = ?", roundID, contestID).
		First(&round).Error; err != nil {
		return nil, err
	}
	return &round, nil
}

func (r *AWDRepository) FindRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*model.AWDRound, error) {
	var round model.AWDRound
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND round_number = ?", contestID, roundNumber).
		First(&round).Error; err != nil {
		return nil, err
	}
	return &round, nil
}

func (r *AWDRepository) FindRunningRound(ctx context.Context, contestID int64) (*model.AWDRound, error) {
	var round model.AWDRound
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND status = ?", contestID, model.AWDRoundStatusRunning).
		Order("round_number DESC, id DESC").
		First(&round).Error; err != nil {
		return nil, err
	}
	return &round, nil
}

func (r *AWDRepository) ListSchedulableAWDContests(ctx context.Context, now, recentCutoff time.Time, limit int) ([]model.Contest, error) {
	var contests []model.Contest
	query := r.dbWithContext(ctx).
		Where("mode = ?", model.ContestModeAWD).
		Where("status IN ?", []string{
			model.ContestStatusRegistration,
			model.ContestStatusRunning,
			model.ContestStatusFrozen,
			model.ContestStatusEnded,
		}).
		Where("start_time <= ?", now).
		Where("end_time > ?", recentCutoff).
		Order("start_time ASC, id ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&contests).Error; err != nil {
		return nil, err
	}
	return contests, nil
}

func (r *AWDRepository) FindTeamsByContest(ctx context.Context, contestID int64) ([]*model.Team, error) {
	var teams []*model.Team
	err := r.dbWithContext(ctx).
		Where("contest_id = ?", contestID).
		Find(&teams).Error
	return teams, err
}

func (r *AWDRepository) ListChallengesByContest(ctx context.Context, contestID int64) ([]model.Challenge, error) {
	var challenges []model.Challenge
	if err := r.dbWithContext(ctx).
		Table("challenges AS c").
		Select("c.*").
		Joins("JOIN contest_challenges AS cc ON cc.challenge_id = c.id").
		Where("cc.contest_id = ?", contestID).
		Order("c.id ASC").
		Scan(&challenges).Error; err != nil {
		return nil, err
	}
	return challenges, nil
}

func (r *AWDRepository) ContestHasChallenge(ctx context.Context, contestID, challengeID int64) (bool, error) {
	var count int64
	if err := r.dbWithContext(ctx).
		Model(&model.ContestChallenge{}).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *AWDRepository) FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	var registration model.ContestRegistration
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		First(&registration).Error; err != nil {
		return nil, err
	}
	return &registration, nil
}

func (r *AWDRepository) FindContestTeamByMember(ctx context.Context, contestID, userID int64) (*model.Team, error) {
	var team model.Team
	if err := r.dbWithContext(ctx).
		Table("teams AS t").
		Select("t.*").
		Joins("JOIN team_members AS tm ON tm.team_id = t.id").
		Where("t.contest_id = ? AND tm.user_id = ?", contestID, userID).
		First(&team).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *AWDRepository) FindChallengeByID(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	var challenge model.Challenge
	if err := r.dbWithContext(ctx).Where("id = ?", challengeID).First(&challenge).Error; err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (r *AWDRepository) ListServiceInstancesByContest(ctx context.Context, contestID int64, challengeIDs []int64) ([]contestapp.AWDServiceInstance, error) {
	if len(challengeIDs) == 0 {
		return nil, nil
	}

	var instances []contestapp.AWDServiceInstance
	if err := r.dbWithContext(ctx).
		Table("instances AS inst").
		Select("COALESCE(inst.team_id, tm.team_id) AS team_id, inst.challenge_id AS challenge_id, inst.access_url AS access_url").
		Joins("LEFT JOIN team_members AS tm ON tm.user_id = inst.user_id AND tm.contest_id = ?", contestID).
		Where("inst.challenge_id IN ?", challengeIDs).
		Where("inst.status = ?", model.InstanceStatusRunning).
		Where("(inst.contest_id = ? AND inst.team_id IS NOT NULL) OR (inst.team_id IS NULL AND tm.team_id IS NOT NULL)", contestID).
		Order("tm.team_id ASC, inst.challenge_id ASC, inst.id ASC").
		Scan(&instances).Error; err != nil {
		return nil, err
	}
	return instances, nil
}

func (r *AWDRepository) UpsertServiceCheck(
	ctx context.Context,
	roundID, teamID, challengeID int64,
	serviceStatus, checkResult string,
	defenseScore int,
	updatedAt time.Time,
) (*model.AWDTeamService, error) {
	record := &model.AWDTeamService{
		RoundID:       roundID,
		TeamID:        teamID,
		ChallengeID:   challengeID,
		ServiceStatus: serviceStatus,
		CheckResult:   checkResult,
		DefenseScore:  defenseScore,
	}
	if err := r.dbWithContext(ctx).
		Where("round_id = ? AND team_id = ? AND challenge_id = ?", roundID, teamID, challengeID).
		Assign(map[string]any{
			"service_status": serviceStatus,
			"check_result":   checkResult,
			"defense_score":  defenseScore,
			"updated_at":     updatedAt,
		}).
		FirstOrCreate(record).Error; err != nil {
		return nil, err
	}
	return record, nil
}

func (r *AWDRepository) UpsertTeamServices(ctx context.Context, records []model.AWDTeamService) error {
	if len(records) == 0 {
		return nil
	}
	return r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "round_id"},
			{Name: "team_id"},
			{Name: "challenge_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"service_status",
			"check_result",
			"defense_score",
			"updated_at",
		}),
	}).Create(&records).Error
}

func (r *AWDRepository) ListServicesByRound(ctx context.Context, roundID int64) ([]model.AWDTeamService, error) {
	var records []model.AWDTeamService
	err := r.dbWithContext(ctx).
		Where("round_id = ?", roundID).
		Order("team_id ASC, challenge_id ASC").
		Find(&records).Error
	return records, err
}

func (r *AWDRepository) CountSuccessfulAttacks(ctx context.Context, roundID, attackerTeamID, victimTeamID, challengeID int64) (int64, error) {
	var count int64
	err := r.dbWithContext(ctx).
		Model(&model.AWDAttackLog{}).
		Where(
			"round_id = ? AND attacker_team_id = ? AND victim_team_id = ? AND challenge_id = ? AND is_success = ?",
			roundID, attackerTeamID, victimTeamID, challengeID, true,
		).
		Count(&count).Error
	return count, err
}

func (r *AWDRepository) CreateAttackLog(ctx context.Context, logRecord *model.AWDAttackLog) error {
	return r.dbWithContext(ctx).Create(logRecord).Error
}

func (r *AWDRepository) ApplyAttackImpactToVictimService(
	ctx context.Context,
	roundID, victimTeamID, challengeID int64,
	scoreGained int,
	updatedAt time.Time,
) error {
	record := &model.AWDTeamService{
		RoundID:        roundID,
		TeamID:         victimTeamID,
		ChallengeID:    challengeID,
		ServiceStatus:  model.AWDServiceStatusCompromised,
		CheckResult:    "{}",
		AttackReceived: 1,
		DefenseScore:   0,
		AttackScore:    scoreGained,
		CreatedAt:      updatedAt,
		UpdatedAt:      updatedAt,
	}
	return r.dbWithContext(ctx).
		Where("round_id = ? AND team_id = ? AND challenge_id = ?", roundID, victimTeamID, challengeID).
		Assign(map[string]any{
			"service_status":  model.AWDServiceStatusCompromised,
			"attack_received": gorm.Expr("attack_received + ?", 1),
			"attack_score":    gorm.Expr("attack_score + ?", scoreGained),
			"defense_score":   0,
			"updated_at":      updatedAt,
		}).
		FirstOrCreate(record).Error
}

func (r *AWDRepository) ListAttackLogsByRound(ctx context.Context, roundID int64) ([]model.AWDAttackLog, error) {
	var logs []model.AWDAttackLog
	err := r.dbWithContext(ctx).
		Where("round_id = ?", roundID).
		Order("created_at ASC, id ASC").
		Find(&logs).Error
	return logs, err
}

func (r *AWDRepository) RecalculateContestTeamScores(ctx context.Context, contestID int64) error {
	return RecalculateAWDContestTeamScores(ctx, r.db, contestID)
}

func (r *AWDRepository) RebuildContestScoreboardCache(ctx context.Context, redis *redislib.Client, contestID int64) error {
	return RebuildContestScoreboardCache(ctx, r.db, redis, contestID)
}
