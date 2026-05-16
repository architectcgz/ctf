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

func (s *Service) recordAWDServiceOperation(ctx context.Context, instanceID, contestID int64, scope practiceports.InstanceScope, operationType, status, requestedBy string, requestedByID *int64, reason string, slaBillable bool) {
	if err := s.repo.WithinAWDServiceOperationTx(ctx, func(txRepo practiceports.PracticeAWDServiceOperationTxRepository) error {
		return createAWDServiceOperation(ctx, txRepo, instanceID, contestID, scope, operationType, status, requestedBy, requestedByID, reason, slaBillable)
	}); err != nil {
		s.logger.Warn("记录 AWD 服务操作失败",
			zap.Int64("contest_id", contestID),
			zap.Int64("instance_id", instanceID),
			zap.String("operation_type", operationType),
			zap.Error(err))
	}
}

func createAWDServiceOperation(ctx context.Context, repo practiceports.PracticeAWDServiceOperationCreateRepository, instanceID, contestID int64, scope practiceports.InstanceScope, operationType, status, requestedBy string, requestedByID *int64, reason string, slaBillable bool) error {
	if repo == nil || instanceID <= 0 || contestID <= 0 || scope.TeamID == nil || *scope.TeamID <= 0 || scope.ServiceID == nil || *scope.ServiceID <= 0 {
		return nil
	}
	now := time.Now().UTC()
	var finishedAt *time.Time
	if isFinishedAWDServiceOperationStatus(status) {
		finishedAt = &now
	}
	return repo.CreateAWDServiceOperation(ctx, &model.AWDServiceOperation{
		ContestID:     contestID,
		TeamID:        *scope.TeamID,
		ServiceID:     *scope.ServiceID,
		InstanceID:    instanceID,
		OperationType: operationType,
		RequestedBy:   requestedBy,
		RequestedByID: requestedByID,
		Reason:        reason,
		SLABillable:   slaBillable,
		Status:        status,
		StartedAt:     now,
		FinishedAt:    finishedAt,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
}

func awdOperationStatusForInstanceStatus(instanceStatus string) string {
	if instanceStatus == model.InstanceStatusRunning {
		return model.AWDServiceOperationStatusSucceeded
	}
	return model.AWDServiceOperationStatusProvisioning
}

func isFinishedAWDServiceOperationStatus(status string) bool {
	return status == model.AWDServiceOperationStatusSucceeded ||
		status == model.AWDServiceOperationStatusRecovered ||
		status == model.AWDServiceOperationStatusFailed
}

func restartCleanupRuntimeView(instance *model.Instance) *model.Instance {
	if instance == nil {
		return nil
	}
	copied := *instance
	copied.HostPort = 0
	details, err := model.DecodeInstanceRuntimeDetails(copied.RuntimeDetails)
	if err != nil {
		return &copied
	}
	for i := range details.Containers {
		details.Containers[i].HostPort = 0
	}
	if raw, err := model.EncodeInstanceRuntimeDetails(details); err == nil {
		copied.RuntimeDetails = raw
	}
	return &copied
}

func (s *Service) GetContestAWDInstanceOrchestration(ctx context.Context, contestID int64) (*dto.AdminAWDInstanceOrchestrationResp, error) {
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
	if contest.Mode != model.ContestModeAWD {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("仅 AWD 赛事支持队伍实例编排"))
	}

	teams, err := s.repo.ListContestTeams(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	services, err := s.repo.ListContestAWDServices(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	instances, err := s.repo.ListContestAWDInstances(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	controls, err := s.listContestAWDScopeControls(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	resp := &dto.AdminAWDInstanceOrchestrationResp{
		ContestID: contestID,
		Teams:     make([]*dto.AdminAWDInstanceTeamResp, 0, len(teams)),
		Services:  make([]*dto.AdminAWDInstanceServiceResp, 0, len(services)),
		Instances: make([]*dto.AdminAWDInstanceItemResp, 0, len(instances)),
		Controls:  make([]*dto.AdminAWDScopeControlResp, 0, len(controls)),
	}
	for _, team := range teams {
		resp.Teams = append(resp.Teams, &dto.AdminAWDInstanceTeamResp{
			TeamID:    team.ID,
			TeamName:  team.Name,
			CaptainID: team.CaptainID,
		})
	}
	for _, service := range services {
		resp.Services = append(resp.Services, &dto.AdminAWDInstanceServiceResp{
			ServiceID:      service.ID,
			AWDChallengeID: service.AWDChallengeID,
			DisplayName:    service.DisplayName,
			IsVisible:      service.IsVisible,
		})
	}
	seen := make(map[string]struct{}, len(instances))
	for _, instance := range instances {
		if instance.TeamID == nil || instance.ServiceID == nil {
			continue
		}
		key := fmt.Sprintf("%d:%d", *instance.TeamID, *instance.ServiceID)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		resp.Instances = append(resp.Instances, &dto.AdminAWDInstanceItemResp{
			TeamID:    *instance.TeamID,
			ServiceID: *instance.ServiceID,
			Instance:  domain.InstanceRespFromModel(instance, s.config.Container.PublicHost, s.config.Container.AccessHost),
		})
	}
	for _, control := range controls {
		if control == nil {
			continue
		}
		resp.Controls = append(resp.Controls, adminAWDScopeControlRecordResp(control))
	}
	return resp, nil
}

func adminAWDScopeControlRecordResp(control *model.AWDScopeControl) *dto.AdminAWDScopeControlResp {
	if control == nil {
		return nil
	}
	resp := &dto.AdminAWDScopeControlResp{
		ScopeType:   control.ScopeType,
		ControlType: control.ControlType,
		TeamID:      control.TeamID,
		Enabled:     true,
		Reason:      control.Reason,
		UpdatedBy:   control.UpdatedBy,
		UpdatedAt:   &control.UpdatedAt,
	}
	if control.ScopeType == model.AWDScopeControlScopeTeamService && control.ServiceID > 0 {
		serviceID := control.ServiceID
		resp.ServiceID = &serviceID
	}
	return resp
}
