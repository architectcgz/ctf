package queries

import (
	"context"
	"strings"
	"time"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
)

type TeacherAWDReviewService struct {
	repo       assessmentports.TeacherAWDReviewRepository
	pagination config.PaginationConfig
}

func NewTeacherAWDReviewService(repo assessmentports.TeacherAWDReviewRepository, pagination config.PaginationConfig) *TeacherAWDReviewService {
	return &TeacherAWDReviewService{
		repo:       repo,
		pagination: pagination,
	}
}

func (s *TeacherAWDReviewService) ListContests(ctx context.Context, requesterID int64, query ListTeacherAWDReviewContestsInput) (*TeacherAWDReviewContestPageResp, error) {
	_ = requesterID
	page := query.Page
	if page < 1 {
		page = 1
	}

	size := query.Size
	defaultPageSize := s.pagination.DefaultPageSize
	if defaultPageSize < 1 {
		defaultPageSize = 20
	}
	maxPageSize := s.pagination.MaxPageSize
	if maxPageSize < 1 {
		maxPageSize = 100
	}
	if size < 1 {
		size = defaultPageSize
	}
	if size > maxPageSize {
		size = maxPageSize
	}

	contests, total, summary, err := s.repo.ListTeacherAWDReviewContests(ctx, assessmentports.TeacherAWDReviewContestFilter{
		Status:  strings.TrimSpace(query.Status),
		Keyword: strings.TrimSpace(query.Keyword),
		Offset:  (page - 1) * size,
		Limit:   size,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	resp := &TeacherAWDReviewContestPageResp{
		List:     teacherAWDReviewMapper.ToTeacherAWDReviewContestResps(contests),
		Total:    total,
		Page:     page,
		PageSize: size,
		Summary: TeacherAWDReviewContestListSummaryResp{
			RunningCount:     summary.RunningCount,
			ExportReadyCount: summary.ExportReadyCount,
		},
	}
	return resp, nil
}

func (s *TeacherAWDReviewService) GetContestArchive(ctx context.Context, requesterID, contestID int64, req GetTeacherAWDReviewArchiveInput) (*TeacherAWDReviewArchiveResp, error) {
	if req.TeamID != nil && req.RoundNumber == nil {
		return nil, apperror.ErrInvalidParams.WithMessage("team_id 需要配合 round 使用")
	}

	contest, err := s.repo.FindTeacherAWDReviewContest(ctx, contestID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if contest == nil {
		return nil, contestcontracts.ErrContestNotFound
	}

	rounds, err := s.repo.ListTeacherAWDReviewRounds(ctx, contestID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	teams, err := s.repo.ListTeacherAWDReviewTeams(ctx, contestID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	selectedTeam, hasSelectedTeam := findTeacherAWDReviewTeam(teams, req.TeamID)
	if req.TeamID != nil && !hasSelectedTeam {
		return nil, apperror.ErrInvalidParams.WithMessage("team_id 无效")
	}

	resp := &TeacherAWDReviewArchiveResp{
		GeneratedAt: time.Now().UTC(),
		Scope: TeacherAWDReviewScopeResp{
			SnapshotType: snapshotTypeForContest(contest.Status),
			RequestedBy:  requesterID,
			RequestedID:  contestID,
		},
		Contest: teacherAWDReviewMapper.ToTeacherAWDReviewContestMetaResp(*contest),
		Rounds:  make([]TeacherAWDReviewRoundResp, 0, len(rounds)),
		Overview: &TeacherAWDReviewOverviewResp{
			RoundCount:       len(rounds),
			TeamCount:        len(teams),
			LatestEvidenceAt: contest.LatestEvidenceAt,
		},
	}

	var (
		selectedRound     *assessmentdomain.TeacherAWDReviewRoundSummary
		selectedRoundResp TeacherAWDReviewRoundResp
		selectedServices  []assessmentdomain.TeacherAWDReviewServiceRecord
		selectedAttacks   []assessmentdomain.TeacherAWDReviewAttackRecord
		selectedTraffic   []assessmentdomain.TeacherAWDReviewTrafficRecord
	)

	for _, round := range rounds {
		services, err := s.repo.ListTeacherAWDReviewRoundServices(ctx, round.ID)
		if err != nil {
			return nil, apperror.ErrInternal.WithCause(err)
		}
		attacks, err := s.repo.ListTeacherAWDReviewRoundAttacks(ctx, round.ID)
		if err != nil {
			return nil, apperror.ErrInternal.WithCause(err)
		}
		traffic, err := s.repo.ListTeacherAWDReviewRoundTraffic(ctx, contestID, round.ID)
		if err != nil {
			return nil, apperror.ErrInternal.WithCause(err)
		}

		roundResp := teacherAWDReviewMapper.ToTeacherAWDReviewRoundResp(round)
		roundResp.ServiceCount = len(services)
		roundResp.AttackCount = len(attacks)
		roundResp.TrafficCount = len(traffic)
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
			return nil, apperror.ErrInvalidParams.WithMessage("round 无效")
		}
		selectedTeams := teams
		if req.TeamID != nil {
			selectedTeams = []assessmentdomain.TeacherAWDReviewTeamSummary{*selectedTeam}
		}
		resp.SelectedRound = &TeacherAWDSelectedRoundResp{
			Round:    selectedRoundResp,
			Teams:    teacherAWDReviewMapper.ToTeacherAWDReviewTeamResps(selectedTeams),
			Services: teacherAWDReviewMapper.ToTeacherAWDReviewServiceResps(selectedServices),
			Attacks:  teacherAWDReviewMapper.ToTeacherAWDReviewAttackResps(selectedAttacks),
			Traffic:  teacherAWDReviewMapper.ToTeacherAWDReviewTrafficResps(selectedTraffic),
		}
	}

	return resp, nil
}

func snapshotTypeForContest(status string) string {
	if status == contestcontracts.ContestStatusEnded {
		return "final"
	}
	return "live"
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
