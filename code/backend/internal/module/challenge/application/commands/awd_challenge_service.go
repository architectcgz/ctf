package commands

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

type AWDChallengeService struct {
	repo challengeports.AWDChallengeCommandRepository
}

func NewAWDChallengeService(repo challengeports.AWDChallengeCommandRepository) *AWDChallengeService {
	return &AWDChallengeService{repo: repo}
}

func (s *AWDChallengeService) CreateChallenge(ctx context.Context, actorUserID int64, req CreateAWDChallengeInput) (*dto.AWDChallengeResp, error) {
	challenge := &model.AWDChallenge{
		Name:            strings.TrimSpace(req.Name),
		Slug:            strings.TrimSpace(req.Slug),
		Category:        strings.TrimSpace(req.Category),
		Difficulty:      strings.TrimSpace(req.Difficulty),
		Description:     strings.TrimSpace(req.Description),
		ServiceType:     model.AWDServiceType(strings.TrimSpace(req.ServiceType)),
		DeploymentMode:  model.AWDDeploymentMode(strings.TrimSpace(req.DeploymentMode)),
		Version:         "v1",
		Status:          model.AWDChallengeStatusDraft,
		ReadinessStatus: model.AWDReadinessStatusPending,
		CreatedBy:       &actorUserID,
	}
	if err := s.repo.CreateAWDChallenge(ctx, challenge); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return domain.AWDChallengeRespFromModel(challenge), nil
}

func (s *AWDChallengeService) UpdateChallenge(ctx context.Context, id int64, req UpdateAWDChallengeInput) (*dto.AWDChallengeResp, error) {
	challenge, err := s.repo.FindAWDChallengeByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	if req.Name != "" {
		challenge.Name = strings.TrimSpace(req.Name)
	}
	if req.Slug != "" {
		challenge.Slug = strings.TrimSpace(req.Slug)
	}
	if req.Category != "" {
		challenge.Category = strings.TrimSpace(req.Category)
	}
	if req.Difficulty != "" {
		challenge.Difficulty = strings.TrimSpace(req.Difficulty)
	}
	if req.Description != "" {
		challenge.Description = strings.TrimSpace(req.Description)
	}
	if req.ServiceType != "" {
		challenge.ServiceType = model.AWDServiceType(strings.TrimSpace(req.ServiceType))
	}
	if req.DeploymentMode != "" {
		challenge.DeploymentMode = model.AWDDeploymentMode(strings.TrimSpace(req.DeploymentMode))
	}
	if req.Status != "" {
		challenge.Status = model.AWDChallengeStatus(strings.TrimSpace(req.Status))
	}

	if err := s.repo.UpdateAWDChallenge(ctx, challenge); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return domain.AWDChallengeRespFromModel(challenge), nil
}

func (s *AWDChallengeService) DeleteChallenge(ctx context.Context, id int64) error {
	if _, err := s.repo.FindAWDChallengeByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}
	if err := s.repo.DeleteAWDChallenge(ctx, id); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}
