package application_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	instancecmd "ctf-platform/internal/module/instance/application/commands"
	instanceqry "ctf-platform/internal/module/instance/application/queries"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/pkg/errcode"
)

type noopRuntimeCleaner struct{}

func (noopRuntimeCleaner) CleanupRuntime(context.Context, *model.Instance) error {
	return nil
}

type runtimeInstanceContextRepo struct {
	findByIDWithContextFn                   func(ctx context.Context, id int64) (*model.Instance, error)
	findUserByIDFn                          func(ctx context.Context, userID int64) (*model.User, error)
	listVisibleByUserFn                     func(ctx context.Context, userID int64) ([]runtimeports.UserVisibleInstanceRow, error)
	updateStatusAndReleasePortWithContextFn func(ctx context.Context, id int64, status string) error
}

func (r *runtimeInstanceContextRepo) FindByID(ctx context.Context, id int64) (*model.Instance, error) {
	if r.findByIDWithContextFn != nil {
		return r.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (r *runtimeInstanceContextRepo) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	if r.findUserByIDFn != nil {
		return r.findUserByIDFn(ctx, userID)
	}
	return nil, nil
}

func (r *runtimeInstanceContextRepo) FindAccessibleByIDForUser(ctx context.Context, instanceID, userID int64) (*model.Instance, error) {
	return nil, nil
}

func (r *runtimeInstanceContextRepo) ListVisibleByUser(ctx context.Context, userID int64) ([]runtimeports.UserVisibleInstanceRow, error) {
	if r.listVisibleByUserFn != nil {
		return r.listVisibleByUserFn(ctx, userID)
	}
	return nil, nil
}

func (r *runtimeInstanceContextRepo) ListTeacherInstances(ctx context.Context, filter runtimeports.TeacherInstanceFilter) ([]runtimeports.TeacherInstanceRow, error) {
	return nil, nil
}

func (r *runtimeInstanceContextRepo) AtomicExtendByID(ctx context.Context, id int64, maxExtends int, duration time.Duration) error {
	return nil
}

func (r *runtimeInstanceContextRepo) UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error {
	if r.updateStatusAndReleasePortWithContextFn != nil {
		return r.updateStatusAndReleasePortWithContextFn(ctx, id, status)
	}
	return nil
}

func TestInstanceServiceGetUserInstancesShowsContestSharedInstanceToTeamMember(t *testing.T) {
	t.Parallel()

	db := newInstanceServiceTestDB(t)
	now := time.Now()
	contestID := int64(501)
	teamID := int64(601)

	seedInstanceServiceChallenge(t, db, &model.Challenge{
		ID:         102,
		Title:      "Shared AWD Challenge",
		Category:   model.DimensionPwn,
		Difficulty: model.ChallengeDifficultyMedium,
		FlagType:   model.FlagTypeDynamic,
		Status:     model.ChallengeStatusPublished,
		Points:     150,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	seedInstanceServiceTeam(t, db, &model.Team{
		ID:         teamID,
		ContestID:  contestID,
		Name:       "Runtime Team",
		CaptainID:  1,
		InviteCode: "runtime",
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	seedInstanceServiceTeamMember(t, db, &model.TeamMember{
		ContestID: contestID,
		TeamID:    teamID,
		UserID:    2,
		JoinedAt:  now,
		CreatedAt: now,
	})
	seedInstanceServiceInstance(t, db, &model.Instance{
		ID:          1002,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 102,
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30002",
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	service := instanceqry.NewInstanceService(runtimeinfrarepo.NewRepository(db), &config.ContainerConfig{})

	items, err := service.GetUserInstances(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetUserInstances() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != 1002 {
		t.Fatalf("expected teammate visible shared instance, got %+v", items)
	}
}

func TestInstanceServiceGetUserInstancesPrefersContestAWDServiceMetadata(t *testing.T) {
	t.Parallel()

	db := newInstanceServiceTestDB(t)
	now := time.Now()
	contestID := int64(701)
	teamID := int64(801)
	serviceID := int64(9701)

	seedInstanceServiceChallenge(t, db, &model.Challenge{
		ID:         201,
		Title:      "Legacy Runtime Challenge",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		FlagType:   model.FlagTypeStatic,
		Status:     model.ChallengeStatusPublished,
		Points:     100,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err := db.Create(&model.Contest{
		ID:        contestID,
		Title:     "AWD Contest",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusRunning,
		StartTime: now.Add(-time.Minute),
		EndTime:   now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create(&model.ContestAWDService{
		ID:              serviceID,
		ContestID:       contestID,
		AWDChallengeID:  202,
		DisplayName:     "Bank Portal",
		Order:           1,
		IsVisible:       true,
		ScoreConfig:     `{"points":300}`,
		RuntimeConfig:   `{"checker_type":"http_standard"}`,
		ServiceSnapshot: `{"name":"Bank Portal","category":"pwn","difficulty":"hard","flag_config":{"flag_type":"dynamic","flag_prefix":"awd"}}`,
		ValidationState: model.AWDCheckerValidationStatePassed,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("create contest awd service: %v", err)
	}
	seedInstanceServiceTeam(t, db, &model.Team{
		ID:         teamID,
		ContestID:  contestID,
		Name:       "Runtime AWD Team",
		CaptainID:  1,
		InviteCode: "runtime-awd",
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	seedInstanceServiceTeamMember(t, db, &model.TeamMember{
		ContestID: contestID,
		TeamID:    teamID,
		UserID:    2,
		JoinedAt:  now,
		CreatedAt: now,
	})
	seedInstanceServiceInstance(t, db, &model.Instance{
		ID:          1201,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 201,
		ServiceID:   &serviceID,
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:31201",
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	service := instanceqry.NewInstanceService(runtimeinfrarepo.NewRepository(db), &config.ContainerConfig{})

	items, err := service.GetUserInstances(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetUserInstances() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 visible awd instance, got %+v", items)
	}
	if items[0].ChallengeID != 202 {
		t.Fatalf("expected awd instance challenge id from contest service, got %+v", items[0])
	}
	if items[0].ChallengeTitle != "Bank Portal" {
		t.Fatalf("expected awd instance title from contest service display name, got %+v", items[0])
	}
	if items[0].Category != model.DimensionPwn || items[0].Difficulty != model.ChallengeDifficultyHard || items[0].FlagType != model.FlagTypeDynamic {
		t.Fatalf("expected awd instance metadata from contest service snapshot, got %+v", items[0])
	}
	if items[0].AccessURL != "" {
		t.Fatalf("expected awd user instance list to hide raw access url, got %q", items[0].AccessURL)
	}
	if items[0].ContestMode != model.ContestModeAWD {
		t.Fatalf("expected awd contest mode in user instance list, got %+v", items[0])
	}
}

func TestInstanceServiceGetUserInstancesFiltersLegacyAWDInstanceWithoutServiceID(t *testing.T) {
	t.Parallel()

	db := newInstanceServiceTestDB(t)
	now := time.Now()
	contestID := int64(703)
	teamID := int64(803)

	seedInstanceServiceChallenge(t, db, &model.Challenge{
		ID:         221,
		Title:      "Legacy AWD Runtime Challenge",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyMedium,
		FlagType:   model.FlagTypeDynamic,
		Status:     model.ChallengeStatusPublished,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err := db.Create(&model.Contest{
		ID:        contestID,
		Title:     "AWD Contest",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusRunning,
		StartTime: now.Add(-time.Minute),
		EndTime:   now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	seedInstanceServiceTeam(t, db, &model.Team{
		ID:         teamID,
		ContestID:  contestID,
		Name:       "Runtime AWD Team",
		CaptainID:  1,
		InviteCode: "runtime-awd-legacy",
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	seedInstanceServiceTeamMember(t, db, &model.TeamMember{
		ContestID: contestID,
		TeamID:    teamID,
		UserID:    2,
		JoinedAt:  now,
		CreatedAt: now,
	})
	seedInstanceServiceInstance(t, db, &model.Instance{
		ID:          1202,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 221,
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:31202",
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	service := instanceqry.NewInstanceService(runtimeinfrarepo.NewRepository(db), &config.ContainerConfig{})

	items, err := service.GetUserInstances(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetUserInstances() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected legacy awd instance without service_id to be filtered out, got %+v", items)
	}
}

func TestInstanceServiceGetUserInstancesHidesControlledAWDInstance(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		scopeType   string
		controlType string
		serviceID   int64
	}{
		{
			name:        "team_retired",
			scopeType:   model.AWDScopeControlScopeTeam,
			controlType: model.AWDScopeControlTypeRetired,
			serviceID:   0,
		},
		{
			name:        "service_disabled",
			scopeType:   model.AWDScopeControlScopeTeamService,
			controlType: model.AWDScopeControlTypeServiceDisabled,
			serviceID:   9703,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			db := newInstanceServiceTestDB(t)
			now := time.Now().UTC()
			contestID := int64(704)
			teamID := int64(804)
			serviceID := int64(9703)

			seedInstanceServiceChallenge(t, db, &model.Challenge{
				ID:         223,
				Title:      "Controlled AWD Runtime Challenge",
				Category:   model.DimensionWeb,
				Difficulty: model.ChallengeDifficultyMedium,
				FlagType:   model.FlagTypeDynamic,
				Status:     model.ChallengeStatusPublished,
				CreatedAt:  now,
				UpdatedAt:  now,
			})
			if err := db.Create(&model.Contest{
				ID:        contestID,
				Title:     "AWD Contest",
				Mode:      model.ContestModeAWD,
				Status:    model.ContestStatusRunning,
				StartTime: now.Add(-time.Minute),
				EndTime:   now.Add(time.Hour),
				CreatedAt: now,
				UpdatedAt: now,
			}).Error; err != nil {
				t.Fatalf("create contest: %v", err)
			}
			if err := db.Create(&model.ContestAWDService{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  223,
				DisplayName:     "Controlled Portal",
				IsVisible:       true,
				ServiceSnapshot: `{"name":"Controlled Portal","category":"web","difficulty":"medium","flag_config":{"flag_type":"dynamic","flag_prefix":"awd"}}`,
				CreatedAt:       now,
				UpdatedAt:       now,
			}).Error; err != nil {
				t.Fatalf("create contest awd service: %v", err)
			}
			seedInstanceServiceTeam(t, db, &model.Team{
				ID:         teamID,
				ContestID:  contestID,
				Name:       "Runtime AWD Team",
				CaptainID:  1,
				InviteCode: "runtime-awd-controlled",
				MaxMembers: 4,
				CreatedAt:  now,
				UpdatedAt:  now,
			})
			seedInstanceServiceTeamMember(t, db, &model.TeamMember{
				ContestID: contestID,
				TeamID:    teamID,
				UserID:    2,
				JoinedAt:  now,
				CreatedAt: now,
			})
			seedInstanceServiceInstance(t, db, &model.Instance{
				ID:          1203,
				UserID:      1,
				ContestID:   &contestID,
				TeamID:      &teamID,
				ChallengeID: 223,
				ServiceID:   &serviceID,
				Status:      model.InstanceStatusRunning,
				AccessURL:   "http://127.0.0.1:31203",
				ExpiresAt:   now.Add(time.Hour),
				MaxExtends:  2,
				CreatedAt:   now,
				UpdatedAt:   now,
			})
			if err := db.Create(&model.AWDScopeControl{
				ContestID:   contestID,
				TeamID:      teamID,
				ScopeType:   tc.scopeType,
				ServiceID:   tc.serviceID,
				ControlType: tc.controlType,
				Reason:      tc.name,
				CreatedAt:   now,
				UpdatedAt:   now,
			}).Error; err != nil {
				t.Fatalf("create awd scope control: %v", err)
			}

			service := instanceqry.NewInstanceService(runtimeinfrarepo.NewRepository(db), &config.ContainerConfig{})
			items, err := service.GetUserInstances(context.Background(), 2)
			if err != nil {
				t.Fatalf("GetUserInstances() error = %v", err)
			}
			if len(items) != 0 {
				t.Fatalf("expected controlled awd instance to be hidden, got %+v", items)
			}
		})
	}
}

func TestInstanceServiceGetUserInstancesIncludesPendingInstance(t *testing.T) {
	t.Parallel()

	db := newInstanceServiceTestDB(t)
	now := time.Now()

	seedInstanceServiceChallenge(t, db, &model.Challenge{
		ID:         103,
		Title:      "Queued Challenge",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		FlagType:   model.FlagTypeStatic,
		Status:     model.ChallengeStatusPublished,
		Points:     120,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	seedInstanceServiceInstance(t, db, &model.Instance{
		ID:          1003,
		UserID:      2,
		ChallengeID: 103,
		Status:      model.InstanceStatusPending,
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	service := instanceqry.NewInstanceService(runtimeinfrarepo.NewRepository(db), &config.ContainerConfig{})

	items, err := service.GetUserInstances(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetUserInstances() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != 1003 || items[0].Status != model.InstanceStatusPending {
		t.Fatalf("expected pending instance to be visible, got %+v", items)
	}
}

func TestInstanceServiceGetUserInstancesIncludesFailedInstance(t *testing.T) {
	t.Parallel()

	db := newInstanceServiceTestDB(t)
	now := time.Now()

	seedInstanceServiceChallenge(t, db, &model.Challenge{
		ID:         104,
		Title:      "Failed Challenge",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		FlagType:   model.FlagTypeStatic,
		Status:     model.ChallengeStatusPublished,
		Points:     120,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	seedInstanceServiceInstance(t, db, &model.Instance{
		ID:          1004,
		UserID:      2,
		ChallengeID: 104,
		Status:      model.InstanceStatusFailed,
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	service := instanceqry.NewInstanceService(runtimeinfrarepo.NewRepository(db), &config.ContainerConfig{})

	items, err := service.GetUserInstances(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetUserInstances() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != 1004 || items[0].Status != model.InstanceStatusFailed {
		t.Fatalf("expected failed instance to be visible, got %+v", items)
	}
}

func TestInstanceServiceGetUserInstancesMarksExpiredRunningInstance(t *testing.T) {
	t.Parallel()

	db := newInstanceServiceTestDB(t)
	now := time.Now()

	seedInstanceServiceChallenge(t, db, &model.Challenge{
		ID:         105,
		Title:      "Expired Challenge",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		FlagType:   model.FlagTypeStatic,
		Status:     model.ChallengeStatusPublished,
		Points:     120,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	seedInstanceServiceInstance(t, db, &model.Instance{
		ID:          1005,
		UserID:      2,
		ChallengeID: 105,
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30005",
		ExpiresAt:   now.Add(-2 * time.Minute),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	service := instanceqry.NewInstanceService(runtimeinfrarepo.NewRepository(db), &config.ContainerConfig{})

	items, err := service.GetUserInstances(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetUserInstances() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != 1005 {
		t.Fatalf("expected expired instance to remain visible, got %+v", items)
	}
	if items[0].Status != model.InstanceStatusExpired {
		t.Fatalf("expected expired status, got %+v", items[0])
	}
}

func TestInstanceServiceGetAccessURLRejectsExpiredRunningInstance(t *testing.T) {
	t.Parallel()

	db := newInstanceServiceTestDB(t)
	now := time.Now()

	seedInstanceServiceChallenge(t, db, &model.Challenge{
		ID:         106,
		Title:      "Expired Access",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		FlagType:   model.FlagTypeStatic,
		Status:     model.ChallengeStatusPublished,
		Points:     120,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	seedInstanceServiceInstance(t, db, &model.Instance{
		ID:          1006,
		UserID:      2,
		ChallengeID: 106,
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30006",
		ExpiresAt:   now.Add(-time.Minute),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	service := instanceqry.NewInstanceService(runtimeinfrarepo.NewRepository(db), &config.ContainerConfig{})

	_, err := service.GetAccessURL(context.Background(), 1006, 2)
	if err == nil || err.Error() != errcode.ErrInstanceExpired.Error() {
		t.Fatalf("expected instance expired error, got %v", err)
	}
}

func TestInstanceServiceGetAccessURLRejectsControlledAWDInstance(t *testing.T) {
	t.Parallel()

	db := newInstanceServiceTestDB(t)
	now := time.Now().UTC()
	contestID := int64(705)
	teamID := int64(805)
	serviceID := int64(9705)

	seedInstanceServiceChallenge(t, db, &model.Challenge{
		ID:         225,
		Title:      "Controlled Access",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		FlagType:   model.FlagTypeStatic,
		Status:     model.ChallengeStatusPublished,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err := db.Create(&model.Contest{
		ID:        contestID,
		Title:     "AWD Contest",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusRunning,
		StartTime: now.Add(-time.Minute),
		EndTime:   now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create(&model.ContestAWDService{
		ID:              serviceID,
		ContestID:       contestID,
		AWDChallengeID:  225,
		DisplayName:     "Controlled Access",
		IsVisible:       true,
		ServiceSnapshot: `{"name":"Controlled Access","category":"web","difficulty":"easy","flag_config":{"flag_type":"static","flag_prefix":"flag"}}`,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("create contest awd service: %v", err)
	}
	seedInstanceServiceTeam(t, db, &model.Team{
		ID:         teamID,
		ContestID:  contestID,
		Name:       "Runtime AWD Team",
		CaptainID:  1,
		InviteCode: "runtime-awd-access",
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	seedInstanceServiceTeamMember(t, db, &model.TeamMember{
		ContestID: contestID,
		TeamID:    teamID,
		UserID:    2,
		JoinedAt:  now,
		CreatedAt: now,
	})
	seedInstanceServiceInstance(t, db, &model.Instance{
		ID:          1205,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 225,
		ServiceID:   &serviceID,
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:31205",
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err := db.Create(&model.AWDScopeControl{
		ContestID:   contestID,
		TeamID:      teamID,
		ScopeType:   model.AWDScopeControlScopeTeamService,
		ServiceID:   serviceID,
		ControlType: model.AWDScopeControlTypeServiceDisabled,
		Reason:      "disabled",
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd scope control: %v", err)
	}

	service := instanceqry.NewInstanceService(runtimeinfrarepo.NewRepository(db), &config.ContainerConfig{})
	_, err := service.GetAccessURL(context.Background(), 1205, 2)
	if err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected controlled awd instance access to be forbidden, got %v", err)
	}
}

func TestInstanceServiceListTeacherInstancesScopesTeacherAndAppliesFilters(t *testing.T) {
	t.Parallel()

	db := newInstanceServiceTestDB(t)
	now := time.Now()

	seedInstanceServiceUser(t, db, &model.User{ID: 1, Username: "teacher-a", Role: model.RoleTeacher, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceUser(t, db, &model.User{ID: 2, Username: "alice", StudentNo: "S-1001", Role: model.RoleStudent, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceUser(t, db, &model.User{ID: 3, Username: "bob", StudentNo: "S-1002", Role: model.RoleStudent, ClassName: "Class B", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceChallenge(t, db, &model.Challenge{ID: 11, Title: "web-101", Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceInstance(t, db, &model.Instance{ID: 101, UserID: 2, ChallengeID: 11, ContainerID: "inst-a", Status: model.InstanceStatusRunning, ExpiresAt: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceInstance(t, db, &model.Instance{ID: 102, UserID: 3, ChallengeID: 11, ContainerID: "inst-b", Status: model.InstanceStatusRunning, ExpiresAt: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceInstance(t, db, &model.Instance{ID: 103, UserID: 2, ChallengeID: 11, ContainerID: "inst-stopped", Status: model.InstanceStatusStopped, ExpiresAt: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now})

	service := instanceqry.NewInstanceService(runtimeinfrarepo.NewRepository(db), &config.ContainerConfig{})

	items, err := service.ListTeacherInstances(context.Background(), 1, model.RoleTeacher, nil)
	if err != nil {
		t.Fatalf("ListTeacherInstances() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 visible instance, got %d (%+v)", len(items), items)
	}
	if items[0].StudentUsername != "alice" || items[0].ClassName != "Class A" {
		t.Fatalf("unexpected item: %+v", items[0])
	}

	filtered, err := service.ListTeacherInstances(context.Background(), 1, model.RoleTeacher, &dto.TeacherInstanceQuery{
		Keyword:   "ali",
		StudentNo: "S-1001",
	})
	if err != nil {
		t.Fatalf("ListTeacherInstances() with filters error = %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != 101 {
		t.Fatalf("unexpected filtered result: %+v", filtered)
	}

	byStudentNoKeyword, err := service.ListTeacherInstances(context.Background(), 1, model.RoleTeacher, &dto.TeacherInstanceQuery{
		Keyword: "1001",
	})
	if err != nil {
		t.Fatalf("ListTeacherInstances() with student_no keyword error = %v", err)
	}
	if len(byStudentNoKeyword) != 1 || byStudentNoKeyword[0].ID != 101 {
		t.Fatalf("expected keyword to match student_no, got %+v", byStudentNoKeyword)
	}
}

func TestInstanceServiceListTeacherInstancesPrefersContestAWDServiceMetadata(t *testing.T) {
	t.Parallel()

	db := newInstanceServiceTestDB(t)
	now := time.Now()
	contestID := int64(702)
	serviceID := int64(9702)

	seedInstanceServiceUser(t, db, &model.User{ID: 1, Username: "teacher-a", Role: model.RoleTeacher, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceUser(t, db, &model.User{ID: 2, Username: "alice", StudentNo: "S-1001", Role: model.RoleStudent, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceChallenge(t, db, &model.Challenge{ID: 211, Title: "Legacy Runtime Challenge", Category: model.DimensionWeb, Difficulty: model.ChallengeDifficultyEasy, FlagType: model.FlagTypeStatic, Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now})
	if err := db.Create(&model.Contest{
		ID:        contestID,
		Title:     "AWD Contest",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusRunning,
		StartTime: now.Add(-time.Minute),
		EndTime:   now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create(&model.ContestAWDService{
		ID:              serviceID,
		ContestID:       contestID,
		AWDChallengeID:  212,
		DisplayName:     "Bank Portal",
		Order:           1,
		IsVisible:       true,
		ScoreConfig:     `{"points":300}`,
		RuntimeConfig:   `{"checker_type":"http_standard"}`,
		ServiceSnapshot: `{"name":"Bank Portal","category":"pwn","difficulty":"hard","flag_config":{"flag_type":"dynamic","flag_prefix":"awd"}}`,
		ValidationState: model.AWDCheckerValidationStatePassed,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("create contest awd service: %v", err)
	}
	seedInstanceServiceInstance(t, db, &model.Instance{
		ID:          1301,
		UserID:      2,
		ContestID:   &contestID,
		ChallengeID: 211,
		ServiceID:   &serviceID,
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:31301",
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	service := instanceqry.NewInstanceService(runtimeinfrarepo.NewRepository(db), &config.ContainerConfig{})

	items, err := service.ListTeacherInstances(context.Background(), 1, model.RoleTeacher, nil)
	if err != nil {
		t.Fatalf("ListTeacherInstances() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 visible awd teacher instance, got %+v", items)
	}
	if items[0].ChallengeID != 212 || items[0].ChallengeTitle != "Bank Portal" {
		t.Fatalf("expected teacher instance metadata from contest awd service, got %+v", items[0])
	}
}

func TestInstanceServiceListTeacherInstancesFiltersLegacyAWDInstanceWithoutServiceID(t *testing.T) {
	t.Parallel()

	db := newInstanceServiceTestDB(t)
	now := time.Now()
	contestID := int64(704)

	seedInstanceServiceUser(t, db, &model.User{ID: 1, Username: "teacher-a", Role: model.RoleTeacher, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceUser(t, db, &model.User{ID: 2, Username: "alice", StudentNo: "S-1001", Role: model.RoleStudent, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceChallenge(t, db, &model.Challenge{ID: 222, Title: "Legacy AWD Runtime Challenge", Category: model.DimensionWeb, Difficulty: model.ChallengeDifficultyMedium, FlagType: model.FlagTypeDynamic, Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now})
	if err := db.Create(&model.Contest{
		ID:        contestID,
		Title:     "AWD Contest",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusRunning,
		StartTime: now.Add(-time.Minute),
		EndTime:   now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	seedInstanceServiceInstance(t, db, &model.Instance{
		ID:          1302,
		UserID:      2,
		ContestID:   &contestID,
		ChallengeID: 222,
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:31302",
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	service := instanceqry.NewInstanceService(runtimeinfrarepo.NewRepository(db), &config.ContainerConfig{})

	items, err := service.ListTeacherInstances(context.Background(), 1, model.RoleTeacher, nil)
	if err != nil {
		t.Fatalf("ListTeacherInstances() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected legacy awd teacher instance without service_id to be filtered out, got %+v", items)
	}
}

func TestInstanceServiceDestroyTeacherInstanceHonorsClassScope(t *testing.T) {
	t.Parallel()

	db := newInstanceServiceTestDB(t)
	now := time.Now()

	seedInstanceServiceUser(t, db, &model.User{ID: 1, Username: "teacher-a", Role: model.RoleTeacher, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceUser(t, db, &model.User{ID: 2, Username: "alice", Role: model.RoleStudent, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceUser(t, db, &model.User{ID: 3, Username: "bob", Role: model.RoleStudent, ClassName: "Class B", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceChallenge(t, db, &model.Challenge{ID: 11, Title: "web-101", Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceInstance(t, db, &model.Instance{ID: 201, UserID: 2, ChallengeID: 11, ContainerID: "inst-a", Status: model.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceInstance(t, db, &model.Instance{ID: 202, UserID: 3, ChallengeID: 11, ContainerID: "inst-b", Status: model.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})

	service := instancecmd.NewInstanceService(
		runtimeinfrarepo.NewRepository(db),
		noopRuntimeCleaner{},
		&config.ContainerConfig{MaxExtends: 2, ExtendDuration: 30 * time.Minute},
		nil,
	)

	if err := service.DestroyTeacherInstance(context.Background(), 202, 1, model.RoleTeacher); err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected forbidden destroy, got %v", err)
	}

	if err := service.DestroyTeacherInstance(context.Background(), 201, 1, model.RoleTeacher); err != nil {
		t.Fatalf("DestroyTeacherInstance() error = %v", err)
	}

	var instance model.Instance
	if err := db.First(&instance, 201).Error; err != nil {
		t.Fatalf("load instance: %v", err)
	}
	if instance.Status != model.InstanceStatusStopped {
		t.Fatalf("expected stopped status, got %s", instance.Status)
	}
}

func newInstanceServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Challenge{}, &model.Instance{}, &model.PortAllocation{}, &model.ContestRegistration{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	if err := db.AutoMigrate(&model.Team{}, &model.TeamMember{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	if err := db.AutoMigrate(&model.Contest{}, &model.ContestAWDService{}); err != nil {
		t.Fatalf("migrate awd tables: %v", err)
	}
	if err := db.AutoMigrate(&model.AWDScopeControl{}); err != nil {
		t.Fatalf("migrate awd scope control tables: %v", err)
	}
	if err := db.AutoMigrate(&model.AWDServiceOperation{}); err != nil {
		t.Fatalf("migrate awd operation tables: %v", err)
	}
	return db
}

func seedInstanceServiceUser(t *testing.T, db *gorm.DB, user *model.User) {
	t.Helper()
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
}

func seedInstanceServiceChallenge(t *testing.T, db *gorm.DB, challenge *model.Challenge) {
	t.Helper()
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
}

func seedInstanceServiceTeam(t *testing.T, db *gorm.DB, team *model.Team) {
	t.Helper()
	if err := db.Create(team).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
}

func seedInstanceServiceTeamMember(t *testing.T, db *gorm.DB, member *model.TeamMember) {
	t.Helper()
	if err := db.Create(member).Error; err != nil {
		t.Fatalf("create team member: %v", err)
	}
}

func seedInstanceServiceInstance(t *testing.T, db *gorm.DB, instance *model.Instance) {
	t.Helper()
	if err := db.Create(instance).Error; err != nil {
		t.Fatalf("create instance: %v", err)
	}
}

type runtimeInstanceContextKey string

func TestInstanceServiceDestroyTeacherInstancePropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := runtimeInstanceContextKey("destroy-teacher")
	expectedCtxValue := "ctx-destroy-teacher"
	findByIDCalled := false
	findRequesterCalled := false
	findOwnerCalled := false
	updateCalled := false
	repo := &runtimeInstanceContextRepo{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Instance, error) {
			findByIDCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-by-id ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Instance{ID: id, UserID: 2, Status: model.InstanceStatusRunning}, nil
		},
		findUserByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-user ctx value %v, got %v", expectedCtxValue, got)
			}
			if userID == 1001 {
				findRequesterCalled = true
				return &model.User{ID: userID, Role: model.RoleTeacher, ClassName: "Class A"}, nil
			}
			findOwnerCalled = true
			return &model.User{ID: userID, Role: model.RoleStudent, ClassName: "Class A"}, nil
		},
		updateStatusAndReleasePortWithContextFn: func(ctx context.Context, id int64, status string) error {
			updateCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected update ctx value %v, got %v", expectedCtxValue, got)
			}
			if id != 201 || status != model.InstanceStatusStopped {
				t.Fatalf("unexpected update args: id=%d status=%s", id, status)
			}
			return nil
		},
	}
	service := instancecmd.NewInstanceService(repo, noopRuntimeCleaner{}, &config.ContainerConfig{}, nil)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.DestroyTeacherInstance(ctx, 201, 1001, model.RoleTeacher); err != nil {
		t.Fatalf("DestroyTeacherInstance() error = %v", err)
	}
	if !findByIDCalled || !findRequesterCalled || !findOwnerCalled || !updateCalled {
		t.Fatalf("expected all repository calls to happen, got findByID=%v requester=%v owner=%v update=%v", findByIDCalled, findRequesterCalled, findOwnerCalled, updateCalled)
	}
}

func TestInstanceServiceDestroyTeacherInstanceDoesNotCreateBackgroundContext(t *testing.T) {
	t.Parallel()

	repo := &runtimeInstanceContextRepo{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Instance, error) {
			if ctx != nil {
				t.Fatalf("expected find-by-id ctx to stay nil, got %v", ctx)
			}
			return &model.Instance{ID: id, UserID: 2, Status: model.InstanceStatusRunning}, nil
		},
		updateStatusAndReleasePortWithContextFn: func(ctx context.Context, id int64, status string) error {
			if ctx != nil {
				t.Fatalf("expected update ctx to stay nil, got %v", ctx)
			}
			return nil
		},
	}
	service := instancecmd.NewInstanceService(repo, noopRuntimeCleaner{}, &config.ContainerConfig{}, nil)

	if err := service.DestroyTeacherInstance(nil, 201, 1001, model.RoleAdmin); err != nil {
		t.Fatalf("DestroyTeacherInstance() error = %v", err)
	}
}

func TestInstanceQueryServiceDoesNotCreateBackgroundContext(t *testing.T) {
	t.Parallel()

	repo := &runtimeInstanceContextRepo{
		listVisibleByUserFn: func(ctx context.Context, userID int64) ([]runtimeports.UserVisibleInstanceRow, error) {
			if ctx != nil {
				t.Fatalf("expected list-visible ctx to stay nil, got %v", ctx)
			}
			return []runtimeports.UserVisibleInstanceRow{}, nil
		},
	}
	service := instanceqry.NewInstanceService(repo, &config.ContainerConfig{})

	if _, err := service.GetUserInstances(nil, 2); err != nil {
		t.Fatalf("GetUserInstances() error = %v", err)
	}
}
