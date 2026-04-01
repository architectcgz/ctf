package queries

import (
	"context"

	"ctf-platform/internal/dto"
)

type Service interface {
	ListClasses(ctx context.Context, requesterID int64, requesterRole string) ([]dto.TeacherClassItem, error)
	ListClassStudents(ctx context.Context, requesterID int64, requesterRole, className string, query *dto.TeacherStudentQuery) ([]dto.TeacherStudentItem, error)
	GetClassSummary(ctx context.Context, requesterID int64, requesterRole, className string) (*dto.TeacherClassSummaryResp, error)
	GetClassTrend(ctx context.Context, requesterID int64, requesterRole, className string) (*dto.TeacherClassTrendResp, error)
	GetClassReview(ctx context.Context, requesterID int64, requesterRole, className string) (*dto.TeacherClassReviewResp, error)
	GetStudentProgress(ctx context.Context, requesterID int64, requesterRole string, studentID int64) (*dto.TeacherProgressResp, error)
	GetStudentRecommendations(ctx context.Context, requesterID int64, requesterRole string, studentID int64, limit int) ([]dto.TeacherRecommendationItem, error)
	GetStudentTimeline(ctx context.Context, requesterID int64, requesterRole string, studentID int64, limit, offset int) (*dto.TimelineResp, error)
	GetStudentEvidence(ctx context.Context, requesterID int64, requesterRole string, studentID int64, query *dto.TeacherEvidenceQuery) (*dto.TeacherEvidenceResp, error)
}
