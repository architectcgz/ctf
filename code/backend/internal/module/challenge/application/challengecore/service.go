package challengecore

import (
	"context"
	"errors"
	"strings"

	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/module/challenge/application/challengecatalog"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	platformevents "ctf-platform/internal/platform/events"
)

type challengeCommandRepository interface {
	challengeports.ChallengeWriteRepository
	challengeports.ChallengeInstanceUsageRepository
}

type ChallengeService struct {
	repo         challengeCommandRepository
	imageRepo    challengeports.ImageQueryRepository
	topologyRepo challengeports.ChallengeTopologyReadRepository
	eventBus     platformevents.Bus
	logger       *zap.Logger
}

func NewChallengeService(
	repo challengeCommandRepository,
	imageRepo challengeports.ImageQueryRepository,
	topologyRepo challengeports.ChallengeTopologyReadRepository,
	logger *zap.Logger,
) *ChallengeService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ChallengeService{
		repo:         repo,
		imageRepo:    imageRepo,
		topologyRepo: topologyRepo,
		logger:       logger,
	}
}

func (s *ChallengeService) SetEventBus(bus platformevents.Bus) *ChallengeService {
	if s == nil {
		return nil
	}
	s.eventBus = bus
	return s
}

func (s *ChallengeService) publishWeakEvent(ctx context.Context, evt platformevents.Event) {
	if s == nil {
		return
	}
	challengecatalog.PublishWeakEvent(ctx, s.logger, s.eventBus, evt)
}

func (s *ChallengeService) publishPublishedCatalogChangedEvent(
	ctx context.Context,
	changeType string,
	before challengecatalog.PublishedState,
	after challengecatalog.PublishedState,
) {
	challengecatalog.PublishPublishedCatalogChangedEvent(ctx, s.logger, s.eventBus, changeType, before, after)
}

func (s *ChallengeService) CreateChallenge(ctx context.Context, actorUserID int64, req CreateChallengeInput) (*challengecontracts.ChallengeResp, error) {
	if req.ImageID > 0 {
		if _, err := s.imageRepo.FindByID(ctx, req.ImageID); err != nil {
			if errors.Is(err, challengeports.ErrChallengeImageNotFound) {
				return nil, apperror.ErrNotFound.WithCause(errors.New(domain.ErrMsgImageNotFound))
			}
			return nil, err
		}
	}

	challenge := &challengeports.ChallengeWriteModel{
		Title:           req.Title,
		Description:     req.Description,
		Category:        req.Category,
		Difficulty:      req.Difficulty,
		Points:          req.Points,
		ImageID:         req.ImageID,
		AttachmentURL:   strings.TrimSpace(req.AttachmentURL),
		InstanceSharing: normalizeInstanceSharing(req.InstanceSharing),
		Status:          challengecontracts.ChallengeStatusDraft,
		CreatedBy:       &actorUserID,
	}

	hints, err := domain.NormalizeHintModels(toChallengeHintReqs(req.Hints))
	if err != nil {
		return nil, err
	}
	if err := s.validateInstanceSharingConfig(ctx, challenge); err != nil {
		return nil, err
	}
	if err := s.repo.CreateWithHints(ctx, challenge, hints); err != nil {
		return nil, err
	}
	return domain.ChallengeRespFromWriteModel(challenge, hints), nil
}

func (s *ChallengeService) UpdateChallenge(ctx context.Context, id int64, req UpdateChallengeInput) error {
	challengeWriteModel, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeCommandChallengeNotFound) {
			return challengecontracts.ErrChallengeNotFound
		}
		return err
	}
	challenge := challengeWriteModel
	before := challengecatalog.PublishedStateFromWriteModel(challenge)

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
			if _, err := s.imageRepo.FindByID(ctx, *req.ImageID); err != nil {
				if errors.Is(err, challengeports.ErrChallengeImageNotFound) {
					return apperror.ErrNotFound.WithCause(errors.New(domain.ErrMsgImageNotFound))
				}
				return err
			}
		}
		challenge.ImageID = *req.ImageID
	}
	if req.AttachmentURL != nil {
		challenge.AttachmentURL = strings.TrimSpace(*req.AttachmentURL)
	}
	if req.InstanceSharing != "" {
		challenge.InstanceSharing = normalizeInstanceSharing(req.InstanceSharing)
	}

	replaceHints := req.Hints != nil
	hints, err := domain.NormalizeHintModels(toChallengeHintReqs(req.Hints))
	if err != nil {
		return err
	}
	if err := s.validateInstanceSharingConfig(ctx, challenge); err != nil {
		return err
	}

	if err := s.repo.UpdateWithHints(ctx, challenge, hints, replaceHints); err != nil {
		return err
	}
	s.publishPublishedCatalogChangedEvent(
		ctx,
		challengecontracts.ChallengeCatalogChangeTypeUpdated,
		before,
		challengecatalog.PublishedStateFromWriteModel(challenge),
	)
	return nil
}

func normalizeInstanceSharing(value string) string {
	switch strings.TrimSpace(value) {
	case challengecontracts.InstanceSharingPerTeam:
		return challengecontracts.InstanceSharingPerTeam
	case challengecontracts.InstanceSharingShared:
		return challengecontracts.InstanceSharingShared
	default:
		return challengecontracts.InstanceSharingPerUser
	}
}

func toChallengeHintReqs(hints []ChallengeHintInput) []challengecontracts.ChallengeHintReq {
	if hints == nil {
		return nil
	}
	resp := make([]challengecontracts.ChallengeHintReq, 0, len(hints))
	for _, hint := range hints {
		resp = append(resp, challengecontracts.ChallengeHintReq{
			Level:   hint.Level,
			Title:   hint.Title,
			Content: hint.Content,
		})
	}
	return resp
}

func (s *ChallengeService) validateInstanceSharingConfig(ctx context.Context, challenge *challengeports.ChallengeWriteModel) error {
	if challenge == nil || challenge.InstanceSharing != challengecontracts.InstanceSharingShared {
		return nil
	}
	if challenge.FlagType == challengecontracts.FlagTypeDynamic {
		return apperror.ErrInvalidParams.WithCause(errors.New("共享实例只适用于无状态题，不支持动态 Flag"))
	}
	if s.topologyRepo == nil || challenge.ID <= 0 {
		return nil
	}
	topology, err := s.topologyRepo.FindChallengeTopologyByChallengeID(ctx, challenge.ID)
	switch {
	case err == nil:
	case errors.Is(err, challengeports.ErrChallengeTopologyNotFound):
		return nil
	default:
		return err
	}

	spec, err := challengecontracts.DecodeTopologySpec(topology.Spec)
	if err != nil {
		return apperror.ErrInvalidParams.WithCause(err)
	}
	for _, node := range spec.Nodes {
		if node.InjectFlag {
			return apperror.ErrInvalidParams.WithCause(errors.New("共享实例只适用于无状态题，不支持带 Flag 注入的拓扑"))
		}
	}
	return nil
}

func (s *ChallengeService) DeleteChallenge(ctx context.Context, id int64) error {
	challengeWriteModel, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeCommandChallengeNotFound) {
			return challengecontracts.ErrChallengeNotFound
		}
		return err
	}
	before := challengecatalog.PublishedStateFromWriteModel(challengeWriteModel)

	hasInstances, err := s.repo.HasRunningInstances(ctx, id)
	if err != nil {
		return err
	}
	if hasInstances {
		return apperror.ErrConflict.WithMessage(domain.ErrMsgHasRunningStudents).
			WithCause(errors.New(domain.ErrMsgHasRunningInstances))
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.publishPublishedCatalogChangedEvent(
		ctx,
		challengecontracts.ChallengeCatalogChangeTypeDeleted,
		before,
		challengecatalog.PublishedState{ID: before.ID},
	)
	return nil
}

func (s *ChallengeService) PublishChallenge(ctx context.Context, id int64) error {
	challengeWriteModel, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeCommandChallengeNotFound) {
			return challengecontracts.ErrChallengeNotFound
		}
		return err
	}
	before := challengecatalog.PublishedStateFromWriteModel(challengeWriteModel)
	challengeWriteModel.Status = challengecontracts.ChallengeStatusPublished
	if err := s.repo.Update(ctx, challengeWriteModel); err != nil {
		return err
	}
	s.publishPublishedCatalogChangedEvent(
		ctx,
		challengecontracts.ChallengeCatalogChangeTypePublished,
		before,
		challengecatalog.PublishedStateFromWriteModel(challengeWriteModel),
	)
	return nil
}
