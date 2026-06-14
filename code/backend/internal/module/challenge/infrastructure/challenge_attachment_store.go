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
	platformstorage "ctf-platform/internal/platform/storage"
)

type ChallengeAttachmentStore struct {
	store     platformstorage.LocalWritableStore
	namespace string
}

var _ challengeports.ChallengeAttachmentStore = (*ChallengeAttachmentStore)(nil)

func NewChallengeAttachmentStore(store platformstorage.LocalWritableStore, namespace string) *ChallengeAttachmentStore {
	return &ChallengeAttachmentStore{store: store, namespace: strings.TrimSpace(namespace)}
}

func (s *ChallengeAttachmentStore) PersistImportedAttachmentBundle(
	ctx context.Context,
	req challengeports.ChallengeImportedAttachmentBundleRequest,
) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if s == nil || s.store == nil {
		return "", fmt.Errorf("challenge attachment store is not configured")
	}
	if len(req.Attachments) == 0 {
		return "", nil
	}
	packageSlug, err := safePathSegment(req.PackageSlug, "package slug")
	if err != nil {
		return "", err
	}

	if len(req.Attachments) == 1 {
		attachment := req.Attachments[0]
		fileName := sanitizeImportedAttachmentName(attachment.Name, attachment.Path)
		storageKey := s.attachmentKey(packageSlug, fileName)
		targetPath, err := s.store.PrepareLocalWrite(ctx, storageKey)
		if err != nil {
			return "", fmt.Errorf("prepare attachment target: %w", err)
		}
		if err := copyFile(attachment.AbsolutePath, targetPath); err != nil {
			return "", fmt.Errorf("copy attachment: %w", err)
		}
		return buildAttachmentURLFromRelativePath(storageKey), nil
	}

	bundleName := sanitizeImportedAttachmentName(packageSlug+"-attachments.zip", packageSlug+"-attachments.zip")
	storageKey := s.attachmentKey(packageSlug, bundleName)
	targetPath, err := s.store.PrepareLocalWrite(ctx, storageKey)
	if err != nil {
		return "", fmt.Errorf("prepare attachment bundle target: %w", err)
	}
	if err := writeImportedAttachmentBundle(targetPath, req.Attachments); err != nil {
		return "", err
	}
	return buildAttachmentURLFromRelativePath(storageKey), nil
}

func (s *ChallengeAttachmentStore) OpenAttachment(ctx context.Context, relativePath string) (*challengeports.ChallengeAttachmentDownload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s == nil || s.store == nil {
		return nil, apperror.ErrServiceUnavailable
	}
	reader, err := s.store.Open(ctx, relativePath)
	if err != nil {
		if errors.Is(err, platformstorage.ErrUnsafeKey) {
			return nil, apperror.ErrInvalidParams.WithCause(err)
		}
		if errors.Is(err, platformstorage.ErrNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	info, err := s.store.Stat(ctx, relativePath)
	if err != nil {
		_ = reader.Close()
		if errors.Is(err, platformstorage.ErrUnsafeKey) {
			return nil, apperror.ErrInvalidParams.WithCause(err)
		}
		if errors.Is(err, platformstorage.ErrNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	return &challengeports.ChallengeAttachmentDownload{
		FileName: filepath.Base(info.Key),
		Reader:   reader,
		Size:     info.Size,
	}, nil
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

func (s *ChallengeAttachmentStore) attachmentKey(packageSlug, fileName string) string {
	base := "imports/" + packageSlug + "/" + fileName
	namespace := strings.Trim(strings.TrimSpace(s.namespace), "/")
	if namespace == "" {
		return base
	}
	return namespace + "/" + base
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
