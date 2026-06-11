package commands

import (
	"context"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
	contestentity "ctf-platform/internal/module/practice/testsupport/contestentity"
	"testing"
	"time"
)

func TestStartContestAWDServiceDoesNotRequireContestChallengeLookup(t *testing.T) {
	t.Parallel()

	teamID := int64(4104)
	contestEnd := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	var createdInstance *instanceentity.Instance
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, contestID int64) (*practiceports.ContestRecord, error) {
			if contestID != 3104 {
				t.Fatalf("unexpected contest id: %d", contestID)
			}
			return &practiceports.ContestRecord{
				ID:      contestID,
				Mode:    practiceports.ContestModeAWD,
				Status:  practiceports.ContestStatusRunning,
				EndTime: contestEnd,
			}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, contestID, serviceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			if contestID != 3104 || serviceID != 7104 {
				t.Fatalf("unexpected awd service lookup: contest=%d service=%d", contestID, serviceID)
			}
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2104,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":104,"instance_sharing":"per_team"},"flag_config":{"flag_type":"static","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestChallengeFn: func(ctx context.Context, contestID, challengeID int64) (*practiceports.ContestChallengeRecord, error) {
			t.Fatalf("unexpected contest challenge lookup for awd start: contest=%d challenge=%d", contestID, challengeID)
			return nil, nil
		},
		findContestRegistrationFn: func(ctx context.Context, contestID, userID int64) (*practiceports.ContestParticipation, error) {
			if contestID != 3104 || userID != 5104 {
				t.Fatalf("unexpected registration lookup: contest=%d user=%d", contestID, userID)
			}
			return &practiceports.ContestParticipation{TeamID: &teamID, Status: contestentity.ContestRegistrationStatusApproved}, nil
		},
		createInstanceFn: func(ctx context.Context, instance *instanceentity.Instance) error {
			copied := *instance
			createdInstance = &copied
			instance.ID = 9104
			return nil
		},
	}

	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			if id != 2104 {
				t.Fatalf("unexpected challenge lookup: %d", id)
			}
			return &challengecontracts.PracticeRuntimeChallenge{
				ID:       id,
				Status:   challengecontracts.ChallengeStatusPublished,
				ImageID:  int64Ptr(104),
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
				PortRangeStart:       30000,
				PortRangeEnd:         30010,
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled: true,
				},
			},
		},
		nil),

		repo, challengeRepo)

	resp, err := service.StartContestAWDService(context.Background(), 5104, 3104, 7104)
	if err != nil {
		t.Fatalf("StartContestAWDService() error = %v", err)
	}
	if resp.ID != 9104 {
		t.Fatalf("expected created awd service instance id, got %+v", resp)
	}
	if resp.ChallengeID != 2104 {
		t.Fatalf("expected awd service challenge id 2104, got %+v", resp)
	}
	if resp.Status != instanceentity.InstanceStatusPending {
		t.Fatalf("expected pending awd service instance, got %+v", resp)
	}
	if createdInstance == nil || !createdInstance.ExpiresAt.Equal(contestEnd) {
		t.Fatalf("expected awd instance expiry to follow contest end time %s, got %+v", contestEnd, createdInstance)
	}
}
