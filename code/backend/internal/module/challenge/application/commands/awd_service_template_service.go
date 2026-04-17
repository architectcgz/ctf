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

type AWDServiceTemplateService struct {
	repo challengeports.AWDServiceTemplateCommandRepository
}

func NewAWDServiceTemplateService(repo challengeports.AWDServiceTemplateCommandRepository) *AWDServiceTemplateService {
	return &AWDServiceTemplateService{repo: repo}
}

func (s *AWDServiceTemplateService) CreateTemplate(_ context.Context, actorUserID int64, req *dto.CreateAWDServiceTemplateReq) (*dto.AWDServiceTemplateResp, error) {
	template := &model.AWDServiceTemplate{
		Name:            strings.TrimSpace(req.Name),
		Slug:            strings.TrimSpace(req.Slug),
		Category:        strings.TrimSpace(req.Category),
		Difficulty:      strings.TrimSpace(req.Difficulty),
		Description:     strings.TrimSpace(req.Description),
		ServiceType:     model.AWDServiceType(strings.TrimSpace(req.ServiceType)),
		DeploymentMode:  model.AWDDeploymentMode(strings.TrimSpace(req.DeploymentMode)),
		Version:         "v1",
		Status:          model.AWDServiceTemplateStatusDraft,
		ReadinessStatus: model.AWDReadinessStatusPending,
		CreatedBy:       &actorUserID,
	}
	if err := s.repo.CreateAWDServiceTemplate(template); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return domain.AWDServiceTemplateRespFromModel(template), nil
}

func (s *AWDServiceTemplateService) UpdateTemplate(_ context.Context, id int64, req *dto.UpdateAWDServiceTemplateReq) (*dto.AWDServiceTemplateResp, error) {
	template, err := s.repo.FindAWDServiceTemplateByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	if req.Name != "" {
		template.Name = strings.TrimSpace(req.Name)
	}
	if req.Slug != "" {
		template.Slug = strings.TrimSpace(req.Slug)
	}
	if req.Category != "" {
		template.Category = strings.TrimSpace(req.Category)
	}
	if req.Difficulty != "" {
		template.Difficulty = strings.TrimSpace(req.Difficulty)
	}
	if req.Description != "" {
		template.Description = strings.TrimSpace(req.Description)
	}
	if req.ServiceType != "" {
		template.ServiceType = model.AWDServiceType(strings.TrimSpace(req.ServiceType))
	}
	if req.DeploymentMode != "" {
		template.DeploymentMode = model.AWDDeploymentMode(strings.TrimSpace(req.DeploymentMode))
	}
	if req.Status != "" {
		template.Status = model.AWDServiceTemplateStatus(strings.TrimSpace(req.Status))
	}

	if err := s.repo.UpdateAWDServiceTemplate(template); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return domain.AWDServiceTemplateRespFromModel(template), nil
}

func (s *AWDServiceTemplateService) DeleteTemplate(_ context.Context, id int64) error {
	if _, err := s.repo.FindAWDServiceTemplateByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}
	if err := s.repo.DeleteAWDServiceTemplate(id); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}
