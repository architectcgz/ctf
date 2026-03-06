package dto

import "time"

type CreateTagReq struct {
	Name      string `json:"name" binding:"required,max=64"`
	Dimension string `json:"dimension" binding:"required,oneof=category technique tool platform"`
}

type TagResp struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Dimension string    `json:"dimension"`
	CreatedAt time.Time `json:"created_at"`
}

type AttachTagsReq struct {
	TagIDs []int64 `json:"tag_ids" binding:"required,min=1"`
}

type TagQuery struct {
	Dimension string `form:"dimension" binding:"omitempty,oneof=category technique tool platform"`
}
