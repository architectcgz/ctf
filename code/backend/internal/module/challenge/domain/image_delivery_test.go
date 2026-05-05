package domain

import "testing"

func TestBuildPlatformImageRefUsesModeNamespace(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		registry string
		mode     string
		slug     string
		tag      string
		want     string
	}{
		{
			name:     "jeopardy",
			registry: "registry.example.edu",
			mode:     ChallengePackageModeJeopardy,
			slug:     "web-demo",
			tag:      "v1",
			want:     "registry.example.edu/jeopardy/web-demo:v1",
		},
		{
			name:     "awd",
			registry: "registry.example.edu",
			mode:     ChallengePackageModeAWD,
			slug:     "awd-demo",
			tag:      "c1",
			want:     "registry.example.edu/awd/awd-demo:c1",
		},
		{
			name:     "local without registry",
			registry: "",
			mode:     ChallengePackageModeJeopardy,
			slug:     "web-demo",
			tag:      "v1",
			want:     "jeopardy/web-demo:v1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BuildPlatformImageRef(tc.registry, tc.mode, tc.slug, tc.tag)
			if err != nil {
				t.Fatalf("BuildPlatformImageRef() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("BuildPlatformImageRef() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestExtractImageTagSuggestionIgnoresRepository(t *testing.T) {
	t.Parallel()

	got := ExtractImageTagSuggestion("registry.example.edu/team/web-demo:v1", "")
	if got != "v1" {
		t.Fatalf("ExtractImageTagSuggestion() = %q, want v1", got)
	}

	got = ExtractImageTagSuggestion("team/web-demo:v1", "c2")
	if got != "c2" {
		t.Fatalf("explicit tag should win, got %q", got)
	}
}

func TestSplitImageRefHandlesRegistryPort(t *testing.T) {
	t.Parallel()

	name, tag, err := SplitImageRef("127.0.0.1:5000/ctf/web-demo:v1")
	if err != nil {
		t.Fatalf("SplitImageRef() error = %v", err)
	}
	if name != "127.0.0.1:5000/ctf/web-demo" || tag != "v1" {
		t.Fatalf("SplitImageRef() = %q:%q", name, tag)
	}
}
