package dto

import "time"

type CreateContestReq struct {
	Title       string    `json:"title" binding:"required,min=1,max=200"`
	Description string    `json:"description" binding:"max=5000"`
	Mode        string    `json:"mode" binding:"required,oneof=jeopardy awd"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required,gtfield=StartTime"`
}

type UpdateContestReq struct {
	Title       *string    `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string    `json:"description" binding:"omitempty,max=5000"`
	Mode        *string    `json:"mode" binding:"omitempty,oneof=jeopardy awd"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Status      *string    `json:"status" binding:"omitempty,oneof=draft registration running frozen ended"`
}

type ContestResp struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Mode        string    `json:"mode"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListContestsReq struct {
	Status *string `form:"status" binding:"omitempty,oneof=draft registration running frozen ended"`
	Page   int     `form:"page" binding:"omitempty,min=1"`
	Size   int     `form:"size" binding:"omitempty,min=1,max=100"`
}
