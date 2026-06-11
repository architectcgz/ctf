package infrastructure

import (
	"archive/zip"
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

	"ctf-platform/internal/apperror"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

const (
	defaultChallengeImportPreviewRoot      = "./data/challenge-import-previews"
	defaultAWDChallengeImportPreviewRoot   = "./data/awd-challenge-import-previews"
	defaultChallengeImportedAttachmentRoot = "./data/challenge-attachments"
	defaultChallengePackageSourceRoot      = "./data/challenge-packages"
	defaultChallengePackageExportRoot      = "./data/challenge-package-exports"
	defaultImportedImageBuildSourceRoot    = "./data/challenge-image-build-sources"
	defaultAWDCheckerArtifactRoot          = "./data/awd-checker-artifacts"
	maxChallengeImportArchiveFiles         = 128
	maxChallengeImportArchiveFileSize      = 16 << 20
	maxChallengeImportArchiveTotalSize     = 64 << 20
)

type ChallengeLocalStorageRoots struct {
	ChallengeImportPreviewRoot    string
	AWDChallengeImportPreviewRoot string
	ChallengeAttachmentRoot       string
	ChallengePackageSourceRoot    string
	ChallengePackageExportRoot    string
	ImageBuildSourceRoot          string
	AWDCheckerArtifactRoot        string
}

func DefaultChallengeLocalStorageRootsFromEnv() ChallengeLocalStorageRoots {
	return ChallengeLocalStorageRoots{
		ChallengeImportPreviewRoot:    challengeImportPreviewRoot(),
		AWDChallengeImportPreviewRoot: awdChallengeImportPreviewRoot(),
		ChallengeAttachmentRoot:       challengeImportedAttachmentRoot(),
		ChallengePackageSourceRoot:    challengePackageSourceRoot(),
		ChallengePackageExportRoot:    challengePackageExportRoot(),
		ImageBuildSourceRoot:          importedImageBuildSourceRoot(),
		AWDCheckerArtifactRoot:        awdCheckerArtifactRoot(),
	}
}

type ChallengeImportPreviewStore struct {
	root string
}

var _ challengeports.ChallengeImportPreviewStore = (*ChallengeImportPreviewStore)(nil)

func NewChallengeImportPreviewStore(root string) *ChallengeImportPreviewStore {
	return &ChallengeImportPreviewStore{root: strings.TrimSpace(root)}
}

func (s *ChallengeImportPreviewStore) rootDir() string {
	if s != nil && strings.TrimSpace(s.root) != "" {
		return strings.TrimSpace(s.root)
	}
	return challengeImportPreviewRoot()
}

func (s *ChallengeImportPreviewStore) CreateWorkspace(
	ctx context.Context,
	id string,
	fileName string,
	reader io.Reader,
) (*challengeports.ChallengeImportPreviewWorkspace, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if reader == nil {
		return nil, apperror.ErrInvalidParams.WithCause(errors.New("缺少题包文件"))
	}
	workspaceID, err := safePathSegment(id, "preview id")
	if err != nil {
		return nil, err
	}
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		fileName = "challenge-package.zip"
	}

	storageRoot := s.rootDir()
	previewDir := filepath.Join(storageRoot, workspaceID)
	archivePath := filepath.Join(previewDir, "package.zip")
	extractDir := filepath.Join(previewDir, "source")
	if err := os.RemoveAll(previewDir); err != nil {
		return nil, fmt.Errorf("clear preview dir: %w", err)
	}
	if err := os.MkdirAll(previewDir, 0o755); err != nil {
		return nil, fmt.Errorf("create preview dir: %w", err)
	}
	if err := writeReaderToFile(archivePath, reader); err != nil {
		_ = os.RemoveAll(previewDir)
		return nil, err
	}

	sourceRoot, err := extractChallengeImportArchive(archivePath, extractDir)
	if err != nil {
		_ = os.RemoveAll(previewDir)
		return nil, err
	}
	return &challengeports.ChallengeImportPreviewWorkspace{
		ID:          workspaceID,
		FileName:    fileName,
		ArchivePath: archivePath,
		SourceDir:   sourceRoot,
	}, nil
}

func (s *ChallengeImportPreviewStore) SaveRecord(ctx context.Context, record *challengeports.ChallengeImportPreviewRecord) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if record == nil {
		return apperror.ErrInvalidParams.WithCause(errors.New("缺少导入预览记录"))
	}
	workspaceID, err := safePathSegment(record.ID, "preview id")
	if err != nil {
		return err
	}
	previewDir := filepath.Join(s.rootDir(), workspaceID)
	if err := os.MkdirAll(previewDir, 0o755); err != nil {
		return fmt.Errorf("create preview dir: %w", err)
	}
	content, err := json.MarshalIndent(storedChallengeImportPreviewRecordFromPort(record), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(previewDir, "preview.json"), content, 0o644)
}

func (s *ChallengeImportPreviewStore) LoadRecord(ctx context.Context, id string) (*challengeports.ChallengeImportPreviewRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	workspaceID, err := safePathSegment(id, "preview id")
	if err != nil {
		return nil, err
	}
	rootDir := s.rootDir()
	content, err := os.ReadFile(filepath.Join(rootDir, workspaceID, "preview.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	var record storedChallengeImportPreviewRecord
	if err := json.Unmarshal(content, &record); err != nil {
		return nil, fmt.Errorf("parse challenge import preview: %w", err)
	}
	if record.ArchivePath == "" {
		record.ArchivePath = filepath.Join(rootDir, workspaceID, "package.zip")
	}
	return record.toPort(), nil
}

func (s *ChallengeImportPreviewStore) ListRecords(ctx context.Context) ([]*challengeports.ChallengeImportPreviewRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.rootDir())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	records := make([]*challengeports.ChallengeImportPreviewRecord, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		record, err := s.LoadRecord(ctx, entry.Name())
		if err != nil {
			if errors.Is(err, apperror.ErrNotFound) {
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

func (s *ChallengeImportPreviewStore) DeleteWorkspace(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	workspaceID, err := safePathSegment(id, "preview id")
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(s.rootDir(), workspaceID))
}

type storedChallengeImportPreviewRecord struct {
	ID          string                                        `json:"id"`
	FileName    string                                        `json:"file_name"`
	ArchivePath string                                        `json:"archive_path"`
	SourceDir   string                                        `json:"source_dir"`
	CreatedBy   int64                                         `json:"created_by"`
	CreatedAt   time.Time                                     `json:"created_at"`
	Preview     challengecontracts.ChallengeImportPreviewResp `json:"preview"`
}

func storedChallengeImportPreviewRecordFromPort(record *challengeports.ChallengeImportPreviewRecord) storedChallengeImportPreviewRecord {
	return storedChallengeImportPreviewRecord{
		ID:          record.ID,
		FileName:    record.FileName,
		ArchivePath: record.ArchivePath,
		SourceDir:   record.SourceDir,
		CreatedBy:   record.CreatedBy,
		CreatedAt:   record.CreatedAt,
		Preview:     record.Preview,
	}
}

func (r storedChallengeImportPreviewRecord) toPort() *challengeports.ChallengeImportPreviewRecord {
	return &challengeports.ChallengeImportPreviewRecord{
		ID:          r.ID,
		FileName:    r.FileName,
		ArchivePath: r.ArchivePath,
		SourceDir:   r.SourceDir,
		CreatedBy:   r.CreatedBy,
		CreatedAt:   r.CreatedAt,
		Preview:     r.Preview,
	}
}

type AWDChallengeImportPreviewStore struct {
	root string
}

var _ challengeports.AWDChallengeImportPreviewStore = (*AWDChallengeImportPreviewStore)(nil)

func NewAWDChallengeImportPreviewStore(root string) *AWDChallengeImportPreviewStore {
	return &AWDChallengeImportPreviewStore{root: strings.TrimSpace(root)}
}

func (s *AWDChallengeImportPreviewStore) rootDir() string {
	if s != nil && strings.TrimSpace(s.root) != "" {
		return strings.TrimSpace(s.root)
	}
	return awdChallengeImportPreviewRoot()
}

func (s *AWDChallengeImportPreviewStore) CreateWorkspace(
	ctx context.Context,
	id string,
	fileName string,
	reader io.Reader,
) (*challengeports.AWDChallengeImportPreviewWorkspace, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if reader == nil {
		return nil, apperror.ErrInvalidParams.WithCause(errors.New("缺少题包文件"))
	}
	workspaceID, err := safePathSegment(id, "preview id")
	if err != nil {
		return nil, err
	}
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		fileName = "awd-challenge-package.zip"
	}

	previewDir := filepath.Join(s.rootDir(), workspaceID)
	archivePath := filepath.Join(previewDir, "package.zip")
	extractDir := filepath.Join(previewDir, "source")
	if err := os.RemoveAll(previewDir); err != nil {
		return nil, fmt.Errorf("clear awd preview dir: %w", err)
	}
	if err := os.MkdirAll(previewDir, 0o755); err != nil {
		return nil, fmt.Errorf("create awd preview dir: %w", err)
	}
	if err := writeReaderToFile(archivePath, reader); err != nil {
		_ = os.RemoveAll(previewDir)
		return nil, err
	}

	rootDir, err := extractChallengeImportArchive(archivePath, extractDir)
	if err != nil {
		_ = os.RemoveAll(previewDir)
		return nil, err
	}
	return &challengeports.AWDChallengeImportPreviewWorkspace{
		ID:          workspaceID,
		FileName:    fileName,
		ArchivePath: archivePath,
		SourceDir:   rootDir,
	}, nil
}

func (s *AWDChallengeImportPreviewStore) SaveRecord(ctx context.Context, record *challengeports.AWDChallengeImportPreviewRecord) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if record == nil {
		return apperror.ErrInvalidParams.WithCause(errors.New("缺少 AWD 导入预览记录"))
	}
	workspaceID, err := safePathSegment(record.ID, "preview id")
	if err != nil {
		return err
	}
	previewDir := filepath.Join(s.rootDir(), workspaceID)
	if err := os.MkdirAll(previewDir, 0o755); err != nil {
		return fmt.Errorf("create awd preview dir: %w", err)
	}
	content, err := json.MarshalIndent(storedAWDChallengeImportPreviewRecordFromPort(record), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(previewDir, "preview.json"), content, 0o644)
}

func (s *AWDChallengeImportPreviewStore) LoadRecord(ctx context.Context, id string) (*challengeports.AWDChallengeImportPreviewRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	workspaceID, err := safePathSegment(id, "preview id")
	if err != nil {
		return nil, err
	}
	rootDir := s.rootDir()
	content, err := os.ReadFile(filepath.Join(rootDir, workspaceID, "preview.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	var record storedAWDChallengeImportPreviewRecord
	if err := json.Unmarshal(content, &record); err != nil {
		return nil, fmt.Errorf("parse awd challenge import preview: %w", err)
	}
	if record.ArchivePath == "" {
		record.ArchivePath = filepath.Join(rootDir, workspaceID, "package.zip")
	}
	return record.toPort(), nil
}

func (s *AWDChallengeImportPreviewStore) ListRecords(ctx context.Context) ([]*challengeports.AWDChallengeImportPreviewRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.rootDir())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	records := make([]*challengeports.AWDChallengeImportPreviewRecord, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		record, err := s.LoadRecord(ctx, entry.Name())
		if err != nil {
			if errors.Is(err, apperror.ErrNotFound) {
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

func (s *AWDChallengeImportPreviewStore) DeleteWorkspace(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	workspaceID, err := safePathSegment(id, "preview id")
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(s.rootDir(), workspaceID))
}

type storedAWDChallengeImportPreviewRecord struct {
	ID          string                                           `json:"id"`
	FileName    string                                           `json:"file_name"`
	ArchivePath string                                           `json:"archive_path"`
	SourceDir   string                                           `json:"source_dir"`
	CreatedBy   int64                                            `json:"created_by"`
	CreatedAt   time.Time                                        `json:"created_at"`
	Preview     challengecontracts.AWDChallengeImportPreviewResp `json:"preview"`
}

func storedAWDChallengeImportPreviewRecordFromPort(record *challengeports.AWDChallengeImportPreviewRecord) storedAWDChallengeImportPreviewRecord {
	return storedAWDChallengeImportPreviewRecord{
		ID:          record.ID,
		FileName:    record.FileName,
		ArchivePath: record.ArchivePath,
		SourceDir:   record.SourceDir,
		CreatedBy:   record.CreatedBy,
		CreatedAt:   record.CreatedAt,
		Preview:     record.Preview,
	}
}

func (r storedAWDChallengeImportPreviewRecord) toPort() *challengeports.AWDChallengeImportPreviewRecord {
	return &challengeports.AWDChallengeImportPreviewRecord{
		ID:          r.ID,
		FileName:    r.FileName,
		ArchivePath: r.ArchivePath,
		SourceDir:   r.SourceDir,
		CreatedBy:   r.CreatedBy,
		CreatedAt:   r.CreatedAt,
		Preview:     r.Preview,
	}
}

func writeReaderToFile(targetPath string, reader io.Reader) error {
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
		return "", apperror.ErrInvalidParams.WithCause(fmt.Errorf("读取 zip 失败: %w", err))
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
	return resolveExtractedChallengeImportRoot(extractDir)
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
		return apperror.ErrInvalidParams.WithCause(fmt.Errorf("zip 条目不允许符号链接: %s", file.Name))
	}
	if file.FileInfo().IsDir() {
		return nil
	}
	s.fileCount++
	if s.fileCount > maxChallengeImportArchiveFiles {
		return apperror.ErrInvalidParams.WithCause(
			fmt.Errorf("zip 文件数量超过限制，最多允许 %d 个文件", maxChallengeImportArchiveFiles),
		)
	}
	if file.UncompressedSize64 > maxChallengeImportArchiveFileSize {
		return apperror.ErrInvalidParams.WithCause(
			fmt.Errorf("zip 单文件超过限制，最多允许 %d 字节", maxChallengeImportArchiveFileSize),
		)
	}
	s.totalSize += file.UncompressedSize64
	if s.totalSize > maxChallengeImportArchiveTotalSize {
		return apperror.ErrInvalidParams.WithCause(
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
	targetAbs, err := safeJoinedPath(baseDir, relativePath)
	if err != nil {
		return apperror.ErrInvalidParams.WithCause(fmt.Errorf("zip 条目路径非法: %s", relativePath))
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

	_, err = io.Copy(target, source)
	return err
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
		return "", apperror.ErrInvalidParams.WithCause(errors.New("zip 根目录必须直接包含 challenge.yml 或单一题目目录"))
	}

	rootDir := filepath.Join(extractDir, entries[0].Name())
	if _, err := os.Stat(filepath.Join(rootDir, "challenge.yml")); err != nil {
		return "", apperror.ErrInvalidParams.WithCause(errors.New("未找到 challenge.yml"))
	}
	return rootDir, nil
}

func challengeImportPreviewRoot() string {
	if value := strings.TrimSpace(os.Getenv("CHALLENGE_IMPORT_PREVIEW_DIR")); value != "" {
		return value
	}
	return defaultChallengeImportPreviewRoot
}

func awdChallengeImportPreviewRoot() string {
	if value := strings.TrimSpace(os.Getenv("AWD_CHALLENGE_IMPORT_PREVIEW_DIR")); value != "" {
		return value
	}
	return defaultAWDChallengeImportPreviewRoot
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

func importedImageBuildSourceRoot() string {
	if dir := strings.TrimSpace(os.Getenv("CHALLENGE_IMAGE_BUILD_SOURCE_DIR")); dir != "" {
		return dir
	}
	return defaultImportedImageBuildSourceRoot
}

func awdCheckerArtifactRoot() string {
	if value := strings.TrimSpace(os.Getenv("AWD_CHECKER_ARTIFACT_DIR")); value != "" {
		return value
	}
	return defaultAWDCheckerArtifactRoot
}
