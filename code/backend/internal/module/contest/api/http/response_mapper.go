package http

import (
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
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :http
type ContestResponseMapper interface {
	ToStringAnyMap(source map[string]any) map[string]any

	ToAWDCheckerPreviewCommandRespPtr(source *contestdomain.AWDCheckerPreviewResult) *contestcmd.AWDCheckerPreviewResp
	// goverter:ignore LastPreviewResult
	ToContestAWDServiceCommandResp(source contestqry.ContestAWDServiceResult) contestcmd.ContestAWDServiceResp
	ToContestAWDServiceCommandRespPtr(source *contestqry.ContestAWDServiceResult) *contestcmd.ContestAWDServiceResp

	ToContestCommandRespPtr(source *contestqry.ContestResult) *contestcmd.ContestResp
	ToContestCommandResps(source []*contestqry.ContestResult) []*contestcmd.ContestResp
	ToContestListItemRespPtr(source *contestqry.ContestResult) *ContestListItemResp
	ToContestListItemResps(source []*contestqry.ContestResult) []*ContestListItemResp

	ToTeamRespPtr(source *contestqry.TeamResult) *TeamResp
	ToTeamResps(source []*contestqry.TeamResult) []*TeamResp
	ToTeamMemberResps(source []*contestqry.TeamMemberResult) []*TeamMemberResp

	ToMyTeamRespPtr(source *contestqry.MyTeamResult) *MyTeamResp
	ToContestChallengeResps(source []*contestqry.ContestChallengeResult) []*ContestChallengeResp
	ToContestChallengeInfos(source []*contestqry.ContestChallengeInfoResult) []*ContestChallengeInfo

	ToAWDTeamServiceResps(source []contestqry.AWDTeamServiceResult) []*AWDTeamServiceResp
	ToAWDAttackLogResps(source []contestqry.AWDAttackLogResult) []*AWDAttackLogResp
	ToAWDRoundResps(source []contestqry.AWDRoundResult) []*AWDRoundResp

	ToContestAnnouncementResps(source []*contestqry.ContestAnnouncementResult) []*ContestAnnouncementResp
	ToContestMyProgressRespPtr(source *contestqry.ParticipationProgressResult) *ContestMyProgressResp

	ToRegistrationPageRespPtr(source *contestqry.RegistrationPageResult[*contestqry.ContestRegistrationResult]) *PageResult[*ContestRegistrationResp]
	ToAWDReadinessRespPtr(source *contestqry.AWDReadinessResult) *AWDReadinessResp
	ToAWDWorkspaceRespPtr(source *contestqry.AWDWorkspaceResult) *ContestAWDWorkspaceResp

	ToAWDRoundSummaryRespPtr(source *contestqry.AWDRoundSummaryResult) *AWDRoundSummaryResp
	// goverter:ignore RequestID
	ToAWDTrafficEventResp(source contestqry.AWDTrafficEventResult) AWDTrafficEventResp
	ToAWDTrafficEventPageRespPtr(source *contestqry.AWDTrafficEventPageResult) *AWDTrafficEventPageResp
	ToAWDTrafficSummaryRespPtr(source *contestqry.AWDTrafficSummaryResult) *AWDTrafficSummaryResp
	ToScoreboardRespPtr(source *contestqry.ScoreboardResult) *ScoreboardResp
}

var contestResponseMapper ContestResponseMapper
