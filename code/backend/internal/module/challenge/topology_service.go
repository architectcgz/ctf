package challenge

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type TopologyService struct {
	repo         *Repository
	templateRepo *TemplateRepository
	imageRepo    *ImageRepository
}

func NewTopologyService(repo *Repository, templateRepo *TemplateRepository, imageRepo *ImageRepository) *TopologyService {
	return &TopologyService{
		repo:         repo,
		templateRepo: templateRepo,
		imageRepo:    imageRepo,
	}
}

func (s *TopologyService) SaveChallengeTopology(challengeID int64, req *dto.SaveChallengeTopologyReq) (*dto.ChallengeTopologyResp, error) {
	if _, err := s.repo.FindByID(challengeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}

	rawSpec, entryNodeKey, templateID, err := s.resolveTopologyPayload(req)
	if err != nil {
		return nil, err
	}
	if err := s.ensureTopologyImagesExist(rawSpec); err != nil {
		return nil, err
	}

	item := &model.ChallengeTopology{
		ChallengeID:  challengeID,
		TemplateID:   templateID,
		EntryNodeKey: entryNodeKey,
		Spec:         rawSpec,
		UpdatedAt:    time.Now(),
	}
	if err := s.repo.UpsertChallengeTopology(item); err != nil {
		return nil, err
	}
	if templateID != nil {
		if err := s.templateRepo.IncrementUsage(*templateID); err != nil {
			return nil, err
		}
	}
	saved, err := s.repo.FindChallengeTopologyByChallengeID(challengeID)
	if err != nil {
		return nil, err
	}
	return topologyRespFromModel(saved)
}

func (s *TopologyService) GetChallengeTopology(challengeID int64) (*dto.ChallengeTopologyResp, error) {
	if _, err := s.repo.FindByID(challengeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}
	item, err := s.repo.FindChallengeTopologyByChallengeID(challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	return topologyRespFromModel(item)
}

func (s *TopologyService) DeleteChallengeTopology(challengeID int64) error {
	if _, err := s.repo.FindByID(challengeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrChallengeNotFound
		}
		return err
	}
	return s.repo.DeleteChallengeTopologyByChallengeID(challengeID)
}

func (s *TopologyService) CreateTemplate(req *dto.UpsertEnvironmentTemplateReq) (*dto.EnvironmentTemplateResp, error) {
	rawSpec, entryNodeKey, err := buildTopologySpec(req.EntryNodeKey, req.Networks, req.Nodes, req.Links, req.Policies)
	if err != nil {
		return nil, err
	}
	if err := s.ensureTopologyImagesExist(rawSpec); err != nil {
		return nil, err
	}
	item := &model.EnvironmentTemplate{
		Name:         strings.TrimSpace(req.Name),
		Description:  strings.TrimSpace(req.Description),
		EntryNodeKey: entryNodeKey,
		Spec:         rawSpec,
	}
	if err := s.templateRepo.Create(item); err != nil {
		return nil, err
	}
	return templateRespFromModel(item)
}

func (s *TopologyService) UpdateTemplate(id int64, req *dto.UpsertEnvironmentTemplateReq) (*dto.EnvironmentTemplateResp, error) {
	item, err := s.templateRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	rawSpec, entryNodeKey, err := buildTopologySpec(req.EntryNodeKey, req.Networks, req.Nodes, req.Links, req.Policies)
	if err != nil {
		return nil, err
	}
	if err := s.ensureTopologyImagesExist(rawSpec); err != nil {
		return nil, err
	}
	item.Name = strings.TrimSpace(req.Name)
	item.Description = strings.TrimSpace(req.Description)
	item.EntryNodeKey = entryNodeKey
	item.Spec = rawSpec
	item.UpdatedAt = time.Now()
	if err := s.templateRepo.Update(item); err != nil {
		return nil, err
	}
	return templateRespFromModel(item)
}

func (s *TopologyService) GetTemplate(id int64) (*dto.EnvironmentTemplateResp, error) {
	item, err := s.templateRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	return templateRespFromModel(item)
}

func (s *TopologyService) ListTemplates(keyword string) ([]*dto.EnvironmentTemplateResp, error) {
	items, err := s.templateRepo.List(strings.TrimSpace(keyword))
	if err != nil {
		return nil, err
	}
	resp := make([]*dto.EnvironmentTemplateResp, 0, len(items))
	for _, item := range items {
		mapped, mapErr := templateRespFromModel(item)
		if mapErr != nil {
			return nil, mapErr
		}
		resp = append(resp, mapped)
	}
	return resp, nil
}

func (s *TopologyService) DeleteTemplate(id int64) error {
	if _, err := s.templateRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrNotFound
		}
		return err
	}
	return s.templateRepo.Delete(id)
}

func (s *TopologyService) resolveTopologyPayload(req *dto.SaveChallengeTopologyReq) (rawSpec, entryNodeKey string, templateID *int64, err error) {
	if req.TemplateID != nil {
		item, findErr := s.templateRepo.FindByID(*req.TemplateID)
		if findErr != nil {
			if errors.Is(findErr, gorm.ErrRecordNotFound) {
				return "", "", nil, errcode.ErrNotFound.WithCause(errors.New("环境模板不存在"))
			}
			return "", "", nil, findErr
		}
		return item.Spec, item.EntryNodeKey, req.TemplateID, nil
	}

	rawSpec, entryNodeKey, err = buildTopologySpec(req.EntryNodeKey, req.Networks, req.Nodes, req.Links, req.Policies)
	if err != nil {
		return "", "", nil, err
	}
	return rawSpec, entryNodeKey, nil, nil
}

func (s *TopologyService) ensureTopologyImagesExist(rawSpec string) error {
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
		if _, findErr := s.imageRepo.FindByID(node.ImageID); findErr != nil {
			if errors.Is(findErr, gorm.ErrRecordNotFound) {
				return errcode.ErrInvalidParams.WithCause(errors.New("拓扑节点引用的镜像不存在"))
			}
			return findErr
		}
	}
	return nil
}
