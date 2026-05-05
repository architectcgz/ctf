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
	ToAWDCheckerPreviewRespPtr(source *contestdomain.AWDCheckerPreviewResult) *dto.AWDCheckerPreviewResp
	// goverter:ignore LastPreviewResult
	ToContestAWDServiceResp(source contestqry.ContestAWDServiceResult) dto.ContestAWDServiceResp
	ToContestAWDServiceRespPtr(source *contestqry.ContestAWDServiceResult) *dto.ContestAWDServiceResp
	ToContestResp(source contestqry.ContestResult) dto.ContestResp
	ToContestRespPtr(source *contestqry.ContestResult) *dto.ContestResp
	ToContestResps(source []*contestqry.ContestResult) []*dto.ContestResp
	ToTeamResp(source contestqry.TeamResult) dto.TeamResp
	ToTeamRespPtr(source *contestqry.TeamResult) *dto.TeamResp
	ToTeamResps(source []*contestqry.TeamResult) []*dto.TeamResp
	ToTeamMemberResp(source contestqry.TeamMemberResult) dto.TeamMemberResp
	ToTeamMemberResps(source []*contestqry.TeamMemberResult) []*dto.TeamMemberResp
	ToMyTeamResp(source contestqry.MyTeamResult) dto.MyTeamResp
	ToMyTeamRespPtr(source *contestqry.MyTeamResult) *dto.MyTeamResp
	ToContestChallengeResp(source contestqry.ContestChallengeResult) dto.ContestChallengeResp
	ToContestChallengeResps(source []*contestqry.ContestChallengeResult) []*dto.ContestChallengeResp
	ToContestChallengeInfo(source contestqry.ContestChallengeInfoResult) dto.ContestChallengeInfo
	ToContestChallengeInfos(source []*contestqry.ContestChallengeInfoResult) []*dto.ContestChallengeInfo
	ToAWDTeamServiceResp(source contestqry.AWDTeamServiceResult) dto.AWDTeamServiceResp
	ToAWDTeamServiceResps(source []contestqry.AWDTeamServiceResult) []*dto.AWDTeamServiceResp
	ToAWDAttackLogResp(source contestqry.AWDAttackLogResult) dto.AWDAttackLogResp
	ToAWDAttackLogResps(source []contestqry.AWDAttackLogResult) []*dto.AWDAttackLogResp
	ToAWDRoundResp(source contestqry.AWDRoundResult) dto.AWDRoundResp
	ToAWDRoundResps(source []contestqry.AWDRoundResult) []*dto.AWDRoundResp
	ToAWDRoundSummaryResp(source contestqry.AWDRoundSummaryResult) dto.AWDRoundSummaryResp
	ToContestAnnouncementResp(source contestqry.ContestAnnouncementResult) dto.ContestAnnouncementResp
	ToContestAnnouncementResps(source []*contestqry.ContestAnnouncementResult) []*dto.ContestAnnouncementResp
	ToContestSolvedProgressItem(source contestqry.ContestSolvedProgressResult) dto.ContestSolvedProgressItem
	ToContestMyProgressResp(source contestqry.ParticipationProgressResult) dto.ContestMyProgressResp
	ToContestMyProgressRespPtr(source *contestqry.ParticipationProgressResult) *dto.ContestMyProgressResp
	ToContestRegistrationResp(source contestqry.ContestRegistrationResult) dto.ContestRegistrationResp
	ToContestRegistrationResps(source []*contestqry.ContestRegistrationResult) []*dto.ContestRegistrationResp
	ToRegistrationPageResp(source contestqry.RegistrationPageResult[*contestqry.ContestRegistrationResult]) dto.PageResult[*dto.ContestRegistrationResp]
	ToRegistrationPageRespPtr(source *contestqry.RegistrationPageResult[*contestqry.ContestRegistrationResult]) *dto.PageResult[*dto.ContestRegistrationResp]
	ToAWDReadinessItemResp(source contestqry.AWDReadinessItem) dto.AWDReadinessItemResp
	ToAWDReadinessResp(source contestqry.AWDReadinessResult) dto.AWDReadinessResp
	ToAWDReadinessRespPtr(source *contestqry.AWDReadinessResult) *dto.AWDReadinessResp
	ToAWDWorkspaceResp(source contestqry.AWDWorkspaceResult) dto.ContestAWDWorkspaceResp
	ToAWDWorkspaceRespPtr(source *contestqry.AWDWorkspaceResult) *dto.ContestAWDWorkspaceResp
	ToAWDRoundSummaryRespPtr(source *contestqry.AWDRoundSummaryResult) *dto.AWDRoundSummaryResp
	// goverter:ignore RequestID
	ToAWDTrafficEventResp(source contestqry.AWDTrafficEventResult) dto.AWDTrafficEventResp
	ToAWDTrafficEventPageResp(source contestqry.AWDTrafficEventPageResult) dto.AWDTrafficEventPageResp
	ToAWDTrafficEventPageRespPtr(source *contestqry.AWDTrafficEventPageResult) *dto.AWDTrafficEventPageResp
	ToAWDTrafficTrendBucketResp(source contestqry.AWDTrafficTrendBucketResult) dto.AWDTrafficTrendBucketResp
	ToAWDTrafficTopTeamResp(source contestqry.AWDTrafficTopTeamResult) dto.AWDTrafficTopTeamResp
	ToAWDTrafficTopChallengeResp(source contestqry.AWDTrafficTopChallengeResult) dto.AWDTrafficTopChallengeResp
	ToAWDTrafficTopPathResp(source contestqry.AWDTrafficTopPathResult) dto.AWDTrafficTopPathResp
	ToAWDTrafficSummaryResp(source contestqry.AWDTrafficSummaryResult) dto.AWDTrafficSummaryResp
	ToAWDTrafficSummaryRespPtr(source *contestqry.AWDTrafficSummaryResult) *dto.AWDTrafficSummaryResp
	ToScoreboardContestInfo(source contestqry.ScoreboardContestResult) dto.ScoreboardContestInfo
	ToScoreboardItem(source contestqry.ScoreboardItemResult) dto.ScoreboardItem
	ToScoreboardResp(source contestqry.ScoreboardResult) dto.ScoreboardResp
	ToScoreboardRespPtr(source *contestqry.ScoreboardResult) *dto.ScoreboardResp
	ToAWDWorkspaceTeamResp(source contestqry.AWDWorkspaceTeamResult) dto.ContestAWDWorkspaceTeamResp
	ToAWDWorkspaceServiceResp(source contestqry.AWDWorkspaceServiceResult) dto.ContestAWDWorkspaceServiceResp
	ToAWDWorkspaceTargetTeamResp(source contestqry.AWDWorkspaceTargetTeamResult) dto.ContestAWDWorkspaceTargetTeamResp
	ToAWDWorkspaceTargetServiceResp(source contestqry.AWDWorkspaceTargetServiceResult) dto.ContestAWDWorkspaceTargetServiceResp
	ToAWDWorkspaceRecentEventResp(source contestqry.AWDWorkspaceRecentEventResult) dto.ContestAWDWorkspaceRecentEventResp
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
