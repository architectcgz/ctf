package commands

import (
	"context"
	"errors"
	"strings"
	"time"

	"ctf-platform/internal/apperror"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"ctf-platform/internal/module/challenge/domain"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type TopologyService struct {
	repo         topologyCommandRepository
	templateRepo topologyTemplateCommandRepository
	imageRepo    challengeports.ImageQueryRepository
}

type topologyCommandRepository interface {
	challengeports.ChallengeTopologyChallengeLookupRepository
	challengeports.ChallengeTopologyReadRepository
	challengeports.ChallengeTopologyWriteRepository
}

type topologyTemplateCommandRepository interface {
	challengeports.EnvironmentTemplateCommandRepository
	challengeports.EnvironmentTemplateQueryRepository
	challengeports.EnvironmentTemplateUsageRepository
}

func NewTopologyService(repo topologyCommandRepository, templateRepo topologyTemplateCommandRepository, imageRepo challengeports.ImageQueryRepository) *TopologyService {
	return &TopologyService{
		repo:         repo,
		templateRepo: templateRepo,
		imageRepo:    imageRepo,
	}
}

func (s *TopologyService) SaveChallengeTopology(ctx context.Context, challengeID int64, req SaveChallengeTopologyInput) (*challengecontracts.ChallengeTopologyResp, error) {
	challenge, err := s.repo.FindByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeTopologyChallengeNotFound) {
			return nil, challengecontracts.ErrChallengeNotFound
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

	var existing *challengeentity.ChallengeTopology
	existing, err = s.repo.FindChallengeTopologyByChallengeID(ctx, challengeID)
	switch {
	case err == nil:
	case errors.Is(err, challengeports.ErrChallengeTopologyNotFound):
		existing = nil
	default:
		return nil, err
	}

	item := &challengeentity.ChallengeTopology{
		ChallengeID:  challengeID,
		TemplateID:   templateID,
		EntryNodeKey: entryNodeKey,
		Spec:         rawSpec,
		UpdatedAt:    time.Now().UTC(),
	}
	if existing != nil {
		item.SourceType = existing.SourceType
		item.SourcePath = existing.SourcePath
		item.PackageRevisionID = existing.PackageRevisionID
		item.PackageBaselineSpec = existing.PackageBaselineSpec
		item.LastExportRevisionID = existing.LastExportRevisionID
		item.SyncStatus = resolveTopologySyncStatus(rawSpec, existing.PackageBaselineSpec)
	} else {
		item.SourceType = challengeentity.ChallengeTopologySourceTypeManual
		item.SyncStatus = challengeentity.ChallengeTopologySyncStatusClean
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

func resolveTopologySyncStatus(rawSpec string, baselineSpec string) string {
	if strings.TrimSpace(baselineSpec) == "" {
		return challengeentity.ChallengeTopologySyncStatusClean
	}
	if strings.TrimSpace(rawSpec) == strings.TrimSpace(baselineSpec) {
		return challengeentity.ChallengeTopologySyncStatusClean
	}
	return challengeentity.ChallengeTopologySyncStatusDrifted
}

func validateSharedTopologyConstraint(challenge *challengeports.ChallengeTopologyChallenge, rawSpec string) error {
	if challenge == nil || challenge.InstanceSharing != challengecontracts.InstanceSharingShared {
		return nil
	}
	spec, err := challengecontracts.DecodeTopologySpec(rawSpec)
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

func (s *TopologyService) DeleteChallengeTopology(ctx context.Context, challengeID int64) error {
	if _, err := s.repo.FindByID(ctx, challengeID); err != nil {
		if errors.Is(err, challengeports.ErrChallengeTopologyChallengeNotFound) {
			return challengecontracts.ErrChallengeNotFound
		}
		return err
	}
	return s.repo.DeleteChallengeTopologyByChallengeID(ctx, challengeID)
}

func (s *TopologyService) CreateTemplate(ctx context.Context, req UpsertEnvironmentTemplateInput) (*challengecontracts.EnvironmentTemplateResp, error) {
	rawSpec, entryNodeKey, err := domain.BuildTopologySpec(req.EntryNodeKey, req.Networks, req.Nodes, req.Links, req.Policies)
	if err != nil {
		return nil, err
	}
	if err := s.ensureTopologyImagesExist(ctx, rawSpec); err != nil {
		return nil, err
	}
	item := &challengeentity.EnvironmentTemplate{
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

func (s *TopologyService) UpdateTemplate(ctx context.Context, id int64, req UpsertEnvironmentTemplateInput) (*challengecontracts.EnvironmentTemplateResp, error) {
	item, err := s.templateRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeTopologyTemplateNotFound) {
			return nil, apperror.ErrNotFound
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
	item.UpdatedAt = time.Now().UTC()
	if err := s.templateRepo.Update(ctx, item); err != nil {
		return nil, err
	}
	return domain.TemplateRespFromModel(item)
}

func (s *TopologyService) DeleteTemplate(ctx context.Context, id int64) error {
	if _, err := s.templateRepo.FindByID(ctx, id); err != nil {
		if errors.Is(err, challengeports.ErrChallengeTopologyTemplateNotFound) {
			return apperror.ErrNotFound
		}
		return err
	}
	return s.templateRepo.Delete(ctx, id)
}

func (s *TopologyService) resolveTopologyPayload(ctx context.Context, req SaveChallengeTopologyInput) (rawSpec, entryNodeKey string, templateID *int64, err error) {
	if req.TemplateID != nil {
		item, findErr := s.templateRepo.FindByID(ctx, *req.TemplateID)
		if findErr != nil {
			if errors.Is(findErr, challengeports.ErrChallengeTopologyTemplateNotFound) {
				return "", "", nil, apperror.ErrNotFound.WithCause(errors.New("环境模板不存在"))
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
	spec, err := challengecontracts.DecodeTopologySpec(rawSpec)
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
			if errors.Is(findErr, challengeports.ErrChallengeImageNotFound) {
				return apperror.ErrInvalidParams.WithCause(errors.New("拓扑节点引用的镜像不存在"))
			}
			return findErr
		}
	}
	return nil
}
