package commands

import (
	"archive/zip"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	platformevents "ctf-platform/internal/platform/events"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
)

const (
	defaultChallengeImportPreviewRoot      = "./data/challenge-import-previews"
	defaultChallengeImportedAttachmentRoot = "./data/challenge-attachments"
	defaultChallengePackageSourceRoot      = "./data/challenge-packages"
	defaultChallengePackageExportRoot      = "./data/challenge-package-exports"
	maxChallengeImportArchiveFiles         = 128
	maxChallengeImportArchiveFileSize      = 16 << 20
	maxChallengeImportArchiveTotalSize     = 64 << 20
)

type SelfCheckConfig struct {
	RuntimeCreateTimeout     time.Duration
	FlagGlobalSecret         string
	PublishCheckPollInterval time.Duration
	PublishCheckBatchSize    int
}

type ChallengeService struct {
	db           *gorm.DB
	repo         challengeCommandRepository
	imageRepo    challengeports.ImageQueryRepository
	topologyRepo challengeports.ChallengeTopologyReadRepository
	packageRepo  challengeports.ChallengePackageRevisionRepository
	runtimeProbe challengeports.ChallengeRuntimeProbe
	imageBuild   *ImageBuildService
	eventBus     platformevents.Bus
	selfCheckCfg SelfCheckConfig
	logger       *zap.Logger
}

func NewChallengeService(
	db *gorm.DB,
	repo challengeCommandRepository,
	imageRepo challengeports.ImageQueryRepository,
	topologyRepo challengeports.ChallengeTopologyReadRepository,
	packageRepo challengeports.ChallengePackageRevisionRepository,
	runtimeProbe challengeports.ChallengeRuntimeProbe,
	cfg SelfCheckConfig,
	logger *zap.Logger,
) *ChallengeService {
	if logger == nil {
		logger = zap.NewNop()
	}
	if cfg.RuntimeCreateTimeout <= 0 {
		cfg.RuntimeCreateTimeout = 60 * time.Second
	}
	if cfg.PublishCheckPollInterval <= 0 {
		cfg.PublishCheckPollInterval = 2 * time.Second
	}
	if cfg.PublishCheckBatchSize <= 0 {
		cfg.PublishCheckBatchSize = 1
	}
	service := &ChallengeService{
		db:           db,
		repo:         repo,
		imageRepo:    imageRepo,
		topologyRepo: topologyRepo,
		packageRepo:  packageRepo,
		runtimeProbe: runtimeProbe,
		selfCheckCfg: cfg,
		logger:       logger,
	}
	return service
}

type storedChallengeImportPreview struct {
	ID        string                         `json:"id"`
	FileName  string                         `json:"file_name"`
	SourceDir string                         `json:"source_dir"`
	CreatedBy int64                          `json:"created_by"`
	CreatedAt time.Time                      `json:"created_at"`
	Preview   dto.ChallengeImportPreviewResp `json:"preview"`
}

func (s *ChallengeService) PreviewChallengeImport(
	ctx context.Context,
	actorUserID int64,
	fileName string,
	reader io.Reader,
) (*dto.ChallengeImportPreviewResp, error) {
	_ = ctx
	if strings.TrimSpace(fileName) == "" {
		fileName = "challenge-package.zip"
	}

	previewID, err := generateChallengeImportPreviewID()
	if err != nil {
		return nil, err
	}

	previewDir := filepath.Join(challengeImportPreviewRoot(), previewID)
	archivePath := filepath.Join(previewDir, "package.zip")
	extractDir := filepath.Join(previewDir, "source")
	if err := os.MkdirAll(previewDir, 0o755); err != nil {
		return nil, fmt.Errorf("create preview dir: %w", err)
	}

	if err := writeImportUploadArchive(archivePath, reader); err != nil {
		return nil, err
	}

	rootDir, err := extractChallengeImportArchive(archivePath, extractDir)
	if err != nil {
		return nil, err
	}

	parsed, err := domain.ParseChallengePackageDir(rootDir)
	if err != nil {
		return nil, err
	}

	preview := s.buildChallengeImportPreview(previewID, fileName, parsed, time.Now())
	record := storedChallengeImportPreview{
		ID:        previewID,
		FileName:  fileName,
		SourceDir: rootDir,
		CreatedBy: actorUserID,
		CreatedAt: preview.CreatedAt,
		Preview:   *preview,
	}
	if err := saveChallengeImportPreviewRecord(previewDir, &record); err != nil {
		return nil, err
	}
	return preview, nil
}

func (s *ChallengeService) GetChallengeImport(ctx context.Context, actorUserID int64, id string) (*dto.ChallengeImportPreviewResp, error) {
	_ = ctx
	record, err := loadChallengeImportPreviewRecord(id)
	if err != nil {
		return nil, err
	}
	if record.CreatedBy != 0 && record.CreatedBy != actorUserID {
		return nil, errcode.ErrForbidden
	}
	preview := record.Preview
	return &preview, nil
}

func (s *ChallengeService) ListChallengeImports(ctx context.Context, actorUserID int64) ([]dto.ChallengeImportPreviewResp, error) {
	_ = ctx
	records, err := loadChallengeImportPreviewRecords()
	if err != nil {
		return nil, err
	}

	previews := make([]dto.ChallengeImportPreviewResp, 0, len(records))
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

func (s *ChallengeService) CommitChallengeImport(
	ctx context.Context,
	actorUserID int64,
	id string,
) (*dto.ChallengeResp, error) {
	record, err := loadChallengeImportPreviewRecord(id)
	if err != nil {
		return nil, err
	}
	if record.CreatedBy != 0 && record.CreatedBy != actorUserID {
		return nil, errcode.ErrForbidden
	}

	parsed, err := domain.ParseChallengePackageDir(record.SourceDir)
	if err != nil {
		return nil, err
	}

	buildSource, err := persistImportedImageBuildSource(
		domain.ChallengePackageModeJeopardy,
		parsed.Slug,
		record.ID,
		parsed.RootDir,
		parsed.DockerfilePath,
		parsed.BuildContextPath,
	)
	if err != nil {
		return nil, err
	}

	attachmentURL, err := persistImportedAttachmentBundle(parsed)
	if err != nil {
		if buildSource != nil {
			_ = os.RemoveAll(buildSource.RootDir)
		}
		return nil, err
	}

	var challenge *model.Challenge
	cleanupPaths := make([]string, 0, 2)
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := rejectImportedChallengeSlugConflict(tx, parsed.Slug); err != nil {
			return err
		}

		resolvedImageID, err := s.resolveImportedImageIDForCommit(ctx, tx, actorUserID, parsed, buildSource)
		if err != nil {
			return err
		}

		now := time.Now()
		var current model.Challenge
		findErr := findLegacyChallengeForImportedPackageCreate(tx, parsed.Title, parsed.Category, &current)

		switch {
		case errors.Is(findErr, gorm.ErrRecordNotFound):
			current = model.Challenge{
				PackageSlug:    stringPointer(parsed.Slug),
				Title:          parsed.Title,
				Description:    parsed.Description,
				Category:       parsed.Category,
				Difficulty:     parsed.Difficulty,
				Points:         parsed.Points,
				ImageID:        resolvedImageID,
				AttachmentURL:  attachmentURL,
				Status:         model.ChallengeStatusDraft,
				FlagPrefix:     parsed.FlagPrefix,
				TargetProtocol: parsed.RuntimeProtocol,
				TargetPort:     parsed.RuntimePort,
				CreatedBy:      &actorUserID,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			if err := tx.Create(&current).Error; err != nil {
				return fmt.Errorf("create imported challenge %s: %w", parsed.Slug, err)
			}
		case findErr != nil:
			return fmt.Errorf("find imported challenge %s: %w", parsed.Slug, findErr)
		default:
			updates := map[string]any{
				"package_slug":    parsed.Slug,
				"title":           parsed.Title,
				"description":     parsed.Description,
				"category":        parsed.Category,
				"difficulty":      parsed.Difficulty,
				"points":          parsed.Points,
				"image_id":        resolvedImageID,
				"attachment_url":  attachmentURL,
				"status":          model.ChallengeStatusDraft,
				"target_protocol": parsed.RuntimeProtocol,
				"target_port":     parsed.RuntimePort,
				"deleted_at":      nil,
				"updated_at":      now,
			}
			if err := tx.Unscoped().Model(&current).Updates(updates).Error; err != nil {
				return fmt.Errorf("update imported challenge %s: %w", parsed.Slug, err)
			}
		}

		if err := tx.Where("challenge_id = ?", current.ID).Delete(&model.ChallengePublishCheckJob{}).Error; err != nil {
			return fmt.Errorf("clear imported challenge publish check jobs %s: %w", parsed.Slug, err)
		}

		if err := syncImportedChallengeHints(tx, current.ID, parsed.Hints); err != nil {
			return err
		}
		if err := configureImportedFlag(tx, current.ID, parsed.FlagType, parsed.FlagPrefix, parsed.FlagValue); err != nil {
			return err
		}
		if parsed.Topology != nil {
			revision, revisionErr := s.createImportedPackageRevision(tx, actorUserID, &current, *record, parsed)
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
					if parsed.ImageSourceType == domain.ImageSourceTypePlatformBuild && resolvedImageID > 0 {
						return resolvedImageID, nil
					}
					return s.resolveExternalImageRefForCommit(ctx, tx, parsed.Slug, imageRef)
				},
			)
			if topologyErr != nil {
				return topologyErr
			}
			now = time.Now()
			revisionID := revision.ID
			item := &model.ChallengeTopology{
				ChallengeID:          current.ID,
				EntryNodeKey:         entryNodeKey,
				Spec:                 topologySpec,
				SourceType:           model.ChallengeTopologySourceTypePackageImport,
				SourcePath:           parsed.Topology.Source,
				PackageRevisionID:    &revisionID,
				PackageBaselineSpec:  topologySpec,
				SyncStatus:           model.ChallengeTopologySyncStatusClean,
				LastExportRevisionID: nil,
				UpdatedAt:            now,
			}
			if err := upsertChallengeTopologyTx(tx, item); err != nil {
				return err
			}
		}

		challenge = &current
		return nil
	}); err != nil {
		if buildSource != nil {
			_ = os.RemoveAll(buildSource.RootDir)
		}
		for _, cleanupPath := range cleanupPaths {
			if strings.TrimSpace(cleanupPath) == "" {
				continue
			}
			_ = os.RemoveAll(cleanupPath)
		}
		return nil, err
	}

	_ = os.RemoveAll(filepath.Join(challengeImportPreviewRoot(), id))
	return domain.ChallengeRespFromModel(challenge, nil), nil
}

func (s *ChallengeService) buildChallengeImportPreview(
	id string,
	fileName string,
	parsed *domain.ParsedChallengePackage,
	createdAt time.Time,
) *dto.ChallengeImportPreviewResp {
	var imageBuild *ImageBuildService
	if s != nil {
		imageBuild = s.imageBuild
	}
	var logger *zap.Logger
	if s != nil {
		logger = s.logger
	}

	attachments := make([]dto.ChallengeImportAttachmentResp, 0, len(parsed.Attachments))
	for _, attachment := range parsed.Attachments {
		attachments = append(attachments, dto.ChallengeImportAttachmentResp{
			Name: attachment.Name,
			Path: attachment.Path,
		})
	}

	hints := make([]dto.ChallengeHintAdminResp, 0, len(parsed.Hints))
	for _, hint := range parsed.Hints {
		hints = append(hints, dto.ChallengeHintAdminResp{
			Level:   hint.Level,
			Title:   hint.Title,
			Content: hint.Content,
		})
	}

	imageDelivery := dto.ChallengeImportImageDeliveryResp{
		SourceType:   parsed.ImageSourceType,
		SuggestedTag: parsed.SuggestedImageTag,
	}
	if parsed.ImageSourceType == domain.ImageSourceTypePlatformBuild && imageBuild != nil {
		if targetRef, err := imageBuild.BuildPlatformTargetRef(domain.ChallengePackageModeJeopardy, parsed.Slug, parsed.SuggestedImageTag); err == nil {
			imageDelivery.TargetImageRef = targetRef
			imageDelivery.BuildStatus = model.ImageStatusPending
		}
	}
	warnings := append([]string(nil), parsed.Warnings...)
	if challengeImportMissingImageBuildService(imageBuild, parsed.ImageSourceType) {
		warnChallengeImportImageBuildServiceUnavailable(logger, parsed.Slug, parsed.ImageSourceType, "preview")
		warnings = appendChallengeImportImageBuildWarning(warnings, parsed.ImageSourceType)
	}

	return &dto.ChallengeImportPreviewResp{
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
		Flag: dto.ChallengeImportFlagResp{
			Type:   parsed.FlagType,
			Prefix: parsed.FlagPrefix,
		},
		Runtime: dto.ChallengeImportRuntimeResp{
			Type:     parsed.Manifest.Runtime.Type,
			ImageRef: parsed.RuntimeImageRef,
		},
		ImageDelivery: imageDelivery,
		Extensions: dto.ChallengeImportExtensionsResp{
			Topology: dto.ChallengeImportTopologyExtensionResp{
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

func (s *ChallengeService) resolveImportedImageIDForCommit(
	ctx context.Context,
	tx *gorm.DB,
	actorUserID int64,
	parsed *domain.ParsedChallengePackage,
	buildSource *importedImageBuildSource,
) (int64, error) {
	if parsed.ImageSourceType == domain.ImageSourceTypeExternalRef {
		return s.resolveExternalImageRefForCommit(ctx, tx, parsed.Slug, parsed.RuntimeImageRef)
	}
	if parsed.ImageSourceType != domain.ImageSourceTypePlatformBuild {
		return resolveImportedImageID(tx, parsed.Slug, parsed.RuntimeImageRef)
	}
	var imageBuild *ImageBuildService
	var logger *zap.Logger
	if s != nil {
		imageBuild = s.imageBuild
		logger = s.logger
	}
	if challengeImportMissingImageBuildService(imageBuild, parsed.ImageSourceType) {
		warnChallengeImportImageBuildServiceUnavailable(logger, parsed.Slug, parsed.ImageSourceType, "commit")
		return 0, challengeImportImageBuildServiceUnavailableError(parsed.ImageSourceType)
	}
	sourceDir := parsed.RootDir
	dockerfilePath := parsed.DockerfilePath
	contextPath := parsed.BuildContextPath
	if buildSource != nil {
		sourceDir = buildSource.SourceDir
		dockerfilePath = buildSource.DockerfilePath
		contextPath = buildSource.ContextPath
	}
	result, err := imageBuild.CreatePlatformBuildJobInTx(ctx, tx, CreatePlatformBuildJobRequest{
		ChallengeMode:  domain.ChallengePackageModeJeopardy,
		PackageSlug:    parsed.Slug,
		SuggestedTag:   parsed.SuggestedImageTag,
		SourceDir:      sourceDir,
		DockerfilePath: dockerfilePath,
		ContextPath:    contextPath,
		CreatedBy:      actorUserID,
	})
	if err != nil {
		return 0, err
	}
	return result.ImageID, nil
}

func (s *ChallengeService) resolveExternalImageRefForCommit(
	ctx context.Context,
	tx *gorm.DB,
	packageSlug string,
	imageRef string,
) (int64, error) {
	if strings.TrimSpace(imageRef) == "" {
		return 0, nil
	}
	var imageBuild *ImageBuildService
	var logger *zap.Logger
	if s != nil {
		imageBuild = s.imageBuild
		logger = s.logger
	}
	if challengeImportMissingImageBuildService(imageBuild, domain.ImageSourceTypeExternalRef) {
		warnChallengeImportImageBuildServiceUnavailable(logger, packageSlug, domain.ImageSourceTypeExternalRef, "commit")
		return 0, challengeImportImageBuildServiceUnavailableError(domain.ImageSourceTypeExternalRef)
	}
	result, err := imageBuild.VerifyExternalImageRefInTx(ctx, tx, packageSlug, imageRef)
	if err != nil {
		return 0, err
	}
	return result.ImageID, nil
}

func writeImportUploadArchive(targetPath string, reader io.Reader) error {
	file, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("create preview archive: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("save preview archive: %w", err)
	}
	return nil
}

func extractChallengeImportArchive(archivePath, extractDir string) (string, error) {
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", errcode.ErrInvalidParams.WithCause(fmt.Errorf("读取 zip 失败: %w", err))
	}
	defer archive.Close()

	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		return "", fmt.Errorf("create extract dir: %w", err)
	}

	stats := challengeImportArchiveStats{}
	for _, file := range archive.File {
		if err := stats.accept(file); err != nil {
			return "", err
		}
		if err := extractChallengeImportFile(extractDir, file); err != nil {
			return "", err
		}
	}

	rootDir, err := resolveExtractedChallengeImportRoot(extractDir)
	if err != nil {
		return "", err
	}
	return rootDir, nil
}

type challengeImportArchiveStats struct {
	fileCount int
	totalSize uint64
}

func (s *challengeImportArchiveStats) accept(file *zip.File) error {
	if file == nil {
		return nil
	}
	if file.Mode()&os.ModeSymlink != 0 {
		return errcode.ErrInvalidParams.WithCause(fmt.Errorf("zip 条目不允许符号链接: %s", file.Name))
	}
	if file.FileInfo().IsDir() {
		return nil
	}

	s.fileCount++
	if s.fileCount > maxChallengeImportArchiveFiles {
		return errcode.ErrInvalidParams.WithCause(
			fmt.Errorf("zip 文件数量超过限制，最多允许 %d 个文件", maxChallengeImportArchiveFiles),
		)
	}
	if file.UncompressedSize64 > maxChallengeImportArchiveFileSize {
		return errcode.ErrInvalidParams.WithCause(
			fmt.Errorf("zip 单文件超过限制，最多允许 %d 字节", maxChallengeImportArchiveFileSize),
		)
	}

	s.totalSize += file.UncompressedSize64
	if s.totalSize > maxChallengeImportArchiveTotalSize {
		return errcode.ErrInvalidParams.WithCause(
			fmt.Errorf("zip 解包总大小超过限制，最多允许 %d 字节", maxChallengeImportArchiveTotalSize),
		)
	}
	return nil
}

func extractChallengeImportFile(baseDir string, file *zip.File) error {
	relativePath := strings.TrimSpace(file.Name)
	if relativePath == "" {
		return nil
	}

	targetPath := filepath.Clean(filepath.Join(baseDir, relativePath))
	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return err
	}
	prefix := baseAbs + string(os.PathSeparator)
	targetAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}
	if targetAbs != baseAbs && !strings.HasPrefix(targetAbs, prefix) {
		return errcode.ErrInvalidParams.WithCause(fmt.Errorf("zip 条目路径非法: %s", relativePath))
	}

	if file.FileInfo().IsDir() {
		return os.MkdirAll(targetAbs, 0o755)
	}

	if err := os.MkdirAll(filepath.Dir(targetAbs), 0o755); err != nil {
		return err
	}
	source, err := file.Open()
	if err != nil {
		return err
	}
	defer source.Close()

	target, err := os.Create(targetAbs)
	if err != nil {
		return err
	}
	defer target.Close()

	if _, err := io.Copy(target, source); err != nil {
		return err
	}
	return nil
}

func resolveExtractedChallengeImportRoot(extractDir string) (string, error) {
	directManifest := filepath.Join(extractDir, "challenge.yml")
	if _, err := os.Stat(directManifest); err == nil {
		return extractDir, nil
	}

	entries, err := os.ReadDir(extractDir)
	if err != nil {
		return "", err
	}
	if len(entries) != 1 || !entries[0].IsDir() {
		return "", errcode.ErrInvalidParams.WithCause(errors.New("zip 根目录必须直接包含 challenge.yml 或单一题目目录"))
	}

	rootDir := filepath.Join(extractDir, entries[0].Name())
	if _, err := os.Stat(filepath.Join(rootDir, "challenge.yml")); err != nil {
		return "", errcode.ErrInvalidParams.WithCause(errors.New("未找到 challenge.yml"))
	}
	return rootDir, nil
}

func findLegacyChallengeForImportedPackageCreate(
	tx *gorm.DB,
	title string,
	category string,
	challenge *model.Challenge,
) error {
	if challenge == nil {
		return fmt.Errorf("challenge target is nil")
	}

	return tx.Unscoped().
		Where("(package_slug IS NULL OR package_slug = '') AND title = ? AND category = ?", title, category).
		First(challenge).Error
}

func rejectImportedChallengeSlugConflict(tx *gorm.DB, packageSlug string) error {
	slug := strings.TrimSpace(packageSlug)
	if slug == "" {
		return nil
	}

	var existing model.Challenge
	err := tx.Unscoped().
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

func saveChallengeImportPreviewRecord(previewDir string, record *storedChallengeImportPreview) error {
	content, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(previewDir, "preview.json"), content, 0o644)
}

func loadChallengeImportPreviewRecord(id string) (*storedChallengeImportPreview, error) {
	content, err := os.ReadFile(filepath.Join(challengeImportPreviewRoot(), id, "preview.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}

	var record storedChallengeImportPreview
	if err := json.Unmarshal(content, &record); err != nil {
		return nil, fmt.Errorf("parse challenge import preview: %w", err)
	}
	return &record, nil
}

func loadChallengeImportPreviewRecords() ([]*storedChallengeImportPreview, error) {
	root := challengeImportPreviewRoot()
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	records := make([]*storedChallengeImportPreview, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		record, err := loadChallengeImportPreviewRecord(entry.Name())
		if err != nil {
			if errors.Is(err, errcode.ErrNotFound) {
				continue
			}
			return nil, err
		}
		records = append(records, record)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})
	return records, nil
}

func persistImportedAttachmentBundle(parsed *domain.ParsedChallengePackage) (string, error) {
	if parsed == nil || len(parsed.Attachments) == 0 {
		return "", nil
	}

	targetDir := filepath.Join(challengeImportedAttachmentRoot(), "imports", parsed.Slug)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", fmt.Errorf("create attachment dir: %w", err)
	}

	if len(parsed.Attachments) == 1 {
		attachment := parsed.Attachments[0]
		fileName := sanitizeImportedAttachmentName(attachment.Name, attachment.Path)
		targetPath := filepath.Join(targetDir, fileName)
		if err := copyImportedAttachmentFile(attachment.AbsolutePath, targetPath); err != nil {
			return "", err
		}
		return buildAttachmentURLFromRelativePath(filepath.ToSlash(filepath.Join("imports", parsed.Slug, fileName))), nil
	}

	bundleName := sanitizeImportedAttachmentName(parsed.Slug+"-attachments.zip", parsed.Slug+"-attachments.zip")
	targetPath := filepath.Join(targetDir, bundleName)
	if err := writeImportedAttachmentBundle(targetPath, parsed.Attachments); err != nil {
		return "", err
	}
	return buildAttachmentURLFromRelativePath(filepath.ToSlash(filepath.Join("imports", parsed.Slug, bundleName))), nil
}

func copyImportedAttachmentFile(sourcePath, targetPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open attachment: %w", err)
	}
	defer source.Close()

	target, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("create attachment target: %w", err)
	}
	defer target.Close()

	if _, err := io.Copy(target, source); err != nil {
		return fmt.Errorf("copy attachment: %w", err)
	}
	return nil
}

func writeImportedAttachmentBundle(
	targetPath string,
	attachments []domain.ParsedChallengePackageAttachment,
) error {
	target, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("create attachment bundle: %w", err)
	}
	defer target.Close()

	archiveWriter := zip.NewWriter(target)

	for _, attachment := range attachments {
		header, err := zip.FileInfoHeader(&fileInfoAdapter{name: sanitizeImportedAttachmentName(attachment.Name, attachment.Path)})
		if err != nil {
			return err
		}
		header.Name = sanitizeImportedAttachmentName(attachment.Name, attachment.Path)
		header.Method = zip.Deflate

		writer, err := archiveWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		source, err := os.Open(attachment.AbsolutePath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(writer, source); err != nil {
			source.Close()
			return err
		}
		source.Close()
	}

	return archiveWriter.Close()
}

func buildAttachmentURLFromRelativePath(relativePath string) string {
	cleanRel := path.Clean("/" + strings.ReplaceAll(relativePath, "\\", "/"))
	cleanRel = strings.TrimPrefix(cleanRel, "/")

	segments := []string{"/api/v1/challenges/attachments"}
	for _, part := range strings.Split(cleanRel, "/") {
		part = strings.TrimSpace(part)
		if part == "" || part == "." || part == ".." {
			continue
		}
		segments = append(segments, part)
	}
	return strings.Join(segments, "/")
}

func sanitizeImportedAttachmentName(name, fallback string) string {
	candidate := strings.TrimSpace(name)
	if candidate == "" {
		candidate = fallback
	}
	candidate = filepath.Base(strings.ReplaceAll(candidate, "\\", "/"))
	if candidate == "." || candidate == string(filepath.Separator) || candidate == "" {
		return "attachment.bin"
	}
	return candidate
}

func configureImportedFlag(
	tx *gorm.DB,
	challengeID int64,
	flagType string,
	prefix string,
	value string,
) error {
	switch flagType {
	case model.FlagTypeStatic:
		return configureImportedStaticFlag(tx, challengeID, prefix, value)
	case model.FlagTypeDynamic:
		return configureImportedDynamicFlag(tx, challengeID, prefix)
	case model.FlagTypeRegex:
		return configureImportedRegexFlag(tx, challengeID, prefix, value)
	case model.FlagTypeManualReview:
		return configureImportedManualReviewFlag(tx, challengeID, prefix)
	default:
		return errcode.ErrInvalidParams.WithCause(errors.New("不支持的 flag 类型"))
	}
}

func configureImportedStaticFlag(tx *gorm.DB, challengeID int64, prefix, value string) error {
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return fmt.Errorf("generate salt for challenge %d: %w", challengeID, err)
	}
	return tx.Model(&model.Challenge{}).
		Where("id = ?", challengeID).
		Updates(map[string]any{
			"flag_type":   model.FlagTypeStatic,
			"flag_salt":   salt,
			"flag_hash":   crypto.HashStaticFlag(value, salt),
			"flag_regex":  "",
			"flag_prefix": prefix,
			"updated_at":  time.Now(),
		}).Error
}

func configureImportedDynamicFlag(tx *gorm.DB, challengeID int64, prefix string) error {
	return tx.Model(&model.Challenge{}).
		Where("id = ?", challengeID).
		Updates(map[string]any{
			"flag_type":   model.FlagTypeDynamic,
			"flag_salt":   "",
			"flag_hash":   "",
			"flag_regex":  "",
			"flag_prefix": prefix,
			"updated_at":  time.Now(),
		}).Error
}

func configureImportedRegexFlag(tx *gorm.DB, challengeID int64, prefix, value string) error {
	compiled, err := regexp.Compile(strings.TrimSpace(value))
	if err != nil {
		return errcode.ErrInvalidParams.WithCause(fmt.Errorf("regex flag 无效: %w", err))
	}
	return tx.Model(&model.Challenge{}).
		Where("id = ?", challengeID).
		Updates(map[string]any{
			"flag_type":   model.FlagTypeRegex,
			"flag_salt":   "",
			"flag_hash":   "",
			"flag_regex":  compiled.String(),
			"flag_prefix": prefix,
			"updated_at":  time.Now(),
		}).Error
}

func configureImportedManualReviewFlag(tx *gorm.DB, challengeID int64, prefix string) error {
	return tx.Model(&model.Challenge{}).
		Where("id = ?", challengeID).
		Updates(map[string]any{
			"flag_type":   model.FlagTypeManualReview,
			"flag_salt":   "",
			"flag_hash":   "",
			"flag_regex":  "",
			"flag_prefix": prefix,
			"updated_at":  time.Now(),
		}).Error
}

func syncImportedChallengeHints(
	tx *gorm.DB,
	challengeID int64,
	hints []domain.ParsedChallengePackageHint,
) error {
	if err := tx.Where("challenge_id = ?", challengeID).Delete(&model.ChallengeHint{}).Error; err != nil {
		return fmt.Errorf("delete hints for challenge %d: %w", challengeID, err)
	}
	if len(hints) == 0 {
		return nil
	}

	now := time.Now()
	records := make([]model.ChallengeHint, 0, len(hints))
	for _, hint := range hints {
		records = append(records, model.ChallengeHint{
			ChallengeID: challengeID,
			Level:       hint.Level,
			Title:       hint.Title,
			Content:     hint.Content,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}
	return tx.Create(&records).Error
}

func resolveImportedImageID(tx *gorm.DB, slug, imageRef string) (int64, error) {
	ref := strings.TrimSpace(imageRef)
	if ref == "" {
		return 0, nil
	}
	name, tag, err := splitImportedImageRef(ref)
	if err != nil {
		return 0, fmt.Errorf("invalid image ref for %s: %w", slug, err)
	}

	var image model.Image
	findErr := tx.Unscoped().
		Where("name = ? AND tag = ?", name, tag).
		First(&image).Error
	switch {
	case errors.Is(findErr, gorm.ErrRecordNotFound):
		image = model.Image{
			Name:        name,
			Tag:         tag,
			Description: fmt.Sprintf("Imported from challenge pack %s", slug),
			Status:      model.ImageStatusAvailable,
			Size:        0,
		}
		if err := tx.Create(&image).Error; err != nil {
			return 0, fmt.Errorf("create image %s:%s for %s: %w", name, tag, slug, err)
		}
		return image.ID, nil
	case findErr != nil:
		return 0, fmt.Errorf("find image %s:%s for %s: %w", name, tag, slug, findErr)
	default:
		if err := tx.Model(&image).Updates(map[string]any{
			"status":     model.ImageStatusAvailable,
			"deleted_at": nil,
			"updated_at": time.Now(),
		}).Error; err != nil {
			return 0, fmt.Errorf("update image %s:%s for %s: %w", name, tag, slug, err)
		}
		return image.ID, nil
	}
}

func splitImportedImageRef(imageRef string) (string, string, error) {
	trimmed := strings.TrimSpace(imageRef)
	if trimmed == "" {
		return "", "", fmt.Errorf("empty image ref")
	}

	lastSlash := strings.LastIndex(trimmed, "/")
	lastColon := strings.LastIndex(trimmed, ":")
	if lastColon > lastSlash {
		name := strings.TrimSpace(trimmed[:lastColon])
		tag := strings.TrimSpace(trimmed[lastColon+1:])
		if name == "" || tag == "" {
			return "", "", fmt.Errorf("invalid image ref %q", imageRef)
		}
		return name, tag, nil
	}
	return trimmed, "latest", nil
}

func challengeImportPreviewRoot() string {
	if value := strings.TrimSpace(os.Getenv("CHALLENGE_IMPORT_PREVIEW_DIR")); value != "" {
		return value
	}
	return defaultChallengeImportPreviewRoot
}

func challengeImportedAttachmentRoot() string {
	if value := strings.TrimSpace(os.Getenv("CHALLENGE_ATTACHMENT_STORAGE_DIR")); value != "" {
		return value
	}
	return defaultChallengeImportedAttachmentRoot
}

func challengePackageSourceRoot() string {
	if value := strings.TrimSpace(os.Getenv("CHALLENGE_PACKAGE_SOURCE_DIR")); value != "" {
		return value
	}
	return defaultChallengePackageSourceRoot
}

func challengePackageExportRoot() string {
	if value := strings.TrimSpace(os.Getenv("CHALLENGE_PACKAGE_EXPORT_DIR")); value != "" {
		return value
	}
	return defaultChallengePackageExportRoot
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

type fileInfoAdapter struct {
	name string
}

func (f *fileInfoAdapter) Name() string       { return f.name }
func (f *fileInfoAdapter) Size() int64        { return 0 }
func (f *fileInfoAdapter) Mode() os.FileMode  { return 0o644 }
func (f *fileInfoAdapter) ModTime() time.Time { return time.Time{} }
func (f *fileInfoAdapter) IsDir() bool        { return false }
func (f *fileInfoAdapter) Sys() any           { return nil }
