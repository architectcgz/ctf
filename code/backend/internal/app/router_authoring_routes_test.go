package app

import (
	"os"
	"strings"
	"testing"
)

func TestAuthoringChallengeRoutesAreExtractedIntoDedicatedRegistrarFile(t *testing.T) {
	t.Parallel()

	mainRoutes, err := os.ReadFile("router_routes.go")
	if err != nil {
		t.Fatalf("read router_routes.go: %v", err)
	}

	for _, marker := range []string{
		`adminAuthoring.POST("/challenge-imports",`,
		`adminAuthoring.POST("/challenges",`,
		`adminAuthoring.PUT("/challenges/:id/flag",`,
	} {
		if strings.Contains(string(mainRoutes), marker) {
			t.Fatalf("expected %q to move out of router_routes.go", marker)
		}
	}

	challengeRoutes, err := os.ReadFile("router_authoring_challenge_routes.go")
	if err != nil {
		t.Fatalf("read router_authoring_challenge_routes.go: %v", err)
	}

	for _, marker := range []string{
		"func registerAuthoringChallengeRoutes(",
		`adminAuthoring.POST("/challenge-imports",`,
		`adminAuthoring.POST("/challenges",`,
		`adminAuthoring.PUT("/challenges/:id/flag",`,
	} {
		if !strings.Contains(string(challengeRoutes), marker) {
			t.Fatalf("router_authoring_challenge_routes.go should contain %q", marker)
		}
	}
}

func TestAuthoringAssetRoutesAreExtractedIntoDedicatedRegistrarFile(t *testing.T) {
	t.Parallel()

	mainRoutes, err := os.ReadFile("router_routes.go")
	if err != nil {
		t.Fatalf("read router_routes.go: %v", err)
	}

	for _, marker := range []string{
		`adminAuthoring.POST("/images",`,
		`adminAuthoring.GET("/environment-templates",`,
	} {
		if strings.Contains(string(mainRoutes), marker) {
			t.Fatalf("expected %q to move out of router_routes.go", marker)
		}
	}

	assetRoutes, err := os.ReadFile("router_authoring_asset_routes.go")
	if err != nil {
		t.Fatalf("read router_authoring_asset_routes.go: %v", err)
	}

	for _, marker := range []string{
		"func registerAuthoringAssetRoutes(",
		`adminAuthoring.POST("/images",`,
		`adminAuthoring.GET("/environment-templates",`,
	} {
		if !strings.Contains(string(assetRoutes), marker) {
			t.Fatalf("router_authoring_asset_routes.go should contain %q", marker)
		}
	}
}

func TestAuthoringAWDRoutesAreExtractedIntoDedicatedRegistrarFile(t *testing.T) {
	t.Parallel()

	mainRoutes, err := os.ReadFile("router_routes.go")
	if err != nil {
		t.Fatalf("read router_routes.go: %v", err)
	}

	for _, marker := range []string{
		`adminAuthoring.GET("/awd-challenges",`,
		`adminAuthoring.POST("/awd-challenge-imports",`,
	} {
		if strings.Contains(string(mainRoutes), marker) {
			t.Fatalf("expected %q to move out of router_routes.go", marker)
		}
	}

	awdRoutes, err := os.ReadFile("router_authoring_awd_routes.go")
	if err != nil {
		t.Fatalf("read router_authoring_awd_routes.go: %v", err)
	}

	for _, marker := range []string{
		"func registerAuthoringAWDRoutes(",
		`adminAuthoring.GET("/awd-challenges",`,
		`adminAuthoring.POST("/awd-challenge-imports",`,
	} {
		if !strings.Contains(string(awdRoutes), marker) {
			t.Fatalf("router_authoring_awd_routes.go should contain %q", marker)
		}
	}
}
