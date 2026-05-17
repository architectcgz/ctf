package commands

import (
	"encoding/json"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func contestRespFromModel(contest *model.Contest) *dto.ContestResp {
	resp := contestResponseMapperInst.ToContestRespBasePtr(contestdomain.CloneContestWithEffectiveSchedule(contest))
	if resp == nil {
		return nil
	}
	resp.StartTime = contestdomain.NormalizeContestTime(resp.StartTime)
	resp.EndTime = contestdomain.NormalizeContestTime(resp.EndTime)
	resp.FreezeTime = contestdomain.NormalizeContestTimePtr(resp.FreezeTime)
	resp.CreatedAt = contestdomain.NormalizeContestTime(resp.CreatedAt)
	resp.UpdatedAt = contestdomain.NormalizeContestTime(resp.UpdatedAt)
	return resp
}

func contestChallengeRespFromModel(cc *model.ContestChallenge, challenge *model.Challenge) *dto.ContestChallengeResp {
	resp := contestResponseMapperInst.ToContestChallengeRespBasePtr(cc)
	if resp == nil {
		return nil
	}
	if challenge != nil {
		resp.Title = challenge.Title
		resp.Category = challenge.Category
		resp.Difficulty = challenge.Difficulty
	}
	return resp
}

func contestAWDServiceRespFromModel(item *model.ContestAWDService) *dto.ContestAWDServiceResp {
	resp := contestResponseMapperInst.ToContestAWDServiceRespBasePtr(item)
	if resp == nil {
		return nil
	}
	runtimeConfig := sanitizeContestAWDServiceRuntimeConfig(contestdomain.ParseAWDCheckerConfig(item.RuntimeConfig))
	snapshot, _ := model.DecodeContestAWDServiceSnapshot(item.ServiceSnapshot)
	resp.Title = snapshot.Name
	resp.Category = snapshot.Category
	resp.Difficulty = snapshot.Difficulty
	resp.ScoreConfig = contestdomain.ParseAWDCheckerConfig(item.ScoreConfig)
	resp.RuntimeConfig = runtimeConfig
	resp.ValidationState = contestdomain.NormalizeAWDCheckerValidationState(string(item.ValidationState))
	resp.LastPreviewResult = awdPreviewResultMapper.ToDTOPtr(contestdomain.ParseAWDCheckerPreviewResult(item.LastPreviewResult))
	return resp
}

func teamRespFromModel(team *model.Team, memberCount int) *dto.TeamResp {
	resp := contestResponseMapperInst.ToTeamRespBasePtr(team)
	if resp == nil {
		return nil
	}
	resp.MemberCount = memberCount
	return resp
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
