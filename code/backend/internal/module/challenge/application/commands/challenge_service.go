package commands

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

type ChallengeService struct {
	repo      challengeports.ChallengeRepository
	imageRepo challengeports.ImageRepository
}

func NewChallengeService(repo challengeports.ChallengeRepository, imageRepo challengeports.ImageRepository) *ChallengeService {
	return &ChallengeService{
		repo:      repo,
		imageRepo: imageRepo,
	}
}

func (s *ChallengeService) CreateChallenge(req *dto.CreateChallengeReq) (*dto.ChallengeResp, error) {
	if req.ImageID > 0 {
		if _, err := s.imageRepo.FindByID(req.ImageID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errcode.ErrNotFound.WithCause(errors.New(domain.ErrMsgImageNotFound))
			}
			return nil, err
		}
	}

	challenge := &model.Challenge{
		Title:         req.Title,
		Description:   req.Description,
		Category:      req.Category,
		Difficulty:    req.Difficulty,
		Points:        req.Points,
		ImageID:       req.ImageID,
		AttachmentURL: strings.TrimSpace(req.AttachmentURL),
		Status:        model.ChallengeStatusDraft,
	}

	hints, err := domain.NormalizeHintModels(req.Hints)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateWithHints(challenge, hints); err != nil {
		return nil, err
	}
	return domain.ChallengeRespFromModel(challenge, hints), nil
}

func (s *ChallengeService) UpdateChallenge(id int64, req *dto.UpdateChallengeReq) error {
	challenge, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrChallengeNotFound
		}
		return err
	}

	if req.Title != "" {
		challenge.Title = req.Title
	}
	if req.Description != "" {
		challenge.Description = req.Description
	}
	if req.Category != "" {
		challenge.Category = req.Category
	}
	if req.Difficulty != "" {
		challenge.Difficulty = req.Difficulty
	}
	if req.Points > 0 {
		challenge.Points = req.Points
	}
	if req.ImageID != nil {
		if *req.ImageID > 0 {
			if _, err := s.imageRepo.FindByID(*req.ImageID); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errcode.ErrNotFound.WithCause(errors.New(domain.ErrMsgImageNotFound))
				}
				return err
			}
		}
		challenge.ImageID = *req.ImageID
	}
	if req.AttachmentURL != nil {
		challenge.AttachmentURL = strings.TrimSpace(*req.AttachmentURL)
	}

	replaceHints := req.Hints != nil
	hints, err := domain.NormalizeHintModels(req.Hints)
	if err != nil {
		return err
	}

	return s.repo.UpdateWithHints(challenge, hints, replaceHints)
}

func (s *ChallengeService) DeleteChallenge(id int64) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrChallengeNotFound
		}
		return err
	}

	hasInstances, err := s.repo.HasRunningInstances(id)
	if err != nil {
		return err
	}
	if hasInstances {
		return errcode.ErrConflict.WithCause(errors.New(domain.ErrMsgHasRunningInstances))
	}
	return s.repo.Delete(id)
}

func (s *ChallengeService) PublishChallenge(id int64) error {
	challenge, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrChallengeNotFound
		}
		return err
	}

	challenge.Status = model.ChallengeStatusPublished
	return s.repo.Update(challenge)
}
