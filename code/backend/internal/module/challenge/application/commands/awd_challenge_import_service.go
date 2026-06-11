package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"ctf-platform/internal/module/challenge/domain"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type AWDChallengeImportService struct {
	repo           challengeports.AWDChallengeCommandRepository
	previewStore   challengeports.AWDChallengeImportPreviewStore
	packageStorage challengeports.ChallengePackageStorage
	checkerStore   challengeports.AWDCheckerArtifactStore
	txRunner       challengeports.AWDChallengeImportTxRunner
	imageBuild     *ImageBuildService
	logger         *zap.Logger
}

func NewAWDChallengeImportService(
	repo challengeports.AWDChallengeCommandRepository,
	previewStore challengeports.AWDChallengeImportPreviewStore,
	packageStorage challengeports.ChallengePackageStorage,
	checkerStore challengeports.AWDCheckerArtifactStore,
	imageBuild ...*ImageBuildService,
) *AWDChallengeImportService {
	service := &AWDChallengeImportService{
		repo:           repo,
		previewStore:   previewStore,
		packageStorage: packageStorage,
		checkerStore:   checkerStore,
		logger:         zap.NewNop(),
	}
	if len(imageBuild) > 0 {
		service.imageBuild = imageBuild[0]
	}
	return service
}

func (s *AWDChallengeImportService) SetLogger(logger *zap.Logger) {
	if s != nil && logger != nil {
		s.logger = logger
	}
}

func (s *AWDChallengeImportService) SetTxRunner(runner challengeports.AWDChallengeImportTxRunner) *AWDChallengeImportService {
	if s == nil {
		return nil
	}
	s.txRunner = runner
	return s
}

func (s *AWDChallengeImportService) PreviewImport(
	ctx context.Context,
	actorUserID int64,
	fileName string,
	reader io.Reader,
) (*challengecontracts.AWDChallengeImportPreviewResp, error) {
	if strings.TrimSpace(fileName) == "" {
		fileName = "awd-challenge-package.zip"
	}
	if s.previewStore == nil {
		return nil, fmt.Errorf("awd challenge import preview store is not configured")
	}

	previewID, err := generateChallengeImportPreviewID()
	if err != nil {
		return nil, err
	}

	workspace, err := s.previewStore.CreateWorkspace(ctx, previewID, fileName, reader)
	if err != nil {
		return nil, err
	}
	keepWorkspace := false
	defer func() {
		if !keepWorkspace {
			_ = s.previewStore.DeleteWorkspace(ctx, workspace.ID)
		}
	}()

	parsed, err := domain.ParseAWDChallengePackageDir(workspace.SourceDir)
	if err != nil {
		return nil, err
	}

	preview := s.buildAWDChallengeImportPreview(previewID, fileName, parsed, time.Now().UTC())
	record := &challengeports.AWDChallengeImportPreviewRecord{
		ID:          workspace.ID,
		FileName:    workspace.FileName,
		ArchivePath: workspace.ArchivePath,
		SourceDir:   workspace.SourceDir,
		CreatedBy:   actorUserID,
		CreatedAt:   preview.CreatedAt,
		Preview:     *preview,
	}
	if err := s.previewStore.SaveRecord(ctx, record); err != nil {
		return nil, err
	}
	keepWorkspace = true
	return preview, nil
}

func (s *AWDChallengeImportService) ListImports(ctx context.Context, actorUserID int64) ([]challengecontracts.AWDChallengeImportPreviewResp, error) {
	if s.previewStore == nil {
		return nil, fmt.Errorf("awd challenge import preview store is not configured")
	}
	records, err := s.previewStore.ListRecords(ctx)
	if err != nil {
		return nil, err
	}

	previews := make([]challengecontracts.AWDChallengeImportPreviewResp, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		if record.CreatedBy != 0 && record.CreatedBy != actorUserID {
			continue
		}
		previews = append(previews, record.Preview)
	}
	return previews, nil
}

func (s *AWDChallengeImportService) GetImport(
	ctx context.Context,
	actorUserID int64,
	id string,
) (*challengecontracts.AWDChallengeImportPreviewResp, error) {
	if s.previewStore == nil {
		return nil, fmt.Errorf("awd challenge import preview store is not configured")
	}
	record, err := s.previewStore.LoadRecord(ctx, id)
	if err != nil {
		return nil, err
	}
	if record.CreatedBy != 0 && record.CreatedBy != actorUserID {
		return nil, apperror.ErrForbidden
	}
	preview := record.Preview
	return &preview, nil
}

func (s *AWDChallengeImportService) CommitImport(
	ctx context.Context,
	actorUserID int64,
	id string,
) (*challengecontracts.AWDChallengeResp, error) {
	if s.previewStore == nil {
		return nil, fmt.Errorf("awd challenge import preview store is not configured")
	}
	if s.packageStorage == nil {
		return nil, fmt.Errorf("challenge package storage is not configured")
	}
	if s.checkerStore == nil {
		return nil, fmt.Errorf("awd checker artifact store is not configured")
	}
	if s.txRunner == nil {
		return nil, fmt.Errorf("awd challenge import tx runner is not configured")
	}
	record, err := s.previewStore.LoadRecord(ctx, id)
	if err != nil {
		return nil, err
	}
	if record.CreatedBy != 0 && record.CreatedBy != actorUserID {
		return nil, apperror.ErrForbidden
	}

	parsed, err := domain.ParseAWDChallengePackageDir(record.SourceDir)
	if err != nil {
		return nil, err
	}

	buildSource, err := s.packageStorage.PersistImportedImageBuildSource(ctx, challengeports.ChallengeImportedImageBuildSourceRequest{
		ChallengeMode:  domain.ChallengePackageModeAWD,
		PackageSlug:    parsed.Slug,
		PreviewID:      record.ID,
		RootDir:        parsed.RootDir,
		DockerfilePath: parsed.DockerfilePath,
		ContextPath:    parsed.BuildContextPath,
	})
	if err != nil {
		return nil, err
	}

	var challenge *challengeentity.AWDChallenge
	if err := s.txRunner.WithinAWDChallengeImportTransaction(ctx, func(store challengeports.AWDChallengeImportTxStore) error {
		if err := store.RejectImportedAWDChallengeSlugConflict(ctx, parsed.Slug); err != nil {
			return err
		}

		resolvedImageID, resolvedImageRef, err := s.resolveAWDImportedImageForCommit(ctx, store, actorUserID, parsed, buildSource)
		if err != nil {
			return err
		}

		runtimeConfig := cloneAWDChallengeConfig(parsed.RuntimeConfig)
		if strings.TrimSpace(resolvedImageRef) != "" {
			runtimeConfig["image_ref"] = resolvedImageRef
		}
		if resolvedImageID > 0 {
			runtimeConfig["image_id"] = resolvedImageID
		}

		now := time.Now().UTC()
		var current challengeentity.AWDChallenge
		checkerConfigWithArtifact, err := s.checkerStore.PersistScriptCheckerArtifact(ctx, awdCheckerArtifactPersistRequestFromParsed(parsed))
		if err != nil {
			return err
		}
		flagConfigRaw, err := marshalAWDChallengeConfig(parsed.FlagConfig)
		if err != nil {
			return err
		}
		accessConfigRaw, err := marshalAWDChallengeConfig(parsed.AccessConfig)
		if err != nil {
			return err
		}
		runtimeConfigRaw, err := marshalAWDChallengeConfig(runtimeConfig)
		if err != nil {
			return err
		}

		current = challengeentity.AWDChallenge{
			Name:             parsed.Title,
			Slug:             parsed.Slug,
			Category:         parsed.Category,
			Difficulty:       parsed.Difficulty,
			Description:      parsed.Description,
			ServiceType:      challengeentity.AWDServiceType(parsed.ServiceType),
			DeploymentMode:   challengeentity.AWDDeploymentMode(parsed.DeploymentMode),
			Version:          parsed.Version,
			Status:           challengeentity.AWDChallengeStatusPublished,
			CheckerType:      challengeentity.AWDCheckerType(parsed.CheckerType),
			CheckerConfig:    checkerConfigWithArtifact,
			FlagMode:         parsed.FlagMode,
			FlagConfig:       flagConfigRaw,
			DefenseEntryMode: parsed.DefenseEntryMode,
			AccessConfig:     accessConfigRaw,
			RuntimeConfig:    runtimeConfigRaw,
			ReadinessStatus:  challengeentity.AWDReadinessStatusPending,
			ReadinessReport:  "",
			LastVerifiedAt:   nil,
			LastVerifiedBy:   nil,
			CreatedBy:        &actorUserID,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := store.CreateImportedAWDChallenge(ctx, &current); err != nil {
			return fmt.Errorf("create imported awd challenge %s: %w", parsed.Slug, err)
		}

		challenge = &current
		return nil
	}); err != nil {
		if buildSource != nil {
			_ = s.packageStorage.DeletePath(ctx, buildSource.RootDir)
		}
		return nil, err
	}

	_ = s.previewStore.DeleteWorkspace(ctx, id)
	return domain.AWDChallengeRespFromModel(challenge), nil
}

func (s *AWDChallengeImportService) buildAWDChallengeImportPreview(
	id string,
	fileName string,
	parsed *domain.ParsedAWDChallengePackage,
	createdAt time.Time,
) *challengecontracts.AWDChallengeImportPreviewResp {
	if parsed == nil {
		return nil
	}
	var imageBuild *ImageBuildService
	var logger *zap.Logger
	if s != nil {
		imageBuild = s.imageBuild
		logger = s.logger
	}

	imageDelivery := challengecontracts.ChallengeImportImageDeliveryResp{
		SourceType:   parsed.ImageSourceType,
		SuggestedTag: parsed.SuggestedImageTag,
	}
	if parsed.ImageSourceType == domain.ImageSourceTypePlatformBuild && imageBuild != nil {
		if targetRef, err := imageBuild.BuildPlatformTargetRef(domain.ChallengePackageModeAWD, parsed.Slug, parsed.SuggestedImageTag); err == nil {
			imageDelivery.TargetImageRef = targetRef
			imageDelivery.BuildStatus = challengeentity.ImageStatusPending
		}
	}
	warnings := append([]string(nil), parsed.Warnings...)
	if challengeImportMissingImageBuildService(imageBuild, parsed.ImageSourceType) {
		warnChallengeImportImageBuildServiceUnavailable(logger, parsed.Slug, parsed.ImageSourceType, "preview")
		warnings = appendChallengeImportImageBuildWarning(warnings, parsed.ImageSourceType)
	}
	return &challengecontracts.AWDChallengeImportPreviewResp{
		ID:               id,
		FileName:         fileName,
		Slug:             parsed.Slug,
		Title:            parsed.Title,
		Category:         parsed.Category,
		Difficulty:       parsed.Difficulty,
		Description:      parsed.Description,
		ServiceType:      parsed.ServiceType,
		DeploymentMode:   parsed.DeploymentMode,
		Version:          parsed.Version,
		CheckerType:      parsed.CheckerType,
		CheckerConfig:    cloneAWDChallengeConfig(parsed.CheckerConfig),
		FlagMode:         parsed.FlagMode,
		FlagConfig:       cloneAWDChallengeConfig(parsed.FlagConfig),
		DefenseEntryMode: parsed.DefenseEntryMode,
		AccessConfig:     cloneAWDChallengeConfig(parsed.AccessConfig),
		RuntimeConfig:    cloneAWDChallengeConfig(parsed.RuntimeConfig),
		ImageDelivery:    imageDelivery,
		Warnings:         warnings,
		CreatedAt:        createdAt,
	}
}

func (s *AWDChallengeImportService) resolveAWDImportedImageForCommit(
	ctx context.Context,
	store challengeports.AWDChallengeImportTxStore,
	actorUserID int64,
	parsed *domain.ParsedAWDChallengePackage,
	buildSource *importedImageBuildSource,
) (int64, string, error) {
	var imageBuild *ImageBuildService
	var logger *zap.Logger
	if s != nil {
		imageBuild = s.imageBuild
		logger = s.logger
	}
	if parsed.ImageSourceType == domain.ImageSourceTypeExternalRef {
		if challengeImportMissingImageBuildService(imageBuild, parsed.ImageSourceType) {
			warnChallengeImportImageBuildServiceUnavailable(logger, parsed.Slug, parsed.ImageSourceType, "commit")
			return 0, "", challengeImportImageBuildServiceUnavailableError(parsed.ImageSourceType)
		}
		result, err := store.ResolveExternalImage(ctx, parsed.Slug, parsed.RuntimeImageRef)
		if err != nil {
			return 0, "", err
		}
		return result.ImageID, result.ImageRef, nil
	}
	if parsed.ImageSourceType != domain.ImageSourceTypePlatformBuild {
		result, err := store.ResolveExistingImageRef(ctx, parsed.Slug, parsed.RuntimeImageRef)
		if err != nil {
			return 0, "", err
		}
		return result.ImageID, parsed.RuntimeImageRef, nil
	}
	if challengeImportMissingImageBuildService(imageBuild, parsed.ImageSourceType) {
		warnChallengeImportImageBuildServiceUnavailable(logger, parsed.Slug, parsed.ImageSourceType, "commit")
		return 0, "", challengeImportImageBuildServiceUnavailableError(parsed.ImageSourceType)
	}
	sourceDir := parsed.RootDir
	dockerfilePath := parsed.DockerfilePath
	contextPath := parsed.BuildContextPath
	if buildSource != nil {
		sourceDir = buildSource.SourceDir
		dockerfilePath = buildSource.DockerfilePath
		contextPath = buildSource.ContextPath
	}
	result, err := store.ResolvePlatformBuildImage(ctx, challengeports.ImportedPlatformBuildImageRequest{
		ChallengeMode:  domain.ChallengePackageModeAWD,
		PackageSlug:    parsed.Slug,
		SuggestedTag:   parsed.SuggestedImageTag,
		SourceDir:      sourceDir,
		DockerfilePath: dockerfilePath,
		ContextPath:    contextPath,
		CreatedBy:      actorUserID,
	})
	if err != nil {
		return 0, "", err
	}
	return result.ImageID, result.ImageRef, nil
}

func marshalAWDChallengeConfig(value map[string]any) (string, error) {
	encoded, err := json.Marshal(cloneAWDChallengeConfig(value))
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func cloneAWDChallengeConfig(value map[string]any) map[string]any {
	if len(value) == 0 {
		return map[string]any{}
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return map[string]any{}
	}
	var cloned map[string]any
	if err := json.Unmarshal(encoded, &cloned); err != nil {
		return map[string]any{}
	}
	if cloned == nil {
		return map[string]any{}
	}
	return cloned
}

func awdCheckerArtifactPersistRequestFromParsed(parsed *domain.ParsedAWDChallengePackage) challengeports.AWDCheckerArtifactPersistRequest {
	if parsed == nil {
		return challengeports.AWDCheckerArtifactPersistRequest{}
	}
	files := make([]challengeports.AWDCheckerArtifactFile, 0, len(parsed.CheckerFiles))
	for _, file := range parsed.CheckerFiles {
		files = append(files, challengeports.AWDCheckerArtifactFile{
			Path: file.Path,
			Abs:  file.Abs,
		})
	}
	return challengeports.AWDCheckerArtifactPersistRequest{
		Slug:             parsed.Slug,
		CheckerType:      parsed.CheckerType,
		CheckerConfig:    cloneAWDChallengeConfig(parsed.CheckerConfig),
		CheckerEntryAbs:  parsed.CheckerEntryAbs,
		CheckerEntryPath: parsed.CheckerEntryPath,
		CheckerFiles:     files,
	}
}
