package queries_test

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/model"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	assessmentinfra "ctf-platform/internal/module/assessment/infrastructure"
)

func TestProfileServiceGetSkillProfileHonorsCancellation(t *testing.T) {
	db := setupRecommendationTestDB(t)
	service := assessmentqry.NewProfileService(assessmentinfra.NewRepository(db))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.GetSkillProfile(ctx, 1)
	if err == nil || err != context.Canceled {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestProfileServiceGetSkillProfileReturnsEmptyDimensionsWhenProfileMissing(t *testing.T) {
	db := setupRecommendationTestDB(t)
	service := assessmentqry.NewProfileService(assessmentinfra.NewRepository(db))

	profile, err := service.GetSkillProfile(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetSkillProfile() error = %v", err)
	}
	if profile.UserID != 42 {
		t.Fatalf("expected user id 42, got %+v", profile)
	}
	if profile.UpdatedAt != "" {
		t.Fatalf("expected empty updated_at, got %+v", profile)
	}
	if len(profile.Dimensions) != len(model.AllDimensions) {
		t.Fatalf("expected all dimensions, got %+v", profile.Dimensions)
	}
	for _, item := range profile.Dimensions {
		if item == nil {
			t.Fatalf("expected concrete dimension item, got %+v", profile.Dimensions)
		}
		if item.Score != 0 {
			t.Fatalf("expected zero score for empty profile, got %+v", profile.Dimensions)
		}
	}
}

func TestProfileServiceGetSkillProfileBuildsRFC3339Contract(t *testing.T) {
	db := setupRecommendationTestDB(t)
	service := assessmentqry.NewProfileService(assessmentinfra.NewRepository(db))
	now := time.Date(2026, 5, 17, 8, 30, 0, 0, time.UTC)

	if err := db.Create(&assessmententity.SkillProfile{
		UserID:    7,
		Dimension: model.DimensionWeb,
		Score:     0.75,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed profile: %v", err)
	}

	profile, err := service.GetSkillProfile(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetSkillProfile() error = %v", err)
	}
	if profile.UpdatedAt != now.Format(time.RFC3339) {
		t.Fatalf("expected RFC3339 updated_at, got %+v", profile)
	}
	if len(profile.Dimensions) != len(model.AllDimensions) {
		t.Fatalf("expected all dimensions, got %+v", profile.Dimensions)
	}
}
