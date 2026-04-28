package queries_test

import (
	"context"
	"testing"

	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
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
