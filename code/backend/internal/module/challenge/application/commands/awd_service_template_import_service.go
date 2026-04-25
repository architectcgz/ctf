package commands

import (
	"context"
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

const defaultAWDServiceTemplateImportPreviewRoot = "./data/awd-service-template-import-previews"

type storedAWDServiceTemplateImportPreview struct {
	ID        string                                  `json:"id"`
	FileName  string                                  `json:"file_name"`
	SourceDir string                                  `json:"source_dir"`
	CreatedBy int64                                   `json:"created_by"`
	CreatedAt time.Time                               `json:"created_at"`
	Preview   dto.AWDServiceTemplateImportPreviewResp `json:"preview"`
}

type AWDServiceTemplateImportService struct {
	db   *gorm.DB
	repo challengeports.AWDServiceTemplateCommandRepository
}

func NewAWDServiceTemplateImportService(
	db *gorm.DB,
	repo challengeports.AWDServiceTemplateCommandRepository,
) *AWDServiceTemplateImportService {
	return &AWDServiceTemplateImportService{db: db, repo: repo}
}

func (s *AWDServiceTemplateImportService) PreviewImport(
	ctx context.Context,
	actorUserID int64,
	fileName string,
	reader io.Reader,
) (*dto.AWDServiceTemplateImportPreviewResp, error) {
	if strings.TrimSpace(fileName) == "" {
		fileName = "awd-service-template-package.zip"
	}

	previewID, err := generateChallengeImportPreviewID()
	if err != nil {
		return nil, err
	}

	previewDir := filepath.Join(awdServiceTemplateImportPreviewRoot(), previewID)
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

	parsed, err := domain.ParseAWDServiceTemplatePackageDir(rootDir)
	if err != nil {
		return nil, err
	}

	preview := buildAWDServiceTemplateImportPreview(previewID, fileName, parsed, time.Now())
	record := storedAWDServiceTemplateImportPreview{
		ID:        previewID,
		FileName:  fileName,
		SourceDir: rootDir,
		CreatedBy: actorUserID,
		CreatedAt: preview.CreatedAt,
		Preview:   *preview,
	}
	if err := saveAWDServiceTemplateImportPreviewRecord(previewDir, &record); err != nil {
		return nil, err
	}
	return preview, nil
}

func (s *AWDServiceTemplateImportService) ListImports(ctx context.Context, actorUserID int64) ([]dto.AWDServiceTemplateImportPreviewResp, error) {
	_ = ctx
	records, err := loadAWDServiceTemplateImportPreviewRecords()
	if err != nil {
		return nil, err
	}

	previews := make([]dto.AWDServiceTemplateImportPreviewResp, 0, len(records))
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

func (s *AWDServiceTemplateImportService) GetImport(
	ctx context.Context,
	actorUserID int64,
	id string,
) (*dto.AWDServiceTemplateImportPreviewResp, error) {
	_ = ctx
	record, err := loadAWDServiceTemplateImportPreviewRecord(id)
	if err != nil {
		return nil, err
	}
	if record.CreatedBy != 0 && record.CreatedBy != actorUserID {
		return nil, errcode.ErrForbidden
	}
	preview := record.Preview
	return &preview, nil
}

func (s *AWDServiceTemplateImportService) CommitImport(
	ctx context.Context,
	actorUserID int64,
	id string,
) (*dto.AWDServiceTemplateResp, error) {
	record, err := loadAWDServiceTemplateImportPreviewRecord(id)
	if err != nil {
		return nil, err
	}
	if record.CreatedBy != 0 && record.CreatedBy != actorUserID {
		return nil, errcode.ErrForbidden
	}

	parsed, err := domain.ParseAWDServiceTemplatePackageDir(record.SourceDir)
	if err != nil {
		return nil, err
	}

	var template *model.AWDServiceTemplate
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		resolvedImageID, err := resolveImportedImageID(tx, parsed.Slug, parsed.RuntimeImageRef)
		if err != nil {
			return err
		}

		runtimeConfig := cloneAWDTemplateConfig(parsed.RuntimeConfig)
		if strings.TrimSpace(parsed.RuntimeImageRef) != "" {
			runtimeConfig["image_ref"] = parsed.RuntimeImageRef
		}
		if resolvedImageID > 0 {
			runtimeConfig["image_id"] = resolvedImageID
		}

		now := time.Now()
		var current model.AWDServiceTemplate
		findErr := findAWDServiceTemplateForImportedPackageUpsert(tx, parsed.Slug, &current)
		checkerConfigRaw, err := marshalAWDTemplateConfig(parsed.CheckerConfig)
		if err != nil {
			return err
		}
		flagConfigRaw, err := marshalAWDTemplateConfig(parsed.FlagConfig)
		if err != nil {
			return err
		}
		accessConfigRaw, err := marshalAWDTemplateConfig(parsed.AccessConfig)
		if err != nil {
			return err
		}
		runtimeConfigRaw, err := marshalAWDTemplateConfig(runtimeConfig)
		if err != nil {
			return err
		}

		switch {
		case errors.Is(findErr, gorm.ErrRecordNotFound):
			current = model.AWDServiceTemplate{
				Name:             parsed.Title,
				Slug:             parsed.Slug,
				Category:         parsed.Category,
				Difficulty:       parsed.Difficulty,
				Description:      parsed.Description,
				ServiceType:      model.AWDServiceType(parsed.ServiceType),
				DeploymentMode:   model.AWDDeploymentMode(parsed.DeploymentMode),
				Version:          parsed.Version,
				Status:           model.AWDServiceTemplateStatusPublished,
				CheckerType:      model.AWDCheckerType(parsed.CheckerType),
				CheckerConfig:    checkerConfigRaw,
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
				return fmt.Errorf("create imported awd service template %s: %w", parsed.Slug, err)
			}
		case findErr != nil:
			return fmt.Errorf("find imported awd service template %s: %w", parsed.Slug, findErr)
		default:
			updates := map[string]any{
				"name":               parsed.Title,
				"slug":               parsed.Slug,
				"category":           parsed.Category,
				"difficulty":         parsed.Difficulty,
				"description":        parsed.Description,
				"service_type":       model.AWDServiceType(parsed.ServiceType),
				"deployment_mode":    model.AWDDeploymentMode(parsed.DeploymentMode),
				"version":            parsed.Version,
				"status":             model.AWDServiceTemplateStatusPublished,
				"checker_type":       model.AWDCheckerType(parsed.CheckerType),
				"checker_config":     checkerConfigRaw,
				"flag_mode":          parsed.FlagMode,
				"flag_config":        flagConfigRaw,
				"defense_entry_mode": parsed.DefenseEntryMode,
				"access_config":      accessConfigRaw,
				"runtime_config":     runtimeConfigRaw,
				"readiness_status":   model.AWDReadinessStatusPending,
				"readiness_report":   "",
				"last_verified_at":   nil,
				"last_verified_by":   nil,
				"deleted_at":         nil,
				"updated_at":         now,
			}
			if err := tx.Unscoped().Model(&current).Updates(updates).Error; err != nil {
				return fmt.Errorf("update imported awd service template %s: %w", parsed.Slug, err)
			}
			if err := tx.Where("id = ?", current.ID).First(&current).Error; err != nil {
				return fmt.Errorf("reload imported awd service template %s: %w", parsed.Slug, err)
			}
		}

		template = &current
		return nil
	}); err != nil {
		return nil, err
	}

	_ = os.RemoveAll(filepath.Join(awdServiceTemplateImportPreviewRoot(), id))
	return domain.AWDServiceTemplateRespFromModel(template), nil
}

func buildAWDServiceTemplateImportPreview(
	id string,
	fileName string,
	parsed *domain.ParsedAWDServiceTemplatePackage,
	createdAt time.Time,
) *dto.AWDServiceTemplateImportPreviewResp {
	if parsed == nil {
		return nil
	}
	return &dto.AWDServiceTemplateImportPreviewResp{
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
		CheckerConfig:    cloneAWDTemplateConfig(parsed.CheckerConfig),
		FlagMode:         parsed.FlagMode,
		FlagConfig:       cloneAWDTemplateConfig(parsed.FlagConfig),
		DefenseEntryMode: parsed.DefenseEntryMode,
		AccessConfig:     cloneAWDTemplateConfig(parsed.AccessConfig),
		RuntimeConfig:    cloneAWDTemplateConfig(parsed.RuntimeConfig),
		Warnings:         append([]string(nil), parsed.Warnings...),
		CreatedAt:        createdAt,
	}
}

func findAWDServiceTemplateForImportedPackageUpsert(
	tx *gorm.DB,
	slug string,
	template *model.AWDServiceTemplate,
) error {
	if template == nil {
		return fmt.Errorf("awd service template target is nil")
	}
	return tx.Unscoped().Where("slug = ?", strings.TrimSpace(slug)).First(template).Error
}

func marshalAWDTemplateConfig(value map[string]any) (string, error) {
	encoded, err := json.Marshal(cloneAWDTemplateConfig(value))
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func cloneAWDTemplateConfig(value map[string]any) map[string]any {
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

func saveAWDServiceTemplateImportPreviewRecord(
	previewDir string,
	record *storedAWDServiceTemplateImportPreview,
) error {
	content, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(previewDir, "preview.json"), content, 0o644)
}

func loadAWDServiceTemplateImportPreviewRecord(id string) (*storedAWDServiceTemplateImportPreview, error) {
	content, err := os.ReadFile(filepath.Join(awdServiceTemplateImportPreviewRoot(), id, "preview.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}

	var record storedAWDServiceTemplateImportPreview
	if err := json.Unmarshal(content, &record); err != nil {
		return nil, fmt.Errorf("parse awd service template import preview: %w", err)
	}
	return &record, nil
}

func loadAWDServiceTemplateImportPreviewRecords() ([]*storedAWDServiceTemplateImportPreview, error) {
	root := awdServiceTemplateImportPreviewRoot()
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	records := make([]*storedAWDServiceTemplateImportPreview, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		record, err := loadAWDServiceTemplateImportPreviewRecord(entry.Name())
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

func awdServiceTemplateImportPreviewRoot() string {
	if value := strings.TrimSpace(os.Getenv("AWD_SERVICE_TEMPLATE_IMPORT_PREVIEW_DIR")); value != "" {
		return value
	}
	return defaultAWDServiceTemplateImportPreviewRoot
}
