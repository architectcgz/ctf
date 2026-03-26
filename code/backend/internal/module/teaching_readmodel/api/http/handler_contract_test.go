package http

import (
	"reflect"
	"testing"

	teachingreadmodelqueries "ctf-platform/internal/module/teaching_readmodel/application/queries"
)

func TestQueryServiceImplementsTeachingQuery(t *testing.T) {
	var _ teachingreadmodelqueries.Service = (*teachingreadmodelqueries.QueryService)(nil)
}

func TestHandlerDependsOnTeachingQuery(t *testing.T) {
	t.Parallel()

	want := reflect.TypeOf((*teachingreadmodelqueries.Service)(nil)).Elem()

	if got := reflect.TypeOf(NewHandler).In(0); got != want {
		t.Fatalf("NewHandler() parameter type = %s, want %s", got, want)
	}

	field, ok := reflect.TypeOf(Handler{}).FieldByName("service")
	if !ok {
		t.Fatal("Handler missing service field")
	}
	if field.Type != want {
		t.Fatalf("Handler.service type = %s, want %s", field.Type, want)
	}
}
