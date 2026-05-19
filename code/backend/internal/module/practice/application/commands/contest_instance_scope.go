package commands

import (
	"context"
	"errors"
	"fmt"

	"ctf-platform/internal/model"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/pkg/errcode"
)

func (s *Service) resolveContestChallengeInstanceScope(ctx context.Context, userID, contestID, challengeID int64) (practiceports.InstanceScope, error) {
	if s.contestScope == nil {
		return practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(fmt.Errorf("practice contest scope repository is nil"))
	}
	scope, err := s.resolveContestBaseInstanceScope(ctx, userID, contestID)
	if err != nil {
		return practiceports.InstanceScope{}, err
	}
	if scope.ContestMode == practiceports.ContestModeAWD {
		return practiceports.InstanceScope{}, errcode.ErrInvalidParams.WithCause(
			errors.New("awd 赛事实例启动必须使用 service_id 入口"),
		)
	}
	contestChallenge, err := s.contestScope.FindContestChallenge(ctx, contestID, challengeID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeContestChallengeNotFound) {
			return practiceports.InstanceScope{}, errcode.ErrChallengeNotInContest
		}
		return practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	if !contestChallenge.IsVisible {
		return practiceports.InstanceScope{}, errcode.ErrContestChallengeVisible
	}
	return scope, nil
}

func (s *Service) resolveContestAWDServiceInstanceScope(ctx context.Context, userID, contestID, serviceID int64) (int64, practiceports.InstanceScope, error) {
	if s.contestScope == nil {
		return 0, practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(fmt.Errorf("practice contest scope repository is nil"))
	}
	scope, err := s.resolveContestBaseInstanceScope(ctx, userID, contestID)
	if err != nil {
		return 0, practiceports.InstanceScope{}, err
	}
	subject, err := s.contestScope.FindContestAWDServiceRuntimeSubject(ctx, contestID, serviceID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeContestAWDServiceNotFound) {
			return 0, practiceports.InstanceScope{}, errcode.ErrChallengeNotInContest
		}
		return 0, practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	if subject == nil {
		return 0, practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(fmt.Errorf("practice contest awd runtime subject is nil"))
	}
	if !subject.Visible {
		return 0, practiceports.InstanceScope{}, errcode.ErrContestChallengeVisible
	}
	serviceIDCopy := subject.ServiceID
	scope.ServiceID = &serviceIDCopy
	return subject.ChallengeID, scope, nil
}

func (s *Service) resolveAdminContestAWDServiceInstanceScope(ctx context.Context, contestID, teamID, serviceID int64) (int64, int64, practiceports.InstanceScope, error) {
	contest, err := s.loadAdminContestAWDContest(ctx, contestID)
	if err != nil {
		return 0, 0, practiceports.InstanceScope{}, err
	}
	switch contest.Status {
	case practiceports.ContestStatusRegistration, practiceports.ContestStatusRunning, practiceports.ContestStatusFrozen:
	default:
		if contest.Status == practiceports.ContestStatusEnded {
			return 0, 0, practiceports.InstanceScope{}, errcode.ErrContestEnded
		}
		return 0, 0, practiceports.InstanceScope{}, errcode.ErrContestNotRunning
	}
	return s.resolveAdminContestAWDServiceInstanceScopeWithContest(ctx, contest, contestID, teamID, serviceID)
}

func (s *Service) loadAdminContestAWDContest(ctx context.Context, contestID int64) (*practiceports.ContestRecord, error) {
	if s.contestScope == nil {
		return nil, errcode.ErrInternal.WithCause(fmt.Errorf("practice contest scope repository is nil"))
	}
	contest, err := s.contestScope.FindContestByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if contest.Mode != practiceports.ContestModeAWD {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("仅 AWD 赛事支持队伍实例编排"))
	}
	return contest, nil
}

func (s *Service) resolveAdminContestAWDServiceInstanceScopeWithContest(ctx context.Context, contest *practiceports.ContestRecord, contestID, teamID, serviceID int64) (int64, int64, practiceports.InstanceScope, error) {
	team, err := s.contestScope.FindContestTeam(ctx, contestID, teamID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeContestTeamNotFound) {
			return 0, 0, practiceports.InstanceScope{}, errcode.ErrTeamNotFound
		}
		return 0, 0, practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	if team.CaptainID <= 0 {
		return 0, 0, practiceports.InstanceScope{}, errcode.ErrInvalidParams.WithCause(errors.New("队伍缺少队长用户"))
	}

	subject, err := s.contestScope.FindContestAWDServiceRuntimeSubject(ctx, contestID, serviceID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeContestAWDServiceNotFound) {
			return 0, 0, practiceports.InstanceScope{}, errcode.ErrChallengeNotInContest
		}
		return 0, 0, practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	if subject == nil {
		return 0, 0, practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(fmt.Errorf("practice contest awd runtime subject is nil"))
	}
	if !subject.Visible {
		return 0, 0, practiceports.InstanceScope{}, errcode.ErrContestChallengeVisible
	}

	contestIDCopy := contestID
	teamIDCopy := teamID
	serviceIDCopy := subject.ServiceID
	scope := practiceports.InstanceScope{
		ContestID:     &contestIDCopy,
		ContestMode:   contest.Mode,
		TeamID:        &teamIDCopy,
		ServiceID:     &serviceIDCopy,
		FlagSubjectID: teamID,
		ShareScope:    instancecontracts.ShareScopePerTeam,
	}
	return subject.ChallengeID, team.CaptainID, scope, nil
}

func (s *Service) loadRuntimeSubjectWithScope(ctx context.Context, scope practiceports.InstanceScope, challengeID int64) (*model.Challenge, *practiceports.RuntimeChallengeTopology, error) {
	if scope.ServiceID != nil && scope.ContestID != nil {
		return s.loadContestAWDServiceRuntimeSubject(ctx, *scope.ContestID, *scope.ServiceID)
	}

	if s.runtimeSubject == nil {
		return nil, nil, errcode.ErrInternal.WithCause(fmt.Errorf("practice runtime subject repository is nil"))
	}
	chal, err := s.runtimeSubject.FindByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeChallengeNotFound) {
			return nil, nil, errcode.ErrChallengeNotFound
		}
		return nil, nil, errcode.ErrInternal.WithCause(err)
	}
	topology, err := s.runtimeSubject.FindChallengeTopologyByChallengeID(ctx, chal.ID)
	if err != nil && !errors.Is(err, practiceports.ErrPracticeChallengeTopologyNotFound) {
		return nil, nil, errcode.ErrContainerCreateFailed.WithCause(err)
	}
	return chal, topology, nil
}

func (s *Service) loadRuntimeSubjectForInstance(ctx context.Context, instance *instancecontracts.Instance) (*model.Challenge, *practiceports.RuntimeChallengeTopology, error) {
	if instance != nil && instance.ServiceID != nil && instance.ContestID != nil {
		return s.loadContestAWDServiceRuntimeSubject(ctx, *instance.ContestID, *instance.ServiceID)
	}
	return s.loadRuntimeSubjectWithScope(ctx, practiceports.InstanceScope{}, instance.ChallengeID)
}

func (s *Service) loadContestAWDServiceRuntimeSubject(ctx context.Context, contestID, serviceID int64) (*model.Challenge, *practiceports.RuntimeChallengeTopology, error) {
	if s.contestScope == nil {
		return nil, nil, errcode.ErrInternal.WithCause(fmt.Errorf("practice contest scope repository is nil"))
	}
	subject, err := s.contestScope.FindContestAWDServiceRuntimeSubject(ctx, contestID, serviceID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeContestAWDServiceNotFound) {
			return nil, nil, errcode.ErrChallengeNotInContest
		}
		return nil, nil, errcode.ErrInternal.WithCause(err)
	}
	if subject == nil {
		return nil, nil, errcode.ErrInternal.WithCause(fmt.Errorf("practice contest awd runtime subject is nil"))
	}
	chal := buildContestAWDServiceVirtualChallenge(subject)
	topology := buildContestAWDServiceVirtualTopology(subject)
	return chal, topology, nil
}

func (s *Service) resolveContestBaseInstanceScope(ctx context.Context, userID, contestID int64) (practiceports.InstanceScope, error) {
	if s.contestScope == nil {
		return practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(fmt.Errorf("practice contest scope repository is nil"))
	}

	contest, err := s.contestScope.FindContestByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeContestNotFound) {
			return practiceports.InstanceScope{}, errcode.ErrContestNotFound
		}
		return practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	switch contest.Status {
	case practiceports.ContestStatusRunning, practiceports.ContestStatusFrozen:
	default:
		if contest.Status == practiceports.ContestStatusEnded {
			return practiceports.InstanceScope{}, errcode.ErrContestEnded
		}
		return practiceports.InstanceScope{}, errcode.ErrContestNotRunning
	}

	registration, err := s.contestScope.FindContestRegistration(ctx, contestID, userID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeContestRegistrationNotFound) {
			return practiceports.InstanceScope{}, errcode.ErrNotRegistered
		}
		return practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	switch registration.Status {
	case practiceports.ContestRegistrationStatusApproved:
	case practiceports.ContestRegistrationStatusPending:
		return practiceports.InstanceScope{}, errcode.ErrContestRegistrationPending
	default:
		return practiceports.InstanceScope{}, errcode.ErrRegistrationNotApproved
	}

	contestIDCopy := contestID
	scope := practiceports.InstanceScope{
		ContestID:     &contestIDCopy,
		ContestMode:   contest.Mode,
		FlagSubjectID: userID,
		ShareScope:    instancecontracts.ShareScopePerUser,
	}
	if registration.TeamID != nil && *registration.TeamID > 0 {
		teamID := *registration.TeamID
		scope.TeamID = &teamID
	}

	return scope, nil
}

func resolveEffectiveInstanceScope(chal *model.Challenge, scope practiceports.InstanceScope) practiceports.InstanceScope {
	effective := scope
	effective.FlagSubjectID = scope.FlagSubjectID
	effective.ShareScope = instancecontracts.ShareScopePerUser

	switch {
	case scope.ContestMode == practiceports.ContestModeAWD:
		effective.ShareScope = instancecontracts.ShareScopePerTeam
		if scope.TeamID != nil && *scope.TeamID > 0 {
			effective.FlagSubjectID = *scope.TeamID
		}
	case chal.InstanceSharing == model.InstanceSharingShared:
		effective.ShareScope = instancecontracts.ShareScopeShared
		effective.TeamID = nil
	case chal.InstanceSharing == model.InstanceSharingPerTeam && scope.TeamID != nil && *scope.TeamID > 0:
		effective.ShareScope = instancecontracts.ShareScopePerTeam
		effective.FlagSubjectID = *scope.TeamID
	default:
		effective.ShareScope = instancecontracts.ShareScopePerUser
		effective.TeamID = nil
	}

	if effective.ShareScope != instancecontracts.ShareScopePerTeam {
		effective.TeamID = nil
	}
	return effective
}
