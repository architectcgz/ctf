package dto

import "time"

// TimelineEvent 时间线事件
type TimelineEvent struct {
	Type        string    `json:"type"`                 // 事件类型: instance_start, flag_submit, instance_destroy
	ChallengeID int64     `json:"challenge_id"`         // 靶场 ID
	Title       string    `json:"title"`                // 靶场标题
	Timestamp   time.Time `json:"timestamp"`            // 事件时间
	IsCorrect   *bool     `json:"is_correct,omitempty"` // Flag 是否正确（仅 flag_submit）
	Points      *int      `json:"points,omitempty"`     // 获得分数（仅正确提交）
	Detail      string    `json:"detail,omitempty"`     // 更细颗粒度的步骤描述
}

// TimelineResp 时间线响应
type TimelineResp struct {
	Events []TimelineEvent `json:"events"`
}
