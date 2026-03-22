package dto

import "time"

type InstanceResp struct {
	ID               int64     `json:"id"`
	ChallengeID      int64     `json:"challenge_id"`
	Status           string    `json:"status"`
	AccessURL        string    `json:"access_url"`
	ExpiresAt        time.Time `json:"expires_at"`
	ExtendCount      int       `json:"extend_count"`
	MaxExtends       int       `json:"max_extends"`
	RemainingExtends int       `json:"remaining_extends"`
	CreatedAt        time.Time `json:"created_at"`
}

type InstanceInfo struct {
	ID               int64     `json:"id"`
	ChallengeID      int64     `json:"challenge_id"`
	ChallengeTitle   string    `json:"challenge_title,omitempty"`
	Category         string    `json:"category,omitempty"`
	Difficulty       string    `json:"difficulty,omitempty"`
	FlagType         string    `json:"flag_type,omitempty"`
	Status           string    `json:"status"`
	AccessURL        string    `json:"access_url"`
	ExpiresAt        time.Time `json:"expires_at"`
	RemainingTime    int64     `json:"remaining_time"` // 秒
	ExtendCount      int       `json:"extend_count"`
	MaxExtends       int       `json:"max_extends"`
	RemainingExtends int       `json:"remaining_extends"`
	CreatedAt        time.Time `json:"created_at"`
}

type InstanceAccessResp struct {
	AccessURL string `json:"access_url"`
}
