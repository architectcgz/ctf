package commands

import (
	"context"
	"errors"
	"strings"

	"ctf-platform/internal/apperror"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"ctf-platform/internal/module/challenge/domain"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type AWDChallengeService struct {
	repo challengeports.AWDChallengeCommandRepository
}

func NewAWDChallengeService(repo challengeports.AWDChallengeCommandRepository) *AWDChallengeService {
	return &AWDChallengeService{repo: repo}
}

func (s *AWDChallengeService) CreateChallenge(ctx context.Context, actorUserID int64, req CreateAWDChallengeInput) (*challengecontracts.AWDChallengeResp, error) {
	challenge := &challengeentity.AWDChallenge{
		Name:            strings.TrimSpace(req.Name),
		Slug:            strings.TrimSpace(req.Slug),
		Category:        strings.TrimSpace(req.Category),
		Difficulty:      strings.TrimSpace(req.Difficulty),
		Description:     strings.TrimSpace(req.Description),
		ServiceType:     challengeentity.AWDServiceType(strings.TrimSpace(req.ServiceType)),
		DeploymentMode:  challengeentity.AWDDeploymentMode(strings.TrimSpace(req.DeploymentMode)),
		Version:         "v1",
		Status:          challengeentity.AWDChallengeStatusDraft,
		ReadinessStatus: challengeentity.AWDReadinessStatusPending,
		CreatedBy:       &actorUserID,
	}
	if err := s.repo.CreateAWDChallenge(ctx, challenge); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	return domain.AWDChallengeRespFromModel(challenge), nil
}

func (s *AWDChallengeService) UpdateChallenge(ctx context.Context, id int64, req UpdateAWDChallengeInput) (*challengecontracts.AWDChallengeResp, error) {
	challenge, err := s.repo.FindAWDChallengeByID(ctx, id)
	if err != nil {
		if errors.Is(err, challengeports.ErrAWDChallengeNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
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
		challenge.ServiceType = challengeentity.AWDServiceType(strings.TrimSpace(req.ServiceType))
	}
	if req.DeploymentMode != "" {
		challenge.DeploymentMode = challengeentity.AWDDeploymentMode(strings.TrimSpace(req.DeploymentMode))
	}
	if req.Status != "" {
		challenge.Status = challengeentity.AWDChallengeStatus(strings.TrimSpace(req.Status))
	}

	if err := s.repo.UpdateAWDChallenge(ctx, challenge); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	return domain.AWDChallengeRespFromModel(challenge), nil
}

func (s *AWDChallengeService) DeleteChallenge(ctx context.Context, id int64) error {
	if _, err := s.repo.FindAWDChallengeByID(ctx, id); err != nil {
		if errors.Is(err, challengeports.ErrAWDChallengeNotFound) {
			return apperror.ErrNotFound
		}
		return apperror.ErrInternal.WithCause(err)
	}
	if err := s.repo.DeleteAWDChallenge(ctx, id); err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	return nil
}
