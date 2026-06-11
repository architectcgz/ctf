package contracts

import (
	"time"
)

type ChallengeHintAdminResp struct {
	ID      int64  `json:"id"`
	Level   int    `json:"level"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content"`
}

type ChallengeHintReq struct {
	Level   int
	Title   string
	Content string
}

type ChallengeResp struct {
	ID              int64                     `json:"id"`
	Title           string                    `json:"title"`
	Description     string                    `json:"description"`
	Category        string                    `json:"category"`
	Difficulty      string                    `json:"difficulty"`
	Points          int                       `json:"points"`
	ImageID         *int64                    `json:"image_id"`
	AttachmentURL   string                    `json:"attachment_url,omitempty"`
	InstanceSharing string                    `json:"instance_sharing"`
	Hints           []*ChallengeHintAdminResp `json:"hints,omitempty"`
	Status          string                    `json:"status"`
	CreatedBy       *int64                    `json:"created_by,omitempty"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
}

type ChallengeQuery struct {
	Category   string
	Difficulty string
	Status     string
	CreatedBy  *int64
	Keyword    string
	SortBy     string
	Page       int
	Size       int
}

type ChallengeListItem struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Category      string    `json:"category"`
	Difficulty    string    `json:"difficulty"`
	Points        int       `json:"points"`
	SolvedCount   int64     `json:"solved_count"`
	TotalAttempts int64     `json:"total_attempts"`
	IsSolved      bool      `json:"is_solved"`
	CreatedAt     time.Time `json:"created_at"`
}

type ChallengeHintResp struct {
	ID      int64  `json:"id"`
	Level   int    `json:"level"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

type ChallengeDetailResp struct {
	ID              int64                `json:"id"`
	Title           string               `json:"title"`
	Description     string               `json:"description"`
	Category        string               `json:"category"`
	Difficulty      string               `json:"difficulty"`
	Points          int                  `json:"points"`
	NeedTarget      bool                 `json:"need_target"`
	FlagType        string               `json:"flag_type"`
	InstanceSharing string               `json:"instance_sharing"`
	AttachmentURL   string               `json:"attachment_url,omitempty"`
	Hints           []*ChallengeHintResp `json:"hints"`
	SolvedCount     int64                `json:"solved_count"`
	TotalAttempts   int64                `json:"total_attempts"`
	IsSolved        bool                 `json:"is_solved"`
	CreatedAt       time.Time            `json:"created_at"`
}

type ChallengeSelfCheckStepResp struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

type ChallengeSelfCheckPhaseResp struct {
	Passed    bool                         `json:"passed"`
	StartedAt time.Time                    `json:"started_at"`
	EndedAt   time.Time                    `json:"ended_at"`
	Steps     []ChallengeSelfCheckStepResp `json:"steps"`
}

type ChallengeSelfCheckRuntimeResp struct {
	Passed         bool                         `json:"passed"`
	StartedAt      time.Time                    `json:"started_at"`
	EndedAt        time.Time                    `json:"ended_at"`
	AccessURL      string                       `json:"access_url,omitempty"`
	ContainerCount int                          `json:"container_count"`
	NetworkCount   int                          `json:"network_count"`
	Steps          []ChallengeSelfCheckStepResp `json:"steps"`
}

type ChallengeSelfCheckResp struct {
	ChallengeID int64                         `json:"challenge_id"`
	Precheck    ChallengeSelfCheckPhaseResp   `json:"precheck"`
	Runtime     ChallengeSelfCheckRuntimeResp `json:"runtime"`
}

type ChallengePublishCheckJobResp struct {
	ID             int64                   `json:"id"`
	ChallengeID    int64                   `json:"challenge_id"`
	RequestedBy    int64                   `json:"requested_by"`
	Status         string                  `json:"status"`
	Active         bool                    `json:"active"`
	RequestSource  string                  `json:"request_source"`
	FailureSummary string                  `json:"failure_summary,omitempty"`
	StartedAt      *time.Time              `json:"started_at,omitempty"`
	FinishedAt     *time.Time              `json:"finished_at,omitempty"`
	PublishedAt    *time.Time              `json:"published_at,omitempty"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
	Result         *ChallengeSelfCheckResp `json:"result,omitempty"`
}
