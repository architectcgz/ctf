package challengeimport

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/module/challenge/application/challengecatalog"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"ctf-platform/internal/module/challenge/domain"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
	platformevents "ctf-platform/internal/platform/events"
	"ctf-platform/internal/platform/randomstring"
	crypto "ctf-platform/internal/shared/flagcrypto"
)

type ImageBuildService interface {
	BuildPlatformTargetRef(challengeMode string, packageSlug string, suggestedTag string) (string, error)
}

const imageBuildServiceUnavailableBase = "当前后端未启用题包镜像构建/校验服务，请检查 registry 配置"

func requiresImageBuildService(sourceType string) bool {
	switch sourceType {
	case domain.ImageSourceTypePlatformBuild, domain.ImageSourceTypeExternalRef:
		return true
	default:
		return false
	}
}

func imageBuildServiceUnavailableMessage(sourceType string) string {
	switch sourceType {
	case domain.ImageSourceTypePlatformBuild:
		return imageBuildServiceUnavailableBase + "；该题包依赖平台镜像构建，当前无法提交导入。"
	case domain.ImageSourceTypeExternalRef:
		return imageBuildServiceUnavailableBase + "；该题包依赖外部镜像校验，当前无法提交导入。"
	default:
		return imageBuildServiceUnavailableBase
	}
}

func imageBuildServiceUnavailableError(sourceType string) error {
	return apperror.ErrServiceUnavailable.
		WithMessage(imageBuildServiceUnavailableMessage(sourceType)).
		WithCause(errors.New("image build service is not configured"))
}

func normalizeImageBuildService(imageBuild ImageBuildService) ImageBuildService {
	if imageBuildServiceIsNil(imageBuild) {
		return nil
	}
	return imageBuild
}

func imageBuildServiceIsNil(imageBuild ImageBuildService) bool {
	if imageBuild == nil {
		return true
	}
	value := reflect.ValueOf(imageBuild)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func missingImageBuildService(imageBuild ImageBuildService, sourceType string) bool {
	return requiresImageBuildService(sourceType) && imageBuildServiceIsNil(imageBuild)
}

func appendImageBuildWarning(warnings []string, sourceType string) []string {
	if !requiresImageBuildService(sourceType) {
		return warnings
	}

	warning := imageBuildServiceUnavailableMessage(sourceType)
	for _, existing := range warnings {
		if existing == warning {
			return warnings
		}
	}

	result := append([]string(nil), warnings...)
	result = append(result, warning)
	return result
}

func warnImageBuildServiceUnavailable(
	logger *zap.Logger,
	packageSlug string,
	sourceType string,
	action string,
) {
	if logger == nil {
		return
	}
	logger.Warn(
		imageBuildServiceUnavailableMessage(sourceType),
		zap.String("package_slug", packageSlug),
		zap.String("image_source_type", sourceType),
		zap.String("action", action),
	)
}

type ChallengeImportService struct {
	previewStore    challengeports.ChallengeImportPreviewStore
	attachmentStore challengeports.ChallengeAttachmentStore
	packageStorage  challengeports.ChallengePackageStorage
	importTxRunner  challengeports.ChallengeImportTxRunner
	imageBuild      ImageBuildService
	eventBus        platformevents.Bus
	logger          *zap.Logger
}

func NewChallengeImportService(
	previewStore challengeports.ChallengeImportPreviewStore,
	attachmentStore challengeports.ChallengeAttachmentStore,
	packageStorage challengeports.ChallengePackageStorage,
	txRunner challengeports.ChallengeImportTxRunner,
	imageBuild ImageBuildService,
	eventBus platformevents.Bus,
	logger *zap.Logger,
) *ChallengeImportService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ChallengeImportService{
		previewStore:    previewStore,
		attachmentStore: attachmentStore,
		packageStorage:  packageStorage,
		importTxRunner:  txRunner,
		imageBuild:      normalizeImageBuildService(imageBuild),
		eventBus:        eventBus,
		logger:          logger,
	}
}

func (s *ChallengeImportService) SetTxRunner(runner challengeports.ChallengeImportTxRunner) *ChallengeImportService {
	if s == nil {
		return nil
	}
	s.importTxRunner = runner
	return s
}

func (s *ChallengeImportService) SetImageBuildService(service ImageBuildService) {
	if s != nil {
		s.imageBuild = normalizeImageBuildService(service)
	}
}

func (s *ChallengeImportService) SetEventBus(bus platformevents.Bus) *ChallengeImportService {
	if s == nil {
		return nil
	}
	s.eventBus = bus
	return s
}

func (s *ChallengeImportService) PreviewChallengeImport(
	ctx context.Context,
	actorUserID int64,
	fileName string,
	reader io.Reader,
) (*challengecontracts.ChallengeImportPreviewResp, error) {
	if strings.TrimSpace(fileName) == "" {
		fileName = "challenge-package.zip"
	}
	if s.previewStore == nil {
		return nil, fmt.Errorf("challenge import preview store is not configured")
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

	parsed, err := domain.ParseChallengePackageDir(workspace.SourceDir)
	if err != nil {
		return nil, err
	}

	preview := s.buildChallengeImportPreview(previewID, fileName, parsed, time.Now().UTC())
	record := &challengeports.ChallengeImportPreviewRecord{
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

func (s *ChallengeImportService) GetChallengeImport(ctx context.Context, actorUserID int64, id string) (*challengecontracts.ChallengeImportPreviewResp, error) {
	if s.previewStore == nil {
		return nil, fmt.Errorf("challenge import preview store is not configured")
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

func (s *ChallengeImportService) ListChallengeImports(ctx context.Context, actorUserID int64) ([]challengecontracts.ChallengeImportPreviewResp, error) {
	if s.previewStore == nil {
		return nil, fmt.Errorf("challenge import preview store is not configured")
	}
	records, err := s.previewStore.ListRecords(ctx)
	if err != nil {
		return nil, err
	}

	previews := make([]challengecontracts.ChallengeImportPreviewResp, 0, len(records))
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

func (s *ChallengeImportService) CommitChallengeImport(
	ctx context.Context,
	actorUserID int64,
	id string,
) (*challengecontracts.ChallengeResp, error) {
	if s.previewStore == nil {
		return nil, fmt.Errorf("challenge import preview store is not configured")
	}
	record, err := s.previewStore.LoadRecord(ctx, id)
	if err != nil {
		return nil, err
	}
	if record.CreatedBy != 0 && record.CreatedBy != actorUserID {
		return nil, apperror.ErrForbidden
	}

	parsed, err := domain.ParseChallengePackageDir(record.SourceDir)
	if err != nil {
		return nil, err
	}
	if s.packageStorage == nil {
		return nil, fmt.Errorf("challenge package storage is not configured")
	}
	if s.attachmentStore == nil {
		return nil, fmt.Errorf("challenge attachment store is not configured")
	}
	if s.importTxRunner == nil {
		return nil, fmt.Errorf("challenge import tx runner is not configured")
	}

	buildSource, err := s.packageStorage.PersistImportedImageBuildSource(ctx, challengeports.ChallengeImportedImageBuildSourceRequest{
		ChallengeMode:  domain.ChallengePackageModeJeopardy,
		PackageSlug:    parsed.Slug,
		PreviewID:      record.ID,
		RootDir:        parsed.RootDir,
		DockerfilePath: parsed.DockerfilePath,
		ContextPath:    parsed.BuildContextPath,
	})
	if err != nil {
		return nil, err
	}

	attachmentURL, err := s.attachmentStore.PersistImportedAttachmentBundle(ctx, importedAttachmentBundleRequestFromParsed(parsed))
	if err != nil {
		if buildSource != nil {
			_ = s.packageStorage.DeletePath(ctx, buildSource.RootDir)
		}
		return nil, err
	}

	var challenge *challengeports.ImportedChallenge
	cleanupPaths := make([]string, 0, 2)
	var before challengecatalog.PublishedState
	var after challengecatalog.PublishedState
	if err := s.importTxRunner.WithinChallengeImportTransaction(ctx, func(store challengeports.ChallengeImportTxStore) error {
		if err := store.RejectImportedChallengeSlugConflict(ctx, parsed.Slug); err != nil {
			return err
		}

		resolvedImageID, err := s.resolveImportedImageIDForCommit(ctx, store, actorUserID, parsed, buildSource)
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		var current challengeports.ImportedChallenge
		existing, found, findErr := store.FindLegacyChallengeForImportedPackageCreate(ctx, parsed.Title, parsed.Category)

		switch {
		case findErr != nil:
			return findErr
		case !found:
			current = challengeports.ImportedChallenge{
				PackageSlug:    stringPointer(parsed.Slug),
				Title:          parsed.Title,
				Description:    parsed.Description,
				Category:       parsed.Category,
				Difficulty:     parsed.Difficulty,
				Points:         parsed.Points,
				ImageID:        resolvedImageID,
				AttachmentURL:  attachmentURL,
				Status:         challengecontracts.ChallengeStatusDraft,
				FlagPrefix:     parsed.FlagPrefix,
				TargetProtocol: parsed.RuntimeProtocol,
				TargetPort:     parsed.RuntimePort,
				CreatedBy:      &actorUserID,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			if err := store.CreateImportedChallenge(ctx, &current); err != nil {
				return fmt.Errorf("create imported challenge %s: %w", parsed.Slug, err)
			}
			before = challengecatalog.PublishedState{}
			after = challengecatalog.PublishedStateFromImportedChallenge(&current)
		default:
			current = *existing
			before = challengecatalog.PublishedStateFromImportedChallenge(existing)
			current.PackageSlug = stringPointer(parsed.Slug)
			current.Title = parsed.Title
			current.Description = parsed.Description
			current.Category = parsed.Category
			current.Difficulty = parsed.Difficulty
			current.Points = parsed.Points
			current.ImageID = resolvedImageID
			current.AttachmentURL = attachmentURL
			current.Status = challengecontracts.ChallengeStatusDraft
			current.TargetProtocol = parsed.RuntimeProtocol
			current.TargetPort = parsed.RuntimePort
			current.UpdatedAt = now
			updates := map[string]any{
				"package_slug":    parsed.Slug,
				"title":           parsed.Title,
				"description":     parsed.Description,
				"category":        parsed.Category,
				"difficulty":      parsed.Difficulty,
				"points":          parsed.Points,
				"image_id":        resolvedImageID,
				"attachment_url":  attachmentURL,
				"status":          challengecontracts.ChallengeStatusDraft,
				"target_protocol": parsed.RuntimeProtocol,
				"target_port":     parsed.RuntimePort,
				"deleted_at":      nil,
				"updated_at":      now,
			}
			if err := store.UpdateImportedChallenge(ctx, &current, updates); err != nil {
				return fmt.Errorf("update imported challenge %s: %w", parsed.Slug, err)
			}
			after = challengecatalog.PublishedStateFromImportedChallenge(&current)
		}

		if err := store.ClearPublishCheckJobs(ctx, current.ID); err != nil {
			return fmt.Errorf("clear imported challenge publish check jobs %s: %w", parsed.Slug, err)
		}

		if err := store.ReplaceImportedHints(ctx, current.ID, buildImportedChallengeHints(current.ID, parsed.Hints, time.Now().UTC())); err != nil {
			return err
		}
		flagUpdates, err := buildImportedFlagUpdates(parsed.FlagType, parsed.FlagPrefix, parsed.FlagValue, time.Now().UTC())
		if err != nil {
			return err
		}
		if err := store.ApplyImportedFlagUpdates(ctx, current.ID, flagUpdates); err != nil {
			return err
		}
		if parsed.Topology != nil {
			revision, revisionErr := s.createImportedPackageRevision(ctx, store, actorUserID, &current, record, parsed)
			if revisionErr != nil {
				return revisionErr
			}
			cleanupPaths = append(cleanupPaths, revision.SourceDir)
			if strings.TrimSpace(revision.ArchivePath) != "" {
				cleanupPaths = append(cleanupPaths, revision.ArchivePath)
			}

			topologySpec, entryNodeKey, topologyErr := domain.BuildTopologySpecFromImportedPackage(
				parsed.Topology,
				func(imageRef string) (int64, error) {
					if parsed.ImageSourceType == domain.ImageSourceTypePlatformBuild && resolvedImageID != nil {
						return *resolvedImageID, nil
					}
					imageID, err := s.resolveExternalImageRefForCommit(ctx, store, parsed.Slug, imageRef)
					if err != nil {
						return 0, err
					}
					if imageID == nil {
						return 0, nil
					}
					return *imageID, nil
				},
			)
			if topologyErr != nil {
				return topologyErr
			}
			now = time.Now().UTC()
			revisionID := revision.ID
			item := &challengeentity.ChallengeTopology{
				ChallengeID:          current.ID,
				EntryNodeKey:         entryNodeKey,
				Spec:                 topologySpec,
				SourceType:           challengeentity.ChallengeTopologySourceTypePackageImport,
				SourcePath:           parsed.Topology.Source,
				PackageRevisionID:    &revisionID,
				PackageBaselineSpec:  topologySpec,
				SyncStatus:           challengeentity.ChallengeTopologySyncStatusClean,
				LastExportRevisionID: nil,
				UpdatedAt:            now,
			}
			if err := store.UpsertImportedTopology(ctx, item); err != nil {
				return err
			}
		}

		challenge = &current
		return nil
	}); err != nil {
		if buildSource != nil {
			_ = s.packageStorage.DeletePath(ctx, buildSource.RootDir)
		}
		for _, cleanupPath := range cleanupPaths {
			if strings.TrimSpace(cleanupPath) == "" {
				continue
			}
			_ = s.packageStorage.DeletePath(ctx, cleanupPath)
		}
		return nil, err
	}

	_ = s.previewStore.DeleteWorkspace(ctx, id)
	challengecatalog.PublishPublishedCatalogChangedEvent(
		ctx,
		s.logger,
		s.eventBus,
		challengecontracts.ChallengeCatalogChangeTypeImported,
		before,
		after,
	)
	return domain.ChallengeRespFromWriteModel(toImportedChallengeResponseWriteModel(challenge), nil), nil
}

func importedAttachmentBundleRequestFromParsed(parsed *domain.ParsedChallengePackage) challengeports.ChallengeImportedAttachmentBundleRequest {
	if parsed == nil {
		return challengeports.ChallengeImportedAttachmentBundleRequest{}
	}
	attachments := make([]challengeports.ChallengeImportedAttachment, 0, len(parsed.Attachments))
	for _, attachment := range parsed.Attachments {
		attachments = append(attachments, challengeports.ChallengeImportedAttachment{
			Name:         attachment.Name,
			Path:         attachment.Path,
			AbsolutePath: attachment.AbsolutePath,
		})
	}
	return challengeports.ChallengeImportedAttachmentBundleRequest{
		PackageSlug: parsed.Slug,
		Attachments: attachments,
	}
}

func (s *ChallengeImportService) buildChallengeImportPreview(
	id string,
	fileName string,
	parsed *domain.ParsedChallengePackage,
	createdAt time.Time,
) *challengecontracts.ChallengeImportPreviewResp {
	var imageBuild ImageBuildService
	if s != nil {
		imageBuild = s.imageBuild
	}
	var logger *zap.Logger
	if s != nil {
		logger = s.logger
	}

	attachments := make([]challengecontracts.ChallengeImportAttachmentResp, 0, len(parsed.Attachments))
	for _, attachment := range parsed.Attachments {
		attachments = append(attachments, challengecontracts.ChallengeImportAttachmentResp{
			Name: attachment.Name,
			Path: attachment.Path,
		})
	}

	hints := make([]challengecontracts.ChallengeHintAdminResp, 0, len(parsed.Hints))
	for _, hint := range parsed.Hints {
		hints = append(hints, challengecontracts.ChallengeHintAdminResp{
			Level:   hint.Level,
			Title:   hint.Title,
			Content: hint.Content,
		})
	}

	imageDelivery := challengecontracts.ChallengeImportImageDeliveryResp{
		SourceType:   parsed.ImageSourceType,
		SuggestedTag: parsed.SuggestedImageTag,
	}
	if parsed.ImageSourceType == domain.ImageSourceTypePlatformBuild && imageBuild != nil {
		if targetRef, err := imageBuild.BuildPlatformTargetRef(domain.ChallengePackageModeJeopardy, parsed.Slug, parsed.SuggestedImageTag); err == nil {
			imageDelivery.TargetImageRef = targetRef
			imageDelivery.BuildStatus = challengeentity.ImageStatusPending
		}
	}
	warnings := append([]string(nil), parsed.Warnings...)
	if missingImageBuildService(imageBuild, parsed.ImageSourceType) {
		warnImageBuildServiceUnavailable(logger, parsed.Slug, parsed.ImageSourceType, "preview")
		warnings = appendImageBuildWarning(warnings, parsed.ImageSourceType)
	}

	return &challengecontracts.ChallengeImportPreviewResp{
		ID:          id,
		FileName:    fileName,
		Slug:        parsed.Slug,
		Title:       parsed.Title,
		Description: parsed.Description,
		Category:    parsed.Category,
		Difficulty:  parsed.Difficulty,
		Points:      parsed.Points,
		Attachments: attachments,
		Hints:       hints,
		Flag: challengecontracts.ChallengeImportFlagResp{
			Type:   parsed.FlagType,
			Prefix: parsed.FlagPrefix,
		},
		Runtime: challengecontracts.ChallengeImportRuntimeResp{
			Type:     parsed.Manifest.Runtime.Type,
			ImageRef: parsed.RuntimeImageRef,
		},
		ImageDelivery: imageDelivery,
		Extensions: challengecontracts.ChallengeImportExtensionsResp{
			Topology: challengecontracts.ChallengeImportTopologyExtensionResp{
				Source:  parsed.Manifest.Extensions.Topology.Source,
				Enabled: parsed.Manifest.Extensions.Topology.Enabled,
			},
		},
		Topology:     domain.ChallengeImportTopologyRespFromParsed(parsed.Topology),
		PackageFiles: domain.ChallengePackageFileRespList(parsed.PackageFiles),
		Warnings:     warnings,
		CreatedAt:    createdAt,
	}
}

func (s *ChallengeImportService) resolveImportedImageIDForCommit(
	ctx context.Context,
	store challengeports.ChallengeImportTxStore,
	actorUserID int64,
	parsed *domain.ParsedChallengePackage,
	buildSource *challengeports.ChallengeStoredImageBuildSource,
) (*int64, error) {
	if parsed.ImageSourceType == domain.ImageSourceTypeExternalRef {
		return s.resolveExternalImageRefForCommit(ctx, store, parsed.Slug, parsed.RuntimeImageRef)
	}
	if parsed.ImageSourceType != domain.ImageSourceTypePlatformBuild {
		resolution, err := store.ResolveExistingImageRef(ctx, parsed.Slug, parsed.RuntimeImageRef)
		if err != nil {
			return nil, err
		}
		return resolution.ImageID, nil
	}
	var imageBuild ImageBuildService
	var logger *zap.Logger
	if s != nil {
		imageBuild = s.imageBuild
		logger = s.logger
	}
	if missingImageBuildService(imageBuild, parsed.ImageSourceType) {
		warnImageBuildServiceUnavailable(logger, parsed.Slug, parsed.ImageSourceType, "commit")
		return nil, imageBuildServiceUnavailableError(parsed.ImageSourceType)
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
		ChallengeMode:  domain.ChallengePackageModeJeopardy,
		PackageSlug:    parsed.Slug,
		SuggestedTag:   parsed.SuggestedImageTag,
		SourceDir:      sourceDir,
		DockerfilePath: dockerfilePath,
		ContextPath:    contextPath,
		CreatedBy:      actorUserID,
	})
	if err != nil {
		return nil, err
	}
	return result.ImageID, nil
}

func (s *ChallengeImportService) resolveExternalImageRefForCommit(
	ctx context.Context,
	store challengeports.ChallengeImportTxStore,
	packageSlug string,
	imageRef string,
) (*int64, error) {
	if strings.TrimSpace(imageRef) == "" {
		return nil, nil
	}
	var imageBuild ImageBuildService
	var logger *zap.Logger
	if s != nil {
		imageBuild = s.imageBuild
		logger = s.logger
	}
	if missingImageBuildService(imageBuild, domain.ImageSourceTypeExternalRef) {
		warnImageBuildServiceUnavailable(logger, packageSlug, domain.ImageSourceTypeExternalRef, "commit")
		return nil, imageBuildServiceUnavailableError(domain.ImageSourceTypeExternalRef)
	}
	result, err := store.ResolveExternalImage(ctx, packageSlug, imageRef)
	if err != nil {
		return nil, err
	}
	return result.ImageID, nil
}

func buildImportedFlagUpdates(
	flagType string,
	prefix string,
	value string,
	updatedAt time.Time,
) (map[string]any, error) {
	switch flagType {
	case challengecontracts.FlagTypeStatic:
		return buildImportedStaticFlagUpdates(prefix, value, updatedAt)
	case challengecontracts.FlagTypeDynamic:
		return buildImportedDynamicFlagUpdates(prefix, updatedAt), nil
	case challengecontracts.FlagTypeRegex:
		return buildImportedRegexFlagUpdates(prefix, value, updatedAt)
	case challengecontracts.FlagTypeManualReview:
		return buildImportedManualReviewFlagUpdates(prefix, updatedAt), nil
	default:
		return nil, apperror.ErrInvalidParams.WithCause(errors.New("不支持的 flag 类型"))
	}
}

func buildImportedStaticFlagUpdates(prefix string, value string, updatedAt time.Time) (map[string]any, error) {
	salt, err := randomstring.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate salt for imported challenge: %w", err)
	}
	return map[string]any{
		"flag_type":   challengecontracts.FlagTypeStatic,
		"flag_salt":   salt,
		"flag_hash":   crypto.HashStaticFlag(value, salt),
		"flag_regex":  "",
		"flag_prefix": prefix,
		"updated_at":  updatedAt,
	}, nil
}

func buildImportedDynamicFlagUpdates(prefix string, updatedAt time.Time) map[string]any {
	return map[string]any{
		"flag_type":   challengecontracts.FlagTypeDynamic,
		"flag_salt":   "",
		"flag_hash":   "",
		"flag_regex":  "",
		"flag_prefix": prefix,
		"updated_at":  updatedAt,
	}
}

func buildImportedRegexFlagUpdates(prefix string, value string, updatedAt time.Time) (map[string]any, error) {
	compiled, err := regexp.Compile(strings.TrimSpace(value))
	if err != nil {
		return nil, apperror.ErrInvalidParams.WithCause(fmt.Errorf("regex flag 无效: %w", err))
	}
	return map[string]any{
		"flag_type":   challengecontracts.FlagTypeRegex,
		"flag_salt":   "",
		"flag_hash":   "",
		"flag_regex":  compiled.String(),
		"flag_prefix": prefix,
		"updated_at":  updatedAt,
	}, nil
}

func buildImportedManualReviewFlagUpdates(prefix string, updatedAt time.Time) map[string]any {
	return map[string]any{
		"flag_type":   challengecontracts.FlagTypeManualReview,
		"flag_salt":   "",
		"flag_hash":   "",
		"flag_regex":  "",
		"flag_prefix": prefix,
		"updated_at":  updatedAt,
	}
}

func buildImportedChallengeHints(
	challengeID int64,
	hints []domain.ParsedChallengePackageHint,
	now time.Time,
) []challengeentity.ChallengeHint {
	if len(hints) == 0 {
		return nil
	}

	records := make([]challengeentity.ChallengeHint, 0, len(hints))
	for _, hint := range hints {
		records = append(records, challengeentity.ChallengeHint{
			ChallengeID: challengeID,
			Level:       hint.Level,
			Title:       hint.Title,
			Content:     hint.Content,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}
	return records
}

func toImportedChallengeResponseWriteModel(source *challengeports.ImportedChallenge) *challengeports.ChallengeWriteModel {
	if source == nil {
		return nil
	}
	return &challengeports.ChallengeWriteModel{
		ID:             source.ID,
		PackageSlug:    source.PackageSlug,
		Title:          source.Title,
		Description:    source.Description,
		Category:       source.Category,
		Difficulty:     source.Difficulty,
		Points:         source.Points,
		ImageID:        source.ImageID,
		AttachmentURL:  source.AttachmentURL,
		Status:         source.Status,
		FlagPrefix:     source.FlagPrefix,
		TargetProtocol: source.TargetProtocol,
		TargetPort:     source.TargetPort,
		CreatedBy:      source.CreatedBy,
		CreatedAt:      source.CreatedAt,
		UpdatedAt:      source.UpdatedAt,
	}
}

func generateChallengeImportPreviewID() (string, error) {
	token := make([]byte, 12)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}

func stringPointer(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func int64Ptr(value int64) *int64 {
	return &value
}
