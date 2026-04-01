package dto

import "time"

type ChallengeImportAttachmentResp struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type ChallengeImportFlagResp struct {
	Type   string `json:"type"`
	Prefix string `json:"prefix,omitempty"`
}

type ChallengeImportRuntimeResp struct {
	Type     string `json:"type,omitempty"`
	ImageRef string `json:"image_ref,omitempty"`
}

type ChallengeImportTopologyExtensionResp struct {
	Source  string `json:"source,omitempty"`
	Enabled bool   `json:"enabled"`
}

type ChallengeImportExtensionsResp struct {
	Topology ChallengeImportTopologyExtensionResp `json:"topology"`
}

type ChallengeImportPreviewResp struct {
	ID          string                          `json:"id"`
	FileName    string                          `json:"file_name"`
	Slug        string                          `json:"slug"`
	Title       string                          `json:"title"`
	Description string                          `json:"description"`
	Category    string                          `json:"category"`
	Difficulty  string                          `json:"difficulty"`
	Points      int                             `json:"points"`
	Attachments []ChallengeImportAttachmentResp `json:"attachments,omitempty"`
	Hints       []ChallengeHintAdminResp        `json:"hints,omitempty"`
	Flag        ChallengeImportFlagResp         `json:"flag"`
	Runtime     ChallengeImportRuntimeResp      `json:"runtime"`
	Extensions  ChallengeImportExtensionsResp   `json:"extensions"`
	Warnings    []string                        `json:"warnings,omitempty"`
	CreatedAt   time.Time                       `json:"created_at"`
}

type ChallengeImportCommitResp struct {
	Challenge *ChallengeResp `json:"challenge"`
}
