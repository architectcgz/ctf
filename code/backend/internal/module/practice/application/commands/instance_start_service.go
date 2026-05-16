package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/practice/domain"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/pkg/errcode"
)

const (
	adminAWDPrewarmOutcomeStarted = "started"
	adminAWDPrewarmOutcomeReused  = "reused"
	adminAWDPrewarmOutcomeFailed  = "failed"
)

func (s *Service) StartChallenge(ctx context.Context, userID, challengeID int64) (*dto.InstanceResp, error) {
	return s.startPersonalChallenge(ctx, userID, challengeID)
}

func (s *Service) StartContestChallenge(ctx context.Context, userID, contestID, challengeID int64) (*dto.InstanceResp, error) {
	scope, err := s.resolveContestChallengeInstanceScope(ctx, userID, contestID, challengeID)
	if err != nil {
		return nil, err
	}
	return s.startChallengeWithScope(ctx, userID, challengeID, scope)
}

func (s *Service) StartContestAWDService(ctx context.Context, userID, contestID, serviceID int64) (*dto.InstanceResp, error) {
	challengeID, scope, err := s.resolveContestAWDServiceInstanceScope(ctx, userID, contestID, serviceID)
	if err != nil {
		return nil, err
	}
	resp, err := s.startChallengeWithScope(ctx, userID, challengeID, scope)
	if err != nil {
		return nil, err
	}
	s.recordAWDServiceOperation(ctx, resp.ID, contestID, scope, model.AWDServiceOperationTypeStart, awdOperationStatusForInstanceStatus(resp.Status), model.AWDServiceOperationRequestedByUser, &userID, "user_start", true)
	return resp, nil
}

func (s *Service) RestartContestAWDService(ctx context.Context, userID, contestID, serviceID int64) (*dto.InstanceResp, error) {
	challengeID, scope, err := s.resolveContestAWDServiceInstanceScope(ctx, userID, contestID, serviceID)
	if err != nil {
		return nil, err
	}
	scope = resolveEffectiveInstanceScope(&model.Challenge{}, scope)

	var instance *model.Instance
	if err := s.repo.WithinInstanceRestartTx(ctx, func(txRepo practiceports.PracticeInstanceRestartTxRepository) error {
		if err := txRepo.LockInstanceScope(ctx, userID, challengeID, scope); err != nil {
			return err
		}
		existing, err := txRepo.FindScopedRestartableInstance(ctx, userID, challengeID, scope)
		if err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		instance = existing
		return nil
	}); err != nil {
		return nil, err
	}

	if instance == nil {
		resp, err := s.startChallengeWithScope(ctx, userID, challengeID, scope)
		if err != nil {
			return nil, err
		}
		s.recordAWDServiceOperation(ctx, resp.ID, contestID, scope, model.AWDServiceOperationTypeRestart, awdOperationStatusForInstanceStatus(resp.Status), model.AWDServiceOperationRequestedByUser, &userID, "user_restart", true)
		return resp, nil
	}
	if instance.Status == model.InstanceStatusPending || instance.Status == model.InstanceStatusCreating {
		return instanceRespForScope(instance, scope, s.config.Container.PublicHost, s.config.Container.AccessHost), nil
	}

	if err := s.runtimeService.CleanupRuntime(ctx, restartCleanupRuntimeView(instance)); err != nil {
		return nil, errcode.ErrServiceUnavailable.WithCause(err)
	}

	nextStatus := model.InstanceStatusCreating
	if s.schedulerEnabled() {
		nextStatus = model.InstanceStatusPending
	}
	nextExpiresAt, err := s.resolveInstanceExpiresAt(ctx, scope)
	if err != nil {
		return nil, err
	}
	preserveHostPort := requiresPublishedHostPort(scope, s.config.Container.AccessHost)
	if err := s.repo.WithinInstanceRestartTx(ctx, func(txRepo practiceports.PracticeInstanceRestartTxRepository) error {
		if err := txRepo.LockInstanceScope(ctx, userID, challengeID, scope); err != nil {
			return err
		}
		if preserveHostPort && instance.HostPort <= 0 {
			hostPort, err := txRepo.ReserveAvailablePort(ctx, s.config.Container.PortRangeStart, s.config.Container.PortRangeEnd)
			if err != nil {
				return errcode.ErrInternal.WithCause(err)
			}
			if err := txRepo.BindReservedPort(ctx, hostPort, instance.ID); err != nil {
				return errcode.ErrInternal.WithCause(err)
			}
			instance.HostPort = hostPort
		}
		if err := txRepo.ResetInstanceRuntimeForRestart(ctx, instance.ID, nextStatus, nextExpiresAt, preserveHostPort); err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		operationStatus := model.AWDServiceOperationStatusRequested
		if nextStatus == model.InstanceStatusPending {
			operationStatus = model.AWDServiceOperationStatusProvisioning
		}
		if err := createAWDServiceOperation(ctx, txRepo, instance.ID, contestID, scope, model.AWDServiceOperationTypeRestart, operationStatus, model.AWDServiceOperationRequestedByUser, &userID, "user_restart", true); err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	instance.ContainerID = ""
	instance.NetworkID = ""
	instance.RuntimeDetails = ""
	instance.AccessURL = ""
	if !preserveHostPort {
		instance.HostPort = 0
	}
	instance.Status = nextStatus
	instance.ExpiresAt = nextExpiresAt
	instance.DestroyedAt = nil
	if !s.schedulerEnabled() {
		chal, topology, err := s.loadRuntimeSubjectWithScope(ctx, scope, challengeID)
		if err != nil {
			return nil, err
		}
		flag, err := s.buildProvisioningFlag(instance, chal)
		if err != nil {
			return nil, err
		}
		if err := s.provisionInstance(ctx, instance, chal, topology, flag); err != nil {
			return nil, err
		}
	}
	return instanceRespForScope(instance, scope, s.config.Container.PublicHost, s.config.Container.AccessHost), nil
}

func (s *Service) StartAdminContestAWDTeamService(ctx context.Context, contestID, teamID, serviceID int64) (*dto.AdminAWDInstanceItemResp, error) {
	challengeID, ownerUserID, scope, err := s.resolveAdminContestAWDServiceInstanceScope(ctx, contestID, teamID, serviceID)
	if err != nil {
		return nil, err
	}
	instance, err := s.startChallengeWithScope(ctx, ownerUserID, challengeID, scope)
	if err != nil {
		return nil, err
	}
	return &dto.AdminAWDInstanceItemResp{
		TeamID:    teamID,
		ServiceID: serviceID,
		Instance:  instance,
	}, nil
}

func (s *Service) PrewarmAdminContestAWDInstances(ctx context.Context, contestID int64, req *dto.PrewarmAdminContestAWDInstancesReq) (*dto.AdminAWDInstancePrewarmResp, error) {
	contest, err := s.loadAdminContestAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if contest.Status == model.ContestStatusEnded {
		return nil, errcode.ErrContestEnded
	}
	if contest.Status != model.ContestStatusRegistration {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("awd 赛前预热仅支持报名阶段"))
	}
	if req == nil {
		req = &dto.PrewarmAdminContestAWDInstancesReq{}
	}

	teams, err := s.resolveAdminContestAWDPrewarmTeams(ctx, contestID, req.TeamID)
	if err != nil {
		return nil, err
	}
	services, err := s.repo.ListContestAWDServices(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	existingInstances, err := s.repo.ListContestAWDInstances(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	resp := &dto.AdminAWDInstancePrewarmResp{
		ContestID: contestID,
		Results:   make([]*dto.AdminAWDInstancePrewarmItemResp, 0, len(teams)*len(services)),
	}
	for _, team := range teams {
		for _, service := range services {
			if service == nil || !service.IsVisible {
				continue
			}
			item := s.prewarmAdminContestAWDTeamService(
				ctx,
				contestID,
				team,
				service,
				existingInstances,
			)
			resp.Results = append(resp.Results, item)
			resp.Summary.Total++
			switch item.Outcome {
			case adminAWDPrewarmOutcomeStarted:
				resp.Summary.Started++
			case adminAWDPrewarmOutcomeReused:
				resp.Summary.Reused++
			default:
				resp.Summary.Failed++
			}
		}
	}
	return resp, nil
}

func (s *Service) resolveAdminContestAWDPrewarmTeams(ctx context.Context, contestID int64, teamID *int64) ([]*model.Team, error) {
	if teamID != nil {
		team, err := s.repo.FindContestTeam(ctx, contestID, *teamID)
		if err != nil {
			if errors.Is(err, practiceports.ErrPracticeContestTeamNotFound) {
				return nil, errcode.ErrTeamNotFound
			}
			return nil, errcode.ErrInternal.WithCause(err)
		}
		return []*model.Team{team}, nil
	}

	teams, err := s.repo.ListContestTeams(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return teams, nil
}

func (s *Service) prewarmAdminContestAWDTeamService(ctx context.Context, contestID int64, team *model.Team, service *model.ContestAWDService, existingInstances []*model.Instance) *dto.AdminAWDInstancePrewarmItemResp {
	result := &dto.AdminAWDInstancePrewarmItemResp{
		Outcome: adminAWDPrewarmOutcomeFailed,
	}
	if team != nil {
		result.TeamID = team.ID
	}
	if service != nil {
		result.ServiceID = service.ID
	}
	if team == nil || team.ID <= 0 {
		result.ErrorMessage = "队伍不存在"
		return result
	}
	if service == nil || service.ID <= 0 {
		result.ErrorMessage = "服务不存在"
		return result
	}
	if team.CaptainID <= 0 {
		result.ErrorMessage = "队伍缺少队长用户"
		return result
	}
	for _, instance := range existingInstances {
		if instance == nil || instance.TeamID == nil || instance.ServiceID == nil {
			continue
		}
		if *instance.TeamID != team.ID || *instance.ServiceID != service.ID {
			continue
		}
		result.Outcome = adminAWDPrewarmOutcomeReused
		result.Instance = instanceRespForScope(instance, practiceports.InstanceScope{
			ContestMode: model.ContestModeAWD,
		}, s.config.Container.PublicHost, s.config.Container.AccessHost)
		return result
	}

	challengeID, ownerUserID, scope, err := s.resolveAdminContestAWDServiceInstanceScope(ctx, contestID, team.ID, service.ID)
	if err != nil {
		result.ErrorMessage = err.Error()
		return result
	}

	instance, err := s.startChallengeWithScope(ctx, ownerUserID, challengeID, scope)
	if err != nil {
		result.ErrorMessage = err.Error()
		return result
	}
	result.Instance = instance
	result.Outcome = adminAWDPrewarmOutcomeStarted
	return result
}

func (s *Service) startPersonalChallenge(ctx context.Context, userID, challengeID int64) (*dto.InstanceResp, error) {
	return s.startChallengeWithScope(ctx, userID, challengeID, practiceports.InstanceScope{
		FlagSubjectID: userID,
		ShareScope:    model.InstanceSharingPerUser,
	})
}

func (s *Service) startChallengeWithScope(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*dto.InstanceResp, error) {
	chal, topology, err := s.loadRuntimeSubjectWithScope(ctx, scope, challengeID)
	if err != nil {
		return nil, err
	}
	if chal.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrChallengeNotPublish
	}
	if chal.ImageID == 0 {
		if topology == nil {
			return nil, errcode.ErrInvalidParams.WithCause(errors.New(errMsgChallengeNoTarget))
		}
	}
	scope = resolveEffectiveInstanceScope(chal, scope)
	expiresAt, err := s.resolveInstanceExpiresAt(ctx, scope)
	if err != nil {
		return nil, err
	}

	flag, nonce, err := s.buildInstanceFlag(scope.FlagSubjectID, challengeID, chal)
	if err != nil {
		return nil, err
	}

	var (
		instance *model.Instance
		reused   bool
	)
	initialStatus := model.InstanceStatusCreating
	if s.schedulerEnabled() {
		initialStatus = model.InstanceStatusPending
	}
	if err := s.repo.WithinInstanceStartTx(ctx, func(txRepo practiceports.PracticeInstanceStartTxRepository) error {
		if err := txRepo.LockInstanceScope(ctx, userID, challengeID, scope); err != nil {
			return err
		}

		existingInstance, err := txRepo.FindScopedExistingInstance(ctx, userID, challengeID, scope)
		if err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		if existingInstance != nil {
			if scope.ContestMode == model.ContestModeAWD {
				if !existingInstance.ExpiresAt.Equal(expiresAt) {
					if err := txRepo.RefreshInstanceExpiry(ctx, existingInstance.ID, expiresAt); err != nil {
						return errcode.ErrInternal.WithCause(err)
					}
					existingInstance.ExpiresAt = expiresAt
				}
			} else if scope.ShareScope == model.InstanceSharingShared {
				refreshedExpiry := existingInstance.ExpiresAt
				if expiresAt.After(refreshedExpiry) {
					refreshedExpiry = expiresAt
				}
				if err := txRepo.RefreshInstanceExpiry(ctx, existingInstance.ID, refreshedExpiry); err != nil {
					return errcode.ErrInternal.WithCause(err)
				}
				existingInstance.ExpiresAt = refreshedExpiry
			}
			instance = existingInstance
			reused = true
			return nil
		}

		runningCount, err := txRepo.CountScopedRunningInstances(ctx, userID, scope)
		if err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		if runningCount >= s.config.Container.MaxConcurrentPerUser {
			s.logger.Warn("实例数量超限",
				zap.Int64("user_id", userID),
				zap.Int64("challenge_id", challengeID),
				zap.Int("current", runningCount),
				zap.Int("limit", s.config.Container.MaxConcurrentPerUser))
			return errcode.ErrInstanceLimitExceeded
		}

		hostPort := 0
		if requiresPublishedHostPort(scope, s.config.Container.AccessHost) {
			var err error
			hostPort, err = txRepo.ReserveAvailablePort(ctx, s.config.Container.PortRangeStart, s.config.Container.PortRangeEnd)
			if err != nil {
				return errcode.ErrInternal.WithCause(err)
			}
		}

		instance = &model.Instance{
			UserID:      userID,
			ContestID:   scope.ContestID,
			TeamID:      scope.TeamID,
			ChallengeID: challengeID,
			ServiceID:   scope.ServiceID,
			HostPort:    hostPort,
			ShareScope:  scope.ShareScope,
			Status:      initialStatus,
			Nonce:       nonce,
			ExpiresAt:   expiresAt,
			MaxExtends:  s.config.Container.MaxExtends,
		}
		if err := txRepo.CreateInstance(ctx, instance); err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		if hostPort > 0 {
			if err := txRepo.BindReservedPort(ctx, hostPort, instance.ID); err != nil {
				return errcode.ErrInternal.WithCause(err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if reused {
		return instanceRespForScope(instance, scope, s.config.Container.PublicHost, s.config.Container.AccessHost), nil
	}
	if s.schedulerEnabled() {
		s.logger.Info("实例已入启动队列",
			zap.Int64("user_id", userID),
			zap.Int64("challenge_id", challengeID),
			zap.Int64("instance_id", instance.ID))
		return instanceRespForScope(instance, scope, s.config.Container.PublicHost, s.config.Container.AccessHost), nil
	}

	if err := s.provisionInstance(ctx, instance, chal, topology, flag); err != nil {
		return nil, err
	}
	return instanceRespForScope(instance, scope, s.config.Container.PublicHost, s.config.Container.AccessHost), nil
}

func instanceRespForScope(instance *model.Instance, scope practiceports.InstanceScope, publicHost, accessHost string) *dto.InstanceResp {
	resp := domain.InstanceRespFromModel(instance, publicHost, accessHost)
	if scope.ContestMode == model.ContestModeAWD {
		resp.AccessURL = ""
	}
	return resp
}

func (s *Service) resolveInstanceExpiresAt(ctx context.Context, scope practiceports.InstanceScope) (time.Time, error) {
	if scope.ContestMode != model.ContestModeAWD || scope.ContestID == nil || *scope.ContestID <= 0 {
		return time.Now().UTC().Add(s.config.Container.DefaultTTL), nil
	}
	if s.contestScope == nil {
		return time.Time{}, errcode.ErrInternal.WithCause(fmt.Errorf("practice contest scope repository is nil"))
	}

	contest, err := s.contestScope.FindContestByID(ctx, *scope.ContestID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeContestNotFound) {
			return time.Time{}, errcode.ErrContestNotFound
		}
		return time.Time{}, errcode.ErrInternal.WithCause(err)
	}
	if contest == nil {
		return time.Time{}, errcode.ErrContestNotFound
	}
	return practiceContestEffectiveEndTime(contest), nil
}

func practiceContestEffectiveEndTime(contest *model.Contest) time.Time {
	if contest == nil {
		return time.Time{}
	}
	return contest.EndTime.UTC().Add(time.Duration(contest.PausedSeconds) * time.Second)
}
