package dto

import "time"

type CreateTagReq struct {
	Name        string `json:"name" binding:"required,min=2,max=64"`
	Type        string `json:"type" binding:"required,oneof=vulnerability tech_stack knowledge"`
	Description string `json:"description" binding:"max=500"`
}

type TagResp struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type AttachTagsReq struct {
	TagIDs []int64 `json:"tag_ids" binding:"required,min=1"`
}

type TagQuery struct {
	Type string `form:"type" binding:"omitempty,oneof=vulnerability tech_stack knowledge"`
}
