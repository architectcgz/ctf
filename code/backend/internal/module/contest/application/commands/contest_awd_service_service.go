package commands

import (
	"context"
	"errors"
	"strings"

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
	repo          contestports.AWDRepository
	contestRepo   contestports.ContestLookupRepository
	challengeRepo challengecontracts.ContestChallengeContract
	templateRepo  challengeports.AWDServiceTemplateQueryRepository
}

func NewContestAWDServiceService(
	repo contestports.AWDRepository,
	contestRepo contestports.ContestLookupRepository,
	challengeRepo challengecontracts.ContestChallengeContract,
	templateRepo challengeports.AWDServiceTemplateQueryRepository,
) *ContestAWDServiceService {
	return &ContestAWDServiceService{
		repo:          repo,
		contestRepo:   contestRepo,
		challengeRepo: challengeRepo,
		templateRepo:  templateRepo,
	}
}

func (s *ContestAWDServiceService) CreateContestAWDService(ctx context.Context, contestID int64, req *dto.CreateContestAWDServiceReq) (*dto.ContestAWDServiceResp, error) {
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

	challenge, err := s.challengeRepo.FindByID(req.ChallengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	template, err := s.templateRepo.FindAWDServiceTemplateByID(req.TemplateID)
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
	record := &model.ContestAWDService{
		ContestID:       contestID,
		ChallengeID:     req.ChallengeID,
		TemplateID:      &req.TemplateID,
		DisplayName:     firstNonEmpty(req.DisplayName, template.Name, challenge.Title),
		Order:           req.Order,
		IsVisible:       isVisible,
		ScoreConfig:     buildContestAWDServiceScoreConfig(challenge.Points, 0, 0),
		ValidationState: model.AWDCheckerValidationStatePending,
		RuntimeConfig: buildContestAWDServiceRuntimeConfig(
			req.ChallengeID,
			template.CheckerType,
			template.CheckerConfig,
			template.RuntimeConfig,
		),
	}
	if err := s.repo.CreateContestAWDService(ctx, record); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return contestdomain.ContestAWDServiceRespFromModel(record), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
