package commands

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/config"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
	contestentity "ctf-platform/internal/module/practice/testsupport/contestentity"
)

func TestStartContestAWDServiceRefreshesExistingInstanceExpiryToContestEnd(t *testing.T) {
	t.Parallel()

	now := time.Now()
	teamID := int64(4118)
	serviceID := int64(7118)
	contestID := int64(3118)
	userID := int64(5118)
	contestEnd := now.Add(6 * time.Hour).UTC()
	existingInstance := &instanceentity.Instance{
		ID:          9118,
		UserID:      userID,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 2118,
		ServiceID:   &serviceID,
		ShareScope:  instanceentity.ShareScopePerTeam,
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(5 * time.Minute),
		MaxExtends:  2,
	}
	refreshCalled := false
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, gotContestID int64) (*practiceports.ContestRecord, error) {
			return &practiceports.ContestRecord{ID: gotContestID, Mode: practiceports.ContestModeAWD, Status: practiceports.ContestStatusRunning, EndTime: contestEnd}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2118,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":118,"instance_sharing":"per_team"},"flag_config":{"flag_type":"static","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestRegistrationFn: func(ctx context.Context, gotContestID, gotUserID int64) (*practiceports.ContestParticipation, error) {
			return &practiceports.ContestParticipation{TeamID: &teamID, Status: contestentity.ContestRegistrationStatusApproved}, nil
		},
		findScopedExistingInstanceFn: func(ctx context.Context, gotUserID, gotChallengeID int64, scope practiceports.InstanceScope) (*instanceentity.Instance, error) {
			if gotUserID != userID || gotChallengeID != 2118 {
				t.Fatalf("unexpected scoped lookup: user=%d challenge=%d", gotUserID, gotChallengeID)
			}
			return existingInstance, nil
		},
		refreshInstanceExpiryWithContextFn: func(ctx context.Context, instanceID int64, expiresAt time.Time) error {
			refreshCalled = true
			if instanceID != existingInstance.ID {
				t.Fatalf("unexpected refresh instance id: %d", instanceID)
			}
			if !expiresAt.Equal(contestEnd) {
				t.Fatalf("expected expiry refresh to contest end time %s, got %s", contestEnd, expiresAt)
			}
			return nil
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

		repo, &stubPracticeChallengeContract{})

	resp, err := service.StartContestAWDService(context.Background(), userID, contestID, serviceID)
	if err != nil {
		t.Fatalf("StartContestAWDService() error = %v", err)
	}
	if resp.ID != existingInstance.ID {
		t.Fatalf("expected reused awd instance, got %+v", resp)
	}
	if !refreshCalled {
		t.Fatal("expected awd instance expiry to be refreshed")
	}
	if !resp.ExpiresAt.Equal(contestEnd) || !existingInstance.ExpiresAt.Equal(contestEnd) {
		t.Fatalf("expected awd instance expiry to refresh to contest end time %s, resp=%s stored=%s", contestEnd, resp.ExpiresAt, existingInstance.ExpiresAt)
	}
}

func TestStartContestAWDServiceDoesNotRefreshStoppingInstanceExpiry(t *testing.T) {
	t.Parallel()

	now := time.Now()
	teamID := int64(4218)
	serviceID := int64(7218)
	contestID := int64(3218)
	userID := int64(5218)
	contestEnd := now.Add(6 * time.Hour).UTC()
	existingExpiry := now.Add(5 * time.Minute)
	existingInstance := &instanceentity.Instance{
		ID:          9218,
		UserID:      userID,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 2218,
		ServiceID:   &serviceID,
		ShareScope:  instanceentity.ShareScopePerTeam,
		Status:      instanceentity.InstanceStatusStopping,
		ExpiresAt:   existingExpiry,
		MaxExtends:  2,
	}
	refreshCalled := false
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, gotContestID int64) (*practiceports.ContestRecord, error) {
			return &practiceports.ContestRecord{ID: gotContestID, Mode: practiceports.ContestModeAWD, Status: practiceports.ContestStatusRunning, EndTime: contestEnd}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2218,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":118,"instance_sharing":"per_team"},"flag_config":{"flag_type":"static","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestRegistrationFn: func(ctx context.Context, gotContestID, gotUserID int64) (*practiceports.ContestParticipation, error) {
			return &practiceports.ContestParticipation{TeamID: &teamID, Status: contestentity.ContestRegistrationStatusApproved}, nil
		},
		findScopedExistingInstanceFn: func(ctx context.Context, gotUserID, gotChallengeID int64, scope practiceports.InstanceScope) (*instanceentity.Instance, error) {
			if gotUserID != userID || gotChallengeID != 2218 {
				t.Fatalf("unexpected scoped lookup: user=%d challenge=%d", gotUserID, gotChallengeID)
			}
			return existingInstance, nil
		},
		refreshInstanceExpiryWithContextFn: func(ctx context.Context, instanceID int64, expiresAt time.Time) error {
			refreshCalled = true
			return nil
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

		repo, &stubPracticeChallengeContract{})

	resp, err := service.StartContestAWDService(context.Background(), userID, contestID, serviceID)
	if err != nil {
		t.Fatalf("StartContestAWDService() error = %v", err)
	}
	if resp.ID != existingInstance.ID {
		t.Fatalf("expected reused awd stopping instance, got %+v", resp)
	}
	if refreshCalled {
		t.Fatal("stopping awd instance should not refresh expiry")
	}
	if !resp.ExpiresAt.Equal(existingExpiry) || !existingInstance.ExpiresAt.Equal(existingExpiry) {
		t.Fatalf("expected stopping awd expiry to remain unchanged, resp=%s stored=%s want=%s", resp.ExpiresAt, existingInstance.ExpiresAt, existingExpiry)
	}
	if resp.Status != "destroying" {
		t.Fatalf("expected stopping awd instance to be exposed as destroying, got %+v", resp)
	}
	if resp.AccessURL != "" || resp.Access != nil {
		t.Fatalf("expected destroying awd response access to be cleared, got %+v", resp)
	}
}
