package commands

import (
	"encoding/json"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func contestRespFromModel(contest *model.Contest) *dto.ContestResp {
	return &dto.ContestResp{
		ID:          contest.ID,
		Title:       contest.Title,
		Description: contest.Description,
		Mode:        contest.Mode,
		StartTime:   contestdomain.NormalizeContestTime(contest.StartTime),
		EndTime:     contestdomain.NormalizeContestTime(contest.EndTime),
		FreezeTime:  contestdomain.NormalizeContestTimePtr(contest.FreezeTime),
		Status:      contest.Status,
		CreatedAt:   contestdomain.NormalizeContestTime(contest.CreatedAt),
		UpdatedAt:   contestdomain.NormalizeContestTime(contest.UpdatedAt),
	}
}

func contestChallengeRespFromModel(cc *model.ContestChallenge, challenge *model.Challenge) *dto.ContestChallengeResp {
	resp := &dto.ContestChallengeResp{
		ID:          cc.ID,
		ContestID:   cc.ContestID,
		ChallengeID: cc.ChallengeID,
		Points:      cc.Points,
		Order:       cc.Order,
		IsVisible:   cc.IsVisible,
		CreatedAt:   cc.CreatedAt,
	}
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
	runtimeConfig := sanitizeContestAWDServiceRuntimeConfig(contestdomain.ParseAWDCheckerConfig(item.RuntimeConfig))
	snapshot, _ := model.DecodeContestAWDServiceSnapshot(item.ServiceSnapshot)
	return &dto.ContestAWDServiceResp{
		ID:                item.ID,
		ContestID:         item.ContestID,
		AWDChallengeID:    item.AWDChallengeID,
		Title:             snapshot.Name,
		Category:          snapshot.Category,
		Difficulty:        snapshot.Difficulty,
		DisplayName:       item.DisplayName,
		Order:             item.Order,
		IsVisible:         item.IsVisible,
		ScoreConfig:       contestdomain.ParseAWDCheckerConfig(item.ScoreConfig),
		RuntimeConfig:     runtimeConfig,
		ValidationState:   contestdomain.NormalizeAWDCheckerValidationState(string(item.ValidationState)),
		LastPreviewAt:     item.LastPreviewAt,
		LastPreviewResult: awdCheckerPreviewResultToDTO(contestdomain.ParseAWDCheckerPreviewResult(item.LastPreviewResult)),
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}
}

func teamRespFromModel(team *model.Team, memberCount int) *dto.TeamResp {
	return &dto.TeamResp{
		ID:          team.ID,
		ContestID:   team.ContestID,
		Name:        team.Name,
		CaptainID:   team.CaptainID,
		InviteCode:  team.InviteCode,
		MaxMembers:  team.MaxMembers,
		MemberCount: memberCount,
		CreatedAt:   team.CreatedAt,
	}
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
