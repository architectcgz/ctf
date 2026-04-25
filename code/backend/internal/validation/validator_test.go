package validation

import "testing"

func TestIsValidUsername(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		value    string
		expected bool
	}{
		{name: "letters_digits_underscore", value: "alice_123", expected: true},
		{name: "dash_not_allowed", value: "alice-123", expected: false},
		{name: "space_not_allowed", value: "alice 123", expected: false},
		{name: "chinese_not_allowed", value: "测试", expected: false},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if actual := IsValidUsername(testCase.value); actual != testCase.expected {
				t.Fatalf("IsValidUsername(%q) = %v, want %v", testCase.value, actual, testCase.expected)
			}
		})
	}
}

func TestIsValidImageNameAllowsRegistryPaths(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"ctf/web-sqli",
		"registry.example.edu/ctf/awd-bank-portal",
		"registry.local:5000/ctf/web_01",
	}
	for _, name := range validNames {
		if !IsValidImageName(name) {
			t.Fatalf("expected valid image name %q", name)
		}
	}
}

func TestIsValidImageNameRejectsUnsafeCharacters(t *testing.T) {
	t.Parallel()

	invalidNames := []string{
		"CTF/Web",
		"ctf/web;rm",
		"ctf web",
	}
	for _, name := range invalidNames {
		if IsValidImageName(name) {
			t.Fatalf("expected invalid image name %q", name)
		}
	}
}

func TestIsValidImageTag(t *testing.T) {
	t.Parallel()

	if !IsValidImageTag("v1.0-debug") {
		t.Fatal("expected valid image tag")
	}
	if IsValidImageTag("-bad") {
		t.Fatal("expected invalid image tag")
	}
}
