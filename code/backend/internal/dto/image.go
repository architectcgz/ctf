package dto

import "time"

type CreateImageReq struct {
	Name        string `json:"name" binding:"required"`
	Tag         string `json:"tag" binding:"required"`
	Description string `json:"description"`
}

type UpdateImageReq struct {
	Description string `json:"description"`
	Status      string `json:"status" binding:"omitempty,oneof=pending building available failed"`
}

type ImageResp struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Tag         string    `json:"tag"`
	Description string    `json:"description"`
	Size        int64     `json:"size"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ImageQuery struct {
	Name   string `form:"name"`
	Status string `form:"status"`
	Page   int    `form:"page" binding:"omitempty,min=1"`
	Size   int    `form:"size" binding:"omitempty,min=1,max=100"`
}
