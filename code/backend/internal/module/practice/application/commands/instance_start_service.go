package commands

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/practice/domain"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/pkg/errcode"
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
		return instanceRespForScope(instance, scope), nil
	}

	if err := s.runtimeService.CleanupRuntime(ctx, restartCleanupRuntimeView(instance)); err != nil {
		return nil, errcode.ErrServiceUnavailable.WithCause(err)
	}

	nextStatus := model.InstanceStatusCreating
	if s.schedulerEnabled() {
		nextStatus = model.InstanceStatusPending
	}
	nextExpiresAt := time.Now().Add(s.config.Container.DefaultTTL)
	if err := s.repo.WithinInstanceRestartTx(ctx, func(txRepo practiceports.PracticeInstanceRestartTxRepository) error {
		if err := txRepo.LockInstanceScope(ctx, userID, challengeID, scope); err != nil {
			return err
		}
		if err := txRepo.ResetInstanceRuntimeForRestart(ctx, instance.ID, nextStatus, nextExpiresAt); err != nil {
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
	return instanceRespForScope(instance, scope), nil
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
			if scope.ShareScope == model.InstanceSharingShared {
				refreshedExpiry := existingInstance.ExpiresAt
				candidateExpiry := time.Now().Add(s.config.Container.DefaultTTL)
				if candidateExpiry.After(refreshedExpiry) {
					refreshedExpiry = candidateExpiry
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
		if requiresPublishedHostPort(scope) {
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
			ExpiresAt:   time.Now().Add(s.config.Container.DefaultTTL),
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
		return instanceRespForScope(instance, scope), nil
	}
	if s.schedulerEnabled() {
		s.logger.Info("实例已入启动队列",
			zap.Int64("user_id", userID),
			zap.Int64("challenge_id", challengeID),
			zap.Int64("instance_id", instance.ID))
		return instanceRespForScope(instance, scope), nil
	}

	if err := s.provisionInstance(ctx, instance, chal, topology, flag); err != nil {
		return nil, err
	}
	return instanceRespForScope(instance, scope), nil
}

func requiresPublishedHostPort(scope practiceports.InstanceScope) bool {
	return true
}

func instanceRespForScope(instance *model.Instance, scope practiceports.InstanceScope) *dto.InstanceResp {
	resp := domain.InstanceRespFromModel(instance)
	if scope.ContestMode == model.ContestModeAWD {
		resp.AccessURL = ""
	}
	return resp
}
