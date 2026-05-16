package infrastructure

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"

	runtimeports "ctf-platform/internal/module/runtime/ports"
)

func (e *Engine) WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error {
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	if strings.TrimSpace(containerID) == "" {
		return fmt.Errorf("container id is empty")
	}

	resolvedPath, err := e.resolveContainerFilePath(ctx, containerID, filePath)
	if err != nil {
		return err
	}

	dir := path.Dir(resolvedPath)
	if dir == "." || dir == "" {
		dir = "/"
	}

	var archive bytes.Buffer
	tw := tar.NewWriter(&archive)
	header := &tar.Header{
		Name: path.Base(resolvedPath),
		Mode: 0o644,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	if _, err := tw.Write(content); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}

	return cli.CopyToContainer(ctx, containerID, dir, io.NopCloser(bytes.NewReader(archive.Bytes())), container.CopyToContainerOptions{})
}

func (e *Engine) ReadFileFromContainer(ctx context.Context, containerID, filePath string, limit int64) ([]byte, error) {
	cli, err := e.requireClient()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(containerID) == "" {
		return nil, fmt.Errorf("container id is empty")
	}
	if limit <= 0 {
		limit = 256 * 1024
	}

	resolvedPath, err := e.resolveContainerFilePath(ctx, containerID, filePath)
	if err != nil {
		return nil, err
	}

	reader, _, err := cli.CopyFromContainer(ctx, containerID, resolvedPath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	tr := tar.NewReader(reader)
	for {
		header, err := tr.Next()
		if err != nil {
			return nil, err
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			continue
		}
		if header.Size > limit {
			return nil, fmt.Errorf("file exceeds limit")
		}
		var content bytes.Buffer
		if _, err := io.CopyN(&content, tr, limit+1); err != nil && err != io.EOF {
			return nil, err
		}
		if int64(content.Len()) > limit {
			return nil, fmt.Errorf("file exceeds limit")
		}
		return content.Bytes(), nil
	}
}

func (e *Engine) ListDirectoryFromContainer(ctx context.Context, containerID, dirPath string, limit int) ([]runtimeports.ContainerDirectoryEntry, error) {
	cli, err := e.requireClient()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(containerID) == "" {
		return nil, fmt.Errorf("container id is empty")
	}
	if limit <= 0 {
		limit = 300
	}

	resolvedPath, err := e.resolveContainerFilePath(ctx, containerID, dirPath)
	if err != nil {
		return nil, err
	}

	reader, _, err := cli.CopyFromContainer(ctx, containerID, resolvedPath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	rootName := path.Base(path.Clean(resolvedPath))
	entriesByName := make(map[string]runtimeports.ContainerDirectoryEntry)
	tr := tar.NewReader(reader)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		name, entryType, ok := containerDirectoryEntryFromTar(rootName, header)
		if !ok {
			continue
		}
		entry := runtimeports.ContainerDirectoryEntry{
			Name: name,
			Type: entryType,
			Size: header.Size,
		}
		if existing, exists := entriesByName[name]; !exists || existing.Type != "dir" {
			entriesByName[name] = entry
		}
		if len(entriesByName) >= limit {
			break
		}
	}

	entries := make([]runtimeports.ContainerDirectoryEntry, 0, len(entriesByName))
	for _, entry := range entriesByName {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Type != entries[j].Type {
			return entries[i].Type == "dir"
		}
		return entries[i].Name < entries[j].Name
	})
	return entries, nil
}

func (e *Engine) ExecContainerCommand(ctx context.Context, containerID string, command []string, stdin []byte, limit int64) ([]byte, error) {
	cli, err := e.requireClient()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(containerID) == "" {
		return nil, fmt.Errorf("container id is empty")
	}
	if len(command) == 0 {
		return nil, fmt.Errorf("command is empty")
	}
	if limit <= 0 {
		limit = 64 * 1024
	}
	workingDir, err := e.inspectContainerWorkingDir(ctx, containerID)
	if err != nil {
		return nil, err
	}

	execID, err := cli.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		AttachStdin:  len(stdin) > 0,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		Cmd:          command,
		WorkingDir:   workingDir,
	})
	if err != nil {
		return nil, err
	}

	attach, err := cli.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{Tty: false})
	if err != nil {
		return nil, err
	}
	defer attach.Close()

	if len(stdin) > 0 {
		go func() {
			_, _ = attach.Conn.Write(stdin)
			_ = attach.CloseWrite()
		}()
	}

	var output bytes.Buffer
	limited := &limitedBuffer{buffer: &output, limit: limit}
	if _, err := stdcopy.StdCopy(limited, limited, attach.Reader); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func (e *Engine) resolveContainerFilePath(ctx context.Context, containerID, filePath string) (string, error) {
	workingDir, err := e.inspectContainerWorkingDir(ctx, containerID)
	if err != nil {
		return "", err
	}
	return resolveContainerFilePath(workingDir, filePath), nil
}

func (e *Engine) inspectContainerWorkingDir(ctx context.Context, containerID string) (string, error) {
	cli, err := e.requireClient()
	if err != nil {
		return "", err
	}

	info, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", err
	}
	if info.Config == nil {
		return "", nil
	}
	return info.Config.WorkingDir, nil
}

func resolveContainerFilePath(workingDir, filePath string) string {
	cleanFilePath := path.Clean(filePath)
	if path.IsAbs(cleanFilePath) {
		return cleanFilePath
	}

	base := strings.TrimSpace(workingDir)
	if base == "" {
		base = "/"
	}
	if !path.IsAbs(base) {
		base = "/" + base
	}
	return path.Join(path.Clean(base), cleanFilePath)
}

func containerDirectoryEntryFromTar(rootName string, header *tar.Header) (string, string, bool) {
	if header == nil {
		return "", "", false
	}
	name := strings.Trim(path.Clean(header.Name), "/")
	if name == "" || name == "." || name == rootName {
		return "", "", false
	}
	if rootName != "." && strings.HasPrefix(name, rootName+"/") {
		name = strings.TrimPrefix(name, rootName+"/")
	}
	parts := strings.Split(name, "/")
	if len(parts) == 0 || parts[0] == "" || parts[0] == "." {
		return "", "", false
	}

	entryType := "file"
	if len(parts) > 1 || header.Typeflag == tar.TypeDir {
		entryType = "dir"
	} else if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
		entryType = "other"
	}
	return parts[0], entryType, true
}

func (e *Engine) ExecContainerInteractive(ctx context.Context, containerID string, command []string, stdin io.Reader, stdout io.Writer) error {
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	if strings.TrimSpace(containerID) == "" {
		return fmt.Errorf("container id is empty")
	}
	if len(command) == 0 {
		command = []string{"/bin/sh"}
	}

	execID, err := cli.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          command,
	})
	if err != nil {
		return err
	}

	attach, err := cli.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{Tty: true})
	if err != nil {
		return err
	}
	defer attach.Close()

	copyErr := make(chan error, 2)
	go func() {
		_, err := io.Copy(attach.Conn, stdin)
		_ = attach.CloseWrite()
		copyErr <- err
	}()
	go func() {
		_, err := io.Copy(stdout, attach.Reader)
		copyErr <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-copyErr:
		if err != nil && err != io.EOF {
			return err
		}
		return nil
	}
}
