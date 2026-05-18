package http

import (
	"time"

	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestqry "ctf-platform/internal/module/contest/application/queries"
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

	ToCreateContestInput(source CreateContestReq) contestcmd.CreateContestInput
	ToUpdateContestInput(source UpdateContestReq) contestcmd.UpdateContestInput
	ToCreateAnnouncementInput(source CreateContestAnnouncementReq) contestcmd.CreateAnnouncementInput
	ToReviewRegistrationInput(source ReviewContestRegistrationReq) contestcmd.ReviewRegistrationInput
	ToCreateTeamInput(source CreateTeamReq) contestcmd.CreateTeamInput
	ToAddContestChallengeInput(source AddContestChallengeReq) contestcmd.AddContestChallengeInput
	ToUpdateContestChallengeInput(source UpdateContestChallengeReq) contestcmd.UpdateContestChallengeInput
	ToCreateAWDRoundInput(source CreateAWDRoundReq) contestcmd.CreateAWDRoundInput
	ToUpsertServiceCheckInput(source UpsertAWDServiceCheckReq) contestcmd.UpsertServiceCheckInput
	ToRunCurrentRoundChecksInput(source RunCurrentAWDCheckerReq) contestcmd.RunCurrentRoundChecksInput
	ToCreateAttackLogInput(source CreateAWDAttackLogReq) contestcmd.CreateAttackLogInput
	ToSubmitAttackInput(source SubmitAWDAttackReq) contestcmd.SubmitAttackInput
	ToPreviewCheckerInput(source PreviewAWDCheckerReq) contestcmd.PreviewCheckerInput
	ToCreateContestAWDServiceInput(source CreateContestAWDServiceReq) contestcmd.CreateContestAWDServiceInput
	ToUpdateContestAWDServiceInput(source UpdateContestAWDServiceReq) contestcmd.UpdateContestAWDServiceInput
	ToListAWDTrafficEventsInput(source ListAWDTrafficEventsReq) contestqry.ListAWDTrafficEventsInput
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
