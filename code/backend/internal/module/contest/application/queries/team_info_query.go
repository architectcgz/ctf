package queries

import (
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

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

	userMap, err := s.loadTeamUsersByMembers(members)
	if err != nil {
		return nil, nil, err
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

func (s *TeamService) loadTeamUsersByMembers(members []*model.TeamMember) (map[int64]*model.User, error) {
	userIDs := make([]int64, len(members))
	for i, member := range members {
		userIDs[i] = member.UserID
	}

	users, err := s.teamRepo.FindUsersByIDs(userIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	userMap := make(map[int64]*model.User, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}
	return userMap, nil
}
