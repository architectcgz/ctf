package commands

import (
	"time"

	contestentity "ctf-platform/internal/module/contest/entity"
)

type submissionRespSource struct {
	IsCorrect   bool
	Status      string
	Points      int
	SubmittedAt time.Time
}

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :commands
type contestResponseMapper interface {
	ToContestRespBase(source contestentity.Contest) ContestResp
	ToContestRespBasePtr(source *contestentity.Contest) *ContestResp
	ToContestAnnouncementRespBase(source contestentity.ContestAnnouncement) ContestAnnouncementResp
	ToContestAnnouncementRespBasePtr(source *contestentity.ContestAnnouncement) *ContestAnnouncementResp

	// goverter:ignore Title
	// goverter:ignore Category
	// goverter:ignore Difficulty
	ToContestChallengeRespBase(source contestentity.ContestChallenge) ContestChallengeResp
	ToContestChallengeRespBasePtr(source *contestentity.ContestChallenge) *ContestChallengeResp

	// goverter:ignore Title
	// goverter:ignore Category
	// goverter:ignore Difficulty
	// goverter:ignore ScoreConfig
	// goverter:ignore RuntimeConfig
	// goverter:ignore ValidationState
	// goverter:ignore LastPreviewResult
	ToContestAWDServiceRespBase(source contestentity.ContestAWDService) ContestAWDServiceResp
	ToContestAWDServiceRespBasePtr(source *contestentity.ContestAWDService) *ContestAWDServiceResp

	// goverter:ignore MemberCount
	ToTeamRespBase(source contestentity.Team) TeamResp
	ToTeamRespBasePtr(source *contestentity.Team) *TeamResp

	ToAWDRoundRespBase(source contestentity.AWDRound) AWDRoundResp
	ToAWDRoundRespBasePtr(source *contestentity.AWDRound) *AWDRoundResp

	// goverter:ignore TeamName
	// goverter:ignore ServiceName
	// goverter:ignore AWDChallengeTitle
	// goverter:ignore CheckResult
	ToAWDTeamServiceRespBase(source contestentity.AWDTeamService) AWDTeamServiceResp
	ToAWDTeamServiceRespBasePtr(source *contestentity.AWDTeamService) *AWDTeamServiceResp

	// goverter:ignore AttackerTeam
	// goverter:ignore VictimTeam
	// goverter:ignore Source
	ToAWDAttackLogRespBase(source contestentity.AWDAttackLog) AWDAttackLogResp
	ToAWDAttackLogRespBasePtr(source *contestentity.AWDAttackLog) *AWDAttackLogResp

	// goverter:ignore Username
	ToContestRegistrationRespBase(source contestentity.ContestRegistration) ContestRegistrationResp
	ToContestRegistrationRespBasePtr(source *contestentity.ContestRegistration) *ContestRegistrationResp

	// goverter:ignore Message
	// goverter:ignore InstanceShutdownAt
	ToSubmissionResp(source submissionRespSource) SubmissionResp
	ToSubmissionRespPtr(source submissionRespSource) *SubmissionResp
}

var contestResponseMapperInst contestResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}

func CopyTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
