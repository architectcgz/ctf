package infrastructure

import (
	"errors"
	"strings"

	"github.com/docker/docker/errdefs"

	runtimeports "ctf-platform/internal/module/runtime/ports"
)

func normalizeContainerCreateError(err error) error {
	if err == nil {
		return nil
	}
	if isPublishedHostPortConflictError(err) {
		return runtimeports.WrapPublishedHostPortConflict(err)
	}
	return err
}

func normalizeContainerNotFoundError(err error) error {
	if err == nil {
		return nil
	}
	if isRuntimeContainerNotFoundError(err) {
		return runtimeports.WrapRuntimeContainerNotFound(err)
	}
	return err
}

func normalizeNetworkNotFoundError(err error) error {
	if err == nil {
		return nil
	}
	if isRuntimeNetworkNotFoundError(err) {
		return runtimeports.WrapRuntimeNetworkNotFound(err)
	}
	return err
}

func isPublishedHostPortConflictError(err error) bool {
	for current := err; current != nil; current = errors.Unwrap(current) {
		message := strings.ToLower(strings.TrimSpace(current.Error()))
		if message == "" {
			continue
		}
		if strings.Contains(message, "port is already allocated") ||
			strings.Contains(message, "address already in use") ||
			strings.Contains(message, "bind for 0.0.0.0:") ||
			strings.Contains(message, "bind for [::]:") {
			return true
		}
	}
	return false
}

func isRuntimeContainerNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if errdefs.IsNotFound(err) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "no such container")
}

func isRuntimeNetworkNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if errdefs.IsNotFound(err) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "network") && strings.Contains(message, "not found")
}
