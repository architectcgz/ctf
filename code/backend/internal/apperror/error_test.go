package apperror

import (
	"errors"
	"testing"
)

func TestAppErrorWithCauseSupportsErrorsIs(t *testing.T) {
	t.Parallel()

	err := ErrInvalidParams.WithCause(errors.New("bad input"))
	if !errors.Is(err, ErrInvalidParams) {
		t.Fatalf("expected errors.Is to match original app error, got %v", err)
	}
}
