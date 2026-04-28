package dto

import (
	"reflect"
	"testing"
)

func TestPageResultListIsTypedSlice(t *testing.T) {
	t.Parallel()

	resultType := reflect.TypeOf(PageResult[string]{})
	field, ok := resultType.FieldByName("List")
	if !ok {
		t.Fatal("PageResult.List field not found")
	}
	if field.Type.Kind() != reflect.Slice || field.Type.Elem().Kind() != reflect.String {
		t.Fatalf("expected PageResult[T].List to be []T, got %s", field.Type)
	}
}
