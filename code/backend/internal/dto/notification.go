package dto

import "time"

type NotificationReq struct {
	Type    string  `json:"type" binding:"required,oneof=system contest challenge team"`
	Title   string  `json:"title" binding:"required,max=256"`
	Content string  `json:"content" binding:"required"`
	Link    *string `json:"link,omitempty" binding:"omitempty,max=512"`
}

type NotificationQuery struct {
	Type     string `form:"type" binding:"omitempty,oneof=system contest challenge team"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type NotificationInfo struct {
	ID        int64      `json:"id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Content   *string    `json:"content,omitempty"`
	Unread    bool       `json:"unread"`
	Link      *string    `json:"link,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
}
