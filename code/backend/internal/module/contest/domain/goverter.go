package domain

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter gen .

// goverter:converter
// goverter:output:file ./zz_generated.goverter.go
// goverter:extend NormalizeContestTime NormalizeContestTimePtr ParseAWDCheckerConfig ParseAWDCheckerPreviewResult ParseAWDCheckResult
// goverter:skipCopySameType
type ContestResponseMapper interface {
	// goverter:map StartTime StartTime | NormalizeContestTime
	// goverter:map EndTime EndTime | NormalizeContestTime
	// goverter:map FreezeTime FreezeTime | NormalizeContestTimePtr
	// goverter:map CreatedAt CreatedAt | NormalizeContestTime
	// goverter:map UpdatedAt UpdatedAt | NormalizeContestTime
	ToContestResp(source *model.Contest) *dto.ContestResp
	// goverter:ignore Title
	// goverter:ignore Category
	// goverter:ignore Difficulty
	ToContestChallengeResp(source *model.ContestChallenge) *dto.ContestChallengeResp
	// goverter:ignore MemberCount
	ToTeamResp(source *model.Team) *dto.TeamResp
	ToAWDRoundResp(source *model.AWDRound) *dto.AWDRoundResp
	// goverter:ignore ValidationState
	// goverter:ignore Title
	// goverter:ignore Category
	// goverter:ignore Difficulty
	// goverter:map ScoreConfig ScoreConfig | ParseAWDCheckerConfig
	// goverter:map RuntimeConfig RuntimeConfig | ParseAWDCheckerConfig
	// goverter:map LastPreviewResult LastPreviewResult | ParseAWDCheckerPreviewResult
	ToContestAWDServiceResp(source *model.ContestAWDService) *dto.ContestAWDServiceResp
	// goverter:ignore TeamName
	// goverter:ignore ServiceName
	// goverter:ignore AWDChallengeTitle
	// goverter:map CheckResult CheckResult | ParseAWDCheckResult
	ToAWDTeamServiceResp(source *model.AWDTeamService) *dto.AWDTeamServiceResp
	// goverter:ignore AttackerTeam
	// goverter:ignore VictimTeam
	// goverter:ignore Source
	ToAWDAttackLogResp(source *model.AWDAttackLog) *dto.AWDAttackLogResp
}

var contestResponseMapperInst ContestResponseMapper
