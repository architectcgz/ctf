package queries

import "time"

type TeacherClassListInput struct {
	Page int
	Size int
}

type TeacherClassInsightInput struct {
	FromDate string
	ToDate   string
}

type TeacherStudentListInput struct {
	Keyword   string
	StudentNo string
}

type TeacherStudentDirectoryInput struct {
	ClassName string
	Keyword   string
	StudentNo string
	Page      int
	Size      int
	SortKey   string
	SortOrder string
}

type TeacherEvidenceInput struct {
	ChallengeID *int64
	ContestID   *int64
	RoundID     *int64
	EventType   string
	From        *time.Time
	To          *time.Time
	Limit       int
	Offset      int
}

type TeacherAttackSessionInput struct {
	Mode        string
	ChallengeID *int64
	ContestID   *int64
	RoundID     *int64
	Result      string
	WithEvents  *bool
	Limit       int
	Offset      int
}
