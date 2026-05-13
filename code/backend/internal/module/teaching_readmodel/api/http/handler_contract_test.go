package http

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	teachingreadmodelqueries "ctf-platform/internal/module/teaching_readmodel/application/queries"
)

func TestQueryServiceImplementsTeachingQuery(t *testing.T) {
	var _ teachingreadmodelqueries.Service = (*teachingreadmodelqueries.QueryService)(nil)
	var _ teachingreadmodelqueries.OverviewService = (*teachingreadmodelqueries.OverviewQueryService)(nil)
	var _ teachingreadmodelqueries.ClassInsightService = (*teachingreadmodelqueries.ClassInsightQueryService)(nil)
	var _ teachingreadmodelqueries.StudentReviewService = (*teachingreadmodelqueries.StudentReviewQueryService)(nil)
}

func TestHandlerDependsOnTeachingQuery(t *testing.T) {
	t.Parallel()

	handlerType := reflect.TypeOf(NewHandler)
	serviceType := reflect.TypeOf((*teachingreadmodelqueries.Service)(nil)).Elem()
	overviewType := reflect.TypeOf((*teachingreadmodelqueries.OverviewService)(nil)).Elem()
	classInsightType := reflect.TypeOf((*teachingreadmodelqueries.ClassInsightService)(nil)).Elem()

	if got := handlerType.NumIn(); got != 4 {
		t.Fatalf("NewHandler() parameter count = %d, want 4", got)
	}
	if got := handlerType.In(0); got != serviceType {
		t.Fatalf("NewHandler() first parameter type = %s, want %s", got, serviceType)
	}
	if got := handlerType.In(1); got != overviewType {
		t.Fatalf("NewHandler() second parameter type = %s, want %s", got, overviewType)
	}
	if got := handlerType.In(2); got != classInsightType {
		t.Fatalf("NewHandler() third parameter type = %s, want %s", got, classInsightType)
	}
	if got := handlerType.In(3).Name(); got != "StudentReviewService" {
		t.Fatalf("NewHandler() fourth parameter name = %s, want StudentReviewService", got)
	}
	if got := handlerType.In(3).PkgPath(); got != "ctf-platform/internal/module/teaching_readmodel/application/queries" {
		t.Fatalf("NewHandler() fourth parameter pkg = %s, want teaching_readmodel/application/queries", got)
	}

	field, ok := reflect.TypeOf(Handler{}).FieldByName("service")
	if !ok {
		t.Fatal("Handler missing service field")
	}
	if field.Type != serviceType {
		t.Fatalf("Handler.service type = %s, want %s", field.Type, serviceType)
	}

	overviewField, ok := reflect.TypeOf(Handler{}).FieldByName("overviewService")
	if !ok {
		t.Fatal("Handler missing overviewService field")
	}
	if overviewField.Type != overviewType {
		t.Fatalf("Handler.overviewService type = %s, want %s", overviewField.Type, overviewType)
	}

	classInsightField, ok := reflect.TypeOf(Handler{}).FieldByName("classInsightService")
	if !ok {
		t.Fatal("Handler missing classInsightService field")
	}
	if classInsightField.Type != classInsightType {
		t.Fatalf("Handler.classInsightService type = %s, want %s", classInsightField.Type, classInsightType)
	}

	studentReviewField, ok := reflect.TypeOf(Handler{}).FieldByName("studentReviewService")
	if !ok {
		t.Fatal("Handler missing studentReviewService field")
	}
	if got := studentReviewField.Type.Name(); got != "StudentReviewService" {
		t.Fatalf("Handler.studentReviewService name = %s, want StudentReviewService", got)
	}
	if got := studentReviewField.Type.PkgPath(); got != "ctf-platform/internal/module/teaching_readmodel/application/queries" {
		t.Fatalf("Handler.studentReviewService pkg = %s, want teaching_readmodel/application/queries", got)
	}
}

func TestHandlerRoutesStudentEndpointsThroughStudentReviewService(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("handler.go"))
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"h.studentReviewService.GetStudentProgress(",
		"h.studentReviewService.GetStudentRecommendations(",
		"h.studentReviewService.GetStudentTimeline(",
		"h.studentReviewService.GetStudentEvidence(",
		"h.studentReviewService.GetStudentAttackSessions(",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("handler.go must route student review endpoint through dedicated service marker %s", marker)
		}
	}
}
