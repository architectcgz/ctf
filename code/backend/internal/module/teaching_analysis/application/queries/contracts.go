package queries

import (
	"context"
)

type OverviewService interface {
	GetOverview(ctx context.Context, requesterID int64, requesterRole string) (*TeacherOverviewResp, error)
}

type ClassInsightService interface {
	GetClassSummary(ctx context.Context, requesterID int64, requesterRole, className string, query *TeacherClassInsightInput) (*TeacherClassSummary, error)
	GetClassTrend(ctx context.Context, requesterID int64, requesterRole, className string, query *TeacherClassInsightInput) (*TeacherClassTrend, error)
	GetClassReview(ctx context.Context, requesterID int64, requesterRole, className string, query *TeacherClassInsightInput) (*TeacherClassReview, error)
}

type StudentReviewService interface {
	GetStudentProgress(ctx context.Context, requesterID int64, requesterRole string, studentID int64) (*TeacherProgressResp, error)
	GetStudentRecommendations(ctx context.Context, requesterID int64, requesterRole string, studentID int64, limit int) (*TeacherRecommendationResp, error)
	GetStudentTimeline(ctx context.Context, requesterID int64, requesterRole string, studentID int64, limit, offset int) (*TimelineResp, error)
	GetStudentEvidence(ctx context.Context, requesterID int64, requesterRole string, studentID int64, query *TeacherEvidenceInput) (*TeacherEvidenceResp, error)
	GetStudentAttackSessions(ctx context.Context, requesterID int64, requesterRole string, studentID int64, query *TeacherAttackSessionInput) (*TeacherAttackSessionResp, error)
}

type Service interface {
	ListClasses(ctx context.Context, requesterID int64, requesterRole string, query *TeacherClassListInput) ([]TeacherClassItem, int64, int, int, error)
	ListStudents(ctx context.Context, requesterID int64, requesterRole string, query *TeacherStudentDirectoryInput) ([]TeacherStudentItem, int64, int, int, error)
	ListClassStudents(ctx context.Context, requesterID int64, requesterRole, className string, query *TeacherStudentListInput) ([]TeacherStudentItem, error)
}
