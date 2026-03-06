package contest

import (
	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge"
	"ctf-platform/pkg/errcode"
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
	var contest model.Contest
	if err := s.db.First(&contest, contestID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrContestNotFound
		}
		return nil, err
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

	var cc model.ContestChallenge
	if err := s.db.Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).First(&cc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrChallengeNotInContest
		}
		return nil, err
	}

	var existingSub model.Submission
	err := s.db.Where("user_id = ? AND challenge_id = ? AND contest_id = ? AND is_correct = ?",
		userID, challengeID, contestID, true).First(&existingSub).Error
	if err == nil {
		return nil, errcode.ErrAlreadySolved
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	isCorrect, err := s.flagService.ValidateFlag(userID, challengeID, flag, "")
	if err != nil {
		return nil, err
	}

	score := 0
	if isCorrect {
		var chal model.Challenge
		if err := s.db.First(&chal, challengeID).Error; err != nil {
			return nil, err
		}

		if cc.ContestScore != nil {
			score = *cc.ContestScore
		} else {
			score = chal.Points
		}

		isFirstBlood := cc.FirstBloodBy == nil
		if isFirstBlood {
			bonus := int(float64(score) * s.cfg.Contest.FirstBloodBonus)
			score += bonus
		}

		if err := s.updateScoreAndFirstBlood(userID, contestID, challengeID, score, isFirstBlood, reg.TeamID); err != nil {
			return nil, err
		}
	}

	submission := &model.Submission{
		UserID:      userID,
		ChallengeID: challengeID,
		ContestID:   &contestID,
		TeamID:      reg.TeamID,
		Flag:        flag,
		IsCorrect:   isCorrect,
		Score:       score,
		SubmittedAt: now,
	}

	if err := s.db.Create(submission).Error; err != nil {
		return nil, err
	}

	message := "Flag 错误"
	if isCorrect {
		message = "恭喜，Flag 正确！"
	}

	return &dto.SubmissionResp{
		IsCorrect:   isCorrect,
		Message:     message,
		Points:      score,
		SubmittedAt: submission.SubmittedAt,
	}, nil
}

func (s *SubmissionService) updateScoreAndFirstBlood(userID, contestID, challengeID int64, score int, isFirstBlood bool, teamID *int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if isFirstBlood {
			if err := tx.Model(&model.ContestChallenge{}).
				Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
				Update("first_blood_by", userID).Error; err != nil {
				return err
			}
		}

		if teamID != nil {
			now := time.Now()
			if err := tx.Model(&model.Team{}).
				Where("id = ?", *teamID).
				Updates(map[string]interface{}{
					"total_score":   gorm.Expr("total_score + ?", score),
					"last_solve_at": now,
				}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
