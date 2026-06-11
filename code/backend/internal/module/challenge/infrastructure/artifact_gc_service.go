package infrastructure

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	challengeports "ctf-platform/internal/module/challenge/ports"
)

const (
	ArtifactGCKindPreview          = "preview"
	ArtifactGCKindAttachment       = "attachment"
	ArtifactGCKindImageBuildSource = "image_build_source"
	ArtifactGCKindAWDChecker       = "awd_checker_artifact"

	ArtifactGCReasonExpired    = "expired"
	ArtifactGCReasonReferenced = "referenced"
	ArtifactGCReasonRecent     = "recent"
)

type ArtifactGCConfig struct {
	Now                       time.Time
	PreviewRoots              []string
	PreviewRetention          time.Duration
	AttachmentRoot            string
	AttachmentRetention       time.Duration
	ImageBuildSourceRoot      string
	ImageBuildSourceRetention time.Duration
	AWDCheckerArtifactRoot    string
	AWDCheckerRetention       time.Duration
}

type ArtifactGCRoots struct {
	PreviewRoots           []string
	AttachmentRoot         string
	ImageBuildSourceRoot   string
	AWDCheckerArtifactRoot string
}

type ArtifactReferenceReader interface {
	ListArtifactReferences(ctx context.Context) (challengeports.ArtifactReferences, error)
}

type ArtifactGCService struct {
	config     ArtifactGCConfig
	references ArtifactReferenceReader
}

type ArtifactGCCandidate struct {
	Kind      string
	Path      string
	Reason    string
	Protected bool
	SizeBytes int64
}

type ArtifactGCReport struct {
	Candidates   []ArtifactGCCandidate
	DeletedCount int
}

type ArtifactGCExecution struct {
	DryRun bool
}

func NewArtifactGCService(config ArtifactGCConfig, references ArtifactReferenceReader) *ArtifactGCService {
	if config.Now.IsZero() {
		config.Now = time.Now().UTC()
	}
	if config.PreviewRetention <= 0 {
		config.PreviewRetention = 24 * time.Hour
	}
	if config.AttachmentRetention <= 0 {
		config.AttachmentRetention = 30 * 24 * time.Hour
	}
	if config.ImageBuildSourceRetention <= 0 {
		config.ImageBuildSourceRetention = 7 * 24 * time.Hour
	}
	if config.AWDCheckerRetention <= 0 {
		config.AWDCheckerRetention = 30 * 24 * time.Hour
	}
	return &ArtifactGCService{config: config, references: references}
}

func DefaultArtifactGCConfig(now time.Time, roots ArtifactGCRoots) ArtifactGCConfig {
	return ArtifactGCConfig{
		Now:                       now,
		PreviewRoots:              append([]string(nil), roots.PreviewRoots...),
		PreviewRetention:          24 * time.Hour,
		AttachmentRoot:            roots.AttachmentRoot,
		AttachmentRetention:       30 * 24 * time.Hour,
		ImageBuildSourceRoot:      roots.ImageBuildSourceRoot,
		ImageBuildSourceRetention: 7 * 24 * time.Hour,
		AWDCheckerArtifactRoot:    roots.AWDCheckerArtifactRoot,
		AWDCheckerRetention:       30 * 24 * time.Hour,
	}
}

func (s *ArtifactGCService) PlanFiles(ctx context.Context) (*ArtifactGCReport, error) {
	if s == nil {
		return nil, fmt.Errorf("artifact gc service is not configured")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	refs, err := s.loadReferences(ctx)
	if err != nil {
		return nil, err
	}
	report := &ArtifactGCReport{}
	for _, root := range s.config.PreviewRoots {
		if err := s.collectPreviewCandidates(root, refs, report); err != nil {
			return nil, err
		}
	}
	if err := s.collectFileTreeCandidates(ArtifactGCKindAttachment, s.config.AttachmentRoot, s.config.AttachmentRetention, refs.attachmentSet, report); err != nil {
		return nil, err
	}
	if err := s.collectDirectoryTreeCandidates(ArtifactGCKindImageBuildSource, s.config.ImageBuildSourceRoot, s.config.ImageBuildSourceRetention, refs.imageBuildSourceSet, 3, report); err != nil {
		return nil, err
	}
	if err := s.collectDirectoryTreeCandidates(ArtifactGCKindAWDChecker, s.config.AWDCheckerArtifactRoot, s.config.AWDCheckerRetention, refs.awdCheckerSet, 2, report); err != nil {
		return nil, err
	}
	sort.Slice(report.Candidates, func(i, j int) bool {
		if report.Candidates[i].Kind != report.Candidates[j].Kind {
			return report.Candidates[i].Kind < report.Candidates[j].Kind
		}
		return report.Candidates[i].Path < report.Candidates[j].Path
	})
	return report, nil
}

func (s *ArtifactGCService) CollectFiles(ctx context.Context, execution ArtifactGCExecution) (*ArtifactGCReport, error) {
	report, err := s.PlanFiles(ctx)
	if err != nil {
		return nil, err
	}
	if execution.DryRun {
		return report, nil
	}
	for _, candidate := range report.Candidates {
		if candidate.Protected {
			continue
		}
		if err := s.DeleteFileCandidate(ctx, candidate); err != nil {
			return nil, err
		}
		report.DeletedCount++
	}
	return report, nil
}

func (s *ArtifactGCService) DeleteFileCandidate(ctx context.Context, candidate ArtifactGCCandidate) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if candidate.Protected {
		return nil
	}
	if s == nil {
		return fmt.Errorf("artifact gc service is not configured")
	}
	if !s.pathInsideConfiguredRoots(candidate.Path) {
		return fmt.Errorf("refuse to delete artifact outside configured roots: %s", candidate.Path)
	}
	if s.pathIsConfiguredRoot(candidate.Path) {
		return fmt.Errorf("refuse to delete configured artifact root: %s", candidate.Path)
	}
	return os.RemoveAll(candidate.Path)
}

type artifactReferenceSets struct {
	attachmentSet       map[string]struct{}
	imageBuildSourceSet map[string]struct{}
	awdCheckerSet       map[string]struct{}
}

func (s *ArtifactGCService) loadReferences(ctx context.Context) (artifactReferenceSets, error) {
	if s.references == nil {
		return artifactReferenceSets{
			attachmentSet:       map[string]struct{}{},
			imageBuildSourceSet: map[string]struct{}{},
			awdCheckerSet:       map[string]struct{}{},
		}, nil
	}
	refs, err := s.references.ListArtifactReferences(ctx)
	if err != nil {
		return artifactReferenceSets{}, err
	}
	return artifactReferenceSets{
		attachmentSet:       pathSet(s.resolveAttachmentReferencePaths(refs)),
		imageBuildSourceSet: pathSet(refs.ImageBuildSourceDirs),
		awdCheckerSet:       pathSet(s.resolveAWDCheckerReferenceDirs(ctx, refs)),
	}, nil
}

func (s *ArtifactGCService) resolveAttachmentReferencePaths(refs challengeports.ArtifactReferences) []string {
	paths := append([]string{}, refs.AttachmentPaths...)
	for _, rawURL := range refs.AttachmentURLs {
		rel := attachmentRelativePathFromURL(rawURL)
		if rel == "" || strings.TrimSpace(s.config.AttachmentRoot) == "" {
			continue
		}
		paths = append(paths, filepath.Join(s.config.AttachmentRoot, filepath.FromSlash(rel)))
	}
	return paths
}

func (s *ArtifactGCService) resolveAWDCheckerReferenceDirs(ctx context.Context, refs challengeports.ArtifactReferences) []string {
	dirs := append([]string{}, refs.AWDCheckerDirs...)
	checkerStore := NewAWDCheckerArtifactStore(s.config.AWDCheckerArtifactRoot)
	for _, rawConfig := range refs.AWDCheckerConfigs {
		if dir := checkerStore.ArtifactDirFromConfig(ctx, rawConfig); dir != "" {
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

func attachmentRelativePathFromURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	const prefix = "/api/v1/challenges/attachments/"
	if strings.HasPrefix(rawURL, prefix) {
		return strings.TrimPrefix(rawURL, prefix)
	}
	const noSlashPrefix = "api/v1/challenges/attachments/"
	if strings.HasPrefix(rawURL, noSlashPrefix) {
		return strings.TrimPrefix(rawURL, noSlashPrefix)
	}
	return ""
}

func (s *ArtifactGCService) collectPreviewCandidates(root string, refs artifactReferenceSets, report *ArtifactGCReport) error {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(root, entry.Name())
		info, err := entry.Info()
		if err != nil {
			return err
		}
		report.Candidates = append(report.Candidates, s.candidateForPath(ArtifactGCKindPreview, path, info, s.config.PreviewRetention, nil))
	}
	return nil
}

func (s *ArtifactGCService) collectFileTreeCandidates(kind, root string, retention time.Duration, protected map[string]struct{}, report *ArtifactGCReport) error {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil
	}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		report.Candidates = append(report.Candidates, s.candidateForPath(kind, path, info, retention, protected))
		return nil
	})
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *ArtifactGCService) collectDirectoryTreeCandidates(kind, root string, retention time.Duration, protected map[string]struct{}, depth int, report *ArtifactGCReport) error {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() || filepath.Clean(path) == filepath.Clean(root) {
			return nil
		}
		pathAbs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(rootAbs, pathAbs)
		if err != nil {
			return err
		}
		if rel == "." || strings.HasPrefix(rel, "..") {
			return nil
		}
		if len(strings.Split(rel, string(filepath.Separator))) != depth {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		report.Candidates = append(report.Candidates, s.candidateForPath(kind, path, info, retention, protected))
		return filepath.SkipDir
	})
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *ArtifactGCService) candidateForPath(kind string, path string, info fs.FileInfo, retention time.Duration, protected map[string]struct{}) ArtifactGCCandidate {
	candidate := ArtifactGCCandidate{
		Kind:      kind,
		Path:      filepath.Clean(path),
		SizeBytes: info.Size(),
	}
	if pathMatchesProtectedReference(path, protected) {
		candidate.Protected = true
		candidate.Reason = ArtifactGCReasonReferenced
		return candidate
	}
	if s.config.Now.Sub(info.ModTime()) < retention {
		candidate.Protected = true
		candidate.Reason = ArtifactGCReasonRecent
		return candidate
	}
	candidate.Reason = ArtifactGCReasonExpired
	return candidate
}

func (s *ArtifactGCService) pathInsideConfiguredRoots(path string) bool {
	for _, root := range s.configuredRoots() {
		if pathInsideRoot(path, root) {
			return true
		}
	}
	return false
}

func (s *ArtifactGCService) pathIsConfiguredRoot(path string) bool {
	for _, root := range s.configuredRoots() {
		if sameCleanPath(path, root) {
			return true
		}
	}
	return false
}

func (s *ArtifactGCService) configuredRoots() []string {
	roots := make([]string, 0, len(s.config.PreviewRoots)+4)
	roots = append(roots, s.config.PreviewRoots...)
	roots = appendNonEmptyRoot(roots, s.config.AttachmentRoot)
	roots = appendNonEmptyRoot(roots, s.config.ImageBuildSourceRoot)
	roots = appendNonEmptyRoot(roots, s.config.AWDCheckerArtifactRoot)
	return roots
}

func appendNonEmptyRoot(roots []string, root string) []string {
	if strings.TrimSpace(root) == "" {
		return roots
	}
	return append(roots, root)
}

func pathSet(paths []string) map[string]struct{} {
	result := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			continue
		}
		result[normalizedPath(path)] = struct{}{}
	}
	return result
}

func normalizedPath(path string) string {
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return filepath.Clean(path)
	}
	return abs
}

func sameCleanPath(left string, right string) bool {
	leftAbs, err := filepath.Abs(filepath.Clean(left))
	if err != nil {
		return false
	}
	rightAbs, err := filepath.Abs(filepath.Clean(right))
	if err != nil {
		return false
	}
	return leftAbs == rightAbs
}

func pathMatchesProtectedReference(path string, protected map[string]struct{}) bool {
	if len(protected) == 0 {
		return false
	}
	candidate := normalizedPath(path)
	for protectedPath := range protected {
		if candidate == protectedPath {
			return true
		}
		if pathInsideRoot(candidate, protectedPath) || pathInsideRoot(protectedPath, candidate) {
			return true
		}
	}
	return false
}

func pathInsideRoot(path string, root string) bool {
	root = strings.TrimSpace(root)
	if root == "" {
		return false
	}
	rootAbs, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return false
	}
	pathAbs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(rootAbs, pathAbs)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}
