package commands

import (
	"context"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/internal/module/challenge/testsupport"
	flagcrypto "ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

type stubChallengeNotificationSender struct {
	calls []stubChallengeNotificationCall
}

type stubChallengeNotificationCall struct {
	userID         int64
	challengeID    int64
	challengeTitle string
	passed         bool
	failureSummary string
}

func (s *stubChallengeNotificationSender) SendChallengePublishCheckResult(_ context.Context, userID int64, challengeID int64, challengeTitle string, passed bool, failureSummary string) error {
	s.calls = append(s.calls, stubChallengeNotificationCall{
		userID:         userID,
		challengeID:    challengeID,
		challengeTitle: challengeTitle,
		passed:         passed,
		failureSummary: failureSummary,
	})
	return nil
}

func newTestService(repo challengeports.ChallengeCommandRepository, imageRepo challengeports.ImageRepository) *ChallengeService {
	return NewChallengeService(nil, repo, imageRepo, nil, nil, SelfCheckConfig{}, zap.NewNop())
}

func TestServiceCreateChallengeSuccess(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	// 创建测试镜像
	db.Create(&model.Image{ID: 1, Name: "test-image"})

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := newTestService(repo, imageRepo)

	resp, err := service.CreateChallenge(1001, &dto.CreateChallengeReq{
		Title:            "Test Challenge",
		Description:      "Test",
		Category:         "web",
		Difficulty:       "easy",
		Points:           100,
		ImageID:          1,
		InstanceSharing:  model.InstanceSharingPerUser,
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

	_, err := service.CreateChallenge(1001, &dto.CreateChallengeReq{
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

	resp, err := service.CreateChallenge(1001, &dto.CreateChallengeReq{
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

	err = service.UpdateChallenge(challenge.ID, &dto.UpdateChallengeReq{
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
	service := NewChallengeService(nil, repo, imageRepo, repo, nil, SelfCheckConfig{}, zap.NewNop())

	err = service.UpdateChallenge(challenge.ID, &dto.UpdateChallengeReq{
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

	err := service.DeleteChallenge(challenge.ID)
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

	err := service.PublishChallenge(challenge.ID)
	if err != nil {
		t.Fatalf("PublishChallenge() error = %v", err)
	}

	published, findErr := repo.FindByID(challenge.ID)
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
	notifier := &stubChallengeNotificationSender{}
	service := NewChallengeService(db, repo, imageRepo, repo, probe, SelfCheckConfig{
		PublishCheckBatchSize: 1,
	}, zap.NewNop(), notifier)

	job, err := service.RequestPublishCheck(context.Background(), teacher.ID, challenge.ID)
	if err != nil {
		t.Fatalf("RequestPublishCheck() error = %v", err)
	}
	if job.Status != "queued" || !job.Active {
		t.Fatalf("unexpected requested job status: %s", job.Status)
	}

	service.dispatchPublishCheckJobs(context.Background())

	published, err := repo.FindByID(challenge.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if published.Status != model.ChallengeStatusPublished {
		t.Fatalf("expected published challenge status, got %s", published.Status)
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

	if len(notifier.calls) != 1 {
		t.Fatalf("expected 1 notification, got %+v", notifier.calls)
	}
	if !notifier.calls[0].passed || notifier.calls[0].challengeID != challenge.ID || notifier.calls[0].userID != teacher.ID {
		t.Fatalf("unexpected notification payload: %+v", notifier.calls[0])
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
	notifier := &stubChallengeNotificationSender{}
	service := NewChallengeService(db, repo, imageRepo, repo, &fakeChallengeRuntimeProbe{}, SelfCheckConfig{
		PublishCheckBatchSize: 1,
	}, zap.NewNop(), notifier)

	if _, err := service.RequestPublishCheck(context.Background(), teacher.ID, challenge.ID); err != nil {
		t.Fatalf("RequestPublishCheck() error = %v", err)
	}

	service.dispatchPublishCheckJobs(context.Background())

	stored, err := repo.FindByID(challenge.ID)
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

	if len(notifier.calls) != 1 {
		t.Fatalf("expected 1 notification, got %+v", notifier.calls)
	}
	if notifier.calls[0].passed {
		t.Fatalf("expected failure notification, got %+v", notifier.calls[0])
	}
	if notifier.calls[0].failureSummary == "" {
		t.Fatalf("expected failure summary in notification, got %+v", notifier.calls[0])
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
	notifier := &stubChallengeNotificationSender{}
	service := NewChallengeService(db, repo, imageRepo, repo, probe, SelfCheckConfig{
		PublishCheckBatchSize: 1,
	}, zap.NewNop(), notifier)

	if _, err := service.RequestPublishCheck(context.Background(), teacher.ID, challenge.ID); err != nil {
		t.Fatalf("RequestPublishCheck() error = %v", err)
	}

	service.dispatchPublishCheckJobs(context.Background())

	published, err := repo.FindByID(challenge.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if published.Status != model.ChallengeStatusPublished {
		t.Fatalf("expected attachment-only challenge to publish, got %s", published.Status)
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
	service := NewChallengeService(db, repo, imageRepo, repo, nil, SelfCheckConfig{}, zap.NewNop())

	latest, err := service.GetLatestPublishCheck(context.Background(), challenge.ID)
	if err == nil || err.Error() != errcode.ErrNotFound.Error() {
		t.Fatalf("expected not found for stale publish check job, got latest=%+v err=%v", latest, err)
	}
}
