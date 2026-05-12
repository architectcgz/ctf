package http

import (
	"reflect"
	"testing"

	teachingreadmodelqueries "ctf-platform/internal/module/teaching_readmodel/application/queries"
)

func TestQueryServiceImplementsTeachingQuery(t *testing.T) {
	var _ teachingreadmodelqueries.Service = (*teachingreadmodelqueries.QueryService)(nil)
	var _ teachingreadmodelqueries.OverviewService = (*teachingreadmodelqueries.OverviewQueryService)(nil)
}

func TestHandlerDependsOnTeachingQuery(t *testing.T) {
	t.Parallel()

	serviceType := reflect.TypeOf((*teachingreadmodelqueries.Service)(nil)).Elem()
	overviewType := reflect.TypeOf((*teachingreadmodelqueries.OverviewService)(nil)).Elem()

	if got := reflect.TypeOf(NewHandler).In(0); got != serviceType {
		t.Fatalf("NewHandler() first parameter type = %s, want %s", got, serviceType)
	}
	if got := reflect.TypeOf(NewHandler).In(1); got != overviewType {
		t.Fatalf("NewHandler() second parameter type = %s, want %s", got, overviewType)
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
}
