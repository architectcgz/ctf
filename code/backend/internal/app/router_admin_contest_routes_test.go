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
		"func registerAdminContestAWDRoutes(",
		`contests := adminOnly.Group("/contests")`,
		`contestByID := contests.Group("/:id")`,
		`awd := contestByID.Group("/awd")`,
		`awdServices := awd.Group("/services")`,
	} {
		if !strings.Contains(string(contestRoutes), marker) {
			t.Fatalf("router_admin_contest_routes.go should contain %q", marker)
		}
	}
}
