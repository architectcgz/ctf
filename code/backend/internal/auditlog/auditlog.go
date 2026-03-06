package auditlog

import "context"

type Entry struct {
	UserID       *int64
	Action       string
	ResourceType string
	ResourceID   *int64
	Detail       map[string]any
	IPAddress    string
	UserAgent    *string
}

type Recorder interface {
	Record(ctx context.Context, entry Entry) error
}
