package contracts

import "time"

type AWDChallengeResp struct {
	ID               int64          `json:"id"`
	Name             string         `json:"name"`
	Slug             string         `json:"slug"`
	Category         string         `json:"category"`
	Difficulty       string         `json:"difficulty"`
	Description      string         `json:"description"`
	ServiceType      string         `json:"service_type"`
	DeploymentMode   string         `json:"deployment_mode"`
	Version          string         `json:"version"`
	Status           string         `json:"status"`
	ReadinessStatus  string         `json:"readiness_status"`
	CheckerType      string         `json:"checker_type,omitempty"`
	CheckerConfig    map[string]any `json:"checker_config,omitempty"`
	FlagMode         string         `json:"flag_mode,omitempty"`
	FlagConfig       map[string]any `json:"flag_config,omitempty"`
	DefenseEntryMode string         `json:"defense_entry_mode,omitempty"`
	AccessConfig     map[string]any `json:"access_config,omitempty"`
	RuntimeConfig    map[string]any `json:"runtime_config,omitempty"`
	CreatedBy        *int64         `json:"created_by,omitempty"`
	LastVerifiedAt   *time.Time     `json:"last_verified_at,omitempty"`
	UpdatedAt        time.Time      `json:"updated_at"`
	CreatedAt        time.Time      `json:"created_at"`
}

type AWDChallengePageResp struct {
	Items []*AWDChallengeResp `json:"items"`
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Size  int                 `json:"size"`
}

type ChallengeImportImageDeliveryResp struct {
	SourceType     string `json:"source_type"`
	TargetImageRef string `json:"target_image_ref,omitempty"`
	BuildStatus    string `json:"build_status,omitempty"`
	SuggestedTag   string `json:"suggested_tag,omitempty"`
	Digest         string `json:"digest,omitempty"`
	LastError      string `json:"last_error,omitempty"`
}

type AWDChallengeImportPreviewResp struct {
	ID               string                           `json:"id"`
	FileName         string                           `json:"file_name"`
	Slug             string                           `json:"slug"`
	Title            string                           `json:"title"`
	Category         string                           `json:"category"`
	Difficulty       string                           `json:"difficulty"`
	Description      string                           `json:"description"`
	ServiceType      string                           `json:"service_type"`
	DeploymentMode   string                           `json:"deployment_mode"`
	Version          string                           `json:"version"`
	CheckerType      string                           `json:"checker_type"`
	CheckerConfig    map[string]any                   `json:"checker_config,omitempty"`
	FlagMode         string                           `json:"flag_mode,omitempty"`
	FlagConfig       map[string]any                   `json:"flag_config,omitempty"`
	DefenseEntryMode string                           `json:"defense_entry_mode,omitempty"`
	AccessConfig     map[string]any                   `json:"access_config,omitempty"`
	RuntimeConfig    map[string]any                   `json:"runtime_config,omitempty"`
	ImageDelivery    ChallengeImportImageDeliveryResp `json:"image_delivery"`
	Warnings         []string                         `json:"warnings,omitempty"`
	CreatedAt        time.Time                        `json:"created_at"`
}

type AWDChallengeImportCommitResp struct {
	Challenge *AWDChallengeResp `json:"challenge"`
}
