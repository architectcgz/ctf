package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/internal/module/challenge/testsupport"
	"gorm.io/gorm"
)

type fakeDockerImageBuilder struct {
	buildErr   error
	pushErr    error
	pullErr    error
	inspectErr error
	calls      []string
}

func (b *fakeDockerImageBuilder) Build(ctx context.Context, contextPath, dockerfilePath, localRef string) error {
	b.calls = append(b.calls, "build")
	return b.buildErr
}

func (b *fakeDockerImageBuilder) Tag(ctx context.Context, sourceRef, targetRef string) error {
	b.calls = append(b.calls, "tag")
	return nil
}

func (b *fakeDockerImageBuilder) Push(ctx context.Context, targetRef string) error {
	b.calls = append(b.calls, "push")
	return b.pushErr
}

func (b *fakeDockerImageBuilder) Pull(ctx context.Context, targetRef string) error {
	b.calls = append(b.calls, "pull")
	return b.pullErr
}

func (b *fakeDockerImageBuilder) Inspect(ctx context.Context, targetRef string) (challengeports.ImageInspectResult, error) {
	b.calls = append(b.calls, "inspect")
	return challengeports.ImageInspectResult{Size: 12345}, b.inspectErr
}

type fakeRegistryVerifier struct {
	digest string
	err    error
}

func (v fakeRegistryVerifier) CheckManifest(ctx context.Context, imageRef string) (string, error) {
	if v.err != nil {
		return "", v.err
	}
	return v.digest, nil
}

type imageBuildRepoStub struct{}

func (imageBuildRepoStub) Create(context.Context, *model.Image) error {
	return nil
}

func (imageBuildRepoStub) FindByID(context.Context, int64) (*model.Image, error) {
	return nil, nil
}

func (imageBuildRepoStub) FindByNameTag(context.Context, string, string) (*model.Image, error) {
	return nil, nil
}

func (imageBuildRepoStub) Update(context.Context, *model.Image) error {
	return nil
}

func (imageBuildRepoStub) Delete(context.Context, int64) error {
	return nil
}

func (imageBuildRepoStub) CreateImageBuildJob(context.Context, *model.ImageBuildJob) error {
	return nil
}

func (imageBuildRepoStub) FindImageBuildJobByID(context.Context, int64) (*model.ImageBuildJob, error) {
	return nil, nil
}

func (imageBuildRepoStub) ListPendingImageBuildJobs(context.Context, int) ([]*model.ImageBuildJob, error) {
	return nil, nil
}

func (imageBuildRepoStub) TryStartImageBuildJob(context.Context, int64, time.Time) (bool, error) {
	return false, nil
}

func (imageBuildRepoStub) UpdateImageBuildJob(context.Context, *model.ImageBuildJob) error {
	return nil
}

type fakeImageBuildTxStore struct {
	findByNameTagResult *model.Image
	findByNameTagErr    error
	createdImage        *model.Image
	createdJob          *model.ImageBuildJob
	updatedImage        *model.Image
}

func (s *fakeImageBuildTxStore) FindByNameTag(ctx context.Context, name, tag string) (*model.Image, error) {
	return s.findByNameTagResult, s.findByNameTagErr
}

func (s *fakeImageBuildTxStore) CreateImage(ctx context.Context, image *model.Image) error {
	s.createdImage = image
	image.ID = 41
	return nil
}

func (s *fakeImageBuildTxStore) CreateImageBuildJob(ctx context.Context, job *model.ImageBuildJob) error {
	s.createdJob = job
	job.ID = 99
	return nil
}

func (s *fakeImageBuildTxStore) UpdateImage(ctx context.Context, image *model.Image, updates map[string]any) error {
	s.updatedImage = image
	return nil
}

func TestImageBuildServiceCreatePlatformBuildJobInTxUsesTxStore(t *testing.T) {
	service := NewImageBuildService(imageBuildRepoStub{}, ImageBuildConfig{Registry: "127.0.0.1:5000"})
	store := &fakeImageBuildTxStore{findByNameTagErr: challengeports.ErrChallengeImageNotFound}

	result, err := service.CreatePlatformBuildJobInTx(context.Background(), store, CreatePlatformBuildJobRequest{
		ChallengeMode:  "jeopardy",
		PackageSlug:    "web-demo",
		SuggestedTag:   "v1",
		SourceDir:      "/tmp/web-demo",
		DockerfilePath: "/tmp/web-demo/docker/Dockerfile",
		ContextPath:    "/tmp/web-demo/docker",
		CreatedBy:      1001,
	})
	if err != nil {
		t.Fatalf("CreatePlatformBuildJobInTx() error = %v", err)
	}
	if result.ImageID != 41 || result.JobID != 99 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if store.createdImage == nil || store.createdJob == nil || store.updatedImage == nil {
		t.Fatalf("tx store did not receive expected writes: %+v", store)
	}
}

func TestImageBuildServiceCreatePlatformBuildJobCreatesPendingImageAndJob(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewImageRepository(db)
	service := NewImageBuildService(challengeinfra.NewImageBuildRepository(repo), ImageBuildConfig{Registry: "127.0.0.1:5000"})

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

	service := NewImageBuildService(challengeinfra.NewImageBuildRepository(repo), ImageBuildConfig{Registry: "127.0.0.1:5000"})
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

func TestImageBuildServiceProcessImageBuildJobMarksImageAvailable(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewImageRepository(db)
	builder := &fakeDockerImageBuilder{}
	service := NewImageBuildService(
		challengeinfra.NewImageBuildRepository(repo),
		ImageBuildConfig{Registry: "127.0.0.1:5000", BuildTimeout: 0},
		WithImageBuildDockerBuilder(builder),
		WithImageBuildRegistryVerifier(fakeRegistryVerifier{digest: "sha256:demo"}),
	)

	created, err := service.CreatePlatformBuildJob(context.Background(), CreatePlatformBuildJobRequest{
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

	if err := service.ProcessImageBuildJob(context.Background(), created.JobID); err != nil {
		t.Fatalf("ProcessImageBuildJob() error = %v", err)
	}

	job, err := repo.FindImageBuildJobByID(context.Background(), created.JobID)
	if err != nil {
		t.Fatalf("FindImageBuildJobByID() error = %v", err)
	}
	if job.Status != model.ImageBuildJobStatusAvailable ||
		job.TargetDigest != "sha256:demo" ||
		job.FinishedAt == nil {
		t.Fatalf("unexpected available job: %+v", job)
	}

	image, err := repo.FindByID(context.Background(), created.ImageID)
	if err != nil {
		t.Fatalf("FindByID(image) error = %v", err)
	}
	if image.Status != model.ImageStatusAvailable ||
		image.Digest != "sha256:demo" ||
		image.Size != 12345 ||
		image.VerifiedAt == nil {
		t.Fatalf("unexpected available image: %+v", image)
	}

	wantCalls := []string{"build", "push", "pull", "inspect"}
	if len(builder.calls) != len(wantCalls) {
		t.Fatalf("builder calls = %+v, want %+v", builder.calls, wantCalls)
	}
	for i := range wantCalls {
		if builder.calls[i] != wantCalls[i] {
			t.Fatalf("builder calls = %+v, want %+v", builder.calls, wantCalls)
		}
	}
}

func TestImageBuildServiceProcessImageBuildJobMarksFailures(t *testing.T) {
	cases := []struct {
		name      string
		builder   *fakeDockerImageBuilder
		verifier  fakeRegistryVerifier
		errorText string
	}{
		{
			name:      "build",
			builder:   &fakeDockerImageBuilder{buildErr: errors.New("build failed")},
			verifier:  fakeRegistryVerifier{digest: "sha256:demo"},
			errorText: "build failed",
		},
		{
			name:      "push",
			builder:   &fakeDockerImageBuilder{pushErr: errors.New("push failed")},
			verifier:  fakeRegistryVerifier{digest: "sha256:demo"},
			errorText: "push failed",
		},
		{
			name:      "manifest",
			builder:   &fakeDockerImageBuilder{},
			verifier:  fakeRegistryVerifier{err: errors.New("manifest failed")},
			errorText: "manifest failed",
		},
		{
			name:      "inspect",
			builder:   &fakeDockerImageBuilder{inspectErr: errors.New("inspect failed")},
			verifier:  fakeRegistryVerifier{digest: "sha256:demo"},
			errorText: "inspect failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := testsupport.SetupTestDB(t)
			repo := challengeinfra.NewImageRepository(db)
			service := NewImageBuildService(
				challengeinfra.NewImageBuildRepository(repo),
				ImageBuildConfig{Registry: "127.0.0.1:5000"},
				WithImageBuildDockerBuilder(tc.builder),
				WithImageBuildRegistryVerifier(tc.verifier),
			)
			created, err := service.CreatePlatformBuildJob(context.Background(), CreatePlatformBuildJobRequest{
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

			err = service.ProcessImageBuildJob(context.Background(), created.JobID)
			if err == nil || err.Error() != tc.errorText {
				t.Fatalf("ProcessImageBuildJob() error = %v, want %q", err, tc.errorText)
			}

			job, err := repo.FindImageBuildJobByID(context.Background(), created.JobID)
			if err != nil {
				t.Fatalf("FindImageBuildJobByID() error = %v", err)
			}
			if job.Status != model.ImageBuildJobStatusFailed || job.ErrorSummary != tc.errorText {
				t.Fatalf("unexpected failed job: %+v", job)
			}

			image, err := repo.FindByID(context.Background(), created.ImageID)
			if err != nil {
				t.Fatalf("FindByID(image) error = %v", err)
			}
			if image.Status != model.ImageStatusFailed || image.LastError != tc.errorText {
				t.Fatalf("unexpected failed image: %+v", image)
			}
		})
	}
}

func TestImageBuildServiceVerifyExternalImageRefInTxMarksImageAvailable(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewImageRepository(db)
	builder := &fakeDockerImageBuilder{}
	service := NewImageBuildService(
		challengeinfra.NewImageBuildRepository(repo),
		ImageBuildConfig{Registry: "127.0.0.1:5000"},
		WithImageBuildDockerBuilder(builder),
		WithImageBuildRegistryVerifier(fakeRegistryVerifier{digest: "sha256:external"}),
	)

	var result *VerifyExternalImageRefResult
	if err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		result, err = service.VerifyExternalImageRefInTx(
			context.Background(),
			newImageBuildTxStore(tx),
			"web-external",
			"registry.example.edu/team/web-demo:v1",
		)
		return err
	}); err != nil {
		t.Fatalf("VerifyExternalImageRefInTx() error = %v", err)
	}
	if result.ImageID == 0 || result.Digest != "sha256:external" || result.Size != 12345 {
		t.Fatalf("unexpected verify result: %+v", result)
	}

	image, err := repo.FindByID(context.Background(), result.ImageID)
	if err != nil {
		t.Fatalf("FindByID(image) error = %v", err)
	}
	if image.Name != "registry.example.edu/team/web-demo" ||
		image.Tag != "v1" ||
		image.Status != model.ImageStatusAvailable ||
		image.SourceType != model.ImageSourceTypeExternalRef ||
		image.Digest != "sha256:external" ||
		image.VerifiedAt == nil ||
		image.BuildJobID != nil {
		t.Fatalf("unexpected external image: %+v", image)
	}

	wantCalls := []string{"pull", "inspect"}
	if len(builder.calls) != len(wantCalls) {
		t.Fatalf("builder calls = %+v, want %+v", builder.calls, wantCalls)
	}
	for i := range wantCalls {
		if builder.calls[i] != wantCalls[i] {
			t.Fatalf("builder calls = %+v, want %+v", builder.calls, wantCalls)
		}
	}
}

func TestImageBuildServiceVerifyExternalImageRefInTxReturnsErrorOnManifestFailure(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewImageRepository(db)
	service := NewImageBuildService(
		challengeinfra.NewImageBuildRepository(repo),
		ImageBuildConfig{Registry: "127.0.0.1:5000"},
		WithImageBuildDockerBuilder(&fakeDockerImageBuilder{}),
		WithImageBuildRegistryVerifier(fakeRegistryVerifier{err: errors.New("manifest failed")}),
	)

	err := db.Transaction(func(tx *gorm.DB) error {
		_, err := service.VerifyExternalImageRefInTx(
			context.Background(),
			newImageBuildTxStore(tx),
			"web-external",
			"registry.example.edu/team/web-demo:v1",
		)
		return err
	})
	if err == nil || err.Error() != "manifest failed" {
		t.Fatalf("VerifyExternalImageRefInTx() error = %v, want manifest failed", err)
	}

	var count int64
	if err := db.Model(&model.Image{}).Where("name = ?", "registry.example.edu/team/web-demo").Count(&count).Error; err != nil {
		t.Fatalf("count images: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected failed external verify transaction to roll back image, got %d", count)
	}
}
