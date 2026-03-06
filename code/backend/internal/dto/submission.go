package dto

import "time"

// SubmitFlagReq Flag 提交请求
type SubmitFlagReq struct {
	Flag string `json:"flag" binding:"required"`
}

// SubmissionResp Flag 提交响应
type SubmissionResp struct {
	IsCorrect   bool      `json:"is_correct"`
	Message     string    `json:"message"`
	Points      int       `json:"points,omitempty"`
	SubmittedAt time.Time `json:"submitted_at"`
}
