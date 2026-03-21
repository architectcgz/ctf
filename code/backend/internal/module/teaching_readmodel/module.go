package teaching_readmodel

import (
	"context"
	"strings"
	"time"

	"ctf-platform/internal/dto"
	teacherModule "ctf-platform/internal/module/teacher"
	"ctf-platform/pkg/errcode"
)

type Module struct {
	repo *teacherModule.Repository
}

func NewModule(repo *teacherModule.Repository) *Module {
	return &Module{repo: repo}
}

func (m *Module) GetClassSummary(ctx context.Context, className string) (*dto.TeacherClassSummaryResp, error) {
	normalized := strings.TrimSpace(className)
	if normalized == "" {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "class_name 不能为空", errcode.ErrInvalidParams.HTTPStatus)
	}

	summary, err := m.repo.GetClassSummary(ctx, normalized, time.Now().AddDate(0, 0, -7))
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return summary, nil
}
