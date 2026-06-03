package app

import (
	"os"
	"strings"
	"testing"
)

func TestAdminOpsRoutesAreExtractedIntoDedicatedRegistrarFile(t *testing.T) {
	t.Parallel()

	mainRoutes, err := os.ReadFile("router_routes.go")
	if err != nil {
		t.Fatalf("read router_routes.go: %v", err)
	}

	for _, marker := range []string{
		`adminOnly.GET("/audit-logs",`,
		`adminOnly.GET("/dashboard",`,
		`adminOnly.GET("/cheat-detection",`,
		`adminOnly.POST("/notifications",`,
	} {
		if strings.Contains(string(mainRoutes), marker) {
			t.Fatalf("expected %q to move out of router_routes.go", marker)
		}
	}

	opsRoutes, err := os.ReadFile("router_admin_ops_routes.go")
	if err != nil {
		t.Fatalf("read router_admin_ops_routes.go: %v", err)
	}

	for _, marker := range []string{
		"func registerAdminOpsRoutes(",
		`adminOnly.GET("/audit-logs",`,
		`adminOnly.POST("/notifications",`,
	} {
		if !strings.Contains(string(opsRoutes), marker) {
			t.Fatalf("router_admin_ops_routes.go should contain %q", marker)
		}
	}
}

func TestAdminIdentityRoutesAreExtractedIntoDedicatedRegistrarFile(t *testing.T) {
	t.Parallel()

	mainRoutes, err := os.ReadFile("router_routes.go")
	if err != nil {
		t.Fatalf("read router_routes.go: %v", err)
	}

	for _, marker := range []string{
		`adminOnly.GET("/users",`,
		`adminOnly.POST("/users",`,
		`adminOnly.GET("/users/:id/sessions",`,
		`adminOnly.DELETE("/users/:id/sessions",`,
	} {
		if strings.Contains(string(mainRoutes), marker) {
			t.Fatalf("expected %q to move out of router_routes.go", marker)
		}
	}

	identityRoutes, err := os.ReadFile("router_admin_identity_routes.go")
	if err != nil {
		t.Fatalf("read router_admin_identity_routes.go: %v", err)
	}

	for _, marker := range []string{
		"func registerAdminIdentityRoutes(",
		`users := adminOnly.Group("/users")`,
		`userByID := users.Group("/:id")`,
		`userSessions := userByID.Group("/sessions")`,
	} {
		if !strings.Contains(string(identityRoutes), marker) {
			t.Fatalf("router_admin_identity_routes.go should contain %q", marker)
		}
	}
}
