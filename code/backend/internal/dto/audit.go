package dto

import "time"

type AuditLogQuery struct {
	UserID       *int64 `form:"user_id" binding:"omitempty,min=1"`
	ActorUserID  *int64 `form:"actor_user_id" binding:"omitempty,min=1"`
	Action       string `form:"action" binding:"omitempty,oneof=login logout create update delete submit admin_op"`
	ResourceType string `form:"resource_type"`
	ResourceID   *int64 `form:"resource_id" binding:"omitempty,min=1"`
	StartTime    string `form:"start_time"`
	EndTime      string `form:"end_time"`
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PageSize     int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type AuditLogItem struct {
	ID            int64          `json:"id"`
	Action        string         `json:"action"`
	ResourceType  string         `json:"resource_type"`
	ResourceID    *int64         `json:"resource_id,omitempty"`
	ActorUserID   *int64         `json:"actor_user_id,omitempty"`
	ActorUsername string         `json:"actor_username"`
	IP            *string        `json:"ip,omitempty"`
	UserAgent     *string        `json:"user_agent,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	Detail        map[string]any `json:"detail,omitempty"`
}
