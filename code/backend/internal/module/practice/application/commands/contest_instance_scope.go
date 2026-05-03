package commands

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/pkg/errcode"
)

func (s *Service) resolveContestChallengeInstanceScope(ctx context.Context, userID, contestID, challengeID int64) (practiceports.InstanceScope, error) {
	scope, err := s.resolveContestBaseInstanceScope(ctx, userID, contestID)
	if err != nil {
		return practiceports.InstanceScope{}, err
	}
	if scope.ContestMode == model.ContestModeAWD {
		return practiceports.InstanceScope{}, errcode.ErrInvalidParams.WithCause(
			errors.New("awd 赛事实例启动必须使用 service_id 入口"),
		)
	}
	contestChallenge, err := s.repo.FindContestChallenge(ctx, contestID, challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
	scope, err := s.resolveContestBaseInstanceScope(ctx, userID, contestID)
	if err != nil {
		return 0, practiceports.InstanceScope{}, err
	}
	service, err := s.repo.FindContestAWDService(ctx, contestID, serviceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, practiceports.InstanceScope{}, errcode.ErrChallengeNotInContest
		}
		return 0, practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	if !service.IsVisible {
		return 0, practiceports.InstanceScope{}, errcode.ErrContestChallengeVisible
	}
	serviceIDCopy := service.ID
	scope.ServiceID = &serviceIDCopy
	return service.AWDChallengeID, scope, nil
}

func (s *Service) resolveAdminContestAWDServiceInstanceScope(ctx context.Context, contestID, teamID, serviceID int64) (int64, int64, practiceports.InstanceScope, error) {
	contest, err := s.repo.FindContestByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, practiceports.InstanceScope{}, errcode.ErrContestNotFound
		}
		return 0, 0, practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	if contest.Mode != model.ContestModeAWD {
		return 0, 0, practiceports.InstanceScope{}, errcode.ErrInvalidParams.WithCause(errors.New("仅 AWD 赛事支持队伍实例编排"))
	}
	switch contest.Status {
	case model.ContestStatusRunning, model.ContestStatusFrozen:
	default:
		if contest.Status == model.ContestStatusEnded {
			return 0, 0, practiceports.InstanceScope{}, errcode.ErrContestEnded
		}
		return 0, 0, practiceports.InstanceScope{}, errcode.ErrContestNotRunning
	}

	team, err := s.repo.FindContestTeam(ctx, contestID, teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, practiceports.InstanceScope{}, errcode.ErrTeamNotFound
		}
		return 0, 0, practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	if team.CaptainID <= 0 {
		return 0, 0, practiceports.InstanceScope{}, errcode.ErrInvalidParams.WithCause(errors.New("队伍缺少队长用户"))
	}

	service, err := s.repo.FindContestAWDService(ctx, contestID, serviceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, practiceports.InstanceScope{}, errcode.ErrChallengeNotInContest
		}
		return 0, 0, practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	if !service.IsVisible {
		return 0, 0, practiceports.InstanceScope{}, errcode.ErrContestChallengeVisible
	}

	contestIDCopy := contestID
	teamIDCopy := teamID
	serviceIDCopy := service.ID
	scope := practiceports.InstanceScope{
		ContestID:     &contestIDCopy,
		ContestMode:   contest.Mode,
		TeamID:        &teamIDCopy,
		ServiceID:     &serviceIDCopy,
		FlagSubjectID: teamID,
		ShareScope:    model.InstanceSharingPerTeam,
	}
	return service.AWDChallengeID, team.CaptainID, scope, nil
}

func (s *Service) loadRuntimeSubjectWithScope(ctx context.Context, scope practiceports.InstanceScope, challengeID int64) (*model.Challenge, *model.ChallengeTopology, error) {
	if scope.ServiceID != nil && scope.ContestID != nil {
		return s.loadContestAWDServiceRuntimeSubject(ctx, *scope.ContestID, *scope.ServiceID)
	}

	chal, err := s.challengeRepo.FindByID(ctx, challengeID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, errcode.ErrChallengeNotFound
		}
		return nil, nil, errcode.ErrInternal.WithCause(err)
	}
	topology, err := s.challengeRepo.FindChallengeTopologyByChallengeID(ctx, chal.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, errcode.ErrContainerCreateFailed.WithCause(err)
	}
	return chal, topology, nil
}

func (s *Service) loadRuntimeSubjectForInstance(ctx context.Context, instance *model.Instance) (*model.Challenge, *model.ChallengeTopology, error) {
	if instance != nil && instance.ServiceID != nil && instance.ContestID != nil {
		return s.loadContestAWDServiceRuntimeSubject(ctx, *instance.ContestID, *instance.ServiceID)
	}
	return s.loadRuntimeSubjectWithScope(ctx, practiceports.InstanceScope{}, instance.ChallengeID)
}

func (s *Service) loadContestAWDServiceRuntimeSubject(ctx context.Context, contestID, serviceID int64) (*model.Challenge, *model.ChallengeTopology, error) {
	service, err := s.repo.FindContestAWDService(ctx, contestID, serviceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errcode.ErrChallengeNotInContest
		}
		return nil, nil, errcode.ErrInternal.WithCause(err)
	}
	snapshot, err := model.DecodeContestAWDServiceSnapshot(service.ServiceSnapshot)
	if err != nil {
		return nil, nil, errcode.ErrInternal.WithCause(err)
	}
	chal := buildContestAWDServiceVirtualChallenge(service, snapshot)
	topology, err := buildContestAWDServiceVirtualTopology(service, snapshot)
	if err != nil {
		return nil, nil, errcode.ErrInternal.WithCause(err)
	}
	return chal, topology, nil
}

func (s *Service) resolveContestBaseInstanceScope(ctx context.Context, userID, contestID int64) (practiceports.InstanceScope, error) {
	if s.repo == nil {
		return practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(fmt.Errorf("practice repository is nil"))
	}

	contest, err := s.repo.FindContestByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return practiceports.InstanceScope{}, errcode.ErrContestNotFound
		}
		return practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	switch contest.Status {
	case model.ContestStatusRunning, model.ContestStatusFrozen:
	default:
		if contest.Status == model.ContestStatusEnded {
			return practiceports.InstanceScope{}, errcode.ErrContestEnded
		}
		return practiceports.InstanceScope{}, errcode.ErrContestNotRunning
	}

	registration, err := s.repo.FindContestRegistration(ctx, contestID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return practiceports.InstanceScope{}, errcode.ErrNotRegistered
		}
		return practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	switch registration.Status {
	case model.ContestRegistrationStatusApproved:
	case model.ContestRegistrationStatusPending:
		return practiceports.InstanceScope{}, errcode.ErrContestRegistrationPending
	default:
		return practiceports.InstanceScope{}, errcode.ErrRegistrationNotApproved
	}

	contestIDCopy := contestID
	scope := practiceports.InstanceScope{
		ContestID:     &contestIDCopy,
		ContestMode:   contest.Mode,
		FlagSubjectID: userID,
		ShareScope:    model.InstanceSharingPerUser,
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
	effective.ShareScope = model.InstanceSharingPerUser

	switch {
	case scope.ContestMode == model.ContestModeAWD:
		effective.ShareScope = model.InstanceSharingPerTeam
		if scope.TeamID != nil && *scope.TeamID > 0 {
			effective.FlagSubjectID = *scope.TeamID
		}
	case chal.InstanceSharing == model.InstanceSharingShared:
		effective.ShareScope = model.InstanceSharingShared
		effective.TeamID = nil
	case chal.InstanceSharing == model.InstanceSharingPerTeam && scope.TeamID != nil && *scope.TeamID > 0:
		effective.ShareScope = model.InstanceSharingPerTeam
		effective.FlagSubjectID = *scope.TeamID
	default:
		effective.ShareScope = model.InstanceSharingPerUser
		effective.TeamID = nil
	}

	if effective.ShareScope != model.InstanceSharingPerTeam {
		effective.TeamID = nil
	}
	return effective
}
