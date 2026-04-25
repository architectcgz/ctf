package auditlog

import "testing"

func TestWithControlDoesNotCreateBackgroundContext(t *testing.T) {
	t.Parallel()

	if ctx := WithControl(nil, &Control{}); ctx != nil {
		t.Fatalf("expected nil context to stay nil, got %v", ctx)
	}
}
