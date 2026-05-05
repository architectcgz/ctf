package commands

import (
	"context"

	"ctf-platform/internal/dto"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	"ctf-platform/pkg/errcode"
)

type AWDReviewExportBuilder interface {
	BuildArchive(ctx context.Context, requesterID, contestID int64, roundNumber *int) (*dto.TeacherAWDReviewArchiveResp, error)
}

type awdReviewArchiveReader interface {
	GetContestArchive(ctx context.Context, requesterID, contestID int64, req assessmentqry.GetTeacherAWDReviewArchiveInput) (*dto.TeacherAWDReviewArchiveResp, error)
}

type teacherAWDReviewExportBuilder struct {
	reader awdReviewArchiveReader
}

func NewAWDReviewExportBuilder(reader awdReviewArchiveReader) AWDReviewExportBuilder {
	return &teacherAWDReviewExportBuilder{reader: reader}
}

func (b *teacherAWDReviewExportBuilder) BuildArchive(ctx context.Context, requesterID, contestID int64, roundNumber *int) (*dto.TeacherAWDReviewArchiveResp, error) {
	if b == nil || b.reader == nil {
		return nil, errcode.New(errcode.ErrServiceUnavailable.Code, "教师 AWD 复盘导出暂不可用", errcode.ErrServiceUnavailable.HTTPStatus)
	}

	return b.reader.GetContestArchive(ctx, requesterID, contestID, assessmentqry.GetTeacherAWDReviewArchiveInput{
		RoundNumber: roundNumber,
	})
}
