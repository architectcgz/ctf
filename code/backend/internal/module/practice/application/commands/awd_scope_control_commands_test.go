package commands_test

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"ctf-platform/internal/model"
	practicecmd "ctf-platform/internal/module/practice/application/commands"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/pkg/errcode"
)

func TestServiceAWDControlLifecycleGuards(t *testing.T) {
	type action func(t *testing.T, service *practicecmd.Service, contestID, teamID, serviceID, userID int64) error

	for _, tc := range []struct {
		name        string
		controlType string
		action      action
		wantErr     error
	}{
		{
			name:        "user_start_rejects_retired_team",
			controlType: model.AWDScopeControlTypeRetired,
			action: func(t *testing.T, service *practicecmd.Service, contestID, teamID, serviceID, userID int64) error {
				_, err := service.StartContestAWDService(context.Background(), userID, contestID, serviceID)
				return err
			},
			wantErr: errcode.ErrAWDTeamRetired,
		},
		{
			name:        "user_start_rejects_disabled_service",
			controlType: model.AWDScopeControlTypeServiceDisabled,
			action: func(t *testing.T, service *practicecmd.Service, contestID, teamID, serviceID, userID int64) error {
				_, err := service.StartContestAWDService(context.Background(), userID, contestID, serviceID)
				return err
			},
			wantErr: errcode.ErrAWDServiceDisabled,
		},
		{
			name:        "admin_start_rejects_retired_team",
			controlType: model.AWDScopeControlTypeRetired,
			action: func(t *testing.T, service *practicecmd.Service, contestID, teamID, serviceID, userID int64) error {
				_, err := service.StartAdminContestAWDTeamService(context.Background(), contestID, teamID, serviceID)
				return err
			},
			wantErr: errcode.ErrAWDTeamRetired,
		},
		{
			name:        "admin_start_rejects_disabled_service",
			controlType: model.AWDScopeControlTypeServiceDisabled,
			action: func(t *testing.T, service *practicecmd.Service, contestID, teamID, serviceID, userID int64) error {
				_, err := service.StartAdminContestAWDTeamService(context.Background(), contestID, teamID, serviceID)
				return err
			},
			wantErr: errcode.ErrAWDServiceDisabled,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			db := newContestInstanceTestDB(t)
			now := time.Now().UTC()

			contestID := int64(31000)
			teamID := int64(41000)
			serviceID := int64(71000)
			userID := int64(51000)

			seedContestInstanceUser(t, db, userID, now)
			seedContestInstanceChallenge(t, db, 11000, 21000, now)
			seedContestInstanceAWDContest(t, db, contestID, 21000, now)
			seedContestInstanceAWDService(t, db, serviceID, contestID, 21000, now)
			seedContestInstanceTeam(t, db, contestID, teamID, userID, now)
			seedContestInstanceRegistration(t, db, contestID, userID, teamID, model.ContestRegistrationStatusApproved, now)
			seedContestInstanceTeamMember(t, db, contestID, teamID, userID, now)
			if err := ensureContestInstanceServiceIDColumn(db); err != nil {
				t.Fatalf("ensure instances.service_id column: %v", err)
			}

			scopeType := model.AWDScopeControlScopeTeam
			controlServiceID := int64(0)
			if tc.controlType == model.AWDScopeControlTypeServiceDisabled {
				scopeType = model.AWDScopeControlScopeTeamService
				controlServiceID = serviceID
			}
			if err := db.Create(&model.AWDScopeControl{
				ContestID:   contestID,
				TeamID:      teamID,
				ScopeType:   scopeType,
				ServiceID:   controlServiceID,
				ControlType: tc.controlType,
				Reason:      tc.name,
				CreatedAt:   now,
				UpdatedAt:   now,
			}).Error; err != nil {
				t.Fatalf("create awd scope control: %v", err)
			}

			service := newContestInstanceTestService(t, db)
			err := tc.action(t, service, contestID, teamID, serviceID, userID)
			if err == nil || err.Error() != tc.wantErr.Error() {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}

			var count int64
			if err := db.Model(&model.Instance{}).Count(&count).Error; err != nil {
				t.Fatalf("count instances: %v", err)
			}
			if count != 0 {
				t.Fatalf("expected lifecycle guard to block instance creation, count=%d", count)
			}
		})
	}
}

func TestServiceStartContestAWDServiceAllowsManualStartWhenDesiredReconcileSuppressed(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now().UTC()

	contestID := int64(31010)
	teamID := int64(41010)
	serviceID := int64(71010)
	userID := int64(51010)

	seedContestInstanceUser(t, db, userID, now)
	seedContestInstanceChallenge(t, db, 11010, 21010, now)
	seedContestInstanceAWDContest(t, db, contestID, 21010, now)
	seedContestInstanceAWDService(t, db, serviceID, contestID, 21010, now)
	seedContestInstanceTeam(t, db, contestID, teamID, userID, now)
	seedContestInstanceRegistration(t, db, contestID, userID, teamID, model.ContestRegistrationStatusApproved, now)
	seedContestInstanceTeamMember(t, db, contestID, teamID, userID, now)
	if err := ensureContestInstanceServiceIDColumn(db); err != nil {
		t.Fatalf("ensure instances.service_id column: %v", err)
	}
	if err := db.Create(&model.AWDScopeControl{
		ContestID:   contestID,
		TeamID:      teamID,
		ScopeType:   model.AWDScopeControlScopeTeamService,
		ServiceID:   serviceID,
		ControlType: model.AWDScopeControlTypeDesiredReconcileSuppressed,
		Reason:      "manual-suppress",
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd scope control: %v", err)
	}

	service := newContestInstanceTestService(t, db)
	resp, err := service.StartContestAWDService(context.Background(), userID, contestID, serviceID)
	if err != nil {
		t.Fatalf("StartContestAWDService() error = %v", err)
	}
	if resp == nil || resp.ID <= 0 {
		t.Fatalf("expected manual start to bypass desired reconcile suppress, got %+v", resp)
	}

	var instance model.Instance
	if err := db.First(&instance, resp.ID).Error; err != nil {
		t.Fatalf("load instance: %v", err)
	}
	if instance.ServiceID == nil || *instance.ServiceID != serviceID {
		t.Fatalf("expected started awd instance to keep service scope, got %+v", instance)
	}
}

func TestServiceSetAdminContestAWDTeamRetiredStopsActiveInstancesAndClearsDesiredState(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now().UTC()

	contestID := int64(31020)
	teamID := int64(41020)
	captainID := int64(51020)
	adminID := int64(61020)
	serviceAID := int64(71020)
	serviceBID := int64(71021)

	seedContestInstanceUser(t, db, captainID, now)
	seedContestInstanceUser(t, db, adminID, now)
	seedContestInstanceChallenge(t, db, 11020, 21020, now)
	seedContestInstanceChallenge(t, db, 11021, 21021, now)
	seedContestInstanceAWDContest(t, db, contestID, 21020, now)
	seedContestInstanceAWDService(t, db, serviceAID, contestID, 21020, now)
	seedContestInstanceAWDService(t, db, serviceBID, contestID, 21021, now)
	seedContestInstanceTeam(t, db, contestID, teamID, captainID, now)
	seedContestInstanceTeamMember(t, db, contestID, teamID, captainID, now)
	if err := ensureContestInstanceServiceIDColumn(db); err != nil {
		t.Fatalf("ensure instances.service_id column: %v", err)
	}

	stateStore := newAWDControlStateStore(t)
	for _, serviceID := range []int64{serviceAID, serviceBID} {
		if err := stateStore.StoreDesiredAWDReconcileState(context.Background(), contestID, teamID, serviceID, &practiceports.DesiredAWDReconcileState{
			FailureCount:  2,
			LastFailureAt: now,
			NextAttemptAt: now.Add(time.Minute),
			LastError:     "stale-failure",
		}); err != nil {
			t.Fatalf("store desired reconcile state: %v", err)
		}
	}

	for idx, serviceID := range []int64{serviceAID, serviceBID} {
		instanceID := int64(81020 + idx)
		challengeID := int64(21020 + idx)
		if err := db.Create(&model.Instance{
			ID:          instanceID,
			UserID:      captainID,
			ContestID:   &contestID,
			TeamID:      &teamID,
			ChallengeID: challengeID,
			ServiceID:   &serviceID,
			ContainerID: "ctr-active",
			Status:      model.InstanceStatusRunning,
			AccessURL:   "http://127.0.0.1:30001",
			ExpiresAt:   now.Add(time.Hour),
			CreatedAt:   now,
			UpdatedAt:   now,
		}).Error; err != nil {
			t.Fatalf("create awd instance: %v", err)
		}
	}

	service := newContestInstanceTestService(t, db).SetDesiredAWDReconcileStateStore(stateStore)
	resp, err := service.SetAdminContestAWDTeamRetired(context.Background(), contestID, teamID, adminID, true, "retire-team")
	if err != nil {
		t.Fatalf("SetAdminContestAWDTeamRetired() error = %v", err)
	}
	if resp == nil || !resp.Enabled || resp.ControlType != model.AWDScopeControlTypeRetired {
		t.Fatalf("unexpected awd team retirement response: %+v", resp)
	}

	var controls []model.AWDScopeControl
	if err := db.Where("contest_id = ? AND team_id = ?", contestID, teamID).Find(&controls).Error; err != nil {
		t.Fatalf("list awd scope controls: %v", err)
	}
	if len(controls) != 1 || controls[0].ControlType != model.AWDScopeControlTypeRetired {
		t.Fatalf("expected team retirement control persisted, got %+v", controls)
	}

	var instances []model.Instance
	if err := db.Where("contest_id = ? AND team_id = ?", contestID, teamID).Order("id ASC").Find(&instances).Error; err != nil {
		t.Fatalf("list awd instances: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected 2 team instances, got %+v", instances)
	}
	for _, instance := range instances {
		if instance.Status != model.InstanceStatusStopped {
			t.Fatalf("expected team retirement to stop active instance, got %+v", instance)
		}
	}

	for _, serviceID := range []int64{serviceAID, serviceBID} {
		if _, exists, err := stateStore.LoadDesiredAWDReconcileState(context.Background(), contestID, teamID, serviceID); err != nil {
			t.Fatalf("load desired reconcile state: %v", err)
		} else if exists {
			t.Fatalf("expected team retirement to clear desired reconcile state for service %d", serviceID)
		}
	}
}

func TestServiceSetAdminContestAWDTeamServiceDisabledStopsActiveInstanceAndClearsDesiredState(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now().UTC()

	contestID := int64(31030)
	teamID := int64(41030)
	captainID := int64(51030)
	adminID := int64(61030)
	serviceID := int64(71030)

	seedContestInstanceUser(t, db, captainID, now)
	seedContestInstanceUser(t, db, adminID, now)
	seedContestInstanceChallenge(t, db, 11030, 21030, now)
	seedContestInstanceAWDContest(t, db, contestID, 21030, now)
	seedContestInstanceAWDService(t, db, serviceID, contestID, 21030, now)
	seedContestInstanceTeam(t, db, contestID, teamID, captainID, now)
	seedContestInstanceTeamMember(t, db, contestID, teamID, captainID, now)
	if err := ensureContestInstanceServiceIDColumn(db); err != nil {
		t.Fatalf("ensure instances.service_id column: %v", err)
	}

	stateStore := newAWDControlStateStore(t)
	if err := stateStore.StoreDesiredAWDReconcileState(context.Background(), contestID, teamID, serviceID, &practiceports.DesiredAWDReconcileState{
		FailureCount:  3,
		LastFailureAt: now,
		NextAttemptAt: now.Add(time.Minute),
		LastError:     "stale-failure",
	}); err != nil {
		t.Fatalf("store desired reconcile state: %v", err)
	}

	instanceID := int64(81030)
	if err := db.Create(&model.Instance{
		ID:          instanceID,
		UserID:      captainID,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 21030,
		ServiceID:   &serviceID,
		ContainerID: "ctr-active-service",
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30002",
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	service := newContestInstanceTestService(t, db).SetDesiredAWDReconcileStateStore(stateStore)
	resp, err := service.SetAdminContestAWDTeamServiceDisabled(context.Background(), contestID, teamID, serviceID, adminID, true, "disable-service")
	if err != nil {
		t.Fatalf("SetAdminContestAWDTeamServiceDisabled() error = %v", err)
	}
	if resp == nil || !resp.Enabled || resp.ControlType != model.AWDScopeControlTypeServiceDisabled {
		t.Fatalf("unexpected awd service disable response: %+v", resp)
	}
	if resp.ServiceID == nil || *resp.ServiceID != serviceID {
		t.Fatalf("expected service-scoped disable response, got %+v", resp)
	}

	var control model.AWDScopeControl
	if err := db.Where("contest_id = ? AND team_id = ? AND service_id = ? AND control_type = ?", contestID, teamID, serviceID, model.AWDScopeControlTypeServiceDisabled).
		First(&control).Error; err != nil {
		t.Fatalf("load service disable control: %v", err)
	}
	if control.ScopeType != model.AWDScopeControlScopeTeamService {
		t.Fatalf("expected team service control scope, got %+v", control)
	}

	var instance model.Instance
	if err := db.First(&instance, instanceID).Error; err != nil {
		t.Fatalf("load awd instance: %v", err)
	}
	if instance.Status != model.InstanceStatusStopped {
		t.Fatalf("expected service disable to stop active instance, got %+v", instance)
	}

	if _, exists, err := stateStore.LoadDesiredAWDReconcileState(context.Background(), contestID, teamID, serviceID); err != nil {
		t.Fatalf("load desired reconcile state: %v", err)
	} else if exists {
		t.Fatal("expected service disable to clear desired reconcile state")
	}
}

func newAWDControlStateStore(t *testing.T) practiceports.PracticeDesiredAWDReconcileStateStore {
	t.Helper()

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	return practiceinfra.NewDesiredAWDReconcileStateStore(redisClient)
}
