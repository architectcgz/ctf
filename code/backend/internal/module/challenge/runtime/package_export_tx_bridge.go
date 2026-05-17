package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

type challengePackageExportTxRunner struct {
	repo *challengeinfra.Repository
}

func NewChallengePackageExportTxRunner(
	repo *challengeinfra.Repository,
) challengeports.ChallengePackageExportTxRunner {
	if repo == nil {
		return nil
	}
	return &challengePackageExportTxRunner{repo: repo}
}

func (r *challengePackageExportTxRunner) WithinChallengePackageExportTransaction(
	ctx context.Context,
	fn func(store challengeports.ChallengePackageExportTxStore) error,
) error {
	if r == nil || r.repo == nil {
		return fmt.Errorf("challenge package export tx runner is not configured")
	}
	return r.repo.WithinTransaction(ctx, func(txRepo *challengeinfra.Repository) error {
		store := &challengePackageExportTxStore{
			rawRepo:       txRepo,
			challengeRepo: challengeinfra.NewChallengeCommandRepository(txRepo),
			topologyRepo:  challengeinfra.NewTopologyServiceRepository(txRepo),
			packageRepo:   challengeinfra.NewTopologyPackageRevisionRepository(txRepo),
			imageRepo:     challengeinfra.NewImageQueryRepository(challengeinfra.NewImageRepository(txRepo.DB(ctx))),
		}
		return fn(store)
	})
}

type challengePackageExportTxStore struct {
	rawRepo       *challengeinfra.Repository
	challengeRepo *challengeinfra.ChallengeCommandRepository
	topologyRepo  *challengeinfra.TopologyServiceRepository
	packageRepo   *challengeinfra.TopologyPackageRevisionRepository
	imageRepo     *challengeinfra.ImageQueryRepository
}

func (s *challengePackageExportTxStore) FindChallenge(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	return s.challengeRepo.FindByID(ctx, challengeID)
}

func (s *challengePackageExportTxStore) FindTopology(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
	return s.topologyRepo.FindChallengeTopologyByChallengeID(ctx, challengeID)
}

func (s *challengePackageExportTxStore) FindPackageRevisionByID(
	ctx context.Context,
	revisionID int64,
) (*model.ChallengePackageRevision, error) {
	return s.packageRepo.FindChallengePackageRevisionByID(ctx, revisionID)
}

func (s *challengePackageExportTxStore) NextPackageRevisionNo(ctx context.Context, challengeID int64) (int, error) {
	var latest model.ChallengePackageRevision
	err := s.rawRepo.DB(ctx).
		Where("challenge_id = ?", challengeID).
		Order("revision_no DESC, id DESC").
		First(&latest).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return 1, nil
	case err != nil:
		return 0, err
	default:
		return latest.RevisionNo + 1, nil
	}
}

func (s *challengePackageExportTxStore) ListChallengeHints(
	ctx context.Context,
	challengeID int64,
) ([]model.ChallengeHint, error) {
	items, err := s.rawRepo.ListHintsByChallengeID(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	result := make([]model.ChallengeHint, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, *item)
	}
	return result, nil
}

func (s *challengePackageExportTxStore) FindImageRefByID(ctx context.Context, imageID int64) (string, error) {
	image, err := s.imageRepo.FindByID(ctx, imageID)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeImageNotFound) {
			return "", errcode.ErrInvalidParams.WithCause(errors.New("拓扑节点引用的镜像不存在"))
		}
		return "", err
	}
	if strings.TrimSpace(image.Name) == "" {
		return "", errcode.ErrInvalidParams.WithCause(errors.New("镜像记录缺少名称"))
	}
	if strings.TrimSpace(image.Tag) == "" || strings.TrimSpace(image.Tag) == "latest" {
		return strings.TrimSpace(image.Name), nil
	}
	return fmt.Sprintf("%s:%s", strings.TrimSpace(image.Name), strings.TrimSpace(image.Tag)), nil
}

func (s *challengePackageExportTxStore) CreateExportRevision(
	ctx context.Context,
	revision *model.ChallengePackageRevision,
) error {
	return s.rawRepo.CreateChallengePackageRevision(ctx, revision)
}

func (s *challengePackageExportTxStore) MarkTopologyExported(
	ctx context.Context,
	topologyID int64,
	revisionID int64,
	baselineSpec string,
	updatedAt time.Time,
) error {
	return s.rawRepo.DB(ctx).
		Model(&model.ChallengeTopology{}).
		Where("id = ?", topologyID).
		Updates(map[string]any{
			"package_revision_id":     revisionID,
			"package_baseline_spec":   baselineSpec,
			"sync_status":             model.ChallengeTopologySyncStatusClean,
			"last_export_revision_id": revisionID,
			"updated_at":              updatedAt,
		}).Error
}
