package domain

import (
	"strings"
	"time"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type ReportDownload struct {
	Path        string
	FileName    string
	ContentType string
}

type ReportUser struct {
	ID        int64
	Username  string
	Name      string
	ClassName string
	Role      string
}

type PersonalReportStats struct {
	TotalScore    int
	TotalSolved   int
	TotalAttempts int
	Rank          int
}

type ReportDimensionStat struct {
	Dimension string
	Solved    int
	Total     int
}

type ClassDimensionAverage struct {
	Dimension string
	AvgScore  float64
}

type ClassTopStudent struct {
	UserID     int64
	Username   string
	TotalScore int
	Rank       int
}

type ContestExportScoreboardItem struct {
	Rank             int        `json:"rank"`
	TeamID           int64      `json:"team_id"`
	TeamName         string     `json:"team_name"`
	Score            int        `json:"score"`
	SolvedCount      int        `json:"solved_count"`
	LastSubmissionAt *time.Time `json:"last_submission_at,omitempty"`
}

type ContestExportChallengeItem struct {
	ContestChallengeID int64   `json:"contest_challenge_id"`
	ChallengeID        int64   `json:"challenge_id"`
	Title              string  `json:"title"`
	Category           string  `json:"category"`
	Difficulty         string  `json:"difficulty"`
	Points             int     `json:"points"`
	Order              int     `json:"order"`
	IsVisible          bool    `json:"is_visible"`
	SolveCount         int     `json:"solve_count"`
	FirstBloodBy       *int64  `json:"first_blood_by,omitempty"`
	FirstBloodTeamName *string `json:"first_blood_team_name,omitempty"`
}

type ContestExportTeamMember struct {
	UserID    int64      `json:"user_id"`
	Username  string     `json:"username"`
	Name      string     `json:"name,omitempty"`
	ClassName string     `json:"class_name,omitempty"`
	JoinedAt  *time.Time `json:"joined_at,omitempty"`
}

type ContestExportTeamItem struct {
	TeamID          int64                     `json:"team_id"`
	Name            string                    `json:"name"`
	CaptainID       int64                     `json:"captain_id"`
	CaptainUsername string                    `json:"captain_username"`
	MaxMembers      int                       `json:"max_members"`
	TotalScore      int                       `json:"total_score"`
	LastSolveAt     *time.Time                `json:"last_solve_at,omitempty"`
	MemberCount     int                       `json:"member_count"`
	Members         []ContestExportTeamMember `json:"members"`
}

type ReviewArchiveSummary struct {
	TotalChallenges        int        `json:"total_challenges"`
	TotalSolved            int        `json:"total_solved"`
	TotalScore             int        `json:"total_score"`
	Rank                   int        `json:"rank"`
	TotalAttempts          int        `json:"total_attempts"`
	TimelineEventCount     int        `json:"timeline_event_count"`
	EvidenceEventCount     int        `json:"evidence_event_count"`
	WriteupCount           int        `json:"writeup_count"`
	ManualReviewCount      int        `json:"manual_review_count"`
	CorrectSubmissionCount int        `json:"correct_submission_count"`
	LastActivityAt         *time.Time `json:"last_activity_at,omitempty"`
}

type ReviewArchiveObservation struct {
	Code      string  `json:"code"`
	Label     string  `json:"label"`
	Severity  string  `json:"severity"`
	Dimension *string `json:"dimension,omitempty"`
	Summary   string  `json:"summary"`
	Evidence  string  `json:"evidence,omitempty"`
	Action    string  `json:"action,omitempty"`
}

type ReviewArchiveTeacherObservations struct {
	Items []ReviewArchiveObservation `json:"items"`
}

type ReviewArchiveTimelineEvent struct {
	Type              string    `json:"type"`
	ChallengeID       int64     `json:"challenge_id"`
	AWDChallengeID    int64     `json:"awd_challenge_id,omitempty"`
	AWDChallengeTitle string    `json:"awd_challenge_title,omitempty"`
	Title             string    `json:"title"`
	Timestamp         time.Time `json:"timestamp"`
	IsCorrect         *bool     `json:"is_correct,omitempty"`
	Points            *int      `json:"points,omitempty"`
	Detail            string    `json:"detail,omitempty"`
}

type ReviewArchiveEvidenceEvent struct {
	Type              string         `json:"type"`
	ChallengeID       int64          `json:"challenge_id"`
	AWDChallengeID    int64          `json:"awd_challenge_id,omitempty"`
	AWDChallengeTitle string         `json:"awd_challenge_title,omitempty"`
	Category          string         `json:"-"`
	Title             string         `json:"title"`
	Timestamp         time.Time      `json:"timestamp"`
	Detail            string         `json:"detail,omitempty"`
	Meta              map[string]any `json:"meta,omitempty"`
}

type ReviewArchiveWriteupItem struct {
	ID               int64      `json:"id"`
	ChallengeID      int64      `json:"challenge_id"`
	Category         string     `json:"-"`
	ChallengeTitle   string     `json:"challenge_title"`
	Title            string     `json:"title"`
	SubmissionStatus string     `json:"submission_status"`
	VisibilityStatus string     `json:"visibility_status"`
	IsRecommended    bool       `json:"is_recommended"`
	PublishedAt      *time.Time `json:"published_at,omitempty"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type ReviewArchiveManualReviewItem struct {
	ID             int64      `json:"id"`
	ChallengeID    int64      `json:"challenge_id"`
	Category       string     `json:"-"`
	ChallengeTitle string     `json:"challenge_title"`
	Answer         string     `json:"answer"`
	ReviewStatus   string     `json:"review_status"`
	SubmittedAt    time.Time  `json:"submitted_at"`
	ReviewedAt     *time.Time `json:"reviewed_at,omitempty"`
	ReviewComment  string     `json:"review_comment,omitempty"`
	Score          int        `json:"score"`
	ReviewerName   string     `json:"reviewer_name,omitempty"`
}

func ValidateReportAccess(report *model.Report, requesterID int64, role string) error {
	if role == model.RoleAdmin {
		return nil
	}
	if report.UserID == nil || *report.UserID != requesterID {
		return errcode.ErrForbidden
	}
	return nil
}

func FillMissingDimensionAverages(rows []ClassDimensionAverage) []ClassDimensionAverage {
	index := make(map[string]float64, len(rows))
	for _, row := range rows {
		index[row.Dimension] = row.AvgScore
	}

	filled := make([]ClassDimensionAverage, 0, len(model.AllDimensions))
	for _, dimension := range model.AllDimensions {
		filled = append(filled, ClassDimensionAverage{
			Dimension: dimension,
			AvgScore:  index[dimension],
		})
	}
	return filled
}

func NormalizeReportConfig(cfg config.ReportConfig) config.ReportConfig {
	if strings.TrimSpace(cfg.StorageDir) == "" {
		cfg.StorageDir = "storage/exports"
	}
	if cfg.DefaultFormat != model.ReportFormatPDF && cfg.DefaultFormat != model.ReportFormatExcel {
		cfg.DefaultFormat = model.ReportFormatPDF
	}
	if cfg.PersonalTimeout <= 0 {
		cfg.PersonalTimeout = 30 * time.Second
	}
	if cfg.ClassTimeout <= 0 {
		cfg.ClassTimeout = 2 * time.Minute
	}
	if cfg.FileTTL <= 0 {
		cfg.FileTTL = 7 * 24 * time.Hour
	}
	if cfg.MaxWorkers <= 0 {
		cfg.MaxWorkers = 2
	}
	return cfg
}
