package dto

import (
	"time"

	"ctf-platform/internal/model"
)

type InstanceResp struct {
	ID               int64            `json:"id"`
	ChallengeID      int64            `json:"challenge_id"`
	Status           string           `json:"status"`
	ShareScope       model.ShareScope `json:"share_scope"`
	AccessURL        string           `json:"access_url"`
	ExpiresAt        time.Time        `json:"expires_at"`
	ExtendCount      int              `json:"extend_count"`
	MaxExtends       int              `json:"max_extends"`
	RemainingExtends int              `json:"remaining_extends"`
	CreatedAt        time.Time        `json:"created_at"`
}

type InstanceInfo struct {
	ID               int64            `json:"id"`
	ChallengeID      int64            `json:"challenge_id"`
	ChallengeTitle   string           `json:"challenge_title,omitempty"`
	Category         string           `json:"category,omitempty"`
	Difficulty       string           `json:"difficulty,omitempty"`
	FlagType         string           `json:"flag_type,omitempty"`
	Status           string           `json:"status"`
	ShareScope       model.ShareScope `json:"share_scope"`
	AccessURL        string           `json:"access_url"`
	ExpiresAt        time.Time        `json:"expires_at"`
	RemainingTime    int64            `json:"remaining_time"` // 秒
	ExtendCount      int              `json:"extend_count"`
	MaxExtends       int              `json:"max_extends"`
	RemainingExtends int              `json:"remaining_extends"`
	CreatedAt        time.Time        `json:"created_at"`
}

type InstanceAccessResp struct {
	AccessURL string `json:"access_url"`
}

type AWDDefenseSSHAccessResp struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Command      string `json:"command"`
	VSCodeConfig string `json:"vscode_config"`
	ExpiresAt    string `json:"expires_at"`
}

type AWDDefenseFileResp struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Size    int    `json:"size"`
}

type AWDDefenseFileSaveReq struct {
	Path    string `json:"path" binding:"required"`
	Content string `json:"content"`
	Backup  bool   `json:"backup"`
}

type AWDDefenseFileSaveResp struct {
	Path       string `json:"path"`
	Size       int    `json:"size"`
	BackupPath string `json:"backup_path,omitempty"`
}

type AWDDefenseCommandReq struct {
	Command string `json:"command" binding:"required"`
}

type AWDDefenseCommandResp struct {
	Command string `json:"command"`
	Output  string `json:"output"`
}
