package http

import "time"

type PageResult[T any] struct {
	List  []T   `json:"list"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"page_size"`
}

type TagResp struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

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

type FlagResp struct {
	FlagType   string `json:"flag_type"`
	FlagRegex  string `json:"flag_regex,omitempty"`
	FlagPrefix string `json:"flag_prefix,omitempty"`
	Configured bool   `json:"configured"`
}
