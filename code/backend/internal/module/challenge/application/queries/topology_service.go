package queries

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
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
	resp, err := domain.TopologyRespFromModel(item)
	if err != nil {
		return nil, err
	}
	revisions, err := s.repo.ListChallengePackageRevisionsByChallengeID(challengeID)
	if err != nil {
		return nil, err
	}
	if len(revisions) > 0 {
		resp.PackageRevisions = make([]dto.ChallengePackageRevisionResp, 0, len(revisions))
		for _, revision := range revisions {
			if revision == nil {
				continue
			}
			resp.PackageRevisions = append(resp.PackageRevisions, domain.ChallengePackageRevisionRespFromModel(revision))
		}
	}
	if item.PackageRevisionID != nil && *item.PackageRevisionID > 0 {
		revision, findErr := s.repo.FindChallengePackageRevisionByID(*item.PackageRevisionID)
		if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return nil, findErr
		}
		if findErr == nil {
			resp.PackageFiles, err = listChallengePackageFilesFromSourceDir(revision.SourceDir)
			if err != nil {
				return nil, err
			}
		}
	}
	return resp, nil
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

func listChallengePackageFilesFromSourceDir(sourceDir string) ([]dto.ChallengePackageFileResp, error) {
	sourceDir = strings.TrimSpace(sourceDir)
	if sourceDir == "" {
		return nil, nil
	}
	files := make([]dto.ChallengePackageFileResp, 0, 16)
	err := filepath.Walk(sourceDir, func(current string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		relativePath, err := filepath.Rel(sourceDir, current)
		if err != nil {
			return err
		}
		files = append(files, dto.ChallengePackageFileResp{
			Path: filepath.ToSlash(relativePath),
			Size: info.Size(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
	return files, nil
}
