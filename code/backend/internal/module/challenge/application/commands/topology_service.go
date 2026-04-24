package commands

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

type TopologyService struct {
	repo         challengeports.ChallengeTopologyRepository
	templateRepo challengeports.EnvironmentTemplateRepository
	imageRepo    challengeports.ImageRepository
}

func NewTopologyService(repo challengeports.ChallengeTopologyRepository, templateRepo challengeports.EnvironmentTemplateRepository, imageRepo challengeports.ImageRepository) *TopologyService {
	return &TopologyService{
		repo:         repo,
		templateRepo: templateRepo,
		imageRepo:    imageRepo,
	}
}

func (s *TopologyService) SaveChallengeTopology(ctx context.Context, challengeID int64, req *dto.SaveChallengeTopologyReq) (*dto.ChallengeTopologyResp, error) {
	challenge, err := s.repo.FindByIDWithContext(ctx, challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}

	rawSpec, entryNodeKey, templateID, err := s.resolveTopologyPayload(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := validateSharedTopologyConstraint(challenge, rawSpec); err != nil {
		return nil, err
	}
	if err := s.ensureTopologyImagesExist(ctx, rawSpec); err != nil {
		return nil, err
	}

	item := &model.ChallengeTopology{
		ChallengeID:  challengeID,
		TemplateID:   templateID,
		EntryNodeKey: entryNodeKey,
		Spec:         rawSpec,
		UpdatedAt:    time.Now(),
	}
	if err := s.repo.UpsertChallengeTopology(ctx, item); err != nil {
		return nil, err
	}
	if templateID != nil {
		if err := s.templateRepo.IncrementUsage(ctx, *templateID); err != nil {
			return nil, err
		}
	}
	saved, err := s.repo.FindChallengeTopologyByChallengeID(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	return domain.TopologyRespFromModel(saved)
}

func validateSharedTopologyConstraint(challenge *model.Challenge, rawSpec string) error {
	if challenge == nil || challenge.InstanceSharing != model.InstanceSharingShared {
		return nil
	}
	spec, err := model.DecodeTopologySpec(rawSpec)
	if err != nil {
		return errcode.ErrInvalidParams.WithCause(err)
	}
	for _, node := range spec.Nodes {
		if node.InjectFlag {
			return errcode.ErrInvalidParams.WithCause(errors.New("共享实例只适用于无状态题，不支持带 Flag 注入的拓扑"))
		}
	}
	return nil
}

func (s *TopologyService) DeleteChallengeTopology(ctx context.Context, challengeID int64) error {
	if _, err := s.repo.FindByIDWithContext(ctx, challengeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrChallengeNotFound
		}
		return err
	}
	return s.repo.DeleteChallengeTopologyByChallengeID(ctx, challengeID)
}

func (s *TopologyService) CreateTemplate(ctx context.Context, req *dto.UpsertEnvironmentTemplateReq) (*dto.EnvironmentTemplateResp, error) {
	rawSpec, entryNodeKey, err := domain.BuildTopologySpec(req.EntryNodeKey, req.Networks, req.Nodes, req.Links, req.Policies)
	if err != nil {
		return nil, err
	}
	if err := s.ensureTopologyImagesExist(ctx, rawSpec); err != nil {
		return nil, err
	}
	item := &model.EnvironmentTemplate{
		Name:         strings.TrimSpace(req.Name),
		Description:  strings.TrimSpace(req.Description),
		EntryNodeKey: entryNodeKey,
		Spec:         rawSpec,
	}
	if err := s.templateRepo.Create(ctx, item); err != nil {
		return nil, err
	}
	return domain.TemplateRespFromModel(item)
}

func (s *TopologyService) UpdateTemplate(ctx context.Context, id int64, req *dto.UpsertEnvironmentTemplateReq) (*dto.EnvironmentTemplateResp, error) {
	item, err := s.templateRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	rawSpec, entryNodeKey, err := domain.BuildTopologySpec(req.EntryNodeKey, req.Networks, req.Nodes, req.Links, req.Policies)
	if err != nil {
		return nil, err
	}
	if err := s.ensureTopologyImagesExist(ctx, rawSpec); err != nil {
		return nil, err
	}
	item.Name = strings.TrimSpace(req.Name)
	item.Description = strings.TrimSpace(req.Description)
	item.EntryNodeKey = entryNodeKey
	item.Spec = rawSpec
	item.UpdatedAt = time.Now()
	if err := s.templateRepo.Update(ctx, item); err != nil {
		return nil, err
	}
	return domain.TemplateRespFromModel(item)
}

func (s *TopologyService) DeleteTemplate(ctx context.Context, id int64) error {
	if _, err := s.templateRepo.FindByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrNotFound
		}
		return err
	}
	return s.templateRepo.DeleteWithContext(ctx, id)
}

func (s *TopologyService) resolveTopologyPayload(ctx context.Context, req *dto.SaveChallengeTopologyReq) (rawSpec, entryNodeKey string, templateID *int64, err error) {
	if req.TemplateID != nil {
		item, findErr := s.templateRepo.FindByID(ctx, *req.TemplateID)
		if findErr != nil {
			if errors.Is(findErr, gorm.ErrRecordNotFound) {
				return "", "", nil, errcode.ErrNotFound.WithCause(errors.New("环境模板不存在"))
			}
			return "", "", nil, findErr
		}
		return item.Spec, item.EntryNodeKey, req.TemplateID, nil
	}

	rawSpec, entryNodeKey, err = domain.BuildTopologySpec(req.EntryNodeKey, req.Networks, req.Nodes, req.Links, req.Policies)
	if err != nil {
		return "", "", nil, err
	}
	return rawSpec, entryNodeKey, nil, nil
}

func (s *TopologyService) ensureTopologyImagesExist(ctx context.Context, rawSpec string) error {
	spec, err := model.DecodeTopologySpec(rawSpec)
	if err != nil {
		return err
	}
	seen := make(map[int64]struct{}, len(spec.Nodes))
	for _, node := range spec.Nodes {
		if node.ImageID == 0 {
			continue
		}
		if _, exists := seen[node.ImageID]; exists {
			continue
		}
		seen[node.ImageID] = struct{}{}
		if _, findErr := s.imageRepo.FindByID(ctx, node.ImageID); findErr != nil {
			if errors.Is(findErr, gorm.ErrRecordNotFound) {
				return errcode.ErrInvalidParams.WithCause(errors.New("拓扑节点引用的镜像不存在"))
			}
			return findErr
		}
	}
	return nil
}
