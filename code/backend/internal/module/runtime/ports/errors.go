package ports

import "errors"

var (
	ErrRuntimeEngineUnavailable  = errors.New("runtime engine is not configured")
	ErrRuntimeContainerNotFound  = errors.New("runtime container not found")
	ErrRuntimeNetworkNotFound    = errors.New("runtime network not found")
	ErrPublishedHostPortConflict = errors.New("runtime published host port conflict")
)

type runtimeError struct {
	kind  error
	cause error
}

func (e *runtimeError) Error() string {
	if e == nil || e.kind == nil {
		return ""
	}
	if e.cause == nil {
		return e.kind.Error()
	}
	return e.kind.Error() + ": " + e.cause.Error()
}

func (e *runtimeError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func (e *runtimeError) Is(target error) bool {
	return e != nil && e.kind == target
}

func wrapRuntimeError(kind error, err error) error {
	if kind == nil {
		return err
	}
	if err == nil || errors.Is(err, kind) {
		return err
	}
	return &runtimeError{kind: kind, cause: err}
}

func WrapRuntimeContainerNotFound(err error) error {
	return wrapRuntimeError(ErrRuntimeContainerNotFound, err)
}

func WrapRuntimeNetworkNotFound(err error) error {
	return wrapRuntimeError(ErrRuntimeNetworkNotFound, err)
}

func WrapPublishedHostPortConflict(err error) error {
	return wrapRuntimeError(ErrPublishedHostPortConflict, err)
}
