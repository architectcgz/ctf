package commands

import (
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
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
	ToContestRespBase(source model.Contest) ContestResp
	ToContestRespBasePtr(source *model.Contest) *ContestResp
	ToContestAnnouncementRespBase(source model.ContestAnnouncement) ContestAnnouncementResp
	ToContestAnnouncementRespBasePtr(source *model.ContestAnnouncement) *ContestAnnouncementResp

	// goverter:ignore Title
	// goverter:ignore Category
	// goverter:ignore Difficulty
	ToContestChallengeRespBase(source model.ContestChallenge) ContestChallengeResp
	ToContestChallengeRespBasePtr(source *model.ContestChallenge) *ContestChallengeResp

	// goverter:ignore Title
	// goverter:ignore Category
	// goverter:ignore Difficulty
	// goverter:ignore ScoreConfig
	// goverter:ignore RuntimeConfig
	// goverter:ignore ValidationState
	// goverter:ignore LastPreviewResult
	ToContestAWDServiceRespBase(source model.ContestAWDService) dto.ContestAWDServiceResp
	ToContestAWDServiceRespBasePtr(source *model.ContestAWDService) *dto.ContestAWDServiceResp

	// goverter:ignore MemberCount
	ToTeamRespBase(source model.Team) TeamResp
	ToTeamRespBasePtr(source *model.Team) *TeamResp

	ToAWDRoundRespBase(source model.AWDRound) dto.AWDRoundResp
	ToAWDRoundRespBasePtr(source *model.AWDRound) *dto.AWDRoundResp

	// goverter:ignore TeamName
	// goverter:ignore ServiceName
	// goverter:ignore AWDChallengeTitle
	// goverter:ignore CheckResult
	ToAWDTeamServiceRespBase(source model.AWDTeamService) dto.AWDTeamServiceResp
	ToAWDTeamServiceRespBasePtr(source *model.AWDTeamService) *dto.AWDTeamServiceResp

	// goverter:ignore AttackerTeam
	// goverter:ignore VictimTeam
	// goverter:ignore Source
	ToAWDAttackLogRespBase(source model.AWDAttackLog) AWDAttackLogResp
	ToAWDAttackLogRespBasePtr(source *model.AWDAttackLog) *AWDAttackLogResp

	// goverter:ignore Username
	ToContestRegistrationRespBase(source model.ContestRegistration) ContestRegistrationResp
	ToContestRegistrationRespBasePtr(source *model.ContestRegistration) *ContestRegistrationResp

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
