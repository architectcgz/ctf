package queries

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

type TopologyService struct {
	repo         challengeports.ChallengeTopologyRepository
	templateRepo challengeports.EnvironmentTemplateRepository
}

func NewTopologyService(repo challengeports.ChallengeTopologyRepository, templateRepo challengeports.EnvironmentTemplateRepository) *TopologyService {
	return &TopologyService{
		repo:         repo,
		templateRepo: templateRepo,
	}
}

func (s *TopologyService) GetChallengeTopology(challengeID int64) (*dto.ChallengeTopologyResp, error) {
	return s.GetChallengeTopologyWithContext(context.Background(), challengeID)
}

func (s *TopologyService) GetChallengeTopologyWithContext(ctx context.Context, challengeID int64) (*dto.ChallengeTopologyResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, err := s.repo.FindByIDWithContext(ctx, challengeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}
	item, err := s.repo.FindChallengeTopologyByChallengeIDWithContext(ctx, challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	return domain.TopologyRespFromModel(item)
}

func (s *TopologyService) GetTemplate(id int64) (*dto.EnvironmentTemplateResp, error) {
	item, err := s.templateRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	return domain.TemplateRespFromModel(item)
}

func (s *TopologyService) ListTemplates(keyword string) ([]*dto.EnvironmentTemplateResp, error) {
	items, err := s.templateRepo.List(strings.TrimSpace(keyword))
	if err != nil {
		return nil, err
	}
	resp := make([]*dto.EnvironmentTemplateResp, 0, len(items))
	for _, item := range items {
		mapped, mapErr := domain.TemplateRespFromModel(item)
		if mapErr != nil {
			return nil, mapErr
		}
		resp = append(resp, mapped)
	}
	return resp, nil
}
