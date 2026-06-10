package composition

import (
	"context"
	"time"

	contestentity "ctf-platform/internal/module/contest/entity"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	instanceentity "ctf-platform/internal/module/instance/entity"
	instanceports "ctf-platform/internal/module/instance/ports"
)

type instanceProxyTicketReaderAdapter struct {
	instances instanceports.InstanceLookupRepository
	awdScopes awdProxyScopeReader
}

type awdProxyScopeReader interface {
	FindAWDTargetProxyScope(ctx context.Context, userID, contestID, serviceID, victimTeamID int64) (*instanceports.AWDTargetProxyScope, error)
	FindAWDDefenseSSHScope(ctx context.Context, userID, contestID, serviceID int64) (*instanceports.AWDDefenseSSHScope, error)
}

func newInstanceProxyTicketReader(instances instanceports.InstanceLookupRepository, awdScopes awdProxyScopeReader) *instanceProxyTicketReaderAdapter {
	if instances == nil || awdScopes == nil {
		return nil
	}
	return &instanceProxyTicketReaderAdapter{instances: instances, awdScopes: awdScopes}
}

func (a *instanceProxyTicketReaderAdapter) FindByID(ctx context.Context, id int64) (*instanceentity.Instance, error) {
	if a == nil || a.instances == nil {
		return nil, nil
	}
	return a.instances.FindByID(ctx, id)
}

func (a *instanceProxyTicketReaderAdapter) FindAWDTargetProxyScope(ctx context.Context, userID, contestID, serviceID, victimTeamID int64) (*instanceports.AWDTargetProxyScope, error) {
	if a == nil || a.awdScopes == nil {
		return nil, nil
	}
	return a.awdScopes.FindAWDTargetProxyScope(ctx, userID, contestID, serviceID, victimTeamID)
}

func (a *instanceProxyTicketReaderAdapter) FindAWDDefenseSSHScope(ctx context.Context, userID, contestID, serviceID int64) (*instanceports.AWDDefenseSSHScope, error) {
	if a == nil || a.awdScopes == nil {
		return nil, nil
	}
	return a.awdScopes.FindAWDDefenseSSHScope(ctx, userID, contestID, serviceID)
}

type instanceProxyTrafficRecorderAdapter struct {
	awdRepo *contestinfra.AWDRepository
}

func newInstanceProxyTrafficRecorder(awdRepo *contestinfra.AWDRepository) *instanceProxyTrafficRecorderAdapter {
	if awdRepo == nil {
		return nil
	}
	return &instanceProxyTrafficRecorderAdapter{awdRepo: awdRepo}
}

func (a *instanceProxyTrafficRecorderAdapter) RecordRuntimeProxyTrafficEvent(ctx context.Context, instanceID, userID int64, method, requestPath string, statusCode int) error {
	if a == nil || a.awdRepo == nil {
		return nil
	}
	return a.awdRepo.RecordRuntimeProxyTrafficEvent(ctx, instanceID, userID, method, requestPath, statusCode)
}

func (a *instanceProxyTrafficRecorderAdapter) RecordAWDProxyTrafficEvent(ctx context.Context, event instanceports.AWDProxyTrafficEventInput) error {
	if a == nil || a.awdRepo == nil {
		return nil
	}
	return a.awdRepo.RecordAWDProxyTrafficEvent(ctx, contestentity.AWDProxyTrafficEventInput{
		ContestID:      event.ContestID,
		AttackerTeamID: event.AttackerTeamID,
		VictimTeamID:   event.VictimTeamID,
		ServiceID:      event.ServiceID,
		AWDChallengeID: event.AWDChallengeID,
		Method:         event.Method,
		Path:           event.Path,
		StatusCode:     event.StatusCode,
	})
}

type startupRuntimeContestRepositoryAdapter struct {
	repo *contestinfra.Repository
}

func newStartupRuntimeContestRepository(repo *contestinfra.Repository) *startupRuntimeContestRepositoryAdapter {
	if repo == nil {
		return nil
	}
	return &startupRuntimeContestRepositoryAdapter{repo: repo}
}

func (a *startupRuntimeContestRepositoryAdapter) AddPausedDurationToActiveAWDContests(ctx context.Context, activeAt time.Time, recoveryKey string, targetPausedSeconds int64, updatedAt time.Time) ([]*instanceports.ActiveAWDContestPause, error) {
	if a == nil || a.repo == nil {
		return nil, nil
	}
	contests, err := a.repo.AddPausedDurationToActiveAWDContests(ctx, activeAt, recoveryKey, targetPausedSeconds, updatedAt)
	if err != nil {
		return nil, err
	}
	result := make([]*instanceports.ActiveAWDContestPause, 0, len(contests))
	for _, contest := range contests {
		if contest == nil {
			continue
		}
		result = append(result, &instanceports.ActiveAWDContestPause{
			ID:            contest.ID,
			EndTime:       contest.EndTime,
			PausedSeconds: contest.PausedSeconds,
		})
	}
	return result, nil
}
