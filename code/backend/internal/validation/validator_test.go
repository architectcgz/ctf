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
