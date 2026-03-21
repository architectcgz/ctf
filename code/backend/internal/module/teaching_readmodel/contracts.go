package teaching_readmodel

import (
	"context"

	"ctf-platform/internal/dto"
)

type TeachingQuery interface {
	GetClassSummary(ctx context.Context, className string) (*dto.TeacherClassSummaryResp, error)
}
