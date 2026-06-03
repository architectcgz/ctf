package app

import (
	"os"
	"strings"
	"testing"
)

func TestUserSelfRoutesAreExtractedIntoDedicatedRegistrarFile(t *testing.T) {
	t.Parallel()

	mainRoutes, err := os.ReadFile("router_routes.go")
	if err != nil {
		t.Fatalf("read router_routes.go: %v", err)
	}

	for _, marker := range []string{
		`contestGroup := apiV1.Group("/contests")`,
		`protected.GET("/challenges",`,
		`usersGroup := protected.Group("/users")`,
		`protected.POST("/reports/personal",`,
	} {
		if strings.Contains(string(mainRoutes), marker) {
			t.Fatalf("expected %q to move out of router_routes.go", marker)
		}
	}

	selfRoutes, err := os.ReadFile("router_user_self_routes.go")
	if err != nil {
		t.Fatalf("read router_user_self_routes.go: %v", err)
	}

	for _, marker := range []string{
		"func registerUserSelfRoutes(",
		`contestGroup := apiV1.Group("/contests")`,
		`usersGroup := protected.Group("/users")`,
		`protected.POST("/reports/personal",`,
	} {
		if !strings.Contains(string(selfRoutes), marker) {
			t.Fatalf("router_user_self_routes.go should contain %q", marker)
		}
	}
}

func TestTeacherRoutesAreExtractedIntoDedicatedRegistrarFile(t *testing.T) {
	t.Parallel()

	mainRoutes, err := os.ReadFile("router_routes.go")
	if err != nil {
		t.Fatalf("read router_routes.go: %v", err)
	}

	for _, marker := range []string{
		`teacherOrAbove.GET("/overview",`,
		`teacherOrAbove.GET("/awd/reviews",`,
		`teacherOrAbove.POST("/reports/class",`,
		`usersGroup.GET("/:id/skill-profile",`,
	} {
		if strings.Contains(string(mainRoutes), marker) {
			t.Fatalf("expected %q to move out of router_routes.go", marker)
		}
	}

	teacherRoutes, err := os.ReadFile("router_user_teacher_routes.go")
	if err != nil {
		t.Fatalf("read router_user_teacher_routes.go: %v", err)
	}

	for _, marker := range []string{
		"func registerTeacherRoutes(",
		`teacherOrAbove.GET("/overview",`,
		`teacherOrAbove.GET("/awd/reviews",`,
		`protected.GET("/users/:id/skill-profile",`,
		`teacherOrAbove.POST("/reports/class",`,
	} {
		if !strings.Contains(string(teacherRoutes), marker) {
			t.Fatalf("router_user_teacher_routes.go should contain %q", marker)
		}
	}
}
