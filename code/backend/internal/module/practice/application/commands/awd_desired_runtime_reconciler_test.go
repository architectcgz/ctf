package commands

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
	rediskeys "ctf-platform/internal/pkg/redis"
)

func TestReconcileDesiredAWDInstancesCreatesMissingInstance(t *testing.T) {
	t.Parallel()

	contestID := int64(3201)
	teamID := int64(4201)
	serviceID := int64(5201)
	challengeID := int64(6201)
	contestEnd := time.Date(2027, 5, 16, 14, 0, 0, 0, time.UTC)
	contest := &model.Contest{
		ID:      contestID,
		Mode:    model.ContestModeAWD,
		Status:  model.ContestStatusRunning,
		EndTime: contestEnd,
	}
	team := &model.Team{ID: teamID, ContestID: contestID, CaptainID: 7201, Name: "team-a"}
	serviceDef := &model.ContestAWDService{
		ID:              serviceID,
		ContestID:       contestID,
		AWDChallengeID:  challengeID,
		IsVisible:       true,
		ServiceSnapshot: `{"name":"awd-web","category":"web","difficulty":"medium","runtime_config":{"image_id":101,"instance_sharing":"per_team"},"flag_config":{"flag_type":"static","flag_prefix":"flag"}}`,
	}

	var createdInstance *model.Instance
	var operation *model.AWDServiceOperation
	findTeamCalled := false
	findServiceCalled := false
	findExistingCalled := false
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, gotContestID int64) (*model.Contest, error) {
			if gotContestID != contestID {
				t.Fatalf("unexpected contest lookup id: %d", gotContestID)
			}
			return contest, nil
		},
		listDesiredRuntimeAWDContestsFn: func(ctx context.Context) ([]*model.Contest, error) {
			return []*model.Contest{contest}, nil
		},
		listContestTeamsFn: func(ctx context.Context, gotContestID int64) ([]*model.Team, error) {
			if gotContestID != contestID {
				t.Fatalf("unexpected contest id for teams: %d", gotContestID)
			}
			return []*model.Team{team}, nil
		},
		listContestAWDServicesFn: func(ctx context.Context, gotContestID int64) ([]*model.ContestAWDService, error) {
			if gotContestID != contestID {
				t.Fatalf("unexpected contest id for services: %d", gotContestID)
			}
			return []*model.ContestAWDService{serviceDef}, nil
		},
		listContestAWDInstancesFn: func(ctx context.Context, gotContestID int64) ([]*model.Instance, error) {
			if gotContestID != contestID {
				t.Fatalf("unexpected contest id for instances: %d", gotContestID)
			}
			return nil, nil
		},
		findContestTeamFn: func(ctx context.Context, gotContestID, gotTeamID int64) (*model.Team, error) {
			findTeamCalled = true
			if gotContestID != contestID || gotTeamID != teamID {
				t.Fatalf("unexpected team lookup contest=%d team=%d", gotContestID, gotTeamID)
			}
			return team, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*model.ContestAWDService, error) {
			findServiceCalled = true
			if gotContestID != contestID || gotServiceID != serviceID {
				t.Fatalf("unexpected service lookup contest=%d service=%d", gotContestID, gotServiceID)
			}
			return serviceDef, nil
		},
		findScopedExistingInstanceFn: func(ctx context.Context, userID, gotChallengeID int64, scope practiceports.InstanceScope) (*model.Instance, error) {
			findExistingCalled = true
			if userID != team.CaptainID || gotChallengeID != challengeID {
				t.Fatalf("unexpected scoped existing lookup user=%d challenge=%d", userID, gotChallengeID)
			}
			if scope.TeamID == nil || *scope.TeamID != teamID || scope.ServiceID == nil || *scope.ServiceID != serviceID {
				t.Fatalf("unexpected scope for existing lookup: %+v", scope)
			}
			return nil, nil
		},
		countScopedRunningInstancesFn: func(ctx context.Context, userID int64, scope practiceports.InstanceScope) (int, error) {
			return 0, nil
		},
		createInstanceFn: func(ctx context.Context, instance *model.Instance) error {
			copied := *instance
			createdInstance = &copied
			instance.ID = 8201
			return nil
		},
		createAWDServiceOperationFn: func(ctx context.Context, next *model.AWDServiceOperation) error {
			copied := *next
			operation = &copied
			return nil
		},
	}

	service := wirePracticeScopeAdapters(NewService(
		repo,
		nil,
		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 4,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled:                  true,
					PollInterval:             time.Second,
					DesiredReconcileInterval: 15 * time.Second,
				},
			},
		},
		nil), repo, nil)

	if err := service.ReconcileDesiredAWDInstances(context.Background()); err != nil {
		t.Fatalf("ReconcileDesiredAWDInstances() error = %v", err)
	}
	if createdInstance == nil {
		t.Fatalf("expected missing awd instance to be created, findTeam=%v findService=%v findExisting=%v", findTeamCalled, findServiceCalled, findExistingCalled)
	}
	if createdInstance.UserID != team.CaptainID {
		t.Fatalf("expected instance owner %d, got %+v", team.CaptainID, createdInstance)
	}
	if createdInstance.TeamID == nil || *createdInstance.TeamID != teamID || createdInstance.ServiceID == nil || *createdInstance.ServiceID != serviceID {
		t.Fatalf("expected team/service scope on created instance, got %+v", createdInstance)
	}
	if createdInstance.Status != model.InstanceStatusPending {
		t.Fatalf("expected pending instance status, got %+v", createdInstance)
	}
	if !createdInstance.ExpiresAt.Equal(contestEnd) {
		t.Fatalf("expected created instance expiry %s, got %+v", contestEnd, createdInstance)
	}
	if operation == nil {
		t.Fatal("expected system awd service operation to be recorded")
	}
	if operation.OperationType != model.AWDServiceOperationTypeStart || operation.RequestedBy != model.AWDServiceOperationRequestedBySystem {
		t.Fatalf("unexpected desired start operation: %+v", operation)
	}
	if operation.Status != model.AWDServiceOperationStatusProvisioning || operation.Reason != "desired_runtime_reconcile" || operation.SLABillable {
		t.Fatalf("unexpected desired start operation detail: %+v", operation)
	}
}

func TestReconcileDesiredAWDInstancesReusesFailedInstance(t *testing.T) {
	t.Parallel()

	contestID := int64(3301)
	teamID := int64(4301)
	serviceID := int64(5301)
	challengeID := int64(6301)
	contestEnd := time.Date(2027, 5, 16, 15, 0, 0, 0, time.UTC)
	contest := &model.Contest{
		ID:      contestID,
		Mode:    model.ContestModeAWD,
		Status:  model.ContestStatusFrozen,
		EndTime: contestEnd,
	}
	team := &model.Team{ID: teamID, ContestID: contestID, CaptainID: 7301, Name: "team-b"}
	serviceDef := &model.ContestAWDService{
		ID:              serviceID,
		ContestID:       contestID,
		AWDChallengeID:  challengeID,
		IsVisible:       true,
		ServiceSnapshot: `{"name":"awd-pwn","category":"pwn","difficulty":"hard","runtime_config":{"image_id":102,"instance_sharing":"per_team"},"flag_config":{"flag_type":"dynamic","flag_prefix":"flag"}}`,
	}
	failedExpiresAt := contestEnd.Add(-30 * time.Minute)
	failedInstance := &model.Instance{
		ID:          8301,
		UserID:      team.CaptainID,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
		Status:      model.InstanceStatusFailed,
		Nonce:       "nonce-old",
		ExpiresAt:   failedExpiresAt,
	}

	cleanupCalled := false
	resetCalled := false
	var operation *model.AWDServiceOperation
	findRestartableCalled := false
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, gotContestID int64) (*model.Contest, error) {
			if gotContestID != contestID {
				t.Fatalf("unexpected contest lookup id: %d", gotContestID)
			}
			return contest, nil
		},
		listDesiredRuntimeAWDContestsFn: func(ctx context.Context) ([]*model.Contest, error) {
			return []*model.Contest{contest}, nil
		},
		listContestTeamsFn: func(ctx context.Context, gotContestID int64) ([]*model.Team, error) {
			if gotContestID != contestID {
				t.Fatalf("unexpected contest id for teams: %d", gotContestID)
			}
			return []*model.Team{team}, nil
		},
		listContestAWDServicesFn: func(ctx context.Context, gotContestID int64) ([]*model.ContestAWDService, error) {
			if gotContestID != contestID {
				t.Fatalf("unexpected contest id for services: %d", gotContestID)
			}
			return []*model.ContestAWDService{serviceDef}, nil
		},
		listContestAWDInstancesFn: func(ctx context.Context, gotContestID int64) ([]*model.Instance, error) {
			if gotContestID != contestID {
				t.Fatalf("unexpected contest id for instances: %d", gotContestID)
			}
			return nil, nil
		},
		findContestTeamFn: func(ctx context.Context, gotContestID, gotTeamID int64) (*model.Team, error) {
			if gotContestID != contestID || gotTeamID != teamID {
				t.Fatalf("unexpected team lookup contest=%d team=%d", gotContestID, gotTeamID)
			}
			return team, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*model.ContestAWDService, error) {
			if gotContestID != contestID || gotServiceID != serviceID {
				t.Fatalf("unexpected service lookup contest=%d service=%d", gotContestID, gotServiceID)
			}
			return serviceDef, nil
		},
		findScopedRestartableInstanceFn: func(ctx context.Context, userID, gotChallengeID int64, scope practiceports.InstanceScope) (*model.Instance, error) {
			findRestartableCalled = true
			if userID != team.CaptainID || gotChallengeID != challengeID {
				t.Fatalf("unexpected restartable lookup user=%d challenge=%d", userID, gotChallengeID)
			}
			if scope.TeamID == nil || *scope.TeamID != teamID || scope.ServiceID == nil || *scope.ServiceID != serviceID {
				t.Fatalf("unexpected scope for restartable lookup: %+v", scope)
			}
			return failedInstance, nil
		},
		resetInstanceRuntimeForRestartFn: func(ctx context.Context, instanceID int64, status string, expiresAt time.Time, preserveHostPort bool) error {
			resetCalled = true
			if instanceID != failedInstance.ID {
				t.Fatalf("unexpected instance id for reset: %d", instanceID)
			}
			if status != model.InstanceStatusPending {
				t.Fatalf("expected pending status on reset, got %s", status)
			}
			if !expiresAt.Equal(contestEnd) {
				t.Fatalf("expected reset expiry %s, got %s", contestEnd, expiresAt)
			}
			if preserveHostPort {
				t.Fatal("expected awd desired reconcile to avoid preserving host port")
			}
			return nil
		},
		createAWDServiceOperationFn: func(ctx context.Context, next *model.AWDServiceOperation) error {
			copied := *next
			operation = &copied
			return nil
		},
	}

	service := wirePracticeScopeAdapters(NewService(
		repo,
		nil,
		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, instance *model.Instance) error {
				cleanupCalled = true
				if instance.ID != failedInstance.ID {
					t.Fatalf("unexpected cleanup instance: %+v", instance)
				}
				return nil
			},
		},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 4,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled:                  true,
					PollInterval:             time.Second,
					DesiredReconcileInterval: 15 * time.Second,
				},
			},
		},
		nil), repo, nil)

	if err := service.ReconcileDesiredAWDInstances(context.Background()); err != nil {
		t.Fatalf("ReconcileDesiredAWDInstances() error = %v", err)
	}
	if !cleanupCalled {
		t.Fatal("expected failed runtime cleanup before restart")
	}
	if !resetCalled {
		t.Fatalf("expected failed instance to be reset for restart, findRestartable=%v cleanup=%v", findRestartableCalled, cleanupCalled)
	}
	if operation == nil {
		t.Fatal("expected recreate operation to be recorded")
	}
	if operation.InstanceID != failedInstance.ID || operation.OperationType != model.AWDServiceOperationTypeRecreate {
		t.Fatalf("unexpected desired recreate operation: %+v", operation)
	}
	if operation.RequestedBy != model.AWDServiceOperationRequestedBySystem || operation.Status != model.AWDServiceOperationStatusProvisioning {
		t.Fatalf("unexpected desired recreate operation detail: %+v", operation)
	}
	if operation.Reason != "desired_runtime_reconcile" || operation.SLABillable {
		t.Fatalf("unexpected desired recreate operation audit detail: %+v", operation)
	}
	if failedInstance.Nonce != "nonce-old" {
		t.Fatalf("expected failed instance nonce to be preserved, got %+v", failedInstance)
	}
}

func TestReconcileDesiredAWDInstancesBacksOffAfterImmediateFailure(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	contestID := int64(3402)
	teamID := int64(4402)
	serviceID := int64(5402)
	contest := &model.Contest{
		ID:      contestID,
		Mode:    model.ContestModeAWD,
		Status:  model.ContestStatusRunning,
		EndTime: time.Now().UTC().Add(time.Hour),
	}
	team := &model.Team{ID: teamID, ContestID: contestID, CaptainID: 7402, Name: "team-c"}
	serviceDef := &model.ContestAWDService{ID: serviceID, ContestID: contestID, AWDChallengeID: 6402, IsVisible: true}

	var resolveCalls int32
	repo := &stubPracticeRepository{
		listDesiredRuntimeAWDContestsFn: func(context.Context) ([]*model.Contest, error) {
			return []*model.Contest{contest}, nil
		},
		listContestTeamsFn: func(context.Context, int64) ([]*model.Team, error) {
			return []*model.Team{team}, nil
		},
		listContestAWDServicesFn: func(context.Context, int64) ([]*model.ContestAWDService, error) {
			return []*model.ContestAWDService{serviceDef}, nil
		},
		listContestAWDInstancesFn: func(context.Context, int64) ([]*model.Instance, error) {
			return nil, nil
		},
		findContestTeamFn: func(context.Context, int64, int64) (*model.Team, error) {
			return team, nil
		},
		findContestAWDServiceFn: func(context.Context, int64, int64) (*model.ContestAWDService, error) {
			atomic.AddInt32(&resolveCalls, 1)
			return nil, errors.New("bad runtime config")
		},
	}

	stateStore := practiceinfra.NewDesiredAWDReconcileStateStore(redisClient)
	service := wirePracticeScopeAdapters(NewService(
		repo,
		nil,
		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 4,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled:                               true,
					PollInterval:                          time.Second,
					DesiredReconcileInterval:              time.Second,
					DesiredReconcileFailureInitialBackoff: time.Minute,
					DesiredReconcileFailureMaxBackoff:     time.Minute,
					DesiredReconcileSuppressAfterFailures: 3,
					DesiredReconcileSuppressDuration:      10 * time.Minute,
				},
			},
		},
		nil), repo, nil).SetDesiredAWDReconcileStateStore(stateStore)

	if err := service.ReconcileDesiredAWDInstances(context.Background()); err != nil {
		t.Fatalf("ReconcileDesiredAWDInstances() error = %v", err)
	}
	if got := atomic.LoadInt32(&resolveCalls); got != 1 {
		t.Fatalf("expected one reconcile attempt, got %d", got)
	}

	if err := service.ReconcileDesiredAWDInstances(context.Background()); err != nil {
		t.Fatalf("second ReconcileDesiredAWDInstances() error = %v", err)
	}
	if got := atomic.LoadInt32(&resolveCalls); got != 1 {
		t.Fatalf("expected backoff to suppress immediate retry, got %d attempts", got)
	}

	state, exists, err := stateStore.LoadDesiredAWDReconcileState(context.Background(), contestID, teamID, serviceID)
	if err != nil {
		t.Fatalf("LoadDesiredAWDReconcileState() error = %v", err)
	}
	if !exists || state == nil || state.FailureCount != 1 {
		t.Fatalf("unexpected desired reconcile state: exists=%v state=%+v", exists, state)
	}
	if !state.NextAttemptAt.After(state.LastFailureAt) {
		t.Fatalf("expected next attempt after last failure, got %+v", state)
	}
}

func TestReconcileDesiredAWDInstancesSuppressesScopeAfterProvisionFailure(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	contestID := int64(3501)
	teamID := int64(4501)
	serviceID := int64(5501)
	challengeID := int64(6501)
	contestEnd := time.Now().UTC().Add(time.Hour)
	contest := &model.Contest{ID: contestID, Mode: model.ContestModeAWD, Status: model.ContestStatusRunning, EndTime: contestEnd}
	team := &model.Team{ID: teamID, ContestID: contestID, CaptainID: 7501, Name: "team-d"}
	serviceDef := &model.ContestAWDService{
		ID:              serviceID,
		ContestID:       contestID,
		AWDChallengeID:  challengeID,
		IsVisible:       true,
		ServiceSnapshot: `{"name":"awd-web","runtime_config":{"image_id":103,"instance_sharing":"per_team"},"flag_config":{"flag_type":"static","flag_prefix":"flag"}}`,
	}

	var createdInstance *model.Instance
	var operationCalls int32
	repo := &stubPracticeRepository{
		findContestByIDFn: func(context.Context, int64) (*model.Contest, error) { return contest, nil },
		listDesiredRuntimeAWDContestsFn: func(context.Context) ([]*model.Contest, error) {
			return []*model.Contest{contest}, nil
		},
		listContestTeamsFn: func(context.Context, int64) ([]*model.Team, error) {
			return []*model.Team{team}, nil
		},
		listContestAWDServicesFn: func(context.Context, int64) ([]*model.ContestAWDService, error) {
			return []*model.ContestAWDService{serviceDef}, nil
		},
		listContestAWDInstancesFn: func(context.Context, int64) ([]*model.Instance, error) {
			return nil, nil
		},
		findContestTeamFn:       func(context.Context, int64, int64) (*model.Team, error) { return team, nil },
		findContestAWDServiceFn: func(context.Context, int64, int64) (*model.ContestAWDService, error) { return serviceDef, nil },
		findScopedExistingInstanceFn: func(context.Context, int64, int64, practiceports.InstanceScope) (*model.Instance, error) {
			return nil, nil
		},
		countScopedRunningInstancesFn: func(context.Context, int64, practiceports.InstanceScope) (int, error) {
			return 0, nil
		},
		createInstanceFn: func(_ context.Context, instance *model.Instance) error {
			copied := *instance
			copied.ID = 8501
			createdInstance = &copied
			instance.ID = copied.ID
			return nil
		},
		createAWDServiceOperationFn: func(context.Context, *model.AWDServiceOperation) error {
			atomic.AddInt32(&operationCalls, 1)
			return nil
		},
	}

	stateStore := practiceinfra.NewDesiredAWDReconcileStateStore(redisClient)
	service := wirePracticeScopeAdapters(NewService(
		repo,
		nil,
		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 4,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled:                               true,
					PollInterval:                          time.Second,
					DesiredReconcileInterval:              time.Second,
					DesiredReconcileFailureInitialBackoff: time.Minute,
					DesiredReconcileFailureMaxBackoff:     time.Minute,
					DesiredReconcileSuppressAfterFailures: 1,
					DesiredReconcileSuppressDuration:      10 * time.Minute,
				},
			},
		},
		nil), repo, nil).SetDesiredAWDReconcileStateStore(stateStore)

	if err := service.ReconcileDesiredAWDInstances(context.Background()); err != nil {
		t.Fatalf("ReconcileDesiredAWDInstances() error = %v", err)
	}
	if createdInstance == nil {
		t.Fatal("expected desired reconcile to create an instance before failure")
	}
	service.markInstanceFailed(context.Background(), createdInstance)

	if err := service.ReconcileDesiredAWDInstances(context.Background()); err != nil {
		t.Fatalf("second ReconcileDesiredAWDInstances() error = %v", err)
	}
	if got := atomic.LoadInt32(&operationCalls); got != 1 {
		t.Fatalf("expected suppressed scope to avoid extra auto operation, got %d", got)
	}

	state, exists, err := stateStore.LoadDesiredAWDReconcileState(context.Background(), contestID, teamID, serviceID)
	if err != nil {
		t.Fatalf("LoadDesiredAWDReconcileState() error = %v", err)
	}
	if !exists || state == nil || state.FailureCount != 1 {
		t.Fatalf("unexpected desired reconcile state after provision failure: exists=%v state=%+v", exists, state)
	}
	if state.SuppressedUntil.IsZero() {
		t.Fatalf("expected scope to enter suppress window, got %+v", state)
	}
}

func TestReconcileDesiredAWDInstancesIgnoresCorruptedDesiredState(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	contestID := int64(3551)
	teamID := int64(4551)
	serviceID := int64(5551)
	challengeID := int64(6551)
	contestEnd := time.Now().UTC().Add(time.Hour)
	contest := &model.Contest{ID: contestID, Mode: model.ContestModeAWD, Status: model.ContestStatusRunning, EndTime: contestEnd}
	team := &model.Team{ID: teamID, ContestID: contestID, CaptainID: 7551, Name: "team-corrupt"}
	serviceDef := &model.ContestAWDService{
		ID:              serviceID,
		ContestID:       contestID,
		AWDChallengeID:  challengeID,
		IsVisible:       true,
		ServiceSnapshot: `{"name":"awd-web","runtime_config":{"image_id":104,"instance_sharing":"per_team"},"flag_config":{"flag_type":"static","flag_prefix":"flag"}}`,
	}

	if err := redisClient.HSet(context.Background(), rediskeys.DesiredAWDReconcileStateKey(contestID, teamID, serviceID), map[string]any{
		"failure_count": "invalid",
	}).Err(); err != nil {
		t.Fatalf("seed corrupted desired reconcile state: %v", err)
	}

	var createdInstance *model.Instance
	repo := &stubPracticeRepository{
		findContestByIDFn: func(context.Context, int64) (*model.Contest, error) { return contest, nil },
		listDesiredRuntimeAWDContestsFn: func(context.Context) ([]*model.Contest, error) {
			return []*model.Contest{contest}, nil
		},
		listContestTeamsFn: func(context.Context, int64) ([]*model.Team, error) {
			return []*model.Team{team}, nil
		},
		listContestAWDServicesFn: func(context.Context, int64) ([]*model.ContestAWDService, error) {
			return []*model.ContestAWDService{serviceDef}, nil
		},
		listContestAWDInstancesFn: func(context.Context, int64) ([]*model.Instance, error) {
			return nil, nil
		},
		findContestTeamFn:       func(context.Context, int64, int64) (*model.Team, error) { return team, nil },
		findContestAWDServiceFn: func(context.Context, int64, int64) (*model.ContestAWDService, error) { return serviceDef, nil },
		findScopedExistingInstanceFn: func(context.Context, int64, int64, practiceports.InstanceScope) (*model.Instance, error) {
			return nil, nil
		},
		countScopedRunningInstancesFn: func(context.Context, int64, practiceports.InstanceScope) (int, error) {
			return 0, nil
		},
		createInstanceFn: func(_ context.Context, instance *model.Instance) error {
			copied := *instance
			createdInstance = &copied
			instance.ID = 8551
			return nil
		},
	}

	stateStore := practiceinfra.NewDesiredAWDReconcileStateStore(redisClient)
	service := wirePracticeScopeAdapters(NewService(
		repo,
		nil,
		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 4,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled:                  true,
					PollInterval:             time.Second,
					DesiredReconcileInterval: time.Second,
				},
			},
		},
		nil), repo, nil).SetDesiredAWDReconcileStateStore(stateStore)

	if err := service.ReconcileDesiredAWDInstances(context.Background()); err != nil {
		t.Fatalf("ReconcileDesiredAWDInstances() error = %v", err)
	}
	if createdInstance == nil {
		t.Fatal("expected reconcile to continue when desired state cache is corrupted")
	}
}

func TestReconcileDesiredAWDInstancesClearsSuppressedStateWhenScopeAlreadyActive(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	contestID := int64(3601)
	teamID := int64(4601)
	serviceID := int64(5601)
	contestEnd := time.Now().UTC().Add(time.Hour)
	contest := &model.Contest{ID: contestID, Mode: model.ContestModeAWD, Status: model.ContestStatusRunning, EndTime: contestEnd}
	team := &model.Team{ID: teamID, ContestID: contestID, CaptainID: 7601, Name: "team-e"}
	serviceDef := &model.ContestAWDService{ID: serviceID, ContestID: contestID, AWDChallengeID: 6601, IsVisible: true}
	activeInstance := &model.Instance{
		ID:          8601,
		UserID:      team.CaptainID,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 6601,
		ServiceID:   &serviceID,
		Status:      model.InstanceStatusRunning,
		ExpiresAt:   contestEnd,
	}

	stateStore := practiceinfra.NewDesiredAWDReconcileStateStore(redisClient)
	if err := stateStore.StoreDesiredAWDReconcileState(context.Background(), contestID, teamID, serviceID, &practiceports.DesiredAWDReconcileState{
		FailureCount:    3,
		LastFailureAt:   time.Now().UTC().Add(-time.Minute),
		NextAttemptAt:   time.Now().UTC().Add(time.Minute),
		SuppressedUntil: time.Now().UTC().Add(10 * time.Minute),
		LastError:       "stale failure",
	}); err != nil {
		t.Fatalf("StoreDesiredAWDReconcileState() error = %v", err)
	}

	repo := &stubPracticeRepository{
		listDesiredRuntimeAWDContestsFn: func(context.Context) ([]*model.Contest, error) {
			return []*model.Contest{contest}, nil
		},
		listContestTeamsFn: func(context.Context, int64) ([]*model.Team, error) {
			return []*model.Team{team}, nil
		},
		listContestAWDServicesFn: func(context.Context, int64) ([]*model.ContestAWDService, error) {
			return []*model.ContestAWDService{serviceDef}, nil
		},
		listContestAWDInstancesFn: func(context.Context, int64) ([]*model.Instance, error) {
			return []*model.Instance{activeInstance}, nil
		},
	}

	service := wirePracticeScopeAdapters(NewService(
		repo,
		nil,
		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{},
		nil,
		nil,
		&config.Config{Container: config.ContainerConfig{DefaultTTL: time.Hour}},
		nil), repo, nil).SetDesiredAWDReconcileStateStore(stateStore)

	if err := service.ReconcileDesiredAWDInstances(context.Background()); err != nil {
		t.Fatalf("ReconcileDesiredAWDInstances() error = %v", err)
	}

	if _, exists, err := stateStore.LoadDesiredAWDReconcileState(context.Background(), contestID, teamID, serviceID); err != nil {
		t.Fatalf("LoadDesiredAWDReconcileState() error = %v", err)
	} else if exists {
		t.Fatal("expected active scope to clear stale desired reconcile suppress state")
	}
}

func TestRunProvisioningLoopTriggersDesiredAWDReconciliation(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now().UTC()
	contestID := int64(3401)
	teamID := int64(4401)
	serviceID := int64(5401)
	challengeID := int64(6401)
	serviceSnapshot, err := model.EncodeContestAWDServiceSnapshot(model.ContestAWDServiceSnapshot{
		Name:       "awd-loop",
		Category:   "web",
		Difficulty: "easy",
		RuntimeConfig: map[string]any{
			"image_id":         9401,
			"instance_sharing": string(model.InstanceSharingPerTeam),
			"defense_workspace": map[string]any{
				"seed_root":       "runtime/workspace",
				"workspace_roots": []string{"runtime/workspace/app"},
				"writable_roots":  []string{"runtime/workspace/app"},
				"runtime_mounts": []map[string]any{
					{"source": "runtime/workspace/app", "target": "/workspace/app", "mode": "rw"},
				},
			},
		},
		FlagConfig: map[string]any{
			"flag_type":   model.FlagTypeStatic,
			"flag_prefix": "flag",
		},
	})
	if err != nil {
		t.Fatalf("encode awd service snapshot: %v", err)
	}

	if err := db.Create(&model.Image{
		ID:        9401,
		Name:      "ctf/awd-web",
		Tag:       "v1",
		Status:    model.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&model.Contest{
		ID:        contestID,
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusRunning,
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create(&model.Team{
		ID:        teamID,
		ContestID: contestID,
		Name:      "team-loop",
		CaptainID: 7401,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := db.Create(&model.ContestAWDService{
		ID:              serviceID,
		ContestID:       contestID,
		AWDChallengeID:  challengeID,
		IsVisible:       true,
		ServiceSnapshot: serviceSnapshot,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("create awd service: %v", err)
	}

	var createTopologyCalls atomic.Int32
	service := wirePracticeScopeAdapters(NewService(
		practiceinfra.NewRepository(db),
		nil,
		challengeinfra.NewImageRepository(db),
		runtimeinfrarepo.NewRepository(db),
		&stubPracticeRuntimeService{
			createTopologyFn: func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
				switch createTopologyCalls.Add(1) {
				case 1:
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "container-loop",
						NetworkID:          "network-loop",
						AccessURL:          "http://awd-c3401-t4401-s5401:8080",
						RuntimeDetails: model.InstanceRuntimeDetails{
							Networks: []model.InstanceRuntimeNetwork{
								{Key: model.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-3401", NetworkID: "network-loop", Shared: true},
							},
							Containers: []model.InstanceRuntimeContainer{
								{
									NodeKey:         "default",
									ContainerID:     "container-loop",
									ServicePort:     8080,
									IsEntryPoint:    true,
									NetworkAliases:  []string{"awd-c3401-t4401-s5401"},
									ServiceProtocol: model.ChallengeTargetProtocolHTTP,
								},
							},
						},
					}, nil
				case 2:
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "workspace-loop",
						NetworkID:          "network-loop",
						AccessURL:          "tcp://172.30.0.21:22",
						RuntimeDetails: model.InstanceRuntimeDetails{
							Containers: []model.InstanceRuntimeContainer{
								{
									NodeKey:         "workspace",
									ContainerID:     "workspace-loop",
									ServicePort:     22,
									IsEntryPoint:    true,
									ServiceProtocol: model.ChallengeTargetProtocolTCP,
								},
							},
						},
					}, nil
				default:
					t.Fatalf("unexpected topology create call #%d", createTopologyCalls.Load())
					return nil, nil
				}
			},
		},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				CreateTimeout:        time.Second,
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 4,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled:                  true,
					PollInterval:             10 * time.Millisecond,
					DesiredReconcileInterval: 10 * time.Millisecond,
					BatchSize:                1,
					MaxConcurrentStarts:      1,
					MaxActiveInstances:       10,
				},
			},
		},
		nil), practiceinfra.NewRepository(db), nil)
	service.StartBackgroundTasks(context.Background())

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go service.RunProvisioningLoop(runCtx)

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		var instances []model.Instance
		if err := db.Where("contest_id = ? AND team_id = ? AND service_id = ?", contestID, teamID, serviceID).Find(&instances).Error; err == nil {
			if len(instances) == 1 && instances[0].Status == model.InstanceStatusRunning && instances[0].ContainerID == "container-loop" {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}

	var instances []model.Instance
	if err := db.Where("contest_id = ? AND team_id = ? AND service_id = ?", contestID, teamID, serviceID).Find(&instances).Error; err != nil {
		t.Fatalf("load desired reconcile instances: %v", err)
	}
	var operations []model.AWDServiceOperation
	if err := db.Where("contest_id = ? AND team_id = ? AND service_id = ?", contestID, teamID, serviceID).Order("id ASC").Find(&operations).Error; err != nil {
		t.Fatalf("load desired reconcile operations: %v", err)
	}
	t.Fatalf("expected desired reconcile loop to produce running instance, got instances=%+v operations=%+v", instances, operations)
}
