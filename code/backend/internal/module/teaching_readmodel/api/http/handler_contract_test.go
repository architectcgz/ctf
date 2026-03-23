package http

import (
	"reflect"
	"testing"

	teachingreadmodel "ctf-platform/internal/module/teaching_readmodel"
	readmodelapp "ctf-platform/internal/module/teaching_readmodel/application"
)

func TestQueryServiceImplementsTeachingQuery(t *testing.T) {
	var _ teachingreadmodel.TeachingQuery = (*readmodelapp.QueryService)(nil)
}

func TestHandlerDependsOnTeachingQuery(t *testing.T) {
	t.Parallel()

	want := reflect.TypeOf((*teachingreadmodel.TeachingQuery)(nil)).Elem()

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
