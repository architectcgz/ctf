package contest

import (
	"context"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
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

func (r *AWDRepository) WithinTransaction(ctx context.Context, fn func(txRepo *AWDRepository) error) error {
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(r.WithDB(tx))
	})
}

func (r *AWDRepository) CreateRound(ctx context.Context, round *model.AWDRound) error {
	return r.dbWithContext(ctx).Create(round).Error
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

func (r *AWDRepository) FindTeamsByContest(ctx context.Context, contestID int64) ([]*model.Team, error) {
	var teams []*model.Team
	err := r.dbWithContext(ctx).
		Where("contest_id = ?", contestID).
		Find(&teams).Error
	return teams, err
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
	return recalculateAWDContestTeamScores(ctx, r.db, contestID)
}

func (r *AWDRepository) RebuildContestScoreboardCache(ctx context.Context, redis *redislib.Client, contestID int64) error {
	return rebuildContestScoreboardCache(ctx, r.db, redis, contestID)
}

func (r *AWDRepository) RunRoundServiceChecks(
	ctx context.Context,
	redis *redislib.Client,
	cfg config.ContestAWDConfig,
	flagSecret string,
	contest *model.Contest,
	round *model.AWDRound,
	source string,
	log *zap.Logger,
) error {
	checker := NewAWDRoundUpdater(r.db, redis, cfg, flagSecret, nil, log)
	return checker.RunRoundServiceChecks(ctx, contest, round, source)
}

func (r *AWDRepository) EnsureActiveRoundMaterialized(
	ctx context.Context,
	redis *redislib.Client,
	cfg config.ContestAWDConfig,
	flagSecret string,
	contest *model.Contest,
	now time.Time,
	log *zap.Logger,
) error {
	updater := NewAWDRoundUpdater(r.db, redis, cfg, flagSecret, nil, log)
	activeRound, totalRounds, ok := updater.calculateRoundPlan(contest, now)
	if !ok || activeRound <= 0 {
		return gorm.ErrRecordNotFound
	}
	if err := updater.reconcileRounds(ctx, contest, activeRound, totalRounds); err != nil {
		return err
	}
	return updater.syncRoundFlags(ctx, contest, activeRound, now)
}
