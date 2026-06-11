package commands

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
	contestentity "ctf-platform/internal/module/practice/testsupport/contestentity"
)

func TestStartContestAWDServiceDoesNotReserveHostPort(t *testing.T) {
	t.Parallel()

	teamID := int64(4105)
	contestEnd := time.Date(2026, 5, 15, 13, 0, 0, 0, time.UTC)
	var createdInstance *instanceentity.Instance
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, contestID int64) (*practiceports.ContestRecord, error) {
			return &practiceports.ContestRecord{
				ID:      contestID,
				Mode:    practiceports.ContestModeAWD,
				Status:  practiceports.ContestStatusRunning,
				EndTime: contestEnd,
			}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, contestID, serviceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2105,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":105,"instance_sharing":"per_team"},"flag_config":{"flag_type":"static","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestRegistrationFn: func(ctx context.Context, contestID, userID int64) (*practiceports.ContestParticipation, error) {
			return &practiceports.ContestParticipation{TeamID: &teamID, Status: contestentity.ContestRegistrationStatusApproved}, nil
		},
		reserveAvailablePortFn: func(ctx context.Context, start, end int) (int, error) {
			t.Fatal("AWD service instances must not reserve a host port")
			return 0, nil
		},
		bindReservedPortFn: func(ctx context.Context, port int, instanceID int64) error {
			t.Fatalf("AWD service instances must not bind a reserved host port: port=%d instance_id=%d", port, instanceID)
			return nil
		},
		createInstanceFn: func(ctx context.Context, instance *instanceentity.Instance) error {
			instance.ID = 9105
			copied := *instance
			createdInstance = &copied
			return nil
		},
	}

	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			return &challengecontracts.PracticeRuntimeChallenge{
				ID:       id,
				Status:   challengecontracts.ChallengeStatusPublished,
				ImageID:  int64Ptr(105),
				FlagType: challengecontracts.FlagTypeStatic,
				FlagHash: "flag{awd-static}",
			}, nil
		},
	}
	service := wirePracticeScopeAdapters(newServiceCore(
		repo,

		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled: true,
				},
			},
		},
		nil),

		repo, challengeRepo)

	resp, err := service.StartContestAWDService(context.Background(), 5105, 3105, 7105)
	if err != nil {
		t.Fatalf("StartContestAWDService() error = %v", err)
	}
	if resp.ID != 9105 {
		t.Fatalf("expected created awd service instance, got %+v", resp)
	}
	if createdInstance == nil {
		t.Fatal("expected instance to be created")
	}
	if createdInstance.HostPort != 0 {
		t.Fatalf("expected no AWD host port reservation, got %d", createdInstance.HostPort)
	}
	if !createdInstance.ExpiresAt.Equal(contestEnd) {
		t.Fatalf("expected awd instance expiry to follow contest end time %s, got %s", contestEnd, createdInstance.ExpiresAt)
	}
}

func TestStartContestAWDServiceReservesHostPortWhenAccessHostConfigured(t *testing.T) {
	t.Parallel()

	teamID := int64(4115)
	contestEnd := time.Date(2026, 5, 15, 14, 0, 0, 0, time.UTC)
	var createdInstance *instanceentity.Instance
	reserved := false
	bound := false
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, contestID int64) (*practiceports.ContestRecord, error) {
			return &practiceports.ContestRecord{
				ID:      contestID,
				Mode:    practiceports.ContestModeAWD,
				Status:  practiceports.ContestStatusRunning,
				EndTime: contestEnd,
			}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, contestID, serviceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2115,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":115,"instance_sharing":"per_team"},"flag_config":{"flag_type":"static","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestRegistrationFn: func(ctx context.Context, contestID, userID int64) (*practiceports.ContestParticipation, error) {
			return &practiceports.ContestParticipation{TeamID: &teamID, Status: contestentity.ContestRegistrationStatusApproved}, nil
		},
		reserveAvailablePortFn: func(ctx context.Context, start, end int) (int, error) {
			reserved = true
			return 30015, nil
		},
		bindReservedPortFn: func(ctx context.Context, port int, instanceID int64) error {
			bound = true
			if port != 30015 || instanceID != 9115 {
				t.Fatalf("unexpected host port bind: port=%d instance_id=%d", port, instanceID)
			}
			return nil
		},
		createInstanceFn: func(ctx context.Context, instance *instanceentity.Instance) error {
			instance.ID = 9115
			copied := *instance
			createdInstance = &copied
			return nil
		},
	}

	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			return &challengecontracts.PracticeRuntimeChallenge{
				ID:       id,
				Status:   challengecontracts.ChallengeStatusPublished,
				ImageID:  int64Ptr(115),
				FlagType: challengecontracts.FlagTypeStatic,
				FlagHash: "flag{awd-static}",
			}, nil
		},
	}
	service := wirePracticeScopeAdapters(newServiceCore(
		repo,

		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				AccessHost:           "host-gateway.internal",
				PortRangeStart:       30000,
				PortRangeEnd:         30020,
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled: true,
				},
			},
		},
		nil),

		repo, challengeRepo)

	resp, err := service.StartContestAWDService(context.Background(), 5115, 3115, 7115)
	if err != nil {
		t.Fatalf("StartContestAWDService() error = %v", err)
	}
	if resp.ID != 9115 {
		t.Fatalf("expected created awd service instance, got %+v", resp)
	}
	if resp.AccessURL != "" {
		t.Fatalf("expected student-facing awd access url to stay hidden, got %+v", resp)
	}
	if !reserved || !bound {
		t.Fatalf("expected awd instance to reserve and bind host port, reserved=%v bound=%v", reserved, bound)
	}
	if createdInstance == nil {
		t.Fatal("expected instance to be created")
	}
	if createdInstance.HostPort != 30015 {
		t.Fatalf("expected reserved AWD host port, got %d", createdInstance.HostPort)
	}
	if !createdInstance.ExpiresAt.Equal(contestEnd) {
		t.Fatalf("expected awd instance expiry to follow contest end time %s, got %s", contestEnd, createdInstance.ExpiresAt)
	}
}
