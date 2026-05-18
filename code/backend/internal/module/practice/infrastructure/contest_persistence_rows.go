package infrastructure

import (
	"time"

	"gorm.io/gorm"

	practiceports "ctf-platform/internal/module/practice/ports"
)

type contestRow struct {
	ID            int64          `gorm:"column:id;primaryKey"`
	Mode          string         `gorm:"column:mode"`
	EndTime       time.Time      `gorm:"column:end_time"`
	PausedSeconds int64          `gorm:"column:paused_seconds"`
	Status        string         `gorm:"column:status"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (contestRow) TableName() string {
	return "contests"
}

func (r contestRow) toRecord() *practiceports.ContestRecord {
	return &practiceports.ContestRecord{
		ID:            r.ID,
		Mode:          r.Mode,
		EndTime:       r.EndTime,
		PausedSeconds: r.PausedSeconds,
		Status:        r.Status,
	}
}

func contestRowsToRecords(items []*contestRow) []*practiceports.ContestRecord {
	result := make([]*practiceports.ContestRecord, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, item.toRecord())
	}
	return result
}

type contestChallengeProjection struct {
	ContestID   int64 `gorm:"column:contest_id"`
	ChallengeID int64 `gorm:"column:challenge_id"`
	IsVisible   bool  `gorm:"column:is_visible"`
}

func (r contestChallengeProjection) toRecord() *practiceports.ContestChallengeRecord {
	return &practiceports.ContestChallengeRecord{
		ContestID:   r.ContestID,
		ChallengeID: r.ChallengeID,
		IsVisible:   r.IsVisible,
	}
}

type contestAWDServiceRow struct {
	ID              int64          `gorm:"column:id;primaryKey"`
	ContestID       int64          `gorm:"column:contest_id"`
	AWDChallengeID  int64          `gorm:"column:awd_challenge_id"`
	DisplayName     string         `gorm:"column:display_name"`
	ServiceSnapshot string         `gorm:"column:service_snapshot;type:text"`
	ScoreConfig     string         `gorm:"column:score_config;type:text"`
	IsVisible       bool           `gorm:"column:is_visible"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (contestAWDServiceRow) TableName() string {
	return "contest_awd_services"
}

func (r contestAWDServiceRow) toRecord() *practiceports.ContestAWDServiceRecord {
	return &practiceports.ContestAWDServiceRecord{
		ID:              r.ID,
		ContestID:       r.ContestID,
		AWDChallengeID:  r.AWDChallengeID,
		DisplayName:     r.DisplayName,
		ServiceSnapshot: r.ServiceSnapshot,
		ScoreConfig:     r.ScoreConfig,
		IsVisible:       r.IsVisible,
	}
}

func contestAWDServiceRowsToRecords(items []*contestAWDServiceRow) []*practiceports.ContestAWDServiceRecord {
	result := make([]*practiceports.ContestAWDServiceRecord, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, item.toRecord())
	}
	return result
}

type contestTeamRow struct {
	ID        int64          `gorm:"column:id;primaryKey"`
	ContestID int64          `gorm:"column:contest_id"`
	Name      string         `gorm:"column:name"`
	CaptainID int64          `gorm:"column:captain_id"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (contestTeamRow) TableName() string {
	return "teams"
}

func (r contestTeamRow) toRecord() *practiceports.ContestTeamRecord {
	return &practiceports.ContestTeamRecord{
		ID:        r.ID,
		ContestID: r.ContestID,
		Name:      r.Name,
		CaptainID: r.CaptainID,
	}
}

func contestTeamRowsToRecords(items []*contestTeamRow) []*practiceports.ContestTeamRecord {
	result := make([]*practiceports.ContestTeamRecord, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, item.toRecord())
	}
	return result
}

type contestRegistrationProjection struct {
	TeamID *int64 `gorm:"column:team_id"`
	Status string `gorm:"column:status"`
}

type submissionRow struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	UserID        int64      `gorm:"column:user_id"`
	ChallengeID   int64      `gorm:"column:challenge_id"`
	ContestID     *int64     `gorm:"column:contest_id"`
	TeamID        *int64     `gorm:"column:team_id"`
	Flag          string     `gorm:"column:flag"`
	IsCorrect     bool       `gorm:"column:is_correct"`
	ReviewStatus  string     `gorm:"column:review_status"`
	ReviewedBy    *int64     `gorm:"column:reviewed_by"`
	ReviewedAt    *time.Time `gorm:"column:reviewed_at"`
	ReviewComment string     `gorm:"column:review_comment"`
	Score         int        `gorm:"column:score"`
	SubmittedAt   time.Time  `gorm:"column:submitted_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (submissionRow) TableName() string {
	return "submissions"
}

func (r submissionRow) toRecord() *practiceports.SubmissionRecord {
	return &practiceports.SubmissionRecord{
		ID:            r.ID,
		UserID:        r.UserID,
		ChallengeID:   r.ChallengeID,
		ContestID:     r.ContestID,
		TeamID:        r.TeamID,
		Flag:          r.Flag,
		IsCorrect:     r.IsCorrect,
		ReviewStatus:  r.ReviewStatus,
		ReviewedBy:    r.ReviewedBy,
		ReviewedAt:    r.ReviewedAt,
		ReviewComment: r.ReviewComment,
		Score:         r.Score,
		SubmittedAt:   r.SubmittedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

func submissionRowsToRecords(items []submissionRow) []practiceports.SubmissionRecord {
	result := make([]practiceports.SubmissionRecord, 0, len(items))
	for i := range items {
		result = append(result, *items[i].toRecord())
	}
	return result
}

func submissionRowFromRecord(submission *practiceports.SubmissionRecord) *submissionRow {
	if submission == nil {
		return nil
	}
	return &submissionRow{
		ID:            submission.ID,
		UserID:        submission.UserID,
		ChallengeID:   submission.ChallengeID,
		ContestID:     submission.ContestID,
		TeamID:        submission.TeamID,
		Flag:          submission.Flag,
		IsCorrect:     submission.IsCorrect,
		ReviewStatus:  submission.ReviewStatus,
		ReviewedBy:    submission.ReviewedBy,
		ReviewedAt:    submission.ReviewedAt,
		ReviewComment: submission.ReviewComment,
		Score:         submission.Score,
		SubmittedAt:   submission.SubmittedAt,
		UpdatedAt:     submission.UpdatedAt,
	}
}
