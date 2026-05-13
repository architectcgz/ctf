package commands

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeports "ctf-platform/internal/module/challenge/ports"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type ContestAWDServiceService struct {
	repo                 contestports.AWDServiceStore
	contestRepo          contestports.ContestLookupRepository
	contestChallengeRepo contestChallengeRelationRepository
	challengeRepo        challengecontracts.ContestChallengeContract
	awdChallengeRepo     challengeports.AWDChallengeQueryRepository
	previewTokenStore    contestports.AWDCheckerPreviewTokenStore
}

type contestChallengeRelationRepository interface {
	contestports.ContestChallengeWriteRepository
}

func NewContestAWDServiceService(
	repo contestports.AWDServiceStore,
	contestRepo contestports.ContestLookupRepository,
	contestChallengeRepo contestChallengeRelationRepository,
	challengeRepo challengecontracts.ContestChallengeContract,
	awdChallengeRepo challengeports.AWDChallengeQueryRepository,
	previewTokenStore contestports.AWDCheckerPreviewTokenStore,
) *ContestAWDServiceService {
	return &ContestAWDServiceService{
		repo:                 repo,
		contestRepo:          contestRepo,
		contestChallengeRepo: contestChallengeRepo,
		challengeRepo:        challengeRepo,
		awdChallengeRepo:     awdChallengeRepo,
		previewTokenStore:    previewTokenStore,
	}
}

func (s *ContestAWDServiceService) CreateContestAWDService(ctx context.Context, contestID int64, req CreateContestAWDServiceInput) (*dto.ContestAWDServiceResp, error) {
	contest, err := s.ensureMutableAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if req.Points < 1 || req.Points > contestdomain.AWDMaxServiceDisplayPoint {
		return nil, errcode.ErrInvalidParams
	}

	awdChallenge, err := s.awdChallengeRepo.FindAWDChallengeByID(ctx, req.AWDChallengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	isVisible := true
	if req.IsVisible != nil {
		isVisible = *req.IsVisible
	}
	checkerType, checkerConfig, slaScore, defenseScore, err := normalizeContestAWDServiceRuntimeFields(
		contest,
		awdChallenge.CheckerType,
		awdChallenge.CheckerConfig,
		req.CheckerType,
		req.CheckerConfig,
		0,
		0,
		req.AWDSLAScore,
		req.AWDDefenseScore,
	)
	if err != nil {
		return nil, err
	}
	previewToken := ""
	if req.AWDCheckerPreviewToken != nil {
		previewToken = *req.AWDCheckerPreviewToken
	}
	checkerTokenEnv := strings.TrimSpace(readStringFromAny(parseContestAWDServiceJSONMap(awdChallenge.RuntimeConfig)["checker_token_env"]))
	validationState, lastPreviewAt, lastPreviewResult, err := consumeCheckerPreviewValidationState(
		ctx,
		s.previewTokenStore,
		contestID,
		0,
		awdChallenge.ID,
		checkerType,
		checkerConfig,
		checkerTokenEnv,
		previewToken,
	)
	if err != nil {
		var appErr *errcode.AppError
		if errors.As(err, &appErr) {
			return nil, err
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if err := ensureCheckerPreviewTokenConsumed(previewToken, lastPreviewResult); err != nil {
		return nil, err
	}
	record := &model.ContestAWDService{
		ContestID:         contestID,
		AWDChallengeID:    req.AWDChallengeID,
		DisplayName:       firstNonEmpty(req.DisplayName, awdChallenge.Name),
		ServiceSnapshot:   buildContestAWDServiceSnapshot(awdChallenge),
		Order:             req.Order,
		IsVisible:         isVisible,
		ScoreConfig:       buildContestAWDServiceScoreConfig(req.Points, slaScore, defenseScore),
		ValidationState:   validationState,
		LastPreviewAt:     lastPreviewAt,
		LastPreviewResult: lastPreviewResult,
		RuntimeConfig: buildContestAWDServiceRuntimeConfig(
			checkerType,
			checkerConfig,
			awdChallenge.RuntimeConfig,
		),
	}
	if err := s.repo.CreateContestAWDService(ctx, record); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return contestAWDServiceRespFromModel(record), nil
}

func (s *ContestAWDServiceService) UpdateContestAWDService(ctx context.Context, contestID, serviceID int64, req UpdateContestAWDServiceInput) error {
	contest, err := s.ensureMutableAWDContest(ctx, contestID)
	if err != nil {
		return err
	}

	stored, err := s.repo.FindContestAWDServiceByContestAndID(ctx, contestID, serviceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}

	displayName := stored.DisplayName
	if req.DisplayName != nil {
		displayName = strings.TrimSpace(*req.DisplayName)
	}
	order := stored.Order
	if req.Order != nil {
		order = *req.Order
	}
	isVisible := stored.IsVisible
	if req.IsVisible != nil {
		isVisible = *req.IsVisible
	}

	updates := map[string]any{
		"display_name": displayName,
		"order":        order,
		"is_visible":   isVisible,
	}

	defaultCheckerType, defaultCheckerConfig := parseContestAWDServiceRuntimeChecker(stored.RuntimeConfig)
	extraRuntimeConfig := parseContestAWDChallengeRuntimeConfig(stored.RuntimeConfig)
	currentPoints, ok := parseContestAWDServiceScore(stored.ScoreConfig, "points")
	if !ok {
		currentPoints = 0
	}
	currentSLAScore, _ := parseContestAWDServiceScore(stored.ScoreConfig, "awd_sla_score")
	currentDefenseScore, _ := parseContestAWDServiceScore(stored.ScoreConfig, "awd_defense_score")

	if req.AWDChallengeID != nil {
		awdChallenge, err := s.awdChallengeRepo.FindAWDChallengeByID(ctx, *req.AWDChallengeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errcode.ErrNotFound
			}
			return errcode.ErrInternal.WithCause(err)
		}
		updates["awd_challenge_id"] = *req.AWDChallengeID
		updates["service_snapshot"] = buildContestAWDServiceSnapshot(awdChallenge)
		defaultCheckerType = awdChallenge.CheckerType
		defaultCheckerConfig = awdChallenge.CheckerConfig
		extraRuntimeConfig = awdChallenge.RuntimeConfig
		if req.DisplayName == nil || strings.TrimSpace(displayName) == "" {
			updates["display_name"] = firstNonEmpty(awdChallenge.Name)
		}
	}

	checkerType, checkerConfig, slaScore, defenseScore, err := normalizeContestAWDServiceRuntimeFields(
		contest,
		defaultCheckerType,
		defaultCheckerConfig,
		req.CheckerType,
		req.CheckerConfig,
		currentSLAScore,
		currentDefenseScore,
		req.AWDSLAScore,
		req.AWDDefenseScore,
	)
	if err != nil {
		return err
	}
	if req.AWDChallengeID != nil || req.CheckerType != nil || req.CheckerConfig != nil {
		updates["runtime_config"] = buildContestAWDServiceRuntimeConfig(
			checkerType,
			checkerConfig,
			extraRuntimeConfig,
		)
	}
	if req.Points != nil || req.AWDSLAScore != nil || req.AWDDefenseScore != nil {
		if req.Points != nil {
			if *req.Points < 1 || *req.Points > contestdomain.AWDMaxServiceDisplayPoint {
				return errcode.ErrInvalidParams
			}
			currentPoints = *req.Points
		}
		updates["score_config"] = buildContestAWDServiceScoreConfig(currentPoints, slaScore, defenseScore)
	}
	previewToken := ""
	if req.AWDCheckerPreviewToken != nil {
		previewToken = *req.AWDCheckerPreviewToken
	}
	validationState, lastPreviewAt, lastPreviewResult, validationChanged, err := buildContestAWDServiceValidationUpdate(
		ctx,
		s.previewTokenStore,
		stored,
		contestID,
		checkerType,
		checkerConfig,
		strings.TrimSpace(readStringFromAny(parseContestAWDServiceJSONMap(extraRuntimeConfig)["checker_token_env"])),
		previewToken,
	)
	if err != nil {
		var appErr *errcode.AppError
		if errors.As(err, &appErr) {
			return err
		}
		return errcode.ErrInternal.WithCause(err)
	}
	if validationChanged {
		updates["awd_checker_validation_state"] = validationState
		updates["awd_checker_last_preview_at"] = lastPreviewAt
		updates["awd_checker_last_preview_result"] = lastPreviewResult
	}

	if err := s.repo.UpdateContestAWDServiceByContestAndID(ctx, contestID, serviceID, updates); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *ContestAWDServiceService) DeleteContestAWDService(ctx context.Context, contestID, serviceID int64) error {
	_, err := s.ensureMutableAWDContest(ctx, contestID)
	if err != nil {
		return err
	}

	_, err = s.repo.FindContestAWDServiceByContestAndID(ctx, contestID, serviceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}

	if err := s.repo.DeleteContestAWDServiceByContestAndID(ctx, contestID, serviceID); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *ContestAWDServiceService) ensureMutableAWDContest(ctx context.Context, contestID int64) (*model.Contest, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if contest.Mode != model.ContestModeAWD {
		return nil, errcode.ErrInvalidParams
	}
	return contest, nil
}

func (s *ContestAWDServiceService) ensureContestChallengeRelation(ctx context.Context, contestID, challengeID int64, points, order int, isVisible bool) (bool, error) {
	if s.contestChallengeRepo == nil {
		return false, nil
	}
	exists, err := s.contestChallengeRepo.Exists(ctx, contestID, challengeID)
	if err != nil {
		return false, errcode.ErrInternal.WithCause(err)
	}
	if exists {
		return false, nil
	}
	relation := &model.ContestChallenge{
		ContestID:   contestID,
		ChallengeID: challengeID,
		Points:      points,
		Order:       order,
		IsVisible:   isVisible,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := s.contestChallengeRepo.AddChallenge(ctx, relation); err != nil {
		return false, errcode.ErrInternal.WithCause(err)
	}
	return true, nil
}

func (s *ContestAWDServiceService) syncContestChallengeRelation(ctx context.Context, contest *model.Contest, challengeID int64, order int, isVisible bool) error {
	if s.contestChallengeRepo == nil || contest == nil {
		return nil
	}
	exists, err := s.contestChallengeRepo.Exists(ctx, contest.ID, challengeID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if !exists {
		challenge, err := s.challengeRepo.FindByID(ctx, challengeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errcode.ErrChallengeNotFound
			}
			return errcode.ErrInternal.WithCause(err)
		}
		_, err = s.ensureContestChallengeRelation(ctx, contest.ID, challengeID, challenge.Points, order, isVisible)
		return err
	}
	if err := s.contestChallengeRepo.UpdateChallenge(ctx, contest.ID, challengeID, map[string]any{
		"order":      order,
		"is_visible": isVisible,
	}); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
