package http

import (
	"time"

	"ctf-platform/internal/dto"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:extend ConvertAny
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./request_mapper_gen.go
// goverter:output:package :http
type ContestRequestMapper interface {
	ToStringAnyMap(source map[string]any) map[string]any
	ToCreateContestInput(source dto.CreateContestReq) contestcmd.CreateContestInput
	ToUpdateContestInput(source dto.UpdateContestReq) contestcmd.UpdateContestInput
	ToCreateAnnouncementInput(source dto.CreateContestAnnouncementReq) contestcmd.CreateAnnouncementInput
	ToReviewRegistrationInput(source dto.ReviewContestRegistrationReq) contestcmd.ReviewRegistrationInput
	ToCreateTeamInput(source dto.CreateTeamReq) contestcmd.CreateTeamInput
	ToAddContestChallengeInput(source dto.AddContestChallengeReq) contestcmd.AddContestChallengeInput
	ToUpdateContestChallengeInput(source dto.UpdateContestChallengeReq) contestcmd.UpdateContestChallengeInput
	ToCreateAWDRoundInput(source dto.CreateAWDRoundReq) contestcmd.CreateAWDRoundInput
	ToUpsertServiceCheckInput(source dto.UpsertAWDServiceCheckReq) contestcmd.UpsertServiceCheckInput
	ToRunCurrentRoundChecksInput(source dto.RunCurrentAWDCheckerReq) contestcmd.RunCurrentRoundChecksInput
	ToCreateAttackLogInput(source dto.CreateAWDAttackLogReq) contestcmd.CreateAttackLogInput
	ToSubmitAttackInput(source dto.SubmitAWDAttackReq) contestcmd.SubmitAttackInput
	ToPreviewCheckerInput(source dto.PreviewAWDCheckerReq) contestcmd.PreviewCheckerInput
	ToCreateContestAWDServiceInput(source dto.CreateContestAWDServiceReq) contestcmd.CreateContestAWDServiceInput
	ToUpdateContestAWDServiceInput(source dto.UpdateContestAWDServiceReq) contestcmd.UpdateContestAWDServiceInput
	ToListAWDTrafficEventsInput(source dto.ListAWDTrafficEventsReq) contestqry.ListAWDTrafficEventsInput
	ToAWDCheckerPreviewResp(source contestdomain.AWDCheckerPreviewResult) dto.AWDCheckerPreviewResp
}

var contestRequestMapper ContestRequestMapper

func ConvertAny(source any) any {
	return source
}

func CopyTime(source time.Time) time.Time {
	return source
}

func CopyTimePtr(source *time.Time) *time.Time {
	if source == nil {
		return nil
	}
	value := *source
	return &value
}
