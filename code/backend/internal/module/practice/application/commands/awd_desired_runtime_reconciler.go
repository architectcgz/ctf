package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
)

const desiredAWDReconcileLastErrorLimit = 256

func (s *Service) ReconcileDesiredAWDInstances(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	contests, err := s.repo.ListDesiredRuntimeAWDContests(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	for _, contest := range contests {
		if contest == nil || !practiceContestEffectiveEndTime(contest).After(now) {
			continue
		}
		if err := s.reconcileDesiredAWDContest(ctx, contest, now); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) reconcileDesiredAWDContest(ctx context.Context, contest *model.Contest, activeAt time.Time) error {
	teams, err := s.repo.ListContestTeams(ctx, contest.ID)
	if err != nil {
		return err
	}
	services, err := s.repo.ListContestAWDServices(ctx, contest.ID)
	if err != nil {
		return err
	}
	instances, err := s.repo.ListContestAWDInstances(ctx, contest.ID)
	if err != nil {
		return err
	}
	controls, err := s.listContestAWDScopeControls(ctx, contest.ID)
	if err != nil {
		return err
	}
	controlIndex := buildAWDContestControlIndex(controls)

	activeScopes := make(map[string]struct{}, len(instances))
	for _, instance := range instances {
		if !isDesiredAWDInstanceActive(instance, activeAt) {
			continue
		}
		activeScopes[awdDesiredRuntimeScopeKey(*instance.TeamID, *instance.ServiceID)] = struct{}{}
	}

	for _, team := range teams {
		if team == nil || team.ID <= 0 {
			continue
		}
		for _, service := range services {
			if service == nil || service.ID <= 0 || !service.IsVisible {
				continue
			}
			controlState := controlIndex.state(team.ID, service.ID)
			if controlState.suppressesDesiredReconcile() {
				s.clearDesiredAWDReconcileFailure(ctx, contest.ID, team.ID, service.ID)
				continue
			}
			scopeKey := awdDesiredRuntimeScopeKey(team.ID, service.ID)
			if _, ok := activeScopes[scopeKey]; ok {
				s.clearDesiredAWDReconcileFailure(ctx, contest.ID, team.ID, service.ID)
				continue
			}
			shouldAttempt, err := s.shouldAttemptDesiredAWDReconcile(ctx, contest.ID, team.ID, service.ID, activeAt)
			if err != nil {
				return err
			}
			if !shouldAttempt {
				continue
			}
			if err := s.ensureDesiredAdminContestAWDTeamService(ctx, contest, team, service); err != nil {
				s.recordDesiredAWDReconcileFailure(ctx, contest.ID, team.ID, service.ID, err, activeAt)
				s.logger.Warn("reconcile_desired_awd_team_service_failed",
					zap.Int64("contest_id", contest.ID),
					zap.Int64("team_id", team.ID),
					zap.Int64("service_id", service.ID),
					zap.Error(err))
				continue
			}
			activeScopes[scopeKey] = struct{}{}
		}
	}
	return nil
}

func (s *Service) ensureDesiredAdminContestAWDTeamService(ctx context.Context, contest *model.Contest, team *model.Team, service *model.ContestAWDService) error {
	if contest == nil || team == nil || service == nil {
		return nil
	}
	challengeID, ownerUserID, scope, err := s.resolveAdminContestAWDServiceInstanceScopeWithContest(ctx, contest, contest.ID, team.ID, service.ID)
	if err != nil {
		return err
	}

	_, err = s.restartOrStartScopedAWDService(ctx, awdScopedRuntimeRequest{
		OwnerUserID:  ownerUserID,
		ContestID:    contest.ID,
		ChallengeID:  challengeID,
		Scope:        scope,
		NoopIfActive: true,
		Audit: awdScopedRuntimeAudit{
			StartOperationType:   model.AWDServiceOperationTypeStart,
			RestartOperationType: model.AWDServiceOperationTypeRecreate,
			RequestedBy:          model.AWDServiceOperationRequestedBySystem,
			Reason:               "desired_runtime_reconcile",
			SLABillable:          false,
		},
	})
	return err
}

func isDesiredAWDInstanceActive(instance *model.Instance, activeAt time.Time) bool {
	if instance == nil || instance.TeamID == nil || instance.ServiceID == nil {
		return false
	}
	if activeAt.IsZero() {
		activeAt = time.Now().UTC()
	}
	if !instance.ExpiresAt.After(activeAt) {
		return false
	}
	switch instance.Status {
	case model.InstanceStatusPending, model.InstanceStatusCreating, model.InstanceStatusRunning:
		return true
	default:
		return false
	}
}

func awdDesiredRuntimeScopeKey(teamID, serviceID int64) string {
	return fmt.Sprintf("%d:%d", teamID, serviceID)
}

func (s *Service) shouldAttemptDesiredAWDReconcile(ctx context.Context, contestID, teamID, serviceID int64, now time.Time) (bool, error) {
	if s == nil || s.desiredState == nil || contestID <= 0 || teamID <= 0 || serviceID <= 0 {
		return true, nil
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	state, exists, err := s.desiredState.LoadDesiredAWDReconcileState(ctx, contestID, teamID, serviceID)
	if err != nil {
		s.logger.Warn("load_desired_awd_reconcile_state_failed",
			zap.Int64("contest_id", contestID),
			zap.Int64("team_id", teamID),
			zap.Int64("service_id", serviceID),
			zap.Error(err))
		return true, nil
	}
	if !exists || state == nil {
		return true, nil
	}
	if !state.SuppressedUntil.IsZero() && state.SuppressedUntil.After(now) {
		return false, nil
	}
	if !state.NextAttemptAt.IsZero() && state.NextAttemptAt.After(now) {
		return false, nil
	}
	return true, nil
}

func (s *Service) recordDesiredAWDReconcileFailure(ctx context.Context, contestID, teamID, serviceID int64, cause error, failedAt time.Time) {
	if s == nil || s.desiredState == nil || contestID <= 0 || teamID <= 0 || serviceID <= 0 {
		return
	}
	if failedAt.IsZero() {
		failedAt = time.Now().UTC()
	}

	nextState := &practiceports.DesiredAWDReconcileState{
		FailureCount:  1,
		LastFailureAt: failedAt.UTC(),
		LastError:     truncateDesiredAWDReconcileError(cause),
	}
	if current, exists, err := s.desiredState.LoadDesiredAWDReconcileState(ctx, contestID, teamID, serviceID); err != nil {
		s.logger.Warn("load_desired_awd_reconcile_state_failed",
			zap.Int64("contest_id", contestID),
			zap.Int64("team_id", teamID),
			zap.Int64("service_id", serviceID),
			zap.Error(err))
	} else if exists && current != nil && current.FailureCount > 0 {
		nextState.FailureCount = current.FailureCount + 1
	}

	nextState.NextAttemptAt = failedAt.Add(s.desiredAWDReconcileFailureBackoff(nextState.FailureCount))
	if threshold := s.desiredAWDReconcileSuppressAfterFailures(); threshold > 0 && nextState.FailureCount >= threshold {
		nextState.SuppressedUntil = failedAt.Add(s.desiredAWDReconcileSuppressDuration())
		if nextState.NextAttemptAt.Before(nextState.SuppressedUntil) {
			nextState.NextAttemptAt = nextState.SuppressedUntil
		}
	}

	if err := s.desiredState.StoreDesiredAWDReconcileState(ctx, contestID, teamID, serviceID, nextState); err != nil {
		s.logger.Warn("store_desired_awd_reconcile_state_failed",
			zap.Int64("contest_id", contestID),
			zap.Int64("team_id", teamID),
			zap.Int64("service_id", serviceID),
			zap.Error(err))
	}
}

func (s *Service) clearDesiredAWDReconcileFailure(ctx context.Context, contestID, teamID, serviceID int64) {
	if s == nil || s.desiredState == nil || contestID <= 0 || teamID <= 0 || serviceID <= 0 {
		return
	}
	if err := s.desiredState.DeleteDesiredAWDReconcileState(ctx, contestID, teamID, serviceID); err != nil {
		s.logger.Warn("delete_desired_awd_reconcile_state_failed",
			zap.Int64("contest_id", contestID),
			zap.Int64("team_id", teamID),
			zap.Int64("service_id", serviceID),
			zap.Error(err))
	}
}

func (s *Service) desiredAWDReconcileFailureBackoff(failureCount int) time.Duration {
	if failureCount <= 1 {
		return s.desiredAWDReconcileFailureInitialBackoff()
	}
	backoff := s.desiredAWDReconcileFailureInitialBackoff()
	maxBackoff := s.desiredAWDReconcileFailureMaxBackoff()
	for attempt := 1; attempt < failureCount; attempt++ {
		if backoff >= maxBackoff {
			return maxBackoff
		}
		backoff *= 2
		if backoff >= maxBackoff {
			return maxBackoff
		}
	}
	return backoff
}

func truncateDesiredAWDReconcileError(cause error) string {
	if cause == nil {
		return ""
	}
	message := strings.TrimSpace(cause.Error())
	if len(message) <= desiredAWDReconcileLastErrorLimit {
		return message
	}
	return message[:desiredAWDReconcileLastErrorLimit]
}
