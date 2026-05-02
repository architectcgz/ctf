package commands

import (
	"encoding/json"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func contestRespFromModel(contest *model.Contest) *dto.ContestResp {
	mapped := contestResponseMapperInst.ToContestRespBase(*contest)
	mapped.StartTime = contestdomain.NormalizeContestTime(mapped.StartTime)
	mapped.EndTime = contestdomain.NormalizeContestTime(mapped.EndTime)
	mapped.FreezeTime = contestdomain.NormalizeContestTimePtr(mapped.FreezeTime)
	mapped.CreatedAt = contestdomain.NormalizeContestTime(mapped.CreatedAt)
	mapped.UpdatedAt = contestdomain.NormalizeContestTime(mapped.UpdatedAt)
	return &mapped
}

func contestChallengeRespFromModel(cc *model.ContestChallenge, challenge *model.Challenge) *dto.ContestChallengeResp {
	mapped := contestResponseMapperInst.ToContestChallengeRespBase(*cc)
	resp := &mapped
	if challenge != nil {
		resp.Title = challenge.Title
		resp.Category = challenge.Category
		resp.Difficulty = challenge.Difficulty
	}
	return resp
}

func contestAWDServiceRespFromModel(item *model.ContestAWDService) *dto.ContestAWDServiceResp {
	if item == nil {
		return nil
	}
	mapped := contestResponseMapperInst.ToContestAWDServiceRespBase(*item)
	runtimeConfig := sanitizeContestAWDServiceRuntimeConfig(contestdomain.ParseAWDCheckerConfig(item.RuntimeConfig))
	snapshot, _ := model.DecodeContestAWDServiceSnapshot(item.ServiceSnapshot)
	mapped.Title = snapshot.Name
	mapped.Category = snapshot.Category
	mapped.Difficulty = snapshot.Difficulty
	mapped.ScoreConfig = contestdomain.ParseAWDCheckerConfig(item.ScoreConfig)
	mapped.RuntimeConfig = runtimeConfig
	mapped.ValidationState = contestdomain.NormalizeAWDCheckerValidationState(string(item.ValidationState))
	mapped.LastPreviewResult = awdPreviewResultMapper.ToDTOPtr(contestdomain.ParseAWDCheckerPreviewResult(item.LastPreviewResult))
	return &mapped
}

func teamRespFromModel(team *model.Team, memberCount int) *dto.TeamResp {
	mapped := contestResponseMapperInst.ToTeamRespBase(*team)
	mapped.MemberCount = memberCount
	return &mapped
}

func sanitizeContestAWDServiceRuntimeConfig(runtimeConfig map[string]any) map[string]any {
	if len(runtimeConfig) == 0 {
		return runtimeConfig
	}
	sanitized := make(map[string]any, len(runtimeConfig))
	for key, value := range runtimeConfig {
		if key == "challenge_id" {
			continue
		}
		sanitized[key] = value
	}
	return sanitized
}

func parseContestAWDServiceScoreValue(scoreConfig map[string]any, key string) (int, bool) {
	if scoreConfig == nil {
		return 0, false
	}
	raw, ok := scoreConfig[key]
	if !ok {
		return 0, false
	}
	switch value := raw.(type) {
	case int:
		return value, true
	case int32:
		return int(value), true
	case int64:
		return int(value), true
	case float64:
		return int(value), true
	case json.Number:
		next, err := value.Int64()
		if err != nil {
			return 0, false
		}
		return int(next), true
	default:
		return 0, false
	}
}
