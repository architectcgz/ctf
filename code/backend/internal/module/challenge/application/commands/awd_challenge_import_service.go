package commands

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

const defaultAWDChallengeImportPreviewRoot = "./data/awd-challenge-import-previews"
const defaultAWDCheckerArtifactRoot = "./data/awd-checker-artifacts"

type storedAWDChallengeImportPreview struct {
	ID        string                            `json:"id"`
	FileName  string                            `json:"file_name"`
	SourceDir string                            `json:"source_dir"`
	CreatedBy int64                             `json:"created_by"`
	CreatedAt time.Time                         `json:"created_at"`
	Preview   dto.AWDChallengeImportPreviewResp `json:"preview"`
}

type AWDChallengeImportService struct {
	db         *gorm.DB
	repo       challengeports.AWDChallengeCommandRepository
	imageBuild *ImageBuildService
}

func NewAWDChallengeImportService(
	db *gorm.DB,
	repo challengeports.AWDChallengeCommandRepository,
	imageBuild ...*ImageBuildService,
) *AWDChallengeImportService {
	service := &AWDChallengeImportService{db: db, repo: repo}
	if len(imageBuild) > 0 {
		service.imageBuild = imageBuild[0]
	}
	return service
}

func (s *AWDChallengeImportService) PreviewImport(
	ctx context.Context,
	actorUserID int64,
	fileName string,
	reader io.Reader,
) (*dto.AWDChallengeImportPreviewResp, error) {
	if strings.TrimSpace(fileName) == "" {
		fileName = "awd-challenge-package.zip"
	}

	previewID, err := generateChallengeImportPreviewID()
	if err != nil {
		return nil, err
	}

	previewDir := filepath.Join(awdChallengeImportPreviewRoot(), previewID)
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

	parsed, err := domain.ParseAWDChallengePackageDir(rootDir)
	if err != nil {
		return nil, err
	}

	preview := s.buildAWDChallengeImportPreview(previewID, fileName, parsed, time.Now())
	record := storedAWDChallengeImportPreview{
		ID:        previewID,
		FileName:  fileName,
		SourceDir: rootDir,
		CreatedBy: actorUserID,
		CreatedAt: preview.CreatedAt,
		Preview:   *preview,
	}
	if err := saveAWDChallengeImportPreviewRecord(previewDir, &record); err != nil {
		return nil, err
	}
	return preview, nil
}

func (s *AWDChallengeImportService) ListImports(ctx context.Context, actorUserID int64) ([]dto.AWDChallengeImportPreviewResp, error) {
	_ = ctx
	records, err := loadAWDChallengeImportPreviewRecords()
	if err != nil {
		return nil, err
	}

	previews := make([]dto.AWDChallengeImportPreviewResp, 0, len(records))
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
) (*dto.AWDChallengeImportPreviewResp, error) {
	_ = ctx
	record, err := loadAWDChallengeImportPreviewRecord(id)
	if err != nil {
		return nil, err
	}
	if record.CreatedBy != 0 && record.CreatedBy != actorUserID {
		return nil, errcode.ErrForbidden
	}
	preview := record.Preview
	return &preview, nil
}

func (s *AWDChallengeImportService) CommitImport(
	ctx context.Context,
	actorUserID int64,
	id string,
) (*dto.AWDChallengeResp, error) {
	record, err := loadAWDChallengeImportPreviewRecord(id)
	if err != nil {
		return nil, err
	}
	if record.CreatedBy != 0 && record.CreatedBy != actorUserID {
		return nil, errcode.ErrForbidden
	}

	parsed, err := domain.ParseAWDChallengePackageDir(record.SourceDir)
	if err != nil {
		return nil, err
	}

	var challenge *model.AWDChallenge
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := rejectImportedAWDChallengeSlugConflict(tx, parsed.Slug); err != nil {
			return err
		}

		resolvedImageID, resolvedImageRef, err := s.resolveAWDImportedImageForCommit(ctx, tx, actorUserID, parsed)
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

		now := time.Now()
		var current model.AWDChallenge
		checkerConfigWithArtifact, err := persistAWDCheckerArtifact(parsed)
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

		current = model.AWDChallenge{
			Name:             parsed.Title,
			Slug:             parsed.Slug,
			Category:         parsed.Category,
			Difficulty:       parsed.Difficulty,
			Description:      parsed.Description,
			ServiceType:      model.AWDServiceType(parsed.ServiceType),
			DeploymentMode:   model.AWDDeploymentMode(parsed.DeploymentMode),
			Version:          parsed.Version,
			Status:           model.AWDChallengeStatusPublished,
			CheckerType:      model.AWDCheckerType(parsed.CheckerType),
			CheckerConfig:    checkerConfigWithArtifact,
			FlagMode:         parsed.FlagMode,
			FlagConfig:       flagConfigRaw,
			DefenseEntryMode: parsed.DefenseEntryMode,
			AccessConfig:     accessConfigRaw,
			RuntimeConfig:    runtimeConfigRaw,
			ReadinessStatus:  model.AWDReadinessStatusPending,
			ReadinessReport:  "",
			LastVerifiedAt:   nil,
			LastVerifiedBy:   nil,
			CreatedBy:        &actorUserID,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := tx.Create(&current).Error; err != nil {
			return fmt.Errorf("create imported awd challenge %s: %w", parsed.Slug, err)
		}

		challenge = &current
		return nil
	}); err != nil {
		return nil, err
	}

	_ = os.RemoveAll(filepath.Join(awdChallengeImportPreviewRoot(), id))
	return domain.AWDChallengeRespFromModel(challenge), nil
}

func (s *AWDChallengeImportService) buildAWDChallengeImportPreview(
	id string,
	fileName string,
	parsed *domain.ParsedAWDChallengePackage,
	createdAt time.Time,
) *dto.AWDChallengeImportPreviewResp {
	if parsed == nil {
		return nil
	}
	imageDelivery := dto.ChallengeImportImageDeliveryResp{
		SourceType:   parsed.ImageSourceType,
		SuggestedTag: parsed.SuggestedImageTag,
	}
	if parsed.ImageSourceType == domain.ImageSourceTypePlatformBuild && s != nil && s.imageBuild != nil {
		if targetRef, err := s.imageBuild.BuildPlatformTargetRef(domain.ChallengePackageModeAWD, parsed.Slug, parsed.SuggestedImageTag); err == nil {
			imageDelivery.TargetImageRef = targetRef
			imageDelivery.BuildStatus = model.ImageStatusPending
		}
	}
	return &dto.AWDChallengeImportPreviewResp{
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
		Warnings:         append([]string(nil), parsed.Warnings...),
		CreatedAt:        createdAt,
	}
}

func (s *AWDChallengeImportService) resolveAWDImportedImageForCommit(
	ctx context.Context,
	tx *gorm.DB,
	actorUserID int64,
	parsed *domain.ParsedAWDChallengePackage,
) (int64, string, error) {
	if parsed.ImageSourceType == domain.ImageSourceTypeExternalRef {
		if s == nil || s.imageBuild == nil {
			return 0, "", fmt.Errorf("image build service is not configured")
		}
		result, err := s.imageBuild.VerifyExternalImageRefInTx(ctx, tx, parsed.Slug, parsed.RuntimeImageRef)
		if err != nil {
			return 0, "", err
		}
		return result.ImageID, parsed.RuntimeImageRef, nil
	}
	if parsed.ImageSourceType != domain.ImageSourceTypePlatformBuild {
		imageID, err := resolveImportedImageID(tx, parsed.Slug, parsed.RuntimeImageRef)
		return imageID, parsed.RuntimeImageRef, err
	}
	if s == nil || s.imageBuild == nil {
		return 0, "", fmt.Errorf("image build service is not configured")
	}
	result, err := s.imageBuild.CreatePlatformBuildJobInTx(ctx, tx, CreatePlatformBuildJobRequest{
		ChallengeMode:  domain.ChallengePackageModeAWD,
		PackageSlug:    parsed.Slug,
		SuggestedTag:   parsed.SuggestedImageTag,
		SourceDir:      parsed.RootDir,
		DockerfilePath: parsed.DockerfilePath,
		ContextPath:    parsed.BuildContextPath,
		CreatedBy:      actorUserID,
	})
	if err != nil {
		return 0, "", err
	}
	return result.ImageID, result.TargetRef, nil
}

func rejectImportedAWDChallengeSlugConflict(tx *gorm.DB, slug string) error {
	normalizedSlug := strings.TrimSpace(slug)
	if normalizedSlug == "" {
		return nil
	}

	var existing model.AWDChallenge
	err := tx.Unscoped().
		Select("id", "slug", "name").
		Where("slug = ?", normalizedSlug).
		First(&existing).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil
	case err != nil:
		return fmt.Errorf("check imported awd challenge slug %s: %w", normalizedSlug, err)
	default:
		message := fmt.Sprintf("AWD 题目 slug %s 已被已有题目占用，请改用题目编辑入口更新", normalizedSlug)
		return errcode.New(errcode.ErrConflict.Code, message, errcode.ErrConflict.HTTPStatus)
	}
}

func marshalAWDChallengeConfig(value map[string]any) (string, error) {
	encoded, err := json.Marshal(cloneAWDChallengeConfig(value))
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func persistAWDCheckerArtifact(parsed *domain.ParsedAWDChallengePackage) (string, error) {
	if parsed == nil {
		return "{}", nil
	}
	config := cloneAWDChallengeConfig(parsed.CheckerConfig)
	if parsed.CheckerType != string(model.AWDCheckerTypeScript) {
		return marshalAWDChallengeConfig(config)
	}
	if strings.TrimSpace(parsed.CheckerEntryAbs) == "" || strings.TrimSpace(parsed.CheckerEntryPath) == "" {
		return "", errcode.ErrInvalidParams.WithCause(errors.New("script_checker artifact entry is missing"))
	}
	files := parsed.CheckerFiles
	if len(files) == 0 {
		files = []domain.ParsedAWDCheckerFile{{Path: parsed.CheckerEntryPath, Abs: parsed.CheckerEntryAbs}}
	}
	fileContents := make([][]byte, 0, len(files))
	fileMetadata := make([]map[string]any, 0, len(files))
	digestSeed := sha256.New()
	for _, file := range files {
		content, err := os.ReadFile(file.Abs)
		if err != nil {
			return "", fmt.Errorf("read script checker artifact %s: %w", file.Path, err)
		}
		sum := sha256.Sum256(content)
		fileDigest := hex.EncodeToString(sum[:])
		digestSeed.Write([]byte(file.Path))
		digestSeed.Write([]byte{0})
		digestSeed.Write([]byte(fileDigest))
		digestSeed.Write([]byte{0})
		digestSeed.Write([]byte(fmt.Sprintf("%d", len(content))))
		digestSeed.Write([]byte{0})
		fileContents = append(fileContents, content)
		fileMetadata = append(fileMetadata, map[string]any{
			"path":   file.Path,
			"sha256": fileDigest,
			"size":   len(content),
		})
	}
	digest := hex.EncodeToString(digestSeed.Sum(nil))
	targetDir := filepath.Join(awdCheckerArtifactRoot(), sanitizeAWDCheckerArtifactSegment(parsed.Slug), digest)
	for index, file := range files {
		targetPath := filepath.Join(targetDir, filepath.FromSlash(file.Path))
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o750); err != nil {
			return "", fmt.Errorf("create script checker artifact dir: %w", err)
		}
		if err := os.WriteFile(targetPath, fileContents[index], 0o400); err != nil {
			return "", fmt.Errorf("write script checker artifact: %w", err)
		}
		fileMetadata[index]["storage_path"] = targetPath
	}
	entryArtifact := fileMetadata[0]
	for _, item := range fileMetadata {
		if item["path"] == parsed.CheckerEntryPath {
			entryArtifact = item
			break
		}
	}
	config["artifact"] = map[string]any{
		"entry":        parsed.CheckerEntryPath,
		"storage_path": entryArtifact["storage_path"],
		"sha256":       entryArtifact["sha256"],
		"size":         entryArtifact["size"],
		"digest":       digest,
		"files":        fileMetadata,
	}
	return marshalAWDChallengeConfig(config)
}

func awdCheckerArtifactDirFromConfig(raw string) string {
	var config map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &config); err != nil {
		return ""
	}
	artifact, ok := config["artifact"].(map[string]any)
	if !ok {
		return ""
	}
	storagePath, _ := artifact["storage_path"].(string)
	packagePath, _ := artifact["entry"].(string)
	if strings.TrimSpace(storagePath) == "" {
		if files, ok := artifact["files"].([]any); ok {
			for _, item := range files {
				file, ok := item.(map[string]any)
				if !ok {
					continue
				}
				storagePath, _ = file["storage_path"].(string)
				packagePath, _ = file["path"].(string)
				if strings.TrimSpace(storagePath) != "" && strings.TrimSpace(packagePath) != "" {
					break
				}
			}
		}
	}
	return awdCheckerArtifactDirFromFile(storagePath, packagePath)
}

func awdCheckerArtifactDirFromFile(storagePath, packagePath string) string {
	storagePath = filepath.Clean(strings.TrimSpace(storagePath))
	packagePath = filepath.Clean(strings.TrimSpace(packagePath))
	if storagePath == "." || packagePath == "." || packagePath == "" || filepath.IsAbs(packagePath) {
		return ""
	}
	suffix := filepath.FromSlash(packagePath)
	if !strings.HasSuffix(storagePath, suffix) {
		return ""
	}
	dir := strings.TrimSuffix(storagePath, suffix)
	dir = strings.TrimRight(dir, string(filepath.Separator))
	if dir == "" || !isAWDCheckerArtifactDirInsideRoot(dir) {
		return ""
	}
	return dir
}

func removeAWDCheckerArtifactDir(dir string) error {
	dir = filepath.Clean(strings.TrimSpace(dir))
	if dir == "." || dir == "" || !isAWDCheckerArtifactDirInsideRoot(dir) {
		return nil
	}
	return os.RemoveAll(dir)
}

func isAWDCheckerArtifactDirInsideRoot(dir string) bool {
	root, err := filepath.Abs(awdCheckerArtifactRoot())
	if err != nil {
		return false
	}
	target, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func sanitizeAWDCheckerArtifactSegment(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "unknown"
	}
	var builder strings.Builder
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			builder.WriteRune(r)
			continue
		}
		builder.WriteByte('-')
	}
	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "unknown"
	}
	return result
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

func saveAWDChallengeImportPreviewRecord(
	previewDir string,
	record *storedAWDChallengeImportPreview,
) error {
	content, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(previewDir, "preview.json"), content, 0o644)
}

func loadAWDChallengeImportPreviewRecord(id string) (*storedAWDChallengeImportPreview, error) {
	content, err := os.ReadFile(filepath.Join(awdChallengeImportPreviewRoot(), id, "preview.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}

	var record storedAWDChallengeImportPreview
	if err := json.Unmarshal(content, &record); err != nil {
		return nil, fmt.Errorf("parse awd challenge import preview: %w", err)
	}
	return &record, nil
}

func loadAWDChallengeImportPreviewRecords() ([]*storedAWDChallengeImportPreview, error) {
	root := awdChallengeImportPreviewRoot()
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	records := make([]*storedAWDChallengeImportPreview, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		record, err := loadAWDChallengeImportPreviewRecord(entry.Name())
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

func awdChallengeImportPreviewRoot() string {
	if value := strings.TrimSpace(os.Getenv("AWD_CHALLENGE_IMPORT_PREVIEW_DIR")); value != "" {
		return value
	}
	return defaultAWDChallengeImportPreviewRoot
}

func awdCheckerArtifactRoot() string {
	if value := strings.TrimSpace(os.Getenv("AWD_CHECKER_ARTIFACT_DIR")); value != "" {
		return value
	}
	return defaultAWDCheckerArtifactRoot
}
