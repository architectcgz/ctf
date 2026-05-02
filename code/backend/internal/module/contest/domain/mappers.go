package domain

import (
	"encoding/json"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

func ContestRespFromModel(contest *model.Contest) *dto.ContestResp {
	return contestResponseMapperInst.ToContestResp(contest)
}

func ContestChallengeRespFromModel(cc *model.ContestChallenge, challenge *model.Challenge) *dto.ContestChallengeResp {
	resp := contestResponseMapperInst.ToContestChallengeResp(cc)
	if challenge != nil {
		resp.Title = challenge.Title
		resp.Category = challenge.Category
		resp.Difficulty = challenge.Difficulty
	}
	return resp
}

func ContestAWDServiceRespFromModel(item *model.ContestAWDService) *dto.ContestAWDServiceResp {
	if item == nil {
		return nil
	}
	resp := contestResponseMapperInst.ToContestAWDServiceResp(item)
	runtimeConfig := sanitizeContestAWDServiceRuntimeConfig(resp.RuntimeConfig)
	snapshot, _ := model.DecodeContestAWDServiceSnapshot(item.ServiceSnapshot)
	resp.Title = snapshot.Name
	resp.Category = snapshot.Category
	resp.Difficulty = snapshot.Difficulty
	resp.RuntimeConfig = runtimeConfig
	resp.ValidationState = NormalizeAWDCheckerValidationState(string(item.ValidationState))
	return resp
}

func parseContestAWDServiceCheckerType(runtimeConfig map[string]any) model.AWDCheckerType {
	if runtimeConfig == nil {
		return ""
	}
	raw, ok := runtimeConfig["checker_type"]
	if !ok {
		return ""
	}
	value, ok := raw.(string)
	if !ok {
		return ""
	}
	return NormalizeAWDCheckerType(value)
}

func parseContestAWDServiceCheckerConfig(runtimeConfig map[string]any) map[string]any {
	if runtimeConfig == nil {
		return map[string]any{}
	}
	if raw, ok := runtimeConfig["checker_config_raw"]; ok {
		if value, ok := raw.(string); ok {
			return ParseAWDCheckerConfig(value)
		}
	}
	if raw, ok := runtimeConfig["checker_config"]; ok {
		encoded, err := json.Marshal(raw)
		if err == nil {
			return ParseAWDCheckerConfig(string(encoded))
		}
	}
	return map[string]any{}
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

func parseContestAWDServiceScore(scoreConfig map[string]any, key string) (int, bool) {
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

func TeamRespFromModel(team *model.Team, memberCount int) *dto.TeamResp {
	resp := contestResponseMapperInst.ToTeamResp(team)
	resp.MemberCount = memberCount
	return resp
}
