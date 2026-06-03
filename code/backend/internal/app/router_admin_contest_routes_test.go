package app

import (
	"os"
	"strings"
	"testing"
)

func TestAdminContestRoutesAreExtractedIntoDedicatedRegistrarFile(t *testing.T) {
	t.Parallel()

	mainRoutes, err := os.ReadFile("router_routes.go")
	if err != nil {
		t.Fatalf("read router_routes.go: %v", err)
	}

	for _, marker := range []string{
		`adminOnly.POST("/contests",`,
		`adminOnly.GET("/contests/:id",`,
		`adminOnly.GET("/contests/:id/awd/services",`,
		`adminOnly.POST("/contests/:id/awd/rounds",`,
	} {
		if strings.Contains(string(mainRoutes), marker) {
			t.Fatalf("expected %q to move out of router_routes.go", marker)
		}
	}

	contestRoutes, err := os.ReadFile("router_admin_contest_routes.go")
	if err != nil {
		t.Fatalf("read router_admin_contest_routes.go: %v", err)
	}

	for _, marker := range []string{
		"func registerAdminContestRoutes(",
		`contests := adminOnly.Group("/contests")`,
		`contestByID := contests.Group("/:id")`,
		`registerAdminContestCoreRoutes(contests, contestByID, deps, audit, awdReadinessAudit)`,
		`registerAdminContestChallengeRoutes(contestByID, deps, audit)`,
		`registerAdminContestParticipationRoutes(contestByID, deps, audit)`,
		`registerAdminContestAWDRoutes(contestByID, deps, audit, awdReadinessAudit)`,
	} {
		if !strings.Contains(string(contestRoutes), marker) {
			t.Fatalf("router_admin_contest_routes.go should contain %q", marker)
		}
	}
}

func TestAdminContestCoreRoutesAreExtractedIntoDedicatedRegistrarFile(t *testing.T) {
	t.Parallel()

	contestRoutes, err := os.ReadFile("router_admin_contest_routes.go")
	if err != nil {
		t.Fatalf("read router_admin_contest_routes.go: %v", err)
	}

	for _, marker := range []string{
		`contestByID.POST("/freeze",`,
		`contestByID.POST("/export",`,
	} {
		if strings.Contains(string(contestRoutes), marker) {
			t.Fatalf("expected %q to move out of router_admin_contest_routes.go", marker)
		}
	}

	coreRoutes, err := os.ReadFile("router_admin_contest_core_routes.go")
	if err != nil {
		t.Fatalf("read router_admin_contest_core_routes.go: %v", err)
	}

	for _, marker := range []string{
		"func registerAdminContestCoreRoutes(",
		`contestByID.POST("/freeze",`,
		`contestByID.POST("/export",`,
	} {
		if !strings.Contains(string(coreRoutes), marker) {
			t.Fatalf("router_admin_contest_core_routes.go should contain %q", marker)
		}
	}
}

func TestAdminContestChallengeRoutesAreExtractedIntoDedicatedRegistrarFile(t *testing.T) {
	t.Parallel()

	contestRoutes, err := os.ReadFile("router_admin_contest_routes.go")
	if err != nil {
		t.Fatalf("read router_admin_contest_routes.go: %v", err)
	}

	for _, marker := range []string{
		`contestByID.GET("/challenges",`,
		`contestByID.DELETE("/challenges/:cid",`,
	} {
		if strings.Contains(string(contestRoutes), marker) {
			t.Fatalf("expected %q to move out of router_admin_contest_routes.go", marker)
		}
	}

	challengeRoutes, err := os.ReadFile("router_admin_contest_challenge_routes.go")
	if err != nil {
		t.Fatalf("read router_admin_contest_challenge_routes.go: %v", err)
	}

	for _, marker := range []string{
		"func registerAdminContestChallengeRoutes(",
		`contestByID.GET("/challenges",`,
		`contestByID.DELETE("/challenges/:cid",`,
	} {
		if !strings.Contains(string(challengeRoutes), marker) {
			t.Fatalf("router_admin_contest_challenge_routes.go should contain %q", marker)
		}
	}
}

func TestAdminContestParticipationRoutesAreExtractedIntoDedicatedRegistrarFile(t *testing.T) {
	t.Parallel()

	contestRoutes, err := os.ReadFile("router_admin_contest_routes.go")
	if err != nil {
		t.Fatalf("read router_admin_contest_routes.go: %v", err)
	}

	for _, marker := range []string{
		`contestByID.GET("/registrations",`,
		`contestByID.POST("/announcements",`,
		`contestByID.GET("/scoreboard/live",`,
	} {
		if strings.Contains(string(contestRoutes), marker) {
			t.Fatalf("expected %q to move out of router_admin_contest_routes.go", marker)
		}
	}

	participationRoutes, err := os.ReadFile("router_admin_contest_participation_routes.go")
	if err != nil {
		t.Fatalf("read router_admin_contest_participation_routes.go: %v", err)
	}

	for _, marker := range []string{
		"func registerAdminContestParticipationRoutes(",
		`contestByID.GET("/registrations",`,
		`contestByID.POST("/announcements",`,
		`contestByID.GET("/scoreboard/live",`,
	} {
		if !strings.Contains(string(participationRoutes), marker) {
			t.Fatalf("router_admin_contest_participation_routes.go should contain %q", marker)
		}
	}
}

func TestAdminContestAWDRoutesAreExtractedIntoDedicatedRegistrarFile(t *testing.T) {
	t.Parallel()

	contestRoutes, err := os.ReadFile("router_admin_contest_routes.go")
	if err != nil {
		t.Fatalf("read router_admin_contest_routes.go: %v", err)
	}

	for _, marker := range []string{
		`awd := contestByID.Group("/awd")`,
		`awdServices := awd.Group("/services")`,
	} {
		if strings.Contains(string(contestRoutes), marker) {
			t.Fatalf("expected %q to move out of router_admin_contest_routes.go", marker)
		}
	}

	awdRoutes, err := os.ReadFile("router_admin_contest_awd_routes.go")
	if err != nil {
		t.Fatalf("read router_admin_contest_awd_routes.go: %v", err)
	}

	for _, marker := range []string{
		"func registerAdminContestAWDRoutes(",
		`awd := contestByID.Group("/awd")`,
		`awdServices := awd.Group("/services")`,
	} {
		if !strings.Contains(string(awdRoutes), marker) {
			t.Fatalf("router_admin_contest_awd_routes.go should contain %q", marker)
		}
	}
}
