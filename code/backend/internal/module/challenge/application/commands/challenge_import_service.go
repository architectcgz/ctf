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
	repo                  challengeCommandRepository
	imageRepo             challengeports.ImageQueryRepository
	topologyRepo          challengeports.ChallengeTopologyReadRepository
	packageRepo           challengeports.ChallengePackageRevisionRepository
	runtimeProbe          challengeports.ChallengeRuntimeProbe
	importTxRunner        challengeports.ChallengeImportTxRunner
	packageExportTxRunner challengeports.ChallengePackageExportTxRunner
	imageBuild            *ImageBuildService
	eventBus              platformevents.Bus
	selfCheckCfg          SelfCheckConfig
	logger                *zap.Logger
}

func NewChallengeService(
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

	preview := s.buildChallengeImportPreview(previewID, fileName, parsed, time.Now().UTC())
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

func (s *ChallengeService) SetChallengeImportTxRunner(runner challengeports.ChallengeImportTxRunner) *ChallengeService {
	if s == nil {
		return nil
	}
	s.importTxRunner = runner
	return s
}

func (s *ChallengeService) SetChallengePackageExportTxRunner(runner challengeports.ChallengePackageExportTxRunner) *ChallengeService {
	if s == nil {
		return nil
	}
	s.packageExportTxRunner = runner
	return s
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
	if s.importTxRunner == nil {
		return nil, fmt.Errorf("challenge import tx runner is not configured")
	}
	if err := s.importTxRunner.WithinChallengeImportTransaction(ctx, func(store challengeports.ChallengeImportTxStore) error {
		if err := store.RejectImportedChallengeSlugConflict(ctx, parsed.Slug); err != nil {
			return err
		}

		resolvedImageID, err := s.resolveImportedImageIDForCommit(ctx, store, actorUserID, parsed, buildSource)
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		var current model.Challenge
		existing, found, findErr := store.FindLegacyChallengeForImportedPackageCreate(ctx, parsed.Title, parsed.Category)

		switch {
		case findErr != nil:
			return findErr
		case !found:
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
			if err := store.CreateImportedChallenge(ctx, &current); err != nil {
				return fmt.Errorf("create imported challenge %s: %w", parsed.Slug, err)
			}
		default:
			current = *existing
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
			if err := store.UpdateImportedChallenge(ctx, &current, updates); err != nil {
				return fmt.Errorf("update imported challenge %s: %w", parsed.Slug, err)
			}
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
			revision, revisionErr := s.createImportedPackageRevision(ctx, store, actorUserID, &current, *record, parsed)
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
					return s.resolveExternalImageRefForCommit(ctx, store, parsed.Slug, imageRef)
				},
			)
			if topologyErr != nil {
				return topologyErr
			}
			now = time.Now().UTC()
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
			if err := store.UpsertImportedTopology(ctx, item); err != nil {
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
	store challengeports.ChallengeImportTxStore,
	actorUserID int64,
	parsed *domain.ParsedChallengePackage,
	buildSource *importedImageBuildSource,
) (int64, error) {
	if parsed.ImageSourceType == domain.ImageSourceTypeExternalRef {
		return s.resolveExternalImageRefForCommit(ctx, store, parsed.Slug, parsed.RuntimeImageRef)
	}
	if parsed.ImageSourceType != domain.ImageSourceTypePlatformBuild {
		resolution, err := store.ResolveExistingImageRef(ctx, parsed.Slug, parsed.RuntimeImageRef)
		if err != nil {
			return 0, err
		}
		return resolution.ImageID, nil
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
		return 0, err
	}
	return result.ImageID, nil
}

func (s *ChallengeService) resolveExternalImageRefForCommit(
	ctx context.Context,
	store challengeports.ChallengeImportTxStore,
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
	result, err := store.ResolveExternalImage(ctx, packageSlug, imageRef)
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

func buildImportedFlagUpdates(
	flagType string,
	prefix string,
	value string,
	updatedAt time.Time,
) (map[string]any, error) {
	switch flagType {
	case model.FlagTypeStatic:
		return buildImportedStaticFlagUpdates(prefix, value, updatedAt)
	case model.FlagTypeDynamic:
		return buildImportedDynamicFlagUpdates(prefix, updatedAt), nil
	case model.FlagTypeRegex:
		return buildImportedRegexFlagUpdates(prefix, value, updatedAt)
	case model.FlagTypeManualReview:
		return buildImportedManualReviewFlagUpdates(prefix, updatedAt), nil
	default:
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("不支持的 flag 类型"))
	}
}

func buildImportedStaticFlagUpdates(prefix string, value string, updatedAt time.Time) (map[string]any, error) {
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("generate salt for imported challenge: %w", err)
	}
	return map[string]any{
		"flag_type":   model.FlagTypeStatic,
		"flag_salt":   salt,
		"flag_hash":   crypto.HashStaticFlag(value, salt),
		"flag_regex":  "",
		"flag_prefix": prefix,
		"updated_at":  updatedAt,
	}, nil
}

func buildImportedDynamicFlagUpdates(prefix string, updatedAt time.Time) map[string]any {
	return map[string]any{
		"flag_type":   model.FlagTypeDynamic,
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
		return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("regex flag 无效: %w", err))
	}
	return map[string]any{
		"flag_type":   model.FlagTypeRegex,
		"flag_salt":   "",
		"flag_hash":   "",
		"flag_regex":  compiled.String(),
		"flag_prefix": prefix,
		"updated_at":  updatedAt,
	}, nil
}

func buildImportedManualReviewFlagUpdates(prefix string, updatedAt time.Time) map[string]any {
	return map[string]any{
		"flag_type":   model.FlagTypeManualReview,
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
) []model.ChallengeHint {
	if len(hints) == 0 {
		return nil
	}

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
	return records
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
