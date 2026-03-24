package domain

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

func ContestRespFromModel(contest *model.Contest) *dto.ContestResp {
	return &dto.ContestResp{
		ID:          contest.ID,
		Title:       contest.Title,
		Description: contest.Description,
		Mode:        contest.Mode,
		StartTime:   contest.StartTime,
		EndTime:     contest.EndTime,
		FreezeTime:  contest.FreezeTime,
		Status:      contest.Status,
		CreatedAt:   contest.CreatedAt,
		UpdatedAt:   contest.UpdatedAt,
	}
}

func ContestChallengeRespFromModel(cc *model.ContestChallenge, challenge *model.Challenge) *dto.ContestChallengeResp {
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

func TeamRespFromModel(team *model.Team, memberCount int) *dto.TeamResp {
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
