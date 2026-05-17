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

	ToAWDCheckerPreviewCommandRespPtr(source *contestdomain.AWDCheckerPreviewResult) *contestcmd.AWDCheckerPreviewResp
	// goverter:ignore LastPreviewResult
	ToContestAWDServiceCommandResp(source contestqry.ContestAWDServiceResult) contestcmd.ContestAWDServiceResp
	ToContestAWDServiceCommandRespPtr(source *contestqry.ContestAWDServiceResult) *contestcmd.ContestAWDServiceResp

	ToContestCommandRespPtr(source *contestqry.ContestResult) *contestcmd.ContestResp
	ToContestCommandResps(source []*contestqry.ContestResult) []*contestcmd.ContestResp

	ToTeamRespPtr(source *contestqry.TeamResult) *dto.TeamResp
	ToTeamResps(source []*contestqry.TeamResult) []*dto.TeamResp
	ToTeamMemberResps(source []*contestqry.TeamMemberResult) []*dto.TeamMemberResp

	ToMyTeamRespPtr(source *contestqry.MyTeamResult) *dto.MyTeamResp
	ToContestChallengeResps(source []*contestqry.ContestChallengeResult) []*dto.ContestChallengeResp
	ToContestChallengeInfos(source []*contestqry.ContestChallengeInfoResult) []*dto.ContestChallengeInfo

	ToAWDTeamServiceResps(source []contestqry.AWDTeamServiceResult) []*dto.AWDTeamServiceResp
	ToAWDAttackLogResps(source []contestqry.AWDAttackLogResult) []*dto.AWDAttackLogResp
	ToAWDRoundResps(source []contestqry.AWDRoundResult) []*dto.AWDRoundResp

	ToContestAnnouncementResps(source []*contestqry.ContestAnnouncementResult) []*dto.ContestAnnouncementResp
	ToContestMyProgressRespPtr(source *contestqry.ParticipationProgressResult) *dto.ContestMyProgressResp

	ToRegistrationPageRespPtr(source *contestqry.RegistrationPageResult[*contestqry.ContestRegistrationResult]) *dto.PageResult[*dto.ContestRegistrationResp]
	ToAWDReadinessRespPtr(source *contestqry.AWDReadinessResult) *dto.AWDReadinessResp
	ToAWDWorkspaceRespPtr(source *contestqry.AWDWorkspaceResult) *dto.ContestAWDWorkspaceResp

	ToAWDRoundSummaryRespPtr(source *contestqry.AWDRoundSummaryResult) *dto.AWDRoundSummaryResp
	// goverter:ignore RequestID
	ToAWDTrafficEventResp(source contestqry.AWDTrafficEventResult) dto.AWDTrafficEventResp
	ToAWDTrafficEventPageRespPtr(source *contestqry.AWDTrafficEventPageResult) *dto.AWDTrafficEventPageResp
	ToAWDTrafficSummaryRespPtr(source *contestqry.AWDTrafficSummaryResult) *dto.AWDTrafficSummaryResp
	ToScoreboardRespPtr(source *contestqry.ScoreboardResult) *dto.ScoreboardResp
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
