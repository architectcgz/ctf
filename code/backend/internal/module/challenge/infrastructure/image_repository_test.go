package infrastructure

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/testsupport"
	"gorm.io/gorm"
)

func TestImageRepositoryListExplicitlyFiltersSoftDeletedImages(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("image_repository.go")
	if err != nil {
		t.Fatalf("read image repository: %v", err)
	}

	listStart := strings.Index(string(source), "func (r *ImageRepository) List(")
	if listStart < 0 {
		t.Fatal("ImageRepository.List not found")
	}
	updateStart := strings.Index(string(source)[listStart:], "func (r *ImageRepository) Update(")
	if updateStart < 0 {
		t.Fatal("ImageRepository.Update not found")
	}
	listSource := string(source)[listStart : listStart+updateStart]
	if !strings.Contains(listSource, "deleted_at IS NULL") {
		t.Fatal("ImageRepository.List should explicitly filter deleted_at IS NULL instead of relying on implicit GORM scope")
	}
}

func TestImageRepositoryManagesImageBuildJobs(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := NewImageRepository(db)
	ctx := context.Background()

	job := &model.ImageBuildJob{
		SourceType:     model.ImageSourceTypePlatformBuild,
		ChallengeMode:  "jeopardy",
		PackageSlug:    "web-demo",
		SourceDir:      "/tmp/web-demo",
		DockerfilePath: "/tmp/web-demo/docker/Dockerfile",
		ContextPath:    "/tmp/web-demo/docker",
		TargetRef:      "127.0.0.1:5000/jeopardy/web-demo:v1",
		Status:         model.ImageBuildJobStatusPending,
	}
	if err := repo.CreateImageBuildJob(ctx, job); err != nil {
		t.Fatalf("CreateImageBuildJob() error = %v", err)
	}
	if job.ID == 0 {
		t.Fatal("expected job id to be assigned")
	}

	found, err := repo.FindImageBuildJobByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("FindImageBuildJobByID() error = %v", err)
	}
	if found.PackageSlug != "web-demo" || found.Status != model.ImageBuildJobStatusPending {
		t.Fatalf("unexpected found job: %+v", found)
	}

	pending, err := repo.ListPendingImageBuildJobs(ctx, 1)
	if err != nil {
		t.Fatalf("ListPendingImageBuildJobs() error = %v", err)
	}
	if len(pending) != 1 || pending[0].ID != job.ID {
		t.Fatalf("unexpected pending jobs: %+v", pending)
	}

	startedAt := time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC)
	started, err := repo.TryStartImageBuildJob(ctx, job.ID, startedAt)
	if err != nil {
		t.Fatalf("TryStartImageBuildJob() error = %v", err)
	}
	if !started {
		t.Fatal("expected pending job to start")
	}
	startedAgain, err := repo.TryStartImageBuildJob(ctx, job.ID, startedAt)
	if err != nil {
		t.Fatalf("TryStartImageBuildJob(second) error = %v", err)
	}
	if startedAgain {
		t.Fatal("expected non-pending job not to start again")
	}

	found, err = repo.FindImageBuildJobByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("FindImageBuildJobByID(started) error = %v", err)
	}
	if found.Status != model.ImageBuildJobStatusBuilding || found.StartedAt == nil || !found.StartedAt.Equal(startedAt) {
		t.Fatalf("unexpected started job: %+v", found)
	}

	finishedAt := time.Date(2026, 5, 5, 10, 1, 0, 0, time.UTC)
	found.Status = model.ImageBuildJobStatusFailed
	found.ErrorSummary = "docker build failed"
	found.TargetDigest = "sha256:demo"
	found.FinishedAt = &finishedAt
	if err := repo.UpdateImageBuildJob(ctx, found); err != nil {
		t.Fatalf("UpdateImageBuildJob() error = %v", err)
	}

	updated, err := repo.FindImageBuildJobByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("FindImageBuildJobByID(updated) error = %v", err)
	}
	if updated.Status != model.ImageBuildJobStatusFailed ||
		updated.ErrorSummary != "docker build failed" ||
		updated.TargetDigest != "sha256:demo" ||
		updated.FinishedAt == nil ||
		!updated.FinishedAt.Equal(finishedAt) {
		t.Fatalf("unexpected updated job: %+v", updated)
	}
}

func TestImageRepositoryFindImageBuildJobByIDReturnsRecordNotFound(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := NewImageRepository(db)

	if _, err := repo.FindImageBuildJobByID(context.Background(), 404); err != gorm.ErrRecordNotFound {
		t.Fatalf("FindImageBuildJobByID() error = %v, want record not found", err)
	}
}
