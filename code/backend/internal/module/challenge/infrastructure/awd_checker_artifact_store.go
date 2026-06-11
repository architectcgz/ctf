package infrastructure

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ctf-platform/internal/apperror"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type AWDCheckerArtifactStore struct {
	root string
}

var _ challengeports.AWDCheckerArtifactStore = (*AWDCheckerArtifactStore)(nil)

func NewAWDCheckerArtifactStore(root string) *AWDCheckerArtifactStore {
	return &AWDCheckerArtifactStore{root: strings.TrimSpace(root)}
}

func (s *AWDCheckerArtifactStore) PersistScriptCheckerArtifact(ctx context.Context, req challengeports.AWDCheckerArtifactPersistRequest) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	config := cloneAWDChallengeConfig(req.CheckerConfig)
	if req.CheckerType != string(challengeentity.AWDCheckerTypeScript) {
		return marshalAWDChallengeConfig(config)
	}
	if strings.TrimSpace(req.CheckerEntryAbs) == "" || strings.TrimSpace(req.CheckerEntryPath) == "" {
		return "", apperror.ErrInvalidParams.WithCause(errors.New("script_checker artifact entry is missing"))
	}
	files := req.CheckerFiles
	if len(files) == 0 {
		files = []challengeports.AWDCheckerArtifactFile{{Path: req.CheckerEntryPath, Abs: req.CheckerEntryAbs}}
	}
	fileContents := make([][]byte, 0, len(files))
	fileMetadata := make([]map[string]any, 0, len(files))
	digestSeed := sha256.New()
	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return "", err
		}
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
	targetDir := filepath.Join(s.rootDir(), sanitizeAWDCheckerArtifactSegment(req.Slug), digest)
	for index, file := range files {
		if err := ctx.Err(); err != nil {
			return "", err
		}
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
		if item["path"] == req.CheckerEntryPath {
			entryArtifact = item
			break
		}
	}
	config["artifact"] = map[string]any{
		"entry":        req.CheckerEntryPath,
		"storage_path": entryArtifact["storage_path"],
		"sha256":       entryArtifact["sha256"],
		"size":         entryArtifact["size"],
		"digest":       digest,
		"files":        fileMetadata,
	}
	return marshalAWDChallengeConfig(config)
}

func (s *AWDCheckerArtifactStore) ArtifactDirFromConfig(ctx context.Context, raw string) string {
	if err := ctx.Err(); err != nil {
		return ""
	}
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
	return s.artifactDirFromFile(storagePath, packagePath)
}

func (s *AWDCheckerArtifactStore) RemoveArtifactDir(ctx context.Context, dir string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	dir = filepath.Clean(strings.TrimSpace(dir))
	if dir == "." || dir == "" || !s.artifactDirInsideRoot(dir) {
		return nil
	}
	return os.RemoveAll(dir)
}

func (s *AWDCheckerArtifactStore) rootDir() string {
	if s != nil && strings.TrimSpace(s.root) != "" {
		return strings.TrimSpace(s.root)
	}
	return awdCheckerArtifactRoot()
}

func (s *AWDCheckerArtifactStore) artifactDirFromFile(storagePath, packagePath string) string {
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
	if dir == "" || !s.artifactDirInsideRoot(dir) {
		return ""
	}
	return dir
}

func (s *AWDCheckerArtifactStore) artifactDirInsideRoot(dir string) bool {
	root, err := filepath.Abs(s.rootDir())
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
