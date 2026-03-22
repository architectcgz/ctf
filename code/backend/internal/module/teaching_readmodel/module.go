package teaching_readmodel

import (
	"context"
	"strings"
	"time"

	"ctf-platform/internal/dto"
	readmodelinfra "ctf-platform/internal/module/teaching_readmodel/infrastructure"
	"ctf-platform/pkg/errcode"
)

type Module struct {
	repo *readmodelinfra.Repository
}

func NewModule(repo *readmodelinfra.Repository) *Module {
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
