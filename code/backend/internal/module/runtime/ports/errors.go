package ports

import "errors"

var ErrPublishedHostPortConflict = errors.New("runtime published host port conflict")

type publishedHostPortConflictError struct {
	cause error
}

func (e *publishedHostPortConflictError) Error() string {
	if e == nil || e.cause == nil {
		return ErrPublishedHostPortConflict.Error()
	}
	return ErrPublishedHostPortConflict.Error() + ": " + e.cause.Error()
}

func (e *publishedHostPortConflictError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func (e *publishedHostPortConflictError) Is(target error) bool {
	return target == ErrPublishedHostPortConflict
}

func WrapPublishedHostPortConflict(err error) error {
	if err == nil || errors.Is(err, ErrPublishedHostPortConflict) {
		return err
	}
	return &publishedHostPortConflictError{cause: err}
}
