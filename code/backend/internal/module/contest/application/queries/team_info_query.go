package queries

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

func (s *TeamService) GetTeamInfo(ctx context.Context, teamID int64) (*TeamResult, []*TeamMemberResult, error) {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errcode.ErrTeamNotFound
		}
		return nil, nil, err
	}

	members, err := s.teamRepo.GetMembers(ctx, teamID)
	if err != nil {
		return nil, nil, err
	}

	userMap, err := s.loadTeamUsersByMembers(ctx, members)
	if err != nil {
		return nil, nil, err
	}

	memberResps := make([]*TeamMemberResult, 0, len(members))
	for _, member := range members {
		if user, ok := userMap[member.UserID]; ok {
			memberResps = append(memberResps, &TeamMemberResult{
				UserID:   member.UserID,
				Username: user.Username,
				JoinedAt: member.JoinedAt,
			})
		}
	}

	return teamResultFromModel(team, len(members)), memberResps, nil
}

func (s *TeamService) loadTeamUsersByMembers(ctx context.Context, members []*model.TeamMember) (map[int64]*model.User, error) {
	userIDs := make([]int64, len(members))
	for i, member := range members {
		userIDs[i] = member.UserID
	}

	users, err := s.teamRepo.FindUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	userMap := make(map[int64]*model.User, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}
	return userMap, nil
}
