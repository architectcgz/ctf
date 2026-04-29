package dto

import "time"

type AWDChallengeImportPreviewResp struct {
	ID               string         `json:"id"`
	FileName         string         `json:"file_name"`
	Slug             string         `json:"slug"`
	Title            string         `json:"title"`
	Category         string         `json:"category"`
	Difficulty       string         `json:"difficulty"`
	Description      string         `json:"description"`
	ServiceType      string         `json:"service_type"`
	DeploymentMode   string         `json:"deployment_mode"`
	Version          string         `json:"version"`
	CheckerType      string         `json:"checker_type"`
	CheckerConfig    map[string]any `json:"checker_config,omitempty"`
	FlagMode         string         `json:"flag_mode,omitempty"`
	FlagConfig       map[string]any `json:"flag_config,omitempty"`
	DefenseEntryMode string         `json:"defense_entry_mode,omitempty"`
	AccessConfig     map[string]any `json:"access_config,omitempty"`
	RuntimeConfig    map[string]any `json:"runtime_config,omitempty"`
	Warnings         []string       `json:"warnings,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
}

type AWDChallengeImportCommitResp struct {
	Challenge *AWDChallengeResp `json:"challenge"`
}
