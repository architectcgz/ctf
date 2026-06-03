package app

import (
	"reflect"
	"testing"

	"ctf-platform/internal/app/composition"
	practicehttp "ctf-platform/internal/module/practice/api/http"
	teachinghttp "ctf-platform/internal/module/teaching_query/api/http"
	"go.uber.org/zap"
)

func TestTeacherRoutesAreServedByTeachingQuery(t *testing.T) {
	cfg, db, cache := newAppTestDependencies(t)

	originalBuildTeachingQueryModule := buildTeachingQueryModule
	t.Cleanup(func() {
		buildTeachingQueryModule = originalBuildTeachingQueryModule
	})

	called := false
	buildTeachingQueryModule = func(root *composition.Root, assessment *composition.AssessmentModule, identity *composition.IdentityModule) *composition.TeachingQueryModule {
		module := originalBuildTeachingQueryModule(root, assessment, identity)
		called = true
		if module == nil || module.Handler == nil {
			t.Fatal("expected teaching query module handler")
		}
		if got, want := reflect.TypeOf(module.Handler), reflect.TypeOf(&teachinghttp.Handler{}); got != want {
			t.Fatalf("teaching query handler type = %v, want %v", got, want)
		}
		return module
	}

	router, err := NewRouter(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}
	if router == nil {
		t.Fatal("expected router")
	}
	if !called {
		t.Fatal("expected teaching query module builder to be called")
	}
}

func TestStudentPracticeReadRoutesAreServedByPracticeModule(t *testing.T) {
	cfg, db, cache := newAppTestDependencies(t)

	originalBuildPracticeModule := buildPracticeModule
	t.Cleanup(func() {
		buildPracticeModule = originalBuildPracticeModule
	})

	called := false
	buildPracticeModule = func(root *composition.Root, challenge *composition.ChallengeModule, instance *composition.InstanceModule) *composition.PracticeModule {
		module := originalBuildPracticeModule(root, challenge, instance)
		called = true
		if module == nil || module.Handler == nil {
			t.Fatal("expected practice module handler")
		}
		if got, want := reflect.TypeOf(module.Handler), reflect.TypeOf(&practicehttp.Handler{}); got != want {
			t.Fatalf("practice handler type = %v, want %v", got, want)
		}
		return module
	}

	router, err := NewRouter(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}
	if router == nil {
		t.Fatal("expected router")
	}
	if !called {
		t.Fatal("expected practice module builder to be called")
	}
}
