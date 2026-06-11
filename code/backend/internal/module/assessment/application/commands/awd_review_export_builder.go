package commands

import (
	"context"

	"ctf-platform/internal/apperror"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmentreporting "ctf-platform/internal/module/assessment/application/reporting"
)

type AWDReviewExportBuilder interface {
	BuildArchive(ctx context.Context, requesterID, contestID int64, roundNumber *int) (*assessmentqry.TeacherAWDReviewArchiveResp, error)
}

type awdReviewArchiveReader interface {
	GetContestArchive(ctx context.Context, requesterID, contestID int64, req assessmentqry.GetTeacherAWDReviewArchiveInput) (*assessmentqry.TeacherAWDReviewArchiveResp, error)
}

type teacherAWDReviewExportBuilder struct {
	reader awdReviewArchiveReader
}

func NewAWDReviewExportBuilder(reader awdReviewArchiveReader) AWDReviewExportBuilder {
	return &teacherAWDReviewExportBuilder{reader: reader}
}

func (b *teacherAWDReviewExportBuilder) BuildArchive(ctx context.Context, requesterID, contestID int64, roundNumber *int) (*assessmentqry.TeacherAWDReviewArchiveResp, error) {
	if b == nil || b.reader == nil {
		return nil, apperror.ErrServiceUnavailable.WithMessage("教师 AWD 复盘导出暂不可用")
	}

	archive, err := b.reader.GetContestArchive(ctx, requesterID, contestID, assessmentqry.GetTeacherAWDReviewArchiveInput{
		RoundNumber: roundNumber,
	})
	if err != nil || roundNumber != nil || archive == nil || archive.SelectedRound != nil {
		return archive, err
	}

	focusRound := assessmentreporting.HottestRound(archive.Rounds)
	if focusRound == nil {
		return archive, nil
	}

	return b.reader.GetContestArchive(ctx, requesterID, contestID, assessmentqry.GetTeacherAWDReviewArchiveInput{
		RoundNumber: &focusRound.RoundNumber,
	})
}
