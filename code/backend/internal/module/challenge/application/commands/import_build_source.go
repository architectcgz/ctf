package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultImportedImageBuildSourceRoot = "./data/challenge-image-build-sources"

type importedImageBuildSource struct {
	RootDir        string
	SourceDir      string
	DockerfilePath string
	ContextPath    string
}

func importedImageBuildSourceRoot() string {
	if dir := strings.TrimSpace(os.Getenv("CHALLENGE_IMAGE_BUILD_SOURCE_DIR")); dir != "" {
		return dir
	}
	return defaultImportedImageBuildSourceRoot
}

func persistImportedImageBuildSource(
	mode string,
	slug string,
	previewID string,
	rootDir string,
	dockerfilePath string,
	contextPath string,
) (*importedImageBuildSource, error) {
	if strings.TrimSpace(rootDir) == "" || strings.TrimSpace(dockerfilePath) == "" || strings.TrimSpace(contextPath) == "" {
		return nil, nil
	}

	targetRoot := filepath.Join(importedImageBuildSourceRoot(), strings.TrimSpace(mode), strings.TrimSpace(slug), strings.TrimSpace(previewID))
	sourceDir := filepath.Join(targetRoot, "source")
	if err := copyDirectoryTree(rootDir, sourceDir); err != nil {
		return nil, fmt.Errorf("copy imported image build source: %w", err)
	}

	relDockerfile, err := importedImageBuildRelativePath(rootDir, dockerfilePath)
	if err != nil {
		_ = os.RemoveAll(targetRoot)
		return nil, err
	}
	relContext, err := importedImageBuildRelativePath(rootDir, contextPath)
	if err != nil {
		_ = os.RemoveAll(targetRoot)
		return nil, err
	}

	stableContextPath := sourceDir
	if relContext != "." {
		stableContextPath = filepath.Join(sourceDir, relContext)
	}
	return &importedImageBuildSource{
		RootDir:        targetRoot,
		SourceDir:      sourceDir,
		DockerfilePath: filepath.Join(sourceDir, relDockerfile),
		ContextPath:    stableContextPath,
	}, nil
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
