package contest

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
	"errors"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

type TeamService struct {
	teamRepo *TeamRepository
	userRepo UserRepository
}

type UserRepository interface {
	FindByID(id int64) (*model.User, error)
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
		return nil, errcode.New(14001, "您已加入该竞赛的队伍", 409)
	}

	maxMembers := req.MaxMembers
	if maxMembers == 0 {
		maxMembers = 4
	}

	team := &model.Team{
		ContestID:  req.ContestID,
		Name:       req.Name,
		CaptainID:  captainID,
		InviteCode: generateInviteCode(),
		MaxMembers: maxMembers,
	}

	if err := s.teamRepo.Create(team); err != nil {
		return nil, err
	}

	// 队长自动加入队伍
	member := &model.TeamMember{
		TeamID:   team.ID,
		UserID:   captainID,
		JoinedAt: time.Now(),
	}
	if err := s.teamRepo.AddMember(member); err != nil {
		return nil, err
	}

	return s.toTeamResp(team, 1), nil
}

func (s *TeamService) JoinTeam(userID int64, req *dto.JoinTeamReq) (*dto.TeamResp, error) {
	team, err := s.teamRepo.FindByInviteCode(req.InviteCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(14002, "邀请码无效", 404)
		}
		return nil, err
	}

	// 检查用户是否已在该竞赛的队伍中
	existingTeam, err := s.teamRepo.FindUserTeamInContest(userID, team.ContestID)
	if err == nil && existingTeam.ID > 0 {
		return nil, errcode.New(14003, "您已加入该竞赛的队伍", 409)
	}

	// 检查队伍人数
	count, err := s.teamRepo.GetMemberCount(team.ID)
	if err != nil {
		return nil, err
	}
	if count >= int64(team.MaxMembers) {
		return nil, errcode.New(14004, "队伍人数已满", 403)
	}

	member := &model.TeamMember{
		TeamID:   team.ID,
		UserID:   userID,
		JoinedAt: time.Now(),
	}
	if err := s.teamRepo.AddMember(member); err != nil {
		return nil, err
	}

	return s.toTeamResp(team, int(count)+1), nil
}

func (s *TeamService) LeaveTeam(userID, teamID int64) error {
	team, err := s.teamRepo.FindByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.New(14005, "队伍不存在", 404)
		}
		return err
	}

	// 队长不能直接退出，需要先解散队伍
	if team.CaptainID == userID {
		return errcode.New(14006, "队长不能退出队伍，请先解散队伍", 403)
	}

	return s.teamRepo.RemoveMember(teamID, userID)
}

func (s *TeamService) DismissTeam(captainID, teamID int64) error {
	team, err := s.teamRepo.FindByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.New(14005, "队伍不存在", 404)
		}
		return err
	}

	if team.CaptainID != captainID {
		return errcode.New(14007, "只有队长可以解散队伍", 403)
	}

	return s.teamRepo.Delete(teamID)
}

func (s *TeamService) GetTeamInfo(teamID int64) (*dto.TeamResp, []*dto.TeamMemberResp, error) {
	team, err := s.teamRepo.FindByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errcode.New(14005, "队伍不存在", 404)
		}
		return nil, nil, err
	}

	members, err := s.teamRepo.GetMembers(teamID)
	if err != nil {
		return nil, nil, err
	}

	memberResps := make([]*dto.TeamMemberResp, 0, len(members))
	for _, m := range members {
		user, err := s.userRepo.FindByID(m.UserID)
		if err != nil {
			continue
		}
		memberResps = append(memberResps, &dto.TeamMemberResp{
			UserID:   m.UserID,
			Username: user.Username,
			JoinedAt: m.JoinedAt,
		})
	}

	return s.toTeamResp(team, len(members)), memberResps, nil
}

func (s *TeamService) ListTeams(contestID int64) ([]*dto.TeamResp, error) {
	teams, err := s.teamRepo.ListByContest(contestID)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.TeamResp, 0, len(teams))
	for _, team := range teams {
		count, _ := s.teamRepo.GetMemberCount(team.ID)
		result = append(result, s.toTeamResp(team, int(count)))
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

func generateInviteCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, 6)
	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}
	return string(code)
}
