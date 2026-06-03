package app

import (
	"os"
	"strings"
	"testing"
)

func TestUserContestRoutesAreExtractedIntoDedicatedRegistrarFile(t *testing.T) {
	t.Parallel()

	selfRoutes, err := os.ReadFile("router_user_self_routes.go")
	if err != nil {
		t.Fatalf("read router_user_self_routes.go: %v", err)
	}

	for _, marker := range []string{
		`contestGroup := apiV1.Group("/contests")`,
		`protected.POST("/contests/:id/teams",`,
		`protected.POST("/contests/:id/awd/services/:sid/instances",`,
	} {
		if strings.Contains(string(selfRoutes), marker) {
			t.Fatalf("expected %q to move out of router_user_self_routes.go", marker)
		}
	}

	contestRoutes, err := os.ReadFile("router_user_contest_routes.go")
	if err != nil {
		t.Fatalf("read router_user_contest_routes.go: %v", err)
	}

	for _, marker := range []string{
		"func registerUserContestRoutes(",
		`contestGroup := apiV1.Group("/contests")`,
		`protected.POST("/contests/:id/teams",`,
		`protected.POST("/contests/:id/awd/services/:sid/instances",`,
	} {
		if !strings.Contains(string(contestRoutes), marker) {
			t.Fatalf("router_user_contest_routes.go should contain %q", marker)
		}
	}
}

func TestUserPracticeRoutesAreExtractedIntoDedicatedRegistrarFile(t *testing.T) {
	t.Parallel()

	selfRoutes, err := os.ReadFile("router_user_self_routes.go")
	if err != nil {
		t.Fatalf("read router_user_self_routes.go: %v", err)
	}

	for _, marker := range []string{
		`protected.GET("/challenges",`,
		`protected.GET("/instances",`,
		`apiV1.GET("/instances/:id/proxy",`,
	} {
		if strings.Contains(string(selfRoutes), marker) {
			t.Fatalf("expected %q to move out of router_user_self_routes.go", marker)
		}
	}

	practiceRoutes, err := os.ReadFile("router_user_practice_routes.go")
	if err != nil {
		t.Fatalf("read router_user_practice_routes.go: %v", err)
	}

	for _, marker := range []string{
		"func registerUserPracticeRoutes(",
		`protected.GET("/challenges",`,
		`protected.GET("/instances",`,
		`apiV1.GET("/instances/:id/proxy",`,
	} {
		if !strings.Contains(string(practiceRoutes), marker) {
			t.Fatalf("router_user_practice_routes.go should contain %q", marker)
		}
	}
}

func TestUserSelfServiceRoutesAreExtractedIntoDedicatedRegistrarFile(t *testing.T) {
	t.Parallel()

	selfRoutes, err := os.ReadFile("router_user_self_routes.go")
	if err != nil {
		t.Fatalf("read router_user_self_routes.go: %v", err)
	}

	for _, marker := range []string{
		`usersGroup := protected.Group("/users")`,
		`protected.POST("/reports/personal",`,
		`protected.POST("/reports/class",`,
	} {
		if strings.Contains(string(selfRoutes), marker) {
			t.Fatalf("expected %q to move out of router_user_self_routes.go", marker)
		}
	}

	selfServiceRoutes, err := os.ReadFile("router_user_self_service_routes.go")
	if err != nil {
		t.Fatalf("read router_user_self_service_routes.go: %v", err)
	}

	for _, marker := range []string{
		"func registerUserSelfServiceRoutes(",
		`usersGroup := protected.Group("/users")`,
		`protected.POST("/reports/personal",`,
		`protected.POST("/reports/class",`,
	} {
		if !strings.Contains(string(selfServiceRoutes), marker) {
			t.Fatalf("router_user_self_service_routes.go should contain %q", marker)
		}
	}
}
