package contest

import (
	"ctf-platform/internal/config"
	"ctf-platform/internal/constants"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge"
	"ctf-platform/pkg/errcode"
	"math"
	"time"

	"gorm.io/gorm"
)

type SubmissionService struct {
	db          *gorm.DB
	flagService *challenge.FlagService
	cfg         *config.Config
}

func NewSubmissionService(db *gorm.DB, flagService *challenge.FlagService, cfg *config.Config) *SubmissionService {
	return &SubmissionService{
		db:          db,
		flagService: flagService,
		cfg:         cfg,
	}
}

// SubmitFlagInContest 竞赛中提交 Flag
func (s *SubmissionService) SubmitFlagInContest(userID, contestID, challengeID int64, flag string) (*dto.SubmissionResp, error) {
	// 前置校验
	var contest model.Contest
	if err := s.db.First(&contest, contestID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrContestNotFound
		}
		return nil, err
	}

	if contest.Status != "running" {
		return nil, errcode.ErrContestNotRunning
	}

	now := time.Now()
	if now.Before(contest.StartAt) {
		return nil, errcode.ErrContestNotStarted
	}
	if now.After(contest.EndAt) {
		return nil, errcode.ErrContestEnded
	}

	var reg model.ContestRegistration
	if err := s.db.Where("contest_id = ? AND user_id = ?", contestID, userID).First(&reg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrNotRegistered
		}
		return nil, err
	}

	if reg.Status != "approved" {
		return nil, errcode.ErrRegistrationNotApproved
	}

	var cc model.ContestChallenge
	if err := s.db.Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).First(&cc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrChallengeNotInContest
		}
		return nil, err
	}

	// 验证 Flag
	isCorrect, err := s.flagService.ValidateFlag(userID, challengeID, flag, "")
	if err != nil {
		return nil, err
	}

	var submission *model.Submission
	var finalScore int

	if isCorrect {
		// 在事务中处理首杀、计分和提交记录创建
		err = s.db.Transaction(func(tx *gorm.DB) error {
			// 事务内二次检查是否已解决（防止并发重复提交）
			var count int64
			tx.Model(&model.Submission{}).
				Where("user_id = ? AND challenge_id = ? AND contest_id = ? AND is_correct = ?",
					userID, challengeID, contestID, true).
				Count(&count)

			if count > 0 {
				return errcode.ErrContestChallengeAlreadySolved
			}

			// 获取基础分数
			var chal model.Challenge
			if err := tx.First(&chal, challengeID).Error; err != nil {
				return err
			}

			baseScore := chal.Points
			if cc.ContestScore != nil {
				baseScore = *cc.ContestScore
			}

			// 使用乐观锁抢首杀
			result := tx.Model(&model.ContestChallenge{}).
				Where("contest_id = ? AND challenge_id = ? AND first_blood_by IS NULL",
					contestID, challengeID).
				Update("first_blood_by", userID)

			isFirstBlood := result.RowsAffected > 0

			// 计算最终分数
			finalScore = baseScore
			if isFirstBlood {
				bonus := int(math.Round(float64(baseScore) * s.cfg.Contest.FirstBloodBonus))
				finalScore += bonus
			}

			// 更新团队分数
			if reg.TeamID != nil {
				if err := tx.Model(&model.Team{}).
					Where("id = ?", *reg.TeamID).
					Updates(map[string]interface{}{
						"total_score":   gorm.Expr("total_score + ?", finalScore),
						"last_solve_at": now,
					}).Error; err != nil {
					return err
				}
			}

			// 创建提交记录
			submission = &model.Submission{
				UserID:      userID,
				ChallengeID: challengeID,
				ContestID:   &contestID,
				TeamID:      reg.TeamID,
				Flag:        flag,
				IsCorrect:   true,
				Score:       finalScore,
				SubmittedAt: now,
			}

			return tx.Create(submission).Error
		})

		if err != nil {
			return nil, err
		}
	} else {
		// 错误提交直接创建记录
		submission = &model.Submission{
			UserID:      userID,
			ChallengeID: challengeID,
			ContestID:   &contestID,
			TeamID:      reg.TeamID,
			Flag:        flag,
			IsCorrect:   false,
			Score:       0,
			SubmittedAt: now,
		}

		if err := s.db.Create(submission).Error; err != nil {
			return nil, err
		}
	}

	message := constants.MsgFlagIncorrect
	if isCorrect {
		message = constants.MsgFlagCorrect
	}

	return &dto.SubmissionResp{
		IsCorrect:   isCorrect,
		Message:     message,
		Points:      finalScore,
		SubmittedAt: submission.SubmittedAt,
	}, nil
}
