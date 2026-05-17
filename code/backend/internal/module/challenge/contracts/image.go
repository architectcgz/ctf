package contracts

import "time"

type ImageResp struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Tag           string     `json:"tag"`
	Description   string     `json:"description"`
	Size          int64      `json:"size"`
	SizeFormatted string     `json:"size_formatted"`
	Status        string     `json:"status"`
	Digest        string     `json:"digest,omitempty"`
	SourceType    string     `json:"source_type,omitempty"`
	BuildJobID    *int64     `json:"build_job_id,omitempty"`
	LastError     string     `json:"last_error,omitempty"`
	VerifiedAt    *time.Time `json:"verified_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
