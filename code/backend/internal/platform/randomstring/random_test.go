package randomstring

import "testing"

func TestGenerate(t *testing.T) {
	t.Parallel()

	value, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(value) == 0 {
		t.Fatal("generated string should not be empty")
	}

	value2, err := Generate()
	if err != nil {
		t.Fatalf("Generate() second call error = %v", err)
	}
	if value == value2 {
		t.Fatal("generated strings should be unique")
	}
}
