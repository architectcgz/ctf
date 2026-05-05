package commands

import (
	"context"
	"testing"

	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/testsupport"
)

func TestImageBuildServiceCreatePlatformBuildJobCreatesPendingImageAndJob(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewImageRepository(db)
	service := NewImageBuildService(repo, ImageBuildConfig{Registry: "127.0.0.1:5000"})

	result, err := service.CreatePlatformBuildJob(context.Background(), CreatePlatformBuildJobRequest{
		ChallengeMode:  "jeopardy",
		PackageSlug:    "web-demo",
		SuggestedTag:   "v1",
		SourceDir:      "/tmp/web-demo",
		DockerfilePath: "/tmp/web-demo/docker/Dockerfile",
		ContextPath:    "/tmp/web-demo/docker",
		CreatedBy:      1001,
	})
	if err != nil {
		t.Fatalf("CreatePlatformBuildJob() error = %v", err)
	}
	if result.ImageID == 0 || result.JobID == 0 {
		t.Fatalf("expected image and job ids, got %+v", result)
	}
	if result.TargetRef != "127.0.0.1:5000/jeopardy/web-demo:v1" {
		t.Fatalf("TargetRef = %q", result.TargetRef)
	}

	image, err := repo.FindByID(context.Background(), result.ImageID)
	if err != nil {
		t.Fatalf("FindByID(image) error = %v", err)
	}
	if image.Name != "127.0.0.1:5000/jeopardy/web-demo" ||
		image.Tag != "v1" ||
		image.Status != model.ImageStatusPending ||
		image.SourceType != model.ImageSourceTypePlatformBuild ||
		image.BuildJobID == nil ||
		*image.BuildJobID != result.JobID {
		t.Fatalf("unexpected image: %+v", image)
	}

	job, err := repo.FindImageBuildJobByID(context.Background(), result.JobID)
	if err != nil {
		t.Fatalf("FindImageBuildJobByID() error = %v", err)
	}
	if job.Status != model.ImageBuildJobStatusPending ||
		job.SourceType != model.ImageSourceTypePlatformBuild ||
		job.TargetRef != result.TargetRef ||
		job.CreatedBy == nil ||
		*job.CreatedBy != 1001 {
		t.Fatalf("unexpected job: %+v", job)
	}
}

func TestImageBuildServiceCreatePlatformBuildJobReusesExistingImage(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewImageRepository(db)
	existing := &model.Image{
		Name:       "127.0.0.1:5000/jeopardy/web-demo",
		Tag:        "v1",
		Status:     model.ImageStatusFailed,
		SourceType: model.ImageSourceTypePlatformBuild,
		LastError:  "old failure",
	}
	if err := repo.Create(context.Background(), existing); err != nil {
		t.Fatalf("create existing image: %v", err)
	}

	service := NewImageBuildService(repo, ImageBuildConfig{Registry: "127.0.0.1:5000"})
	result, err := service.CreatePlatformBuildJob(context.Background(), CreatePlatformBuildJobRequest{
		ChallengeMode:  "jeopardy",
		PackageSlug:    "web-demo",
		SuggestedTag:   "v1",
		SourceDir:      "/tmp/web-demo",
		DockerfilePath: "/tmp/web-demo/docker/Dockerfile",
		ContextPath:    "/tmp/web-demo/docker",
	})
	if err != nil {
		t.Fatalf("CreatePlatformBuildJob() error = %v", err)
	}
	if result.ImageID != existing.ID {
		t.Fatalf("expected existing image id %d, got %d", existing.ID, result.ImageID)
	}

	image, err := repo.FindByID(context.Background(), existing.ID)
	if err != nil {
		t.Fatalf("FindByID(image) error = %v", err)
	}
	if image.Status != model.ImageStatusPending || image.LastError != "" || image.BuildJobID == nil {
		t.Fatalf("expected existing image to be reset for new build, got %+v", image)
	}
}
