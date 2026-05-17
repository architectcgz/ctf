package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/domain"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

type imageBuildServiceGetter func() *ImageBuildService

type testImageBuildTxStore struct {
	tx *gorm.DB
}

func newImageBuildTxStore(tx *gorm.DB) *testImageBuildTxStore {
	if tx == nil {
		return nil
	}
	return &testImageBuildTxStore{tx: tx}
}

func (s *testImageBuildTxStore) FindByNameTag(ctx context.Context, name string, tag string) (*model.Image, error) {
	var image model.Image
	err := s.tx.WithContext(ctx).Unscoped().
		Where("name = ? AND tag = ?", name, tag).
		First(&image).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengeImageNotFound
	}
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (s *testImageBuildTxStore) CreateImage(ctx context.Context, image *model.Image) error {
	return s.tx.WithContext(ctx).Create(image).Error
}

func (s *testImageBuildTxStore) CreateImageBuildJob(ctx context.Context, job *model.ImageBuildJob) error {
	return s.tx.WithContext(ctx).Create(job).Error
}

func (s *testImageBuildTxStore) UpdateImage(ctx context.Context, image *model.Image, updates map[string]any) error {
	return s.tx.WithContext(ctx).Unscoped().Model(image).Updates(updates).Error
}

type testChallengeImportTxRunner struct {
	repo   *challengeinfra.Repository
	getter imageBuildServiceGetter
}

func newTestChallengeImportTxRunner(repo *challengeinfra.Repository, getter imageBuildServiceGetter) challengeports.ChallengeImportTxRunner {
	if repo == nil {
		return nil
	}
	return &testChallengeImportTxRunner{repo: repo, getter: getter}
}

func (r *testChallengeImportTxRunner) WithinChallengeImportTransaction(
	ctx context.Context,
	fn func(store challengeports.ChallengeImportTxStore) error,
) error {
	return r.repo.WithinTransaction(ctx, func(txRepo *challengeinfra.Repository) error {
		store := &testChallengeImportTxStore{
			rawRepo: txRepo,
			getter:  r.getter,
		}
		return fn(store)
	})
}

type testChallengeImportTxStore struct {
	rawRepo *challengeinfra.Repository
	getter  imageBuildServiceGetter
}

func (s *testChallengeImportTxStore) tx(ctx context.Context) *gorm.DB {
	return s.rawRepo.DB(ctx)
}

func (s *testChallengeImportTxStore) imageBuild() *ImageBuildService {
	if s.getter == nil {
		return nil
	}
	return s.getter()
}

func (s *testChallengeImportTxStore) RejectImportedChallengeSlugConflict(ctx context.Context, packageSlug string) error {
	slug := strings.TrimSpace(packageSlug)
	if slug == "" {
		return nil
	}
	var existing model.Challenge
	err := s.tx(ctx).Unscoped().
		Select("id", "title", "package_slug").
		Where("package_slug = ?", slug).
		First(&existing).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil
	case err != nil:
		return fmt.Errorf("check imported challenge slug %s: %w", slug, err)
	default:
		message := fmt.Sprintf("题目 slug %s 已被已有题目占用，请改用题目编辑入口更新", slug)
		return errcode.New(errcode.ErrConflict.Code, message, errcode.ErrConflict.HTTPStatus)
	}
}

func (s *testChallengeImportTxStore) FindLegacyChallengeForImportedPackageCreate(
	ctx context.Context,
	title string,
	category string,
) (*model.Challenge, bool, error) {
	var challenge model.Challenge
	err := s.tx(ctx).Unscoped().
		Where("(package_slug IS NULL OR package_slug = '') AND title = ? AND category = ?", title, category).
		First(&challenge).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, false, nil
	case err != nil:
		return nil, false, fmt.Errorf("find imported challenge %s: %w", title, err)
	default:
		return &challenge, true, nil
	}
}

func (s *testChallengeImportTxStore) CreateImportedChallenge(ctx context.Context, challenge *model.Challenge) error {
	return s.tx(ctx).Create(challenge).Error
}

func (s *testChallengeImportTxStore) UpdateImportedChallenge(ctx context.Context, challenge *model.Challenge, updates map[string]any) error {
	return s.tx(ctx).Unscoped().Model(challenge).Updates(updates).Error
}

func (s *testChallengeImportTxStore) ClearPublishCheckJobs(ctx context.Context, challengeID int64) error {
	return s.tx(ctx).Where("challenge_id = ?", challengeID).Delete(&model.ChallengePublishCheckJob{}).Error
}

func (s *testChallengeImportTxStore) ReplaceImportedHints(ctx context.Context, challengeID int64, hints []model.ChallengeHint) error {
	if err := s.tx(ctx).Where("challenge_id = ?", challengeID).Delete(&model.ChallengeHint{}).Error; err != nil {
		return err
	}
	if len(hints) == 0 {
		return nil
	}
	return s.tx(ctx).Create(&hints).Error
}

func (s *testChallengeImportTxStore) ApplyImportedFlagUpdates(ctx context.Context, challengeID int64, updates map[string]any) error {
	return s.tx(ctx).Model(&model.Challenge{}).Where("id = ?", challengeID).Updates(updates).Error
}

func (s *testChallengeImportTxStore) NextChallengePackageRevisionNo(ctx context.Context, challengeID int64) (int, error) {
	var latest model.ChallengePackageRevision
	err := s.tx(ctx).Where("challenge_id = ?", challengeID).Order("revision_no DESC, id DESC").First(&latest).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return 1, nil
	case err != nil:
		return 0, err
	default:
		return latest.RevisionNo + 1, nil
	}
}

func (s *testChallengeImportTxStore) CreateImportedPackageRevision(ctx context.Context, revision *model.ChallengePackageRevision) error {
	return s.rawRepo.CreateChallengePackageRevision(ctx, revision)
}

func (s *testChallengeImportTxStore) UpsertImportedTopology(ctx context.Context, topology *model.ChallengeTopology) error {
	return s.rawRepo.UpsertChallengeTopology(ctx, topology)
}

func (s *testChallengeImportTxStore) ResolvePlatformBuildImage(
	ctx context.Context,
	req challengeports.ImportedPlatformBuildImageRequest,
) (*challengeports.ImportedImageResolution, error) {
	imageBuild := s.imageBuild()
	if imageBuild == nil {
		return nil, fmt.Errorf("image build service is not configured")
	}
	result, err := imageBuild.CreatePlatformBuildJobInTx(ctx, &testImageBuildTxStore{tx: s.tx(ctx)}, CreatePlatformBuildJobRequest{
		ChallengeMode:  req.ChallengeMode,
		PackageSlug:    req.PackageSlug,
		SuggestedTag:   req.SuggestedTag,
		SourceDir:      req.SourceDir,
		DockerfilePath: req.DockerfilePath,
		ContextPath:    req.ContextPath,
		CreatedBy:      req.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return &challengeports.ImportedImageResolution{ImageID: result.ImageID, ImageRef: result.TargetRef}, nil
}

func (s *testChallengeImportTxStore) ResolveExternalImage(
	ctx context.Context,
	packageSlug string,
	imageRef string,
) (*challengeports.ImportedImageResolution, error) {
	imageBuild := s.imageBuild()
	if imageBuild == nil {
		return nil, fmt.Errorf("image build service is not configured")
	}
	result, err := imageBuild.VerifyExternalImageRefInTx(ctx, &testImageBuildTxStore{tx: s.tx(ctx)}, packageSlug, imageRef)
	if err != nil {
		return nil, err
	}
	return &challengeports.ImportedImageResolution{ImageID: result.ImageID, ImageRef: result.ImageRef}, nil
}

func (s *testChallengeImportTxStore) ResolveExistingImageRef(
	ctx context.Context,
	packageSlug string,
	imageRef string,
) (*challengeports.ImportedImageResolution, error) {
	ref := strings.TrimSpace(imageRef)
	if ref == "" {
		return &challengeports.ImportedImageResolution{}, nil
	}
	name, tag, err := domain.SplitImageRef(ref)
	if err != nil {
		return nil, fmt.Errorf("invalid image ref for %s: %w", packageSlug, err)
	}
	tx := s.tx(ctx)
	var image model.Image
	findErr := tx.Unscoped().Where("name = ? AND tag = ?", name, tag).First(&image).Error
	switch {
	case errors.Is(findErr, gorm.ErrRecordNotFound):
		image = model.Image{
			Name:        name,
			Tag:         tag,
			Description: fmt.Sprintf("Imported from challenge pack %s", packageSlug),
			Status:      model.ImageStatusAvailable,
		}
		if err := tx.Create(&image).Error; err != nil {
			return nil, err
		}
	case findErr != nil:
		return nil, findErr
	default:
		if err := tx.Model(&image).Updates(map[string]any{
			"status":     model.ImageStatusAvailable,
			"deleted_at": nil,
			"updated_at": time.Now().UTC(),
		}).Error; err != nil {
			return nil, err
		}
	}
	return &challengeports.ImportedImageResolution{ImageID: image.ID, ImageRef: ref}, nil
}

type testAWDChallengeImportTxRunner struct {
	repo   *challengeinfra.Repository
	getter imageBuildServiceGetter
}

func newTestAWDChallengeImportTxRunner(repo *challengeinfra.Repository, getter imageBuildServiceGetter) challengeports.AWDChallengeImportTxRunner {
	if repo == nil {
		return nil
	}
	return &testAWDChallengeImportTxRunner{repo: repo, getter: getter}
}

func (r *testAWDChallengeImportTxRunner) WithinAWDChallengeImportTransaction(
	ctx context.Context,
	fn func(store challengeports.AWDChallengeImportTxStore) error,
) error {
	return r.repo.WithinTransaction(ctx, func(txRepo *challengeinfra.Repository) error {
		store := &testAWDChallengeImportTxStore{
			rawRepo: txRepo,
			getter:  r.getter,
		}
		return fn(store)
	})
}

type testAWDChallengeImportTxStore struct {
	rawRepo *challengeinfra.Repository
	getter  imageBuildServiceGetter
}

func (s *testAWDChallengeImportTxStore) tx(ctx context.Context) *gorm.DB {
	return s.rawRepo.DB(ctx)
}

func (s *testAWDChallengeImportTxStore) imageBuild() *ImageBuildService {
	if s.getter == nil {
		return nil
	}
	return s.getter()
}

func (s *testAWDChallengeImportTxStore) RejectImportedAWDChallengeSlugConflict(ctx context.Context, slug string) error {
	normalizedSlug := strings.TrimSpace(slug)
	if normalizedSlug == "" {
		return nil
	}
	var existing model.AWDChallenge
	err := s.tx(ctx).Unscoped().
		Select("id", "slug", "name").
		Where("slug = ?", normalizedSlug).
		First(&existing).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil
	case err != nil:
		return err
	default:
		message := fmt.Sprintf("AWD 题目 slug %s 已被已有题目占用，请改用题目编辑入口更新", normalizedSlug)
		return errcode.New(errcode.ErrConflict.Code, message, errcode.ErrConflict.HTTPStatus)
	}
}

func (s *testAWDChallengeImportTxStore) CreateImportedAWDChallenge(ctx context.Context, challenge *model.AWDChallenge) error {
	return s.rawRepo.CreateAWDChallenge(ctx, challenge)
}

func (s *testAWDChallengeImportTxStore) ResolvePlatformBuildImage(
	ctx context.Context,
	req challengeports.ImportedPlatformBuildImageRequest,
) (*challengeports.ImportedImageResolution, error) {
	base := testChallengeImportTxStore{rawRepo: s.rawRepo, getter: s.getter}
	return base.ResolvePlatformBuildImage(ctx, req)
}

func (s *testAWDChallengeImportTxStore) ResolveExternalImage(
	ctx context.Context,
	packageSlug string,
	imageRef string,
) (*challengeports.ImportedImageResolution, error) {
	base := testChallengeImportTxStore{rawRepo: s.rawRepo, getter: s.getter}
	return base.ResolveExternalImage(ctx, packageSlug, imageRef)
}

func (s *testAWDChallengeImportTxStore) ResolveExistingImageRef(
	ctx context.Context,
	packageSlug string,
	imageRef string,
) (*challengeports.ImportedImageResolution, error) {
	base := testChallengeImportTxStore{rawRepo: s.rawRepo, getter: s.getter}
	return base.ResolveExistingImageRef(ctx, packageSlug, imageRef)
}

type testChallengePackageExportTxRunner struct {
	repo *challengeinfra.Repository
}

func newTestChallengePackageExportTxRunner(repo *challengeinfra.Repository) challengeports.ChallengePackageExportTxRunner {
	if repo == nil {
		return nil
	}
	return &testChallengePackageExportTxRunner{repo: repo}
}

func (r *testChallengePackageExportTxRunner) WithinChallengePackageExportTransaction(
	ctx context.Context,
	fn func(store challengeports.ChallengePackageExportTxStore) error,
) error {
	return r.repo.WithinTransaction(ctx, func(txRepo *challengeinfra.Repository) error {
		store := &testChallengePackageExportTxStore{
			rawRepo:       txRepo,
			challengeRepo: challengeinfra.NewChallengeCommandRepository(txRepo),
			topologyRepo:  challengeinfra.NewTopologyServiceRepository(txRepo),
			packageRepo:   challengeinfra.NewTopologyPackageRevisionRepository(txRepo),
			imageRepo:     challengeinfra.NewImageQueryRepository(challengeinfra.NewImageRepository(txRepo.DB(ctx))),
		}
		return fn(store)
	})
}

type testChallengePackageExportTxStore struct {
	rawRepo       *challengeinfra.Repository
	challengeRepo *challengeinfra.ChallengeCommandRepository
	topologyRepo  *challengeinfra.TopologyServiceRepository
	packageRepo   *challengeinfra.TopologyPackageRevisionRepository
	imageRepo     *challengeinfra.ImageQueryRepository
}

func (s *testChallengePackageExportTxStore) FindChallenge(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	return s.challengeRepo.FindByID(ctx, challengeID)
}

func (s *testChallengePackageExportTxStore) FindTopology(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
	return s.topologyRepo.FindChallengeTopologyByChallengeID(ctx, challengeID)
}

func (s *testChallengePackageExportTxStore) FindPackageRevisionByID(ctx context.Context, revisionID int64) (*model.ChallengePackageRevision, error) {
	return s.packageRepo.FindChallengePackageRevisionByID(ctx, revisionID)
}

func (s *testChallengePackageExportTxStore) NextPackageRevisionNo(ctx context.Context, challengeID int64) (int, error) {
	var latest model.ChallengePackageRevision
	err := s.rawRepo.DB(ctx).Where("challenge_id = ?", challengeID).Order("revision_no DESC, id DESC").First(&latest).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return 1, nil
	case err != nil:
		return 0, err
	default:
		return latest.RevisionNo + 1, nil
	}
}

func (s *testChallengePackageExportTxStore) ListChallengeHints(ctx context.Context, challengeID int64) ([]model.ChallengeHint, error) {
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

func (s *testChallengePackageExportTxStore) FindImageRefByID(ctx context.Context, imageID int64) (string, error) {
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

func (s *testChallengePackageExportTxStore) CreateExportRevision(ctx context.Context, revision *model.ChallengePackageRevision) error {
	return s.rawRepo.CreateChallengePackageRevision(ctx, revision)
}

func (s *testChallengePackageExportTxStore) MarkTopologyExported(
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
