package commands

import (
	"context"
	"errors"
	"fmt"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/pkg/errcode"
)

type awdScopeControlSpec struct {
	ContestID   int64
	TeamID      int64
	ServiceID   int64
	ScopeType   string
	ControlType string
}

func (s *Service) SetAdminContestAWDTeamRetired(ctx context.Context, contestID, teamID, actorUserID int64, retired bool, reason string) (*dto.AdminAWDScopeControlResp, error) {
	contest, err := s.loadAdminContestAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if _, err := s.loadAdminContestAWDTeam(ctx, contestID, teamID); err != nil {
		return nil, err
	}

	spec := awdScopeControlSpec{
		ContestID:   contest.ID,
		TeamID:      teamID,
		ScopeType:   model.AWDScopeControlScopeTeam,
		ControlType: model.AWDScopeControlTypeRetired,
	}
	resp, err := s.setAWDScopeControl(ctx, spec, actorUserID, retired, reason)
	if err != nil {
		return nil, err
	}
	if retired {
		s.clearDesiredAWDReconcileFailuresForTeam(ctx, contest.ID, teamID)
		if err := s.stopContestAWDActiveInstances(ctx, contest.ID, func(instance *model.Instance) bool {
			return instance != nil && instance.TeamID != nil && *instance.TeamID == teamID
		}); err != nil {
			return nil, err
		}
	}
	return resp, nil
}

func (s *Service) SetAdminContestAWDTeamServiceDisabled(ctx context.Context, contestID, teamID, serviceID, actorUserID int64, disabled bool, reason string) (*dto.AdminAWDScopeControlResp, error) {
	contest, err := s.validateAdminContestAWDServiceControlScope(ctx, contestID, teamID, serviceID)
	if err != nil {
		return nil, err
	}

	spec := awdScopeControlSpec{
		ContestID:   contest.ID,
		TeamID:      teamID,
		ServiceID:   serviceID,
		ScopeType:   model.AWDScopeControlScopeTeamService,
		ControlType: model.AWDScopeControlTypeServiceDisabled,
	}
	resp, err := s.setAWDScopeControl(ctx, spec, actorUserID, disabled, reason)
	if err != nil {
		return nil, err
	}
	if disabled {
		s.clearDesiredAWDReconcileFailure(ctx, contest.ID, teamID, serviceID)
		if err := s.stopContestAWDActiveInstances(ctx, contest.ID, func(instance *model.Instance) bool {
			return instance != nil &&
				instance.TeamID != nil && *instance.TeamID == teamID &&
				instance.ServiceID != nil && *instance.ServiceID == serviceID
		}); err != nil {
			return nil, err
		}
	}
	return resp, nil
}

func (s *Service) SetAdminContestAWDDesiredReconcileSuppressed(ctx context.Context, contestID, teamID, serviceID, actorUserID int64, suppressed bool, reason string) (*dto.AdminAWDScopeControlResp, error) {
	contest, err := s.validateAdminContestAWDServiceControlScope(ctx, contestID, teamID, serviceID)
	if err != nil {
		return nil, err
	}

	spec := awdScopeControlSpec{
		ContestID:   contest.ID,
		TeamID:      teamID,
		ServiceID:   serviceID,
		ScopeType:   model.AWDScopeControlScopeTeamService,
		ControlType: model.AWDScopeControlTypeDesiredReconcileSuppressed,
	}
	resp, err := s.setAWDScopeControl(ctx, spec, actorUserID, suppressed, reason)
	if err != nil {
		return nil, err
	}
	if suppressed {
		s.clearDesiredAWDReconcileFailure(ctx, contest.ID, teamID, serviceID)
	}
	return resp, nil
}

func (s *Service) setAWDScopeControl(ctx context.Context, spec awdScopeControlSpec, actorUserID int64, enabled bool, reason string) (*dto.AdminAWDScopeControlResp, error) {
	if s == nil || s.repo == nil {
		return nil, errcode.ErrInternal.WithCause(fmt.Errorf("practice awd scope control repository is nil"))
	}
	if spec.ContestID <= 0 || spec.TeamID <= 0 || spec.ScopeType == "" || spec.ControlType == "" {
		return nil, errcode.ErrInvalidParams
	}

	if enabled {
		control := &model.AWDScopeControl{
			ContestID:   spec.ContestID,
			TeamID:      spec.TeamID,
			ScopeType:   spec.ScopeType,
			ServiceID:   spec.ServiceID,
			ControlType: spec.ControlType,
			Reason:      normalizeAWDScopeControlReason(reason),
			UpdatedBy:   &actorUserID,
		}
		if err := s.repo.UpsertAWDScopeControl(ctx, control); err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
	} else {
		if err := s.repo.DeleteAWDScopeControl(ctx, spec.ContestID, spec.TeamID, spec.ScopeType, spec.ControlType, spec.ServiceID); err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
	}

	rows, err := s.repo.ListScopeAWDScopeControls(ctx, spec.ContestID, spec.TeamID, spec.ServiceID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	row := findAWDScopeControlRow(rows, spec.ScopeType, spec.ControlType, spec.ServiceID)
	return adminAWDScopeControlRespFromModel(spec, row), nil
}

func adminAWDScopeControlRespFromModel(spec awdScopeControlSpec, row *model.AWDScopeControl) *dto.AdminAWDScopeControlResp {
	resp := &dto.AdminAWDScopeControlResp{
		ScopeType:   spec.ScopeType,
		ControlType: spec.ControlType,
		TeamID:      spec.TeamID,
		Enabled:     row != nil,
	}
	if spec.ScopeType == model.AWDScopeControlScopeTeamService && spec.ServiceID > 0 {
		serviceID := spec.ServiceID
		resp.ServiceID = &serviceID
	}
	if row != nil {
		resp.Reason = row.Reason
		resp.UpdatedBy = row.UpdatedBy
		resp.UpdatedAt = &row.UpdatedAt
	}
	return resp
}

func findAWDScopeControlRow(rows []*model.AWDScopeControl, scopeType, controlType string, serviceID int64) *model.AWDScopeControl {
	for _, row := range rows {
		if row == nil {
			continue
		}
		if row.ScopeType == scopeType && row.ControlType == controlType && row.ServiceID == serviceID {
			return row
		}
	}
	return nil
}

func (s *Service) validateAdminContestAWDServiceControlScope(ctx context.Context, contestID, teamID, serviceID int64) (*model.Contest, error) {
	contest, err := s.loadAdminContestAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if _, _, _, err := s.resolveAdminContestAWDServiceInstanceScopeWithContest(ctx, contest, contestID, teamID, serviceID); err != nil {
		return nil, err
	}
	return contest, nil
}

func (s *Service) loadAdminContestAWDTeam(ctx context.Context, contestID, teamID int64) (*model.Team, error) {
	if s.contestScope == nil {
		return nil, errcode.ErrInternal.WithCause(fmt.Errorf("practice contest scope repository is nil"))
	}
	team, err := s.contestScope.FindContestTeam(ctx, contestID, teamID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeContestTeamNotFound) {
			return nil, errcode.ErrTeamNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return team, nil
}

func (s *Service) stopContestAWDActiveInstances(ctx context.Context, contestID int64, match func(instance *model.Instance) bool) error {
	if s == nil || s.repo == nil || s.instanceRepo == nil {
		return nil
	}
	instances, err := s.repo.ListContestAWDInstances(ctx, contestID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	for _, instance := range instances {
		if !match(instance) {
			continue
		}
		if s.runtimeService != nil {
			if err := s.runtimeService.CleanupRuntime(ctx, instance); err != nil {
				return errcode.ErrServiceUnavailable.WithCause(err)
			}
		}
		if err := s.instanceRepo.UpdateStatusAndReleasePort(ctx, instance.ID, model.InstanceStatusStopped); err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
	}
	return nil
}

func (s *Service) clearDesiredAWDReconcileFailuresForTeam(ctx context.Context, contestID, teamID int64) {
	if s == nil || s.repo == nil || s.desiredState == nil {
		return
	}
	services, err := s.repo.ListContestAWDServices(ctx, contestID)
	if err != nil {
		return
	}
	for _, service := range services {
		if service == nil || service.ID <= 0 {
			continue
		}
		s.clearDesiredAWDReconcileFailure(ctx, contestID, teamID, service.ID)
	}
}
