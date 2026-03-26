package queries

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type TeamService struct {
	teamRepo    contestports.ContestTeamRepository
	contestRepo contestports.ContestLookupRepository
}

func NewTeamService(teamRepo contestports.ContestTeamRepository, contestRepo contestports.ContestLookupRepository) *TeamService {
	return &TeamService{
		teamRepo:    teamRepo,
		contestRepo: contestRepo,
	}
}

func (s *TeamService) GetTeamInfo(teamID int64) (*dto.TeamResp, []*dto.TeamMemberResp, error) {
	team, err := s.teamRepo.FindByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errcode.ErrTeamNotFound
		}
		return nil, nil, err
	}

	members, err := s.teamRepo.GetMembers(teamID)
	if err != nil {
		return nil, nil, err
	}

	userIDs := make([]int64, len(members))
	for i, member := range members {
		userIDs[i] = member.UserID
	}

	users, err := s.teamRepo.FindUsersByIDs(userIDs)
	if err != nil {
		return nil, nil, errcode.ErrInternal.WithCause(err)
	}

	userMap := make(map[int64]*model.User, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}

	memberResps := make([]*dto.TeamMemberResp, 0, len(members))
	for _, member := range members {
		if user, ok := userMap[member.UserID]; ok {
			memberResps = append(memberResps, &dto.TeamMemberResp{
				UserID:   member.UserID,
				Username: user.Username,
				JoinedAt: member.JoinedAt,
			})
		}
	}

	return contestdomain.TeamRespFromModel(team, len(members)), memberResps, nil
}

func (s *TeamService) ListTeams(ctx context.Context, contestID int64) ([]*dto.TeamResp, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	teams, err := s.teamRepo.ListByContest(contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	teamIDs := make([]int64, len(teams))
	for i, team := range teams {
		teamIDs[i] = team.ID
	}

	countMap, err := s.teamRepo.GetMemberCountBatch(teamIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*dto.TeamResp, 0, len(teams))
	for _, team := range teams {
		result = append(result, contestdomain.TeamRespFromModel(team, countMap[team.ID]))
	}
	return result, nil
}

func (s *TeamService) GetMyTeam(ctx context.Context, contestID, userID int64) (map[string]any, error) {
	team, err := s.teamRepo.FindUserTeamInContest(userID, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	teamResp, members, err := s.GetTeamInfo(team.ID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"id":              teamResp.ID,
		"name":            teamResp.Name,
		"invite_code":     teamResp.InviteCode,
		"captain_user_id": teamResp.CaptainID,
		"members":         members,
	}, nil
}
