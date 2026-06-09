package infrastructure

import (
	"context"
	"testing"
	"time"

	challengeentity "ctf-platform/internal/module/challenge/entity"
	"ctf-platform/internal/module/challenge/testsupport"
)

func TestArtifactReferenceRepositoryListsActiveArtifactReferences(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	ctx := context.Background()
	now := time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC)

	active := &challengeentity.Challenge{
		Title:         "active",
		AttachmentURL: "/api/v1/challenges/attachments/imports/active/asset.zip",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	deleted := &challengeentity.Challenge{
		Title:         "deleted",
		AttachmentURL: "/api/v1/challenges/attachments/imports/deleted/asset.zip",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := db.Create(active).Error; err != nil {
		t.Fatalf("create active challenge: %v", err)
	}
	if err := db.Create(deleted).Error; err != nil {
		t.Fatalf("create deleted challenge: %v", err)
	}
	if err := db.Delete(deleted).Error; err != nil {
		t.Fatalf("soft delete challenge: %v", err)
	}
	if err := db.Create(&challengeentity.ImageBuildJob{
		SourceType:    challengeentity.ImageSourceTypePlatformBuild,
		ChallengeMode: "jeopardy",
		PackageSlug:   "active-build",
		SourceDir:     "/tmp/active-build",
		TargetRef:     "127.0.0.1:5000/jeopardy/active:v1",
		Status:        challengeentity.ImageBuildJobStatusBuilding,
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatalf("create active build job: %v", err)
	}
	if err := db.Create(&challengeentity.ImageBuildJob{
		SourceType:    challengeentity.ImageSourceTypePlatformBuild,
		ChallengeMode: "jeopardy",
		PackageSlug:   "finished-build",
		SourceDir:     "/tmp/finished-build",
		TargetRef:     "127.0.0.1:5000/jeopardy/finished:v1",
		Status:        challengeentity.ImageBuildJobStatusAvailable,
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatalf("create finished build job: %v", err)
	}
	if err := db.Create(&challengeentity.AWDChallenge{
		Name:          "awd",
		Slug:          "awd",
		CheckerConfig: `{"artifact":{"entry":"check.py","storage_path":"/tmp/checker/awd/digest/check.py"}}`,
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatalf("create awd challenge: %v", err)
	}

	repo := NewArtifactReferenceRepository(db)
	refs, err := repo.ListArtifactReferences(ctx)
	if err != nil {
		t.Fatalf("ListArtifactReferences() error = %v", err)
	}
	if len(refs.AttachmentURLs) != 1 || refs.AttachmentURLs[0] != active.AttachmentURL {
		t.Fatalf("unexpected attachment refs: %+v", refs.AttachmentURLs)
	}
	if len(refs.ImageBuildSourceDirs) != 1 || refs.ImageBuildSourceDirs[0] != "/tmp/active-build" {
		t.Fatalf("unexpected build source refs: %+v", refs.ImageBuildSourceDirs)
	}
	if len(refs.AWDCheckerConfigs) != 1 || refs.AWDCheckerConfigs[0] == "" {
		t.Fatalf("unexpected checker refs: %+v", refs.AWDCheckerConfigs)
	}
}
