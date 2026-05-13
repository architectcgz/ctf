package commands

import (
	"context"
	"ctf-platform/internal/model"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"ctf-platform/internal/module/challenge/domain"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/internal/module/challenge/testsupport"
	platformevents "ctf-platform/internal/platform/events"
	flagcrypto "ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newTestService(repo challengeCommandRepository, imageRepo challengeports.ImageQueryRepository) *ChallengeService {
	return NewChallengeService(
		nil,
		challengeinfra.NewChallengeCommandRepository(repo),
		challengeinfra.NewImageQueryRepository(imageRepo),
		nil,
		nil,
		nil,
		SelfCheckConfig{},
		zap.NewNop(),
	)
}

func newDBBackedChallengeService(
	db *gorm.DB,
	repo *challengeinfra.Repository,
	imageRepo *challengeinfra.ImageRepository,
	runtimeProbe challengeports.ChallengeRuntimeProbe,
	cfg SelfCheckConfig,
) *ChallengeService {
	return NewChallengeService(
		db,
		challengeinfra.NewChallengeCommandRepository(repo),
		challengeinfra.NewImageQueryRepository(imageRepo),
		challengeinfra.NewTopologyServiceRepository(repo),
		repo,
		runtimeProbe,
		cfg,
		zap.NewNop(),
	)
}

func TestServiceCreateChallengeSuccess(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	// 创建测试镜像
	db.Create(&model.Image{ID: 1, Name: "test-image"})

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := newTestService(repo, imageRepo)

	resp, err := service.CreateChallenge(context.Background(), 1001, CreateChallengeInput{
		Title:           "Test Challenge",
		Description:     "Test",
		Category:        "web",
		Difficulty:      "easy",
		Points:          100,
		ImageID:         1,
		InstanceSharing: model.InstanceSharingPerUser,
	})

	if err != nil {
		t.Fatalf("CreateChallenge() error = %v", err)
	}
	if resp.Status != model.ChallengeStatusDraft {
		t.Fatalf("unexpected status: %s", resp.Status)
	}
	if resp.CreatedBy == nil || *resp.CreatedBy != 1001 {
		t.Fatalf("unexpected created_by: %+v", resp.CreatedBy)
	}
	if resp.InstanceSharing != model.InstanceSharingPerUser {
		t.Fatalf("unexpected instance sharing: %s", resp.InstanceSharing)
	}
}

func TestServiceCreateChallengeImageNotFound(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := newTestService(repo, imageRepo)

	_, err := service.CreateChallenge(context.Background(), 1001, CreateChallengeInput{
		ImageID: 999,
	})
	if err == nil || err.Error() != errcode.ErrNotFound.Error() {
		t.Fatalf("expected image not found error, got %v", err)
	}
}

func TestServiceCreateChallengeWithoutImageSuccess(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := newTestService(repo, imageRepo)

	resp, err := service.CreateChallenge(context.Background(), 1001, CreateChallengeInput{
		Title:       "No Target Challenge",
		Description: "No target required",
		Category:    "misc",
		Difficulty:  "easy",
		Points:      50,
		ImageID:     0,
	})

	if err != nil {
		t.Fatalf("CreateChallenge() without image error = %v", err)
	}
	if resp.ImageID != 0 {
		t.Fatalf("expected image_id=0, got %d", resp.ImageID)
	}
}

func TestServiceUpdateChallengeRejectsSharedDynamicFlagCombination(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	salt, err := flagcrypto.GenerateSalt()
	if err != nil {
		t.Fatalf("generate salt: %v", err)
	}
	challenge := &model.Challenge{
		Title:           "dynamic-flag",
		Description:     "desc",
		Category:        model.DimensionCrypto,
		Difficulty:      model.ChallengeDifficultyEasy,
		Points:          100,
		Status:          model.ChallengeStatusDraft,
		FlagType:        model.FlagTypeDynamic,
		FlagSalt:        salt,
		FlagPrefix:      "flag",
		InstanceSharing: model.InstanceSharingPerUser,
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := newTestService(repo, imageRepo)

	err = service.UpdateChallenge(context.Background(), challenge.ID, UpdateChallengeInput{
		InstanceSharing: model.InstanceSharingShared,
	})
	if err == nil || err.Error() != errcode.ErrInvalidParams.Error() {
		t.Fatalf("expected invalid params when enabling shared for dynamic flag challenge, got %v", err)
	}
}

func TestServiceUpdateChallengeRejectsSharedInjectFlagTopologyCombination(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &model.Challenge{
		Title:           "inject-flag-topology",
		Description:     "desc",
		Category:        model.DimensionWeb,
		Difficulty:      model.ChallengeDifficultyEasy,
		Points:          100,
		Status:          model.ChallengeStatusDraft,
		FlagType:        model.FlagTypeStatic,
		InstanceSharing: model.InstanceSharingPerUser,
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	rawSpec, err := model.EncodeTopologySpec(model.TopologySpec{
		Nodes: []model.TopologyNode{
			{Key: "web", Name: "Web", ServicePort: 8080, InjectFlag: true},
		},
	})
	if err != nil {
		t.Fatalf("encode topology spec: %v", err)
	}
	if err := db.Create(&model.ChallengeTopology{
		ChallengeID:  challenge.ID,
		EntryNodeKey: "web",
		Spec:         rawSpec,
	}).Error; err != nil {
		t.Fatalf("create topology: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := newDBBackedChallengeService(nil, repo, imageRepo, nil, SelfCheckConfig{})

	err = service.UpdateChallenge(context.Background(), challenge.ID, UpdateChallengeInput{
		InstanceSharing: model.InstanceSharingShared,
	})
	if err == nil || err.Error() != errcode.ErrInvalidParams.Error() {
		t.Fatalf("expected invalid params when enabling shared for inject_flag topology challenge, got %v", err)
	}
}

func TestServiceDeleteChallengeWithRunningInstances(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	// 创建靶场和运行中的实例
	challenge := &model.Challenge{Title: "Test", Status: "draft"}
	db.Create(challenge)
	db.Create(&model.Instance{ChallengeID: challenge.ID, Status: "running"})

	repo := challengeinfra.NewRepository(db)
	service := newTestService(repo, nil)

	err := service.DeleteChallenge(context.Background(), challenge.ID)
	if err == nil {
		t.Fatal("expected running instances error, got nil")
	}
	if err.Error() != domain.ErrMsgHasRunningStudents {
		t.Fatalf("expected running instances error, got %v", err)
	}
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrConflict.Code {
		t.Fatalf("expected conflict app error, got %v", err)
	}
}

func TestServicePublishChallengeNoImage(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &model.Challenge{Title: "Test", ImageID: 0}
	db.Create(challenge)

	repo := challengeinfra.NewRepository(db)
	service := newTestService(repo, nil)

	err := service.PublishChallenge(context.Background(), challenge.ID)
	if err != nil {
		t.Fatalf("PublishChallenge() error = %v", err)
	}

	published, findErr := repo.FindByID(context.Background(), challenge.ID)
	if findErr != nil {
		t.Fatalf("FindByID() error = %v", findErr)
	}
	if published.Status != model.ChallengeStatusPublished {
		t.Fatalf("expected published status, got %s", published.Status)
	}
}

func TestServiceDispatchPublishCheckJobsPublishesChallengeAndNotifiesRequester(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	teacher := &model.User{Username: "teacher", PasswordHash: "x", Role: model.RoleTeacher, Status: model.UserStatusActive}
	if err := db.Create(teacher).Error; err != nil {
		t.Fatalf("create teacher: %v", err)
	}
	image := &model.Image{Name: "ctf/web-demo", Tag: "latest", Status: model.ImageStatusAvailable}
	if err := db.Create(image).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	salt, err := flagcrypto.GenerateSalt()
	if err != nil {
		t.Fatalf("generate salt: %v", err)
	}
	challenge := &model.Challenge{
		Title:      "publish-me",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		ImageID:    image.ID,
		Status:     model.ChallengeStatusDraft,
		CreatedBy:  &teacher.ID,
		FlagType:   model.FlagTypeStatic,
		FlagSalt:   salt,
		FlagHash:   flagcrypto.HashStaticFlag("flag{ok}", salt),
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	probe := &fakeChallengeRuntimeProbe{
		containerResultAccessURL: "http://127.0.0.1:30001",
		containerResultDetails: model.InstanceRuntimeDetails{
			Containers: []model.InstanceRuntimeContainer{{ContainerID: "ctr-1"}},
			Networks:   []model.InstanceRuntimeNetwork{{NetworkID: "net-1"}},
		},
	}
	service := newDBBackedChallengeService(db, repo, imageRepo, probe, SelfCheckConfig{
		PublishCheckBatchSize: 1,
	})
	var publishedEvents []platformevents.Event
	service.SetEventBus(&challengeCommandEventBusStub{
		publishFn: func(ctx context.Context, evt platformevents.Event) error {
			publishedEvents = append(publishedEvents, evt)
			return nil
		},
	})

	job, err := service.RequestPublishCheck(context.Background(), teacher.ID, challenge.ID)
	if err != nil {
		t.Fatalf("RequestPublishCheck() error = %v", err)
	}
	if job.Status != "queued" || !job.Active {
		t.Fatalf("unexpected requested job status: %s", job.Status)
	}

	service.dispatchPublishCheckJobs(context.Background())

	publishedChallenge, err := repo.FindByID(context.Background(), challenge.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if publishedChallenge.Status != model.ChallengeStatusPublished {
		t.Fatalf("expected published challenge status, got %s", publishedChallenge.Status)
	}

	latest, err := service.GetLatestPublishCheck(context.Background(), challenge.ID)
	if err != nil {
		t.Fatalf("GetLatestPublishCheck() error = %v", err)
	}
	if latest.Status != "succeeded" || latest.Active {
		t.Fatalf("expected passed publish check job, got %+v", latest)
	}
	if latest.Result == nil || !latest.Result.Precheck.Passed || !latest.Result.Runtime.Passed {
		t.Fatalf("expected successful self-check result, got %+v", latest.Result)
	}
	if latest.PublishedAt == nil {
		t.Fatalf("expected published_at to be set, got %+v", latest)
	}

	if len(publishedEvents) != 1 {
		t.Fatalf("expected 1 challenge event, got %+v", publishedEvents)
	}
	if publishedEvents[0].Name != challengecontracts.EventPublishCheckFinished {
		t.Fatalf("unexpected event name: %+v", publishedEvents[0])
	}
	payload, ok := publishedEvents[0].Payload.(challengecontracts.PublishCheckFinishedEvent)
	if !ok {
		t.Fatalf("unexpected event payload type: %T", publishedEvents[0].Payload)
	}
	if !payload.Passed || payload.ChallengeID != challenge.ID || payload.UserID != teacher.ID {
		t.Fatalf("unexpected event payload: %+v", payload)
	}
}

func TestServiceDispatchPublishCheckJobsKeepsDraftOnFailureAndNotifiesRequester(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	teacher := &model.User{Username: "teacher", PasswordHash: "x", Role: model.RoleTeacher, Status: model.UserStatusActive}
	if err := db.Create(teacher).Error; err != nil {
		t.Fatalf("create teacher: %v", err)
	}
	salt, err := flagcrypto.GenerateSalt()
	if err != nil {
		t.Fatalf("generate salt: %v", err)
	}
	challenge := &model.Challenge{
		Title:      "no-image",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		Status:     model.ChallengeStatusDraft,
		CreatedBy:  &teacher.ID,
		FlagType:   model.FlagTypeStatic,
		FlagSalt:   salt,
		FlagHash:   flagcrypto.HashStaticFlag("flag{ok}", salt),
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := newDBBackedChallengeService(db, repo, imageRepo, &fakeChallengeRuntimeProbe{}, SelfCheckConfig{
		PublishCheckBatchSize: 1,
	})
	var publishedEvents []platformevents.Event
	service.SetEventBus(&challengeCommandEventBusStub{
		publishFn: func(ctx context.Context, evt platformevents.Event) error {
			publishedEvents = append(publishedEvents, evt)
			return nil
		},
	})

	if _, err := service.RequestPublishCheck(context.Background(), teacher.ID, challenge.ID); err != nil {
		t.Fatalf("RequestPublishCheck() error = %v", err)
	}

	service.dispatchPublishCheckJobs(context.Background())

	stored, err := repo.FindByID(context.Background(), challenge.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if stored.Status != model.ChallengeStatusDraft {
		t.Fatalf("expected challenge to stay draft, got %s", stored.Status)
	}

	latest, err := service.GetLatestPublishCheck(context.Background(), challenge.ID)
	if err != nil {
		t.Fatalf("GetLatestPublishCheck() error = %v", err)
	}
	if latest.Status != model.ChallengePublishCheckStatusFailed || latest.Active {
		t.Fatalf("expected failed publish check job, got %+v", latest)
	}
	if latest.FailureSummary == "" {
		t.Fatalf("expected failure summary, got %+v", latest)
	}

	if len(publishedEvents) != 1 {
		t.Fatalf("expected 1 challenge event, got %+v", publishedEvents)
	}
	payload, ok := publishedEvents[0].Payload.(challengecontracts.PublishCheckFinishedEvent)
	if !ok {
		t.Fatalf("unexpected event payload type: %T", publishedEvents[0].Payload)
	}
	if payload.Passed {
		t.Fatalf("expected failure event, got %+v", payload)
	}
	if payload.FailureSummary == "" {
		t.Fatalf("expected failure summary in event, got %+v", payload)
	}
}

func TestServiceDispatchPublishCheckJobsPublishesAttachmentOnlyChallenge(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	teacher := &model.User{Username: "teacher", PasswordHash: "x", Role: model.RoleTeacher, Status: model.UserStatusActive}
	if err := db.Create(teacher).Error; err != nil {
		t.Fatalf("create teacher: %v", err)
	}
	salt, err := flagcrypto.GenerateSalt()
	if err != nil {
		t.Fatalf("generate salt: %v", err)
	}
	challenge := &model.Challenge{
		Title:         "attachment-only",
		Category:      model.DimensionWeb,
		Difficulty:    model.ChallengeDifficultyEasy,
		Points:        100,
		Status:        model.ChallengeStatusDraft,
		CreatedBy:     &teacher.ID,
		AttachmentURL: "/api/v1/challenges/attachments/imports/web-source-audit-double-wrap-01/source.html",
		FlagType:      model.FlagTypeStatic,
		FlagSalt:      salt,
		FlagHash:      flagcrypto.HashStaticFlag("flag{ok}", salt),
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	probe := &fakeChallengeRuntimeProbe{}
	service := newDBBackedChallengeService(db, repo, imageRepo, probe, SelfCheckConfig{
		PublishCheckBatchSize: 1,
	})
	var publishedEvents []platformevents.Event
	service.SetEventBus(&challengeCommandEventBusStub{
		publishFn: func(ctx context.Context, evt platformevents.Event) error {
			publishedEvents = append(publishedEvents, evt)
			return nil
		},
	})

	if _, err := service.RequestPublishCheck(context.Background(), teacher.ID, challenge.ID); err != nil {
		t.Fatalf("RequestPublishCheck() error = %v", err)
	}

	service.dispatchPublishCheckJobs(context.Background())

	publishedChallenge, err := repo.FindByID(context.Background(), challenge.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if publishedChallenge.Status != model.ChallengeStatusPublished {
		t.Fatalf("expected attachment-only challenge to publish, got %s", publishedChallenge.Status)
	}
	if probe.createContainerCalled || probe.createTopologyCalled {
		t.Fatalf("attachment-only challenge publish check should skip runtime startup")
	}

	latest, err := service.GetLatestPublishCheck(context.Background(), challenge.ID)
	if err != nil {
		t.Fatalf("GetLatestPublishCheck() error = %v", err)
	}
	if latest.Status != "succeeded" || latest.Active {
		t.Fatalf("expected passed publish check job, got %+v", latest)
	}
	if latest.Result == nil || !latest.Result.Precheck.Passed || !latest.Result.Runtime.Passed {
		t.Fatalf("expected successful self-check result, got %+v", latest.Result)
	}
	if len(publishedEvents) != 1 {
		t.Fatalf("expected 1 challenge event, got %+v", publishedEvents)
	}
}

func TestGetLatestPublishCheckIgnoresStaleJobsAfterChallengeUpdate(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	teacher := &model.User{Username: "teacher", PasswordHash: "x", Role: model.RoleTeacher, Status: model.UserStatusActive}
	if err := db.Create(teacher).Error; err != nil {
		t.Fatalf("create teacher: %v", err)
	}

	createdAt := time.Date(2026, 4, 9, 10, 55, 5, 0, time.UTC)
	updatedAt := createdAt.Add(2 * time.Hour)
	challenge := &model.Challenge{
		Title:      "Web-01 源码审计：双层伪装",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		ImageID:    0,
		Status:     model.ChallengeStatusDraft,
		CreatedBy:  &teacher.ID,
		UpdatedAt:  updatedAt,
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Model(challenge).Update("updated_at", updatedAt).Error; err != nil {
		t.Fatalf("update challenge updated_at: %v", err)
	}

	job := &model.ChallengePublishCheckJob{
		ChallengeID:    challenge.ID,
		RequestedBy:    teacher.ID,
		Status:         model.ChallengePublishCheckStatusFailed,
		RequestSource:  "admin_publish",
		FailureSummary: "单容器拉起失败: Error response from daemon: No such image: registry.example.edu/ctf/web-source-audit-double-wrap-01:20260404",
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	}
	if err := db.Create(job).Error; err != nil {
		t.Fatalf("create publish check job: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := newDBBackedChallengeService(db, repo, imageRepo, nil, SelfCheckConfig{})

	latest, err := service.GetLatestPublishCheck(context.Background(), challenge.ID)
	if err == nil || err.Error() != errcode.ErrNotFound.Error() {
		t.Fatalf("expected not found for stale publish check job, got latest=%+v err=%v", latest, err)
	}
}
