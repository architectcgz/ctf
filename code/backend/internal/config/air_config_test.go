package config

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestAirConfigExcludesRuntimeDataDirectories(t *testing.T) {
	airConfigPath := filepath.Join("..", "..", ".air.toml")
	content, err := os.ReadFile(airConfigPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", airConfigPath, err)
	}

	matches := regexp.MustCompile(`(?m)^exclude_dir\s*=\s*\[(.*)\]\s*$`).FindStringSubmatch(string(content))
	if len(matches) != 2 {
		t.Fatal(".air.toml should define build.exclude_dir")
	}

	excludedDirs := matches[1]
	for _, dir := range []string{"storage", "data"} {
		if !strings.Contains(excludedDirs, `"`+dir+`"`) {
			t.Fatalf(".air.toml exclude_dir should include %q, got [%s]", dir, excludedDirs)
		}
	}
}
