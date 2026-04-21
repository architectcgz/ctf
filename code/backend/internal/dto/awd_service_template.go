package dto

import "time"

type AWDServiceTemplateQuery struct {
	Keyword     string `form:"keyword"`
	ServiceType string `form:"service_type"`
	Status      string `form:"status"`
	Page        int    `form:"page" binding:"omitempty,min=1"`
	Size        int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type CreateAWDServiceTemplateReq struct {
	Name           string `json:"name" binding:"required"`
	Slug           string `json:"slug" binding:"required"`
	Category       string `json:"category" binding:"required,oneof=web pwn reverse crypto misc forensics"`
	Difficulty     string `json:"difficulty" binding:"required,oneof=beginner easy medium hard insane"`
	Description    string `json:"description"`
	ServiceType    string `json:"service_type" binding:"required,oneof=web_http binary_tcp multi_container"`
	DeploymentMode string `json:"deployment_mode" binding:"required,oneof=single_container topology"`
}

type UpdateAWDServiceTemplateReq struct {
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	Category       string `json:"category" binding:"omitempty,oneof=web pwn reverse crypto misc forensics"`
	Difficulty     string `json:"difficulty" binding:"omitempty,oneof=beginner easy medium hard insane"`
	Description    string `json:"description"`
	ServiceType    string `json:"service_type" binding:"omitempty,oneof=web_http binary_tcp multi_container"`
	DeploymentMode string `json:"deployment_mode" binding:"omitempty,oneof=single_container topology"`
	Status         string `json:"status" binding:"omitempty,oneof=draft published archived"`
}

type AWDServiceTemplateResp struct {
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

type AWDServiceTemplatePageResp struct {
	Items []*AWDServiceTemplateResp `json:"items"`
	Total int64                     `json:"total"`
	Page  int                       `json:"page"`
	Size  int                       `json:"size"`
}
