package contest

import (
	"context"
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
	teamRepo    *TeamRepository
	contestRepo Repository
}

func NewTeamService(teamRepo *TeamRepository, contestRepo Repository) *TeamService {
	return &TeamService{
		teamRepo:    teamRepo,
		contestRepo: contestRepo,
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, contestID, captainID int64, req *dto.CreateTeamReq) (*dto.TeamResp, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if contest.Status != model.ContestStatusRegistration && contest.Status != model.ContestStatusRunning {
		return nil, errcode.ErrContestTeamUnavailable
	}

	// 检查用户是否已在该竞赛的队伍中
	existingTeam, err := s.teamRepo.FindUserTeamInContest(captainID, contestID)
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
			return nil, errcode.ErrInviteCodeGenerationFailed.WithCause(err)
		}

		team = &model.Team{
			ContestID:  contestID,
			Name:       req.Name,
			CaptainID:  captainID,
			InviteCode: inviteCode,
			MaxMembers: maxMembers,
		}

		err = s.teamRepo.CreateWithMember(team, captainID)
		if err == nil {
			break
		}

		if IsUniqueViolation(err, "uk_teams_invite_code") {
			continue
		}
		if IsUniqueViolation(err, "uk_teams_contest_name") {
			return nil, errcode.ErrTeamNameExists
		}
		if IsUniqueViolation(err, "uk_team_members_contest_user") {
			return nil, errcode.ErrAlreadyInTeam
		}
		if !strings.Contains(err.Error(), "duplicate") && !strings.Contains(err.Error(), "unique") {
			return nil, errcode.ErrInternal.WithCause(err)
		}
	}

	if team == nil || team.ID == 0 {
		return nil, errcode.ErrInviteCodeGenerationFailed
	}

	return s.toTeamResp(team, 1), nil
}

func (s *TeamService) JoinTeam(ctx context.Context, contestID, userID, teamID int64) (*dto.TeamResp, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if contest.Status != model.ContestStatusRegistration && contest.Status != model.ContestStatusRunning {
		return nil, errcode.ErrContestTeamUnavailable
	}

	team, err := s.teamRepo.FindByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrTeamNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if team.ContestID != contestID {
		return nil, errcode.ErrTeamNotFound
	}

	// 检查用户是否已在该竞赛的队伍中
	existingTeam, err := s.teamRepo.FindUserTeamInContest(userID, team.ContestID)
	if err == nil && existingTeam.ID > 0 {
		return nil, errcode.ErrAlreadyInTeam
	}

	// 使用带锁的添加成员方法，防止并发竞态
	err = s.teamRepo.AddMemberWithLock(contestID, team.ID, userID)
	if err != nil {
		if errors.Is(err, ErrTeamFull) {
			return nil, errcode.ErrTeamFull
		}
		if errors.Is(err, ErrAlreadyJoinedContest) {
			return nil, errcode.ErrAlreadyInTeam
		}
		if IsUniqueViolation(err, "uk_team_members_contest_user") {
			return nil, errcode.ErrAlreadyInTeam
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	count, _ := s.teamRepo.GetMemberCount(team.ID)
	return s.toTeamResp(team, int(count)), nil
}

func (s *TeamService) LeaveTeam(_ context.Context, contestID, userID, teamID int64) error {
	team, err := s.teamRepo.FindByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrTeamNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}
	if team.ContestID != contestID {
		return errcode.ErrTeamNotFound
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

func (s *TeamService) DismissTeam(_ context.Context, contestID, captainID, teamID int64) error {
	team, err := s.teamRepo.FindByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrTeamNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}
	if team.ContestID != contestID {
		return errcode.ErrTeamNotFound
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

	users, err := s.teamRepo.FindUsersByIDs(userIDs)
	if err != nil {
		return nil, nil, errcode.ErrInternal.WithCause(err)
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

func (s *TeamService) ListTeams(ctx context.Context, contestID int64) ([]*dto.TeamResp, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	teams, err := s.teamRepo.ListByContest(contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	teamIDs := make([]int64, len(teams))
	for i, t := range teams {
		teamIDs[i] = t.ID
	}

	countMap, err := s.teamRepo.GetMemberCountBatch(teamIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
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
