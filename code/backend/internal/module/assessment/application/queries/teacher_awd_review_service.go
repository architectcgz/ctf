package queries

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	"ctf-platform/pkg/errcode"
)

type TeacherAWDReviewService struct {
	repo assessmentports.TeacherAWDReviewRepository
}

func NewTeacherAWDReviewService(repo assessmentports.TeacherAWDReviewRepository) *TeacherAWDReviewService {
	return &TeacherAWDReviewService{repo: repo}
}

func (s *TeacherAWDReviewService) ListContests(ctx context.Context, requesterID int64) (*dto.TeacherAWDReviewContestListResp, error) {
	_ = requesterID

	contests, err := s.repo.ListTeacherAWDReviewContests(ctx)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	resp := &dto.TeacherAWDReviewContestListResp{
		Contests: make([]dto.TeacherAWDReviewContestResp, 0, len(contests)),
	}
	for _, contest := range contests {
		resp.Contests = append(resp.Contests, dto.TeacherAWDReviewContestResp{
			ID:               contest.ID,
			Title:            contest.Title,
			Mode:             contest.Mode,
			Status:           contest.Status,
			CurrentRound:     contest.CurrentRound,
			RoundCount:       contest.RoundCount,
			TeamCount:        contest.TeamCount,
			LatestEvidenceAt: contest.LatestEvidenceAt,
			ExportReady:      contest.ExportReady,
		})
	}
	return resp, nil
}

func (s *TeacherAWDReviewService) GetContestArchive(ctx context.Context, requesterID, contestID int64, req *dto.GetTeacherAWDReviewArchiveReq) (*dto.TeacherAWDReviewArchiveResp, error) {
	if req == nil {
		req = &dto.GetTeacherAWDReviewArchiveReq{}
	}
	if req.TeamID != nil && req.RoundNumber == nil {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "team_id 需要配合 round 使用", errcode.ErrInvalidParams.HTTPStatus)
	}

	contest, err := s.repo.FindTeacherAWDReviewContest(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if contest == nil {
		return nil, errcode.ErrContestNotFound
	}

	rounds, err := s.repo.ListTeacherAWDReviewRounds(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	teams, err := s.repo.ListTeacherAWDReviewTeams(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	selectedTeam, hasSelectedTeam := findTeacherAWDReviewTeam(teams, req.TeamID)
	if req.TeamID != nil && !hasSelectedTeam {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "team_id 无效", errcode.ErrInvalidParams.HTTPStatus)
	}

	resp := &dto.TeacherAWDReviewArchiveResp{
		GeneratedAt: time.Now().UTC(),
		Scope: dto.TeacherAWDReviewScopeResp{
			SnapshotType: snapshotTypeForContest(contest.Status),
			RequestedBy:  requesterID,
			RequestedID:  contestID,
		},
		Contest: dto.TeacherAWDReviewContestMetaResp{
			ID:               contest.ID,
			Title:            contest.Title,
			Mode:             contest.Mode,
			Status:           contest.Status,
			CurrentRound:     contest.CurrentRound,
			RoundCount:       contest.RoundCount,
			TeamCount:        contest.TeamCount,
			LatestEvidenceAt: contest.LatestEvidenceAt,
			ExportReady:      contest.ExportReady,
		},
		Rounds: make([]dto.TeacherAWDReviewRoundResp, 0, len(rounds)),
		Overview: &dto.TeacherAWDReviewOverviewResp{
			RoundCount:       len(rounds),
			TeamCount:        len(teams),
			LatestEvidenceAt: contest.LatestEvidenceAt,
		},
	}

	var (
		selectedRound     *assessmentdomain.TeacherAWDReviewRoundSummary
		selectedRoundResp dto.TeacherAWDReviewRoundResp
		selectedServices  []assessmentdomain.TeacherAWDReviewServiceRecord
		selectedAttacks   []assessmentdomain.TeacherAWDReviewAttackRecord
		selectedTraffic   []assessmentdomain.TeacherAWDReviewTrafficRecord
	)

	for _, round := range rounds {
		services, err := s.repo.ListTeacherAWDReviewRoundServices(ctx, round.ID)
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		attacks, err := s.repo.ListTeacherAWDReviewRoundAttacks(ctx, round.ID)
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		traffic, err := s.repo.ListTeacherAWDReviewRoundTraffic(ctx, contestID, round.ID)
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}

		roundResp := dto.TeacherAWDReviewRoundResp{
			ID:           round.ID,
			ContestID:    round.ContestID,
			RoundNumber:  round.RoundNumber,
			Status:       round.Status,
			StartedAt:    round.StartedAt,
			EndedAt:      round.EndedAt,
			AttackScore:  round.AttackScore,
			DefenseScore: round.DefenseScore,
			ServiceCount: len(services),
			AttackCount:  len(attacks),
			TrafficCount: len(traffic),
		}
		resp.Rounds = append(resp.Rounds, roundResp)
		resp.Overview.ServiceCount += len(services)
		resp.Overview.AttackCount += len(attacks)
		resp.Overview.TrafficCount += len(traffic)

		if req.RoundNumber != nil && round.RoundNumber == *req.RoundNumber {
			roundCopy := round
			selectedRound = &roundCopy
			selectedRoundResp = roundResp
			selectedServices = services
			selectedAttacks = attacks
			selectedTraffic = traffic
			if req.TeamID != nil {
				selectedServices = filterTeacherAWDReviewServicesByTeam(selectedServices, *req.TeamID)
				selectedAttacks = filterTeacherAWDReviewAttacksByTeam(selectedAttacks, *req.TeamID)
				selectedTraffic = filterTeacherAWDReviewTrafficByTeam(selectedTraffic, *req.TeamID)
				selectedRoundResp.ServiceCount = len(selectedServices)
				selectedRoundResp.AttackCount = len(selectedAttacks)
				selectedRoundResp.TrafficCount = len(selectedTraffic)
			}
		}
	}

	if req.RoundNumber != nil {
		if selectedRound == nil {
			return nil, errcode.New(errcode.ErrInvalidParams.Code, "round 无效", errcode.ErrInvalidParams.HTTPStatus)
		}
		selectedTeams := teams
		if req.TeamID != nil {
			selectedTeams = []assessmentdomain.TeacherAWDReviewTeamSummary{*selectedTeam}
		}
		resp.SelectedRound = &dto.TeacherAWDSelectedRoundResp{
			Round:    selectedRoundResp,
			Teams:    toTeacherAWDReviewTeams(selectedTeams),
			Services: toTeacherAWDReviewServices(selectedServices),
			Attacks:  toTeacherAWDReviewAttacks(selectedAttacks),
			Traffic:  toTeacherAWDReviewTraffic(selectedTraffic),
		}
	}

	return resp, nil
}

func snapshotTypeForContest(status string) string {
	if status == model.ContestStatusEnded {
		return "final"
	}
	return "live"
}

func toTeacherAWDReviewTeams(items []assessmentdomain.TeacherAWDReviewTeamSummary) []dto.TeacherAWDReviewTeamResp {
	resp := make([]dto.TeacherAWDReviewTeamResp, 0, len(items))
	for _, item := range items {
		resp = append(resp, dto.TeacherAWDReviewTeamResp{
			TeamID:      item.TeamID,
			TeamName:    item.TeamName,
			CaptainID:   item.CaptainID,
			TotalScore:  item.TotalScore,
			MemberCount: item.MemberCount,
			LastSolveAt: item.LastSolveAt,
		})
	}
	return resp
}

func toTeacherAWDReviewServices(items []assessmentdomain.TeacherAWDReviewServiceRecord) []dto.TeacherAWDReviewServiceResp {
	resp := make([]dto.TeacherAWDReviewServiceResp, 0, len(items))
	for _, item := range items {
		resp = append(resp, dto.TeacherAWDReviewServiceResp{
			ID:             item.ID,
			RoundID:        item.RoundID,
			TeamID:         item.TeamID,
			TeamName:       item.TeamName,
			ServiceID:      item.ServiceID,
			ChallengeID:    item.ChallengeID,
			ChallengeTitle: item.ChallengeTitle,
			ServiceStatus:  item.ServiceStatus,
			AttackReceived: item.AttackReceived,
			SLAScore:       item.SLAScore,
			DefenseScore:   item.DefenseScore,
			AttackScore:    item.AttackScore,
			UpdatedAt:      item.UpdatedAt,
		})
	}
	return resp
}

func toTeacherAWDReviewAttacks(items []assessmentdomain.TeacherAWDReviewAttackRecord) []dto.TeacherAWDReviewAttackResp {
	resp := make([]dto.TeacherAWDReviewAttackResp, 0, len(items))
	for _, item := range items {
		resp = append(resp, dto.TeacherAWDReviewAttackResp{
			ID:               item.ID,
			RoundID:          item.RoundID,
			AttackerTeamID:   item.AttackerTeamID,
			AttackerTeamName: item.AttackerTeamName,
			VictimTeamID:     item.VictimTeamID,
			VictimTeamName:   item.VictimTeamName,
			ServiceID:        item.ServiceID,
			ChallengeID:      item.ChallengeID,
			ChallengeTitle:   item.ChallengeTitle,
			AttackType:       item.AttackType,
			Source:           item.Source,
			SubmittedFlag:    item.SubmittedFlag,
			IsSuccess:        item.IsSuccess,
			ScoreGained:      item.ScoreGained,
			CreatedAt:        item.CreatedAt,
		})
	}
	return resp
}

func toTeacherAWDReviewTraffic(items []assessmentdomain.TeacherAWDReviewTrafficRecord) []dto.TeacherAWDReviewTrafficResp {
	resp := make([]dto.TeacherAWDReviewTrafficResp, 0, len(items))
	for _, item := range items {
		resp = append(resp, dto.TeacherAWDReviewTrafficResp{
			ID:               item.ID,
			ContestID:        item.ContestID,
			RoundID:          item.RoundID,
			AttackerTeamID:   item.AttackerTeamID,
			AttackerTeamName: item.AttackerTeamName,
			VictimTeamID:     item.VictimTeamID,
			VictimTeamName:   item.VictimTeamName,
			ServiceID:        item.ServiceID,
			ChallengeID:      item.ChallengeID,
			ChallengeTitle:   item.ChallengeTitle,
			Method:           item.Method,
			Path:             item.Path,
			StatusCode:       item.StatusCode,
			Source:           item.Source,
			CreatedAt:        item.CreatedAt,
		})
	}
	return resp
}

func findTeacherAWDReviewTeam(items []assessmentdomain.TeacherAWDReviewTeamSummary, teamID *int64) (*assessmentdomain.TeacherAWDReviewTeamSummary, bool) {
	if teamID == nil {
		return nil, false
	}
	for _, item := range items {
		if item.TeamID == *teamID {
			team := item
			return &team, true
		}
	}
	return nil, false
}

func filterTeacherAWDReviewServicesByTeam(items []assessmentdomain.TeacherAWDReviewServiceRecord, teamID int64) []assessmentdomain.TeacherAWDReviewServiceRecord {
	filtered := make([]assessmentdomain.TeacherAWDReviewServiceRecord, 0, len(items))
	for _, item := range items {
		if item.TeamID == teamID {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func filterTeacherAWDReviewAttacksByTeam(items []assessmentdomain.TeacherAWDReviewAttackRecord, teamID int64) []assessmentdomain.TeacherAWDReviewAttackRecord {
	filtered := make([]assessmentdomain.TeacherAWDReviewAttackRecord, 0, len(items))
	for _, item := range items {
		if item.AttackerTeamID == teamID || item.VictimTeamID == teamID {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func filterTeacherAWDReviewTrafficByTeam(items []assessmentdomain.TeacherAWDReviewTrafficRecord, teamID int64) []assessmentdomain.TeacherAWDReviewTrafficRecord {
	filtered := make([]assessmentdomain.TeacherAWDReviewTrafficRecord, 0, len(items))
	for _, item := range items {
		if item.AttackerTeamID == teamID || item.VictimTeamID == teamID {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
