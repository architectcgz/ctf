package dto

import "time"

// ProgressResp 个人解题进度响应
type ProgressResp struct {
	TotalScore      int              `json:"total_score"`      // 总得分
	TotalSolved     int              `json:"total_solved"`     // 总解题数
	Rank            int              `json:"rank"`             // 排名
	CategoryStats   []CategoryStat   `json:"category_stats"`   // 按分类统计
	DifficultyStats []DifficultyStat `json:"difficulty_stats"` // 按难度统计
}

// CategoryStat 分类统计
type CategoryStat struct {
	Category string `json:"category"` // 分类名称
	Solved   int    `json:"solved"`   // 已完成数
	Total    int    `json:"total"`    // 总数
}

// DifficultyStat 难度统计
type DifficultyStat struct {
	Difficulty string `json:"difficulty"` // 难度
	Solved     int    `json:"solved"`     // 已完成数
	Total      int    `json:"total"`      // 总数
}

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
