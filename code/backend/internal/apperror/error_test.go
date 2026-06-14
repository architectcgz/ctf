package apperror

import (
	"errors"
	"net/http"
	"testing"
)

func TestAppErrorWithCauseSupportsErrorsIs(t *testing.T) {
	t.Parallel()

	err := ErrInvalidParams.WithCause(errors.New("bad input"))
	if !errors.Is(err, ErrInvalidParams) {
		t.Fatalf("expected errors.Is to match original app error, got %v", err)
	}
}

func TestAppErrorWithCausePreservesPublicContract(t *testing.T) {
	t.Parallel()

	cause := errors.New("dial tcp backend refused")
	err := ErrServiceUnavailable.WithCause(cause)

	if !errors.Is(err, ErrServiceUnavailable) {
		t.Fatalf("expected errors.Is to match original app error, got %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("expected wrapped cause to remain inspectable, got %v", err)
	}
	if err.Code != ErrServiceUnavailable.Code {
		t.Fatalf("code = %d, want %d", err.Code, ErrServiceUnavailable.Code)
	}
	if err.Message != ErrServiceUnavailable.Message {
		t.Fatalf("message = %q, want %q", err.Message, ErrServiceUnavailable.Message)
	}
	if status := HTTPStatus(err); status != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", status, http.StatusServiceUnavailable)
	}
}

func TestAppErrorWithMessagePreservesCodeAndHTTPStatus(t *testing.T) {
	t.Parallel()

	cause := errors.New("page_size must be positive")
	err := ErrInvalidParams.WithCause(cause).WithMessage("分页参数无效")

	if !errors.Is(err, ErrInvalidParams) {
		t.Fatalf("expected errors.Is to match original app error, got %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("expected wrapped cause to remain inspectable, got %v", err)
	}
	if err.Code != ErrInvalidParams.Code {
		t.Fatalf("code = %d, want %d", err.Code, ErrInvalidParams.Code)
	}
	if err.Message != "分页参数无效" {
		t.Fatalf("message = %q, want custom message", err.Message)
	}
	if status := HTTPStatus(err); status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
	}
}
