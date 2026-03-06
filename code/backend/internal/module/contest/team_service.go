package contest

import (
	"crypto/rand"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
	"encoding/base32"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type TeamService struct {
	teamRepo *TeamRepository
	userRepo UserRepository
}

type UserRepository interface {
	FindByID(id int64) (*model.User, error)
	FindByIDs(ids []int64) ([]*model.User, error)
}

func NewTeamService(teamRepo *TeamRepository, userRepo UserRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *TeamService) CreateTeam(captainID int64, req *dto.CreateTeamReq) (*dto.TeamResp, error) {
	// 检查用户是否已在该竞赛的队伍中
	existingTeam, err := s.teamRepo.FindUserTeamInContest(captainID, req.ContestID)
	if err == nil && existingTeam.ID > 0 {
		return nil, errcode.ErrAlreadyInTeam
	}

	maxMembers := req.MaxMembers
	if maxMembers == 0 {
		maxMembers = 4
	}

	// 重试生成邀请码并创建队伍
	const maxRetries = 3
	var team *model.Team
	for i := 0; i < maxRetries; i++ {
		inviteCode, err := generateInviteCode()
		if err != nil {
			return nil, err
		}

		team = &model.Team{
			ContestID:  req.ContestID,
			Name:       req.Name,
			CaptainID:  captainID,
			InviteCode: inviteCode,
			MaxMembers: maxMembers,
		}

		err = s.teamRepo.CreateWithMember(team, captainID)
		if err == nil {
			break
		}

		// 检查是否为唯一索引冲突
		if !strings.Contains(err.Error(), "duplicate") && !strings.Contains(err.Error(), "unique") {
			return nil, err
		}
		// 重试
	}

	if team == nil || team.ID == 0 {
		return nil, errors.New("创建队伍失败")
	}

	return s.toTeamResp(team, 1), nil
}

func (s *TeamService) JoinTeam(userID int64, req *dto.JoinTeamReq) (*dto.TeamResp, error) {
	team, err := s.teamRepo.FindByInviteCode(req.InviteCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrInvalidInviteCode
		}
		return nil, err
	}

	// 检查用户是否已在该竞赛的队伍中
	existingTeam, err := s.teamRepo.FindUserTeamInContest(userID, team.ContestID)
	if err == nil && existingTeam.ID > 0 {
		return nil, errcode.ErrAlreadyInTeamDup
	}

	// 使用带锁的添加成员方法，防止并发竞态
	err = s.teamRepo.AddMemberWithLock(team.ID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrInvalidData) {
			return nil, errcode.ErrTeamFull
		}
		return nil, err
	}

	count, _ := s.teamRepo.GetMemberCount(team.ID)
	return s.toTeamResp(team, int(count)), nil
}

func (s *TeamService) LeaveTeam(userID, teamID int64) error {
	team, err := s.teamRepo.FindByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrTeamNotFound
		}
		return err
	}

	if team.CaptainID == userID {
		return errcode.ErrCaptainCannotLeave
	}

	members, err := s.teamRepo.GetMembers(teamID)
	if err != nil {
		return err
	}

	found := false
	for _, m := range members {
		if m.UserID == userID {
			found = true
			break
		}
	}
	if !found {
		return errcode.ErrNotInTeam
	}

	return s.teamRepo.RemoveMember(teamID, userID)
}

func (s *TeamService) DismissTeam(captainID, teamID int64) error {
	team, err := s.teamRepo.FindByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrTeamNotFound
		}
		return err
	}

	if team.CaptainID != captainID {
		return errcode.ErrNotCaptain
	}

	return s.teamRepo.DeleteWithMembers(teamID)
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
	for i, m := range members {
		userIDs[i] = m.UserID
	}

	users, err := s.userRepo.FindByIDs(userIDs)
	if err != nil {
		return nil, nil, err
	}

	userMap := make(map[int64]*model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	memberResps := make([]*dto.TeamMemberResp, 0, len(members))
	for _, m := range members {
		if user, ok := userMap[m.UserID]; ok {
			memberResps = append(memberResps, &dto.TeamMemberResp{
				UserID:   m.UserID,
				Username: user.Username,
				JoinedAt: m.JoinedAt,
			})
		}
	}

	return s.toTeamResp(team, len(members)), memberResps, nil
}

func (s *TeamService) ListTeams(contestID int64) ([]*dto.TeamResp, error) {
	teams, err := s.teamRepo.ListByContest(contestID)
	if err != nil {
		return nil, err
	}

	teamIDs := make([]int64, len(teams))
	for i, t := range teams {
		teamIDs[i] = t.ID
	}

	countMap, err := s.teamRepo.GetMemberCountBatch(teamIDs)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.TeamResp, 0, len(teams))
	for _, team := range teams {
		count := countMap[team.ID]
		result = append(result, s.toTeamResp(team, count))
	}
	return result, nil
}

func (s *TeamService) toTeamResp(team *model.Team, memberCount int) *dto.TeamResp {
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

func generateInviteCode() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	code := base32.StdEncoding.EncodeToString(bytes)
	code = strings.ReplaceAll(code, "=", "")
	if len(code) > 6 {
		code = code[:6]
	}
	return code, nil
}
