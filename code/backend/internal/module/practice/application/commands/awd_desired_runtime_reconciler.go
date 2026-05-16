package commands

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
)

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
			scopeKey := awdDesiredRuntimeScopeKey(team.ID, service.ID)
			if _, ok := activeScopes[scopeKey]; ok {
				continue
			}
			if err := s.ensureDesiredAdminContestAWDTeamService(ctx, contest, team, service); err != nil {
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
		OwnerUserID: ownerUserID,
		ContestID:   contest.ID,
		ChallengeID: challengeID,
		Scope:       scope,
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
