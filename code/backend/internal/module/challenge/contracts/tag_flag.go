package contracts

import "time"

type TagResp struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type FlagResp struct {
	FlagType   string `json:"flag_type"`
	FlagRegex  string `json:"flag_regex,omitempty"`
	FlagPrefix string `json:"flag_prefix,omitempty"`
	Configured bool   `json:"configured"`
}
