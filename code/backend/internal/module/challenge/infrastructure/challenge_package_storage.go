package infrastructure

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ctf-platform/internal/apperror"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ChallengePackageStorageConfig struct {
	SourceRoot           string
	ExportRoot           string
	ImageBuildSourceRoot string
}

type ChallengePackageStorage struct {
	sourceRoot           string
	exportRoot           string
	imageBuildSourceRoot string
}

var _ challengeports.ChallengePackageStorage = (*ChallengePackageStorage)(nil)

func NewChallengePackageStorage(cfg ChallengePackageStorageConfig) *ChallengePackageStorage {
	return &ChallengePackageStorage{
		sourceRoot:           strings.TrimSpace(cfg.SourceRoot),
		exportRoot:           strings.TrimSpace(cfg.ExportRoot),
		imageBuildSourceRoot: strings.TrimSpace(cfg.ImageBuildSourceRoot),
	}
}

func (s *ChallengePackageStorage) sourceRootDir() string {
	if s != nil && strings.TrimSpace(s.sourceRoot) != "" {
		return strings.TrimSpace(s.sourceRoot)
	}
	return challengePackageSourceRoot()
}

func (s *ChallengePackageStorage) exportRootDir() string {
	if s != nil && strings.TrimSpace(s.exportRoot) != "" {
		return strings.TrimSpace(s.exportRoot)
	}
	return challengePackageExportRoot()
}

func (s *ChallengePackageStorage) imageBuildSourceRootDir() string {
	if s != nil && strings.TrimSpace(s.imageBuildSourceRoot) != "" {
		return strings.TrimSpace(s.imageBuildSourceRoot)
	}
	return importedImageBuildSourceRoot()
}

func (s *ChallengePackageStorage) PersistImportedPackageSource(
	ctx context.Context,
	req challengeports.ChallengeImportedPackageSourceRequest,
) (*challengeports.ChallengeStoredPackageSource, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if req.ChallengeID <= 0 || req.RevisionNo <= 0 {
		return nil, apperror.ErrInvalidParams.WithCause(errors.New("题包修订参数无效"))
	}
	packageSlug, err := safePathSegment(req.PackageSlug, "package slug")
	if err != nil {
		return nil, err
	}
	_ = packageSlug

	revisionRoot := filepath.Join(s.sourceRootDir(), fmt.Sprintf("challenge-%d", req.ChallengeID), fmt.Sprintf("r%04d", req.RevisionNo))
	sourceDir := filepath.Join(revisionRoot, "source")
	if err := os.RemoveAll(revisionRoot); err != nil {
		return nil, fmt.Errorf("clear package revision root: %w", err)
	}
	if err := copyDirectoryTree(req.SourceDir, sourceDir); err != nil {
		return nil, fmt.Errorf("copy imported package source: %w", err)
	}

	archivePath := ""
	if strings.TrimSpace(req.PreviewArchivePath) != "" {
		if info, statErr := os.Stat(req.PreviewArchivePath); statErr == nil && !info.IsDir() {
			archivePath = filepath.Join(revisionRoot, sanitizeImportedAttachmentName(req.PreviewArchiveName, "challenge-package.zip"))
			if err := copyFile(req.PreviewArchivePath, archivePath); err != nil {
				return nil, fmt.Errorf("copy imported package archive: %w", err)
			}
		}
	}

	return &challengeports.ChallengeStoredPackageSource{
		RevisionRoot: revisionRoot,
		SourceDir:    sourceDir,
		ArchivePath:  archivePath,
	}, nil
}

func (s *ChallengePackageStorage) PersistImportedImageBuildSource(
	ctx context.Context,
	req challengeports.ChallengeImportedImageBuildSourceRequest,
) (*challengeports.ChallengeStoredImageBuildSource, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.RootDir) == "" || strings.TrimSpace(req.DockerfilePath) == "" || strings.TrimSpace(req.ContextPath) == "" {
		return nil, nil
	}
	mode, err := safePathSegment(req.ChallengeMode, "challenge mode")
	if err != nil {
		return nil, err
	}
	packageSlug, err := safePathSegment(req.PackageSlug, "package slug")
	if err != nil {
		return nil, err
	}
	previewID, err := safePathSegment(req.PreviewID, "preview id")
	if err != nil {
		return nil, err
	}
	relDockerfile, err := importedImageBuildRelativePath(req.RootDir, req.DockerfilePath)
	if err != nil {
		return nil, err
	}
	relContext, err := importedImageBuildRelativePath(req.RootDir, req.ContextPath)
	if err != nil {
		return nil, err
	}

	targetRoot := filepath.Join(s.imageBuildSourceRootDir(), mode, packageSlug, previewID)
	sourceDir := filepath.Join(targetRoot, "source")
	if err := os.RemoveAll(targetRoot); err != nil {
		return nil, fmt.Errorf("clear imported image build source: %w", err)
	}
	if err := copyDirectoryTree(req.RootDir, sourceDir); err != nil {
		return nil, fmt.Errorf("copy imported image build source: %w", err)
	}

	stableContextPath := sourceDir
	if relContext != "." {
		stableContextPath = filepath.Join(sourceDir, relContext)
	}
	return &challengeports.ChallengeStoredImageBuildSource{
		RootDir:        targetRoot,
		SourceDir:      sourceDir,
		DockerfilePath: filepath.Join(sourceDir, relDockerfile),
		ContextPath:    stableContextPath,
	}, nil
}

func (s *ChallengePackageStorage) PrepareExportWorkspace(
	ctx context.Context,
	req challengeports.ChallengePackageExportWorkspaceRequest,
) (*challengeports.ChallengePackageExportWorkspace, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if req.ChallengeID <= 0 || req.RevisionNo <= 0 {
		return nil, apperror.ErrInvalidParams.WithCause(errors.New("题包导出参数无效"))
	}
	packageSlug, err := safePathSegment(req.PackageSlug, "package slug")
	if err != nil {
		return nil, err
	}
	_ = packageSlug

	exportRoot := filepath.Join(s.exportRootDir(), fmt.Sprintf("challenge-%d", req.ChallengeID), fmt.Sprintf("r%04d", req.RevisionNo))
	sourceDir := filepath.Join(exportRoot, "source")
	if err := os.RemoveAll(exportRoot); err != nil {
		return nil, fmt.Errorf("clear export root: %w", err)
	}
	if err := copyDirectoryTree(req.SourceDir, sourceDir); err != nil {
		return nil, fmt.Errorf("copy export package source: %w", err)
	}

	fileName := sanitizeImportedAttachmentName(req.FileName, "challenge-package.zip")
	return &challengeports.ChallengePackageExportWorkspace{
		ExportRoot:  exportRoot,
		SourceDir:   sourceDir,
		ArchivePath: filepath.Join(exportRoot, fileName),
		FileName:    fileName,
	}, nil
}

func (s *ChallengePackageStorage) ReadTextFile(ctx context.Context, rootDir string, relativePath string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	targetPath, err := safeJoinedPath(rootDir, relativePath)
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(targetPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (s *ChallengePackageStorage) WriteTextFile(ctx context.Context, rootDir string, relativePath string, content string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	targetPath, err := safeJoinedPath(rootDir, relativePath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(targetPath, []byte(content), 0o644)
}

func (s *ChallengePackageStorage) BuildExportArchive(
	ctx context.Context,
	workspace challengeports.ChallengePackageExportWorkspace,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return zipDirectory(workspace.SourceDir, workspace.ArchivePath)
}

func (s *ChallengePackageStorage) EnsureArchiveExists(ctx context.Context, archivePath string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	info, err := os.Stat(archivePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", apperror.ErrNotFound
		}
		return "", err
	}
	if info.IsDir() {
		return "", apperror.ErrNotFound
	}
	return filepath.Base(archivePath), nil
}

func (s *ChallengePackageStorage) DeletePath(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(path) == "" {
		return nil
	}
	return os.RemoveAll(path)
}

func copyDirectoryTree(sourceDir string, targetDir string) error {
	if strings.TrimSpace(sourceDir) == "" {
		return apperror.ErrInvalidParams.WithCause(errors.New("source dir is required"))
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}
	return filepath.WalkDir(sourceDir, func(current string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return apperror.ErrInvalidParams.WithCause(fmt.Errorf("路径不允许符号链接: %s", current))
		}
		relativePath, err := filepath.Rel(sourceDir, current)
		if err != nil {
			return err
		}
		if relativePath == "." {
			return nil
		}
		targetPath := filepath.Join(targetDir, relativePath)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.MkdirAll(targetPath, info.Mode().Perm())
		}
		if !info.Mode().IsRegular() {
			return apperror.ErrInvalidParams.WithCause(fmt.Errorf("只允许复制普通文件: %s", current))
		}
		return copyFile(current, targetPath)
	})
}

func copyFile(sourcePath string, targetPath string) error {
	info, err := os.Lstat(sourcePath)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return apperror.ErrInvalidParams.WithCause(fmt.Errorf("路径不允许符号链接: %s", sourcePath))
	}
	if !info.Mode().IsRegular() {
		return apperror.ErrInvalidParams.WithCause(fmt.Errorf("只允许复制普通文件: %s", sourcePath))
	}

	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}
	target, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer target.Close()

	_, err = io.Copy(target, source)
	return err
}

func importedImageBuildRelativePath(rootDir string, targetPath string) (string, error) {
	rel, err := filepath.Rel(rootDir, targetPath)
	if err != nil {
		return "", fmt.Errorf("resolve imported image build path relative to source root: %w", err)
	}
	cleanRel := filepath.Clean(rel)
	if cleanRel == ".." || strings.HasPrefix(cleanRel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("imported image build path %s escapes source root %s", targetPath, rootDir)
	}
	return cleanRel, nil
}

func zipDirectory(sourceDir string, archivePath string) error {
	if err := os.MkdirAll(filepath.Dir(archivePath), 0o755); err != nil {
		return err
	}
	target, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer target.Close()

	writer := zip.NewWriter(target)
	defer writer.Close()

	entries := make([]string, 0, 32)
	if err := filepath.WalkDir(sourceDir, func(current string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return apperror.ErrInvalidParams.WithCause(fmt.Errorf("路径不允许符号链接: %s", current))
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return apperror.ErrInvalidParams.WithCause(fmt.Errorf("只允许打包普通文件: %s", current))
		}
		relativePath, err := filepath.Rel(sourceDir, current)
		if err != nil {
			return err
		}
		entries = append(entries, relativePath)
		return nil
	}); err != nil {
		return err
	}
	sort.Strings(entries)

	for _, relativePath := range entries {
		sourcePath := filepath.Join(sourceDir, relativePath)
		if err := addZipFile(writer, sourceDir, sourcePath); err != nil {
			return err
		}
	}
	return nil
}

func addZipFile(writer *zip.Writer, rootDir string, sourcePath string) error {
	info, err := os.Lstat(sourcePath)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return apperror.ErrInvalidParams.WithCause(fmt.Errorf("路径不允许符号链接: %s", sourcePath))
	}
	if !info.Mode().IsRegular() {
		return apperror.ErrInvalidParams.WithCause(fmt.Errorf("只允许打包普通文件: %s", sourcePath))
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Method = zip.Deflate
	relativePath, err := filepath.Rel(rootDir, sourcePath)
	if err != nil {
		return err
	}
	header.Name = filepath.ToSlash(relativePath)
	fileWriter, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	file, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(fileWriter, file)
	return err
}

func safePathSegment(value string, label string) (string, error) {
	segment := strings.TrimSpace(value)
	if segment == "" {
		return "", apperror.ErrInvalidParams.WithCause(fmt.Errorf("%s is required", label))
	}
	if segment == "." || segment == ".." || strings.ContainsAny(segment, `/\`) {
		return "", apperror.ErrInvalidParams.WithCause(fmt.Errorf("%s contains invalid path segment", label))
	}
	return segment, nil
}

func safeJoinedPath(rootDir string, relativePath string) (string, error) {
	if strings.TrimSpace(rootDir) == "" {
		return "", apperror.ErrInvalidParams.WithCause(errors.New("root dir is required"))
	}
	relativePath = strings.TrimSpace(relativePath)
	if relativePath == "" {
		return "", apperror.ErrInvalidParams.WithCause(errors.New("relative path is required"))
	}
	targetPath := filepath.Clean(filepath.Join(rootDir, relativePath))
	rootAbs, err := filepath.Abs(rootDir)
	if err != nil {
		return "", err
	}
	targetAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return "", err
	}
	prefix := rootAbs + string(os.PathSeparator)
	if targetAbs != rootAbs && !strings.HasPrefix(targetAbs, prefix) {
		return "", apperror.ErrInvalidParams.WithCause(fmt.Errorf("path escapes root: %s", relativePath))
	}
	return targetAbs, nil
}
