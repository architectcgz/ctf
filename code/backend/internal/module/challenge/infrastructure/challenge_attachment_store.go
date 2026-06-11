package infrastructure

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"ctf-platform/internal/apperror"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ChallengeAttachmentStore struct {
	root string
}

var _ challengeports.ChallengeAttachmentStore = (*ChallengeAttachmentStore)(nil)

func NewChallengeAttachmentStore(root string) *ChallengeAttachmentStore {
	return &ChallengeAttachmentStore{root: strings.TrimSpace(root)}
}

func (s *ChallengeAttachmentStore) rootDir() string {
	if s != nil && strings.TrimSpace(s.root) != "" {
		return strings.TrimSpace(s.root)
	}
	return challengeImportedAttachmentRoot()
}

func (s *ChallengeAttachmentStore) PersistImportedAttachmentBundle(
	ctx context.Context,
	req challengeports.ChallengeImportedAttachmentBundleRequest,
) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if len(req.Attachments) == 0 {
		return "", nil
	}
	packageSlug, err := safePathSegment(req.PackageSlug, "package slug")
	if err != nil {
		return "", err
	}

	targetDir := filepath.Join(s.rootDir(), "imports", packageSlug)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", fmt.Errorf("create attachment dir: %w", err)
	}

	if len(req.Attachments) == 1 {
		attachment := req.Attachments[0]
		fileName := sanitizeImportedAttachmentName(attachment.Name, attachment.Path)
		targetPath := filepath.Join(targetDir, fileName)
		if err := copyFile(attachment.AbsolutePath, targetPath); err != nil {
			return "", fmt.Errorf("copy attachment: %w", err)
		}
		return buildAttachmentURLFromRelativePath(filepath.ToSlash(filepath.Join("imports", packageSlug, fileName))), nil
	}

	bundleName := sanitizeImportedAttachmentName(packageSlug+"-attachments.zip", packageSlug+"-attachments.zip")
	targetPath := filepath.Join(targetDir, bundleName)
	if err := writeImportedAttachmentBundle(targetPath, req.Attachments); err != nil {
		return "", err
	}
	return buildAttachmentURLFromRelativePath(filepath.ToSlash(filepath.Join("imports", packageSlug, bundleName))), nil
}

func writeImportedAttachmentBundle(
	targetPath string,
	attachments []challengeports.ChallengeImportedAttachment,
) error {
	target, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("create attachment bundle: %w", err)
	}
	defer target.Close()

	archiveWriter := zip.NewWriter(target)
	for _, attachment := range attachments {
		if err := addImportedAttachmentToZip(archiveWriter, attachment); err != nil {
			_ = archiveWriter.Close()
			return err
		}
	}
	return archiveWriter.Close()
}

func addImportedAttachmentToZip(writer *zip.Writer, attachment challengeports.ChallengeImportedAttachment) error {
	info, err := os.Lstat(attachment.AbsolutePath)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return apperror.ErrInvalidParams.WithCause(errors.New("附件不允许符号链接"))
	}
	if !info.Mode().IsRegular() {
		return apperror.ErrInvalidParams.WithCause(errors.New("附件必须是普通文件"))
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = sanitizeImportedAttachmentName(attachment.Name, attachment.Path)
	header.Method = zip.Deflate

	fileWriter, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	source, err := os.Open(attachment.AbsolutePath)
	if err != nil {
		return err
	}
	defer source.Close()

	_, err = io.Copy(fileWriter, source)
	return err
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
