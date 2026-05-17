package queries

import (
	"context"

	"ctf-platform/internal/dto"
)

type OverviewService interface {
	GetOverview(ctx context.Context, requesterID int64, requesterRole string) (*dto.TeacherOverviewResp, error)
}

type ClassInsightService interface {
	GetClassSummary(ctx context.Context, requesterID int64, requesterRole, className string, query *TeacherClassInsightInput) (*dto.TeacherClassSummaryResp, error)
	GetClassTrend(ctx context.Context, requesterID int64, requesterRole, className string, query *TeacherClassInsightInput) (*dto.TeacherClassTrendResp, error)
	GetClassReview(ctx context.Context, requesterID int64, requesterRole, className string, query *TeacherClassInsightInput) (*dto.TeacherClassReviewResp, error)
}

type StudentReviewService interface {
	GetStudentProgress(ctx context.Context, requesterID int64, requesterRole string, studentID int64) (*dto.TeacherProgressResp, error)
	GetStudentRecommendations(ctx context.Context, requesterID int64, requesterRole string, studentID int64, limit int) (*dto.TeacherRecommendationResp, error)
	GetStudentTimeline(ctx context.Context, requesterID int64, requesterRole string, studentID int64, limit, offset int) (*TimelineResp, error)
	GetStudentEvidence(ctx context.Context, requesterID int64, requesterRole string, studentID int64, query *TeacherEvidenceInput) (*dto.TeacherEvidenceResp, error)
	GetStudentAttackSessions(ctx context.Context, requesterID int64, requesterRole string, studentID int64, query *TeacherAttackSessionInput) (*dto.TeacherAttackSessionResp, error)
}

type Service interface {
	ListClasses(ctx context.Context, requesterID int64, requesterRole string, query *TeacherClassListInput) ([]dto.TeacherClassItem, int64, int, int, error)
	ListStudents(ctx context.Context, requesterID int64, requesterRole string, query *TeacherStudentDirectoryInput) ([]dto.TeacherStudentItem, int64, int, int, error)
	ListClassStudents(ctx context.Context, requesterID int64, requesterRole, className string, query *TeacherStudentListInput) ([]dto.TeacherStudentItem, error)
}
