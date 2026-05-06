package domain

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ctf-platform/pkg/errcode"
)

const (
	defaultHTTPCheckerYAML = `      type: http_standard
      config:
        put_flag:
          method: PUT
          path: /api/flag
          expected_status: 200
          body_template: "{{FLAG}}"
        get_flag:
          method: GET
          path: /api/flag
          expected_status: 200
          expected_substring: "{{FLAG}}"
        havoc:
          method: GET
          path: /healthz
          expected_status: 200
`

	defaultTCPCheckerYAML = `      type: tcp_standard
      config:
        timeout_ms: 3000
        steps:
          - send: "PING\n"
            expect_contains: PONG
          - send_template: "SET_FLAG {{FLAG}}\n"
            expect_contains: OK
          - send: "GET_FLAG\n"
            expect_contains: "{{FLAG}}"
`

	defaultHTTPAccessConfigYAML = `      public_base_url: http://{{TEAM_HOST}}:8080
      service_port: 8080
      exposed_ports:
        - port: 8080
          protocol: tcp
          purpose: http
`

	defaultTCPAccessConfigYAML = `      public_base_url: tcp://{{TEAM_HOST}}:8080
      service_port: 8080
      exposed_ports:
        - port: 8080
          protocol: tcp
          purpose: tcp
`

	defaultDefenseWorkspaceYAML = `      defense_workspace:
        entry_mode: ssh
        seed_root: docker/workspace
        workspace_roots:
          - docker/workspace/src
          - docker/workspace/templates
          - docker/workspace/static
          - docker/workspace/data
        writable_roots:
          - docker/workspace/src
          - docker/workspace/templates
          - docker/workspace/static
        readonly_roots:
          - docker/workspace/data
        runtime_mounts:
          - source: docker/workspace/src
            target: /workspace/src
            mode: rw
          - source: docker/workspace/templates
            target: /workspace/templates
            mode: rw
          - source: docker/workspace/static
            target: /workspace/static
            mode: rw
          - source: docker/workspace/data
            target: /workspace/data
            mode: ro
`

	defaultDefenseScopeYAML = `      defense_scope:
        protected_paths:
          - docker/runtime/app.py
          - docker/runtime/ctf_runtime.py
          - docker/check/check.py
          - challenge.yml
        service_contracts:
          - /health 必须返回 200
`
)

type awdManifestOptions struct {
	Slug               string
	Title              string
	Category           string
	Difficulty         string
	ServiceType        string
	Version            string
	DefenseEntryMode   string
	RuntimeImageBlock  string
	CheckerBlock       string
	AccessConfigBlock  string
	RuntimeConfigBlock string
}

func TestParseAWDChallengePackageDir(t *testing.T) {
	rootDir := t.TempDir()
	writeDefaultAWDPackageLayout(t, rootDir, false, nil)
	writeAWDChallengeManifest(t, rootDir, buildAWDManifest(awdManifestOptions{
		Slug:              "awd-bank-portal-01",
		Title:             "Bank Portal AWD",
		Category:          "web",
		Difficulty:        "hard",
		ServiceType:       "web_http",
		Version:           "v2026.04",
		DefenseEntryMode:  "http",
		RuntimeImageBlock: "  image:\n    ref: registry.example.edu/ctf/awd-bank-portal:v1\n",
		CheckerBlock:      defaultHTTPCheckerYAML,
		AccessConfigBlock: defaultHTTPAccessConfigYAML,
		RuntimeConfigBlock: joinRuntimeConfigBlocks(
			"      checker_token_env: CHECKER_TOKEN\n",
			defaultDefenseWorkspaceYAML,
			defaultDefenseScopeYAML,
		),
	}))

	parsed, err := ParseAWDChallengePackageDir(rootDir)
	if err != nil {
		t.Fatalf("ParseAWDChallengePackageDir() error = %v", err)
	}

	if parsed.Slug != "awd-bank-portal-01" || parsed.Title != "Bank Portal AWD" {
		t.Fatalf("unexpected meta: %+v", parsed)
	}
	if parsed.ServiceType != "web_http" || parsed.DeploymentMode != "single_container" {
		t.Fatalf("unexpected awd service shape: %+v", parsed)
	}
	if parsed.CheckerType != "http_standard" {
		t.Fatalf("unexpected checker type: %+v", parsed)
	}
	if parsed.FlagMode != "dynamic_team" || parsed.DefenseEntryMode != "http" {
		t.Fatalf("unexpected flag/entry mode: %+v", parsed)
	}
	if parsed.AccessConfig["service_port"] != float64(8080) {
		t.Fatalf("unexpected access_config: %+v", parsed.AccessConfig)
	}
	if parsed.RuntimeConfig["instance_sharing"] != "per_team" {
		t.Fatalf("unexpected runtime_config: %+v", parsed.RuntimeConfig)
	}
	if parsed.RuntimeImageRef != "registry.example.edu/ctf/awd-bank-portal:v1" {
		t.Fatalf("unexpected runtime image ref: %s", parsed.RuntimeImageRef)
	}

	defenseWorkspace, ok := parsed.RuntimeConfig["defense_workspace"].(map[string]any)
	if !ok {
		t.Fatalf("expected defense_workspace in runtime_config, got %+v", parsed.RuntimeConfig)
	}
	workspaceRoots, ok := defenseWorkspace["workspace_roots"].([]any)
	if !ok || len(workspaceRoots) != 4 {
		t.Fatalf("unexpected defense_workspace.workspace_roots: %+v", defenseWorkspace)
	}
	if defenseWorkspace["seed_root"] != "docker/workspace" {
		t.Fatalf("unexpected defense_workspace.seed_root: %+v", defenseWorkspace)
	}
	defenseScope, ok := parsed.RuntimeConfig["defense_scope"].(map[string]any)
	if !ok {
		t.Fatalf("expected defense_scope in runtime_config, got %+v", parsed.RuntimeConfig)
	}
	if _, exists := defenseScope["editable_paths"]; exists {
		t.Fatalf("expected defense_scope.editable_paths to be absent, got %+v", defenseScope)
	}
}

func TestParseAWDChallengePackageDirAcceptsTCPStandardChecker(t *testing.T) {
	rootDir := t.TempDir()
	writeDefaultAWDPackageLayout(t, rootDir, false, nil)
	writeAWDChallengeManifest(t, rootDir, buildAWDManifest(awdManifestOptions{
		Slug:              "awd-tcp-length-gate",
		Title:             "TCP Length Gate",
		Category:          "pwn",
		Difficulty:        "medium",
		ServiceType:       "binary_tcp",
		Version:           "v2026.04",
		DefenseEntryMode:  "tcp",
		RuntimeImageBlock: "  image:\n    ref: registry.example.edu/ctf/awd-tcp-length-gate:v1\n",
		CheckerBlock:      defaultTCPCheckerYAML,
		AccessConfigBlock: defaultTCPAccessConfigYAML,
		RuntimeConfigBlock: joinRuntimeConfigBlocks(
			defaultDefenseWorkspaceYAML,
			defaultDefenseScopeYAML,
		),
	}))

	parsed, err := ParseAWDChallengePackageDir(rootDir)
	if err != nil {
		t.Fatalf("ParseAWDChallengePackageDir() error = %v", err)
	}

	if parsed.ServiceType != "binary_tcp" {
		t.Fatalf("ServiceType = %q, want binary_tcp", parsed.ServiceType)
	}
	if parsed.CheckerType != "tcp_standard" {
		t.Fatalf("CheckerType = %q, want tcp_standard", parsed.CheckerType)
	}
	if parsed.RuntimeConfig["service_port"] != float64(8080) {
		t.Fatalf("unexpected runtime_config: %+v", parsed.RuntimeConfig)
	}
	if parsed.CheckerConfig["timeout_ms"] != float64(3000) {
		t.Fatalf("unexpected checker_config: %+v", parsed.CheckerConfig)
	}
}

func TestParseAWDChallengePackageDirAllowsPlatformBuildWithoutRuntimeImageRef(t *testing.T) {
	rootDir := t.TempDir()
	writeDefaultAWDPackageLayout(t, rootDir, true, nil)
	writeAWDChallengeManifest(t, rootDir, buildAWDManifest(awdManifestOptions{
		Slug:              "awd-platform-build",
		Title:             "AWD Platform Build",
		Category:          "web",
		Difficulty:        "hard",
		ServiceType:       "web_http",
		Version:           "v2026.05",
		DefenseEntryMode:  "http",
		RuntimeImageBlock: "  image:\n    tag: c1\n",
		CheckerBlock:      defaultHTTPCheckerYAML,
		AccessConfigBlock: defaultHTTPAccessConfigYAML,
		RuntimeConfigBlock: joinRuntimeConfigBlocks(
			defaultDefenseWorkspaceYAML,
			defaultDefenseScopeYAML,
		),
	}))

	parsed, err := ParseAWDChallengePackageDir(rootDir)
	if err != nil {
		t.Fatalf("ParseAWDChallengePackageDir() error = %v", err)
	}

	if parsed.RuntimeImageRef != "" {
		t.Fatalf("expected no author runtime image ref, got %q", parsed.RuntimeImageRef)
	}
	if parsed.ImageSourceType != ImageSourceTypePlatformBuild {
		t.Fatalf("ImageSourceType = %q, want %q", parsed.ImageSourceType, ImageSourceTypePlatformBuild)
	}
	if parsed.SuggestedImageTag != "c1" {
		t.Fatalf("SuggestedImageTag = %q, want c1", parsed.SuggestedImageTag)
	}
	if !strings.HasSuffix(filepath.ToSlash(parsed.DockerfilePath), "docker/runtime/Dockerfile") {
		t.Fatalf("expected runtime Dockerfile path, got %q", parsed.DockerfilePath)
	}
	if !strings.HasSuffix(filepath.ToSlash(parsed.BuildContextPath), "docker") {
		t.Fatalf("expected docker build context, got %q", parsed.BuildContextPath)
	}
}

func TestParseAWDChallengePackageDirRejectsInvalidDefenseWorkspace(t *testing.T) {
	cases := []struct {
		name               string
		runtimeConfigBlock string
		extraFiles         map[string]string
		wantContains       string
	}{
		{
			name: "missing",
			runtimeConfigBlock: joinRuntimeConfigBlocks(
				defaultDefenseScopeYAML,
			),
			wantContains: "defense_workspace 不能为空",
		},
		{
			name: "protected runtime root",
			runtimeConfigBlock: joinRuntimeConfigBlocks(
				`      defense_workspace:
        entry_mode: ssh
        seed_root: docker/workspace
        workspace_roots:
          - docker/runtime
        writable_roots:
          - docker/runtime
        readonly_roots:
          - docker/workspace/data
        runtime_mounts:
          - source: docker/runtime
            target: /workspace/runtime
            mode: rw
`,
				defaultDefenseScopeYAML,
			),
			wantContains: "不能包含受保护路径: docker/runtime",
		},
		{
			name: "challenge manifest root",
			runtimeConfigBlock: joinRuntimeConfigBlocks(
				`      defense_workspace:
        entry_mode: ssh
        seed_root: docker/workspace
        workspace_roots:
          - challenge.yml
        writable_roots:
          - challenge.yml
        readonly_roots:
          - docker/workspace/data
        runtime_mounts:
          - source: challenge.yml
            target: /workspace/challenge.yml
            mode: rw
`,
				defaultDefenseScopeYAML,
			),
			wantContains: "不能包含受保护路径: challenge.yml",
		},
		{
			name: "single file legacy boundary",
			runtimeConfigBlock: joinRuntimeConfigBlocks(
				`      defense_workspace:
        entry_mode: ssh
        seed_root: docker/workspace
        workspace_roots:
          - docker/challenge_app.py
        writable_roots:
          - docker/challenge_app.py
        readonly_roots:
          - docker/workspace/data
        runtime_mounts:
          - source: docker/challenge_app.py
            target: /workspace/challenge_app.py
            mode: rw
`,
				defaultDefenseScopeYAML,
			),
			extraFiles: map[string]string{
				"docker/challenge_app.py": "print('legacy')\n",
			},
			wantContains: "必须指向目录: docker/challenge_app.py",
		},
		{
			name: "overlap writable and readonly",
			runtimeConfigBlock: joinRuntimeConfigBlocks(
				`      defense_workspace:
        entry_mode: ssh
        seed_root: docker/workspace
        workspace_roots:
          - docker/workspace/src
          - docker/workspace/data
        writable_roots:
          - docker/workspace/src
        readonly_roots:
          - docker/workspace/src
        runtime_mounts:
          - source: docker/workspace/src
            target: /workspace/src
            mode: rw
`,
				defaultDefenseScopeYAML,
			),
			wantContains: "writable_roots 与 readonly_roots 不能重叠",
		},
		{
			name: "mount source outside workspace roots",
			runtimeConfigBlock: joinRuntimeConfigBlocks(
				`      defense_workspace:
        entry_mode: ssh
        seed_root: docker/workspace
        workspace_roots:
          - docker/workspace/src
          - docker/workspace/templates
          - docker/workspace/static
          - docker/workspace/data
        writable_roots:
          - docker/workspace/src
          - docker/workspace/templates
          - docker/workspace/static
        readonly_roots:
          - docker/workspace/data
        runtime_mounts:
          - source: docker/workspace/private
            target: /workspace/private
            mode: rw
`,
				defaultDefenseScopeYAML,
			),
			extraFiles: map[string]string{
				"docker/workspace/private/secret.txt": "secret\n",
			},
			wantContains: "runtime_mounts.source 必须来自 workspace_roots",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rootDir := t.TempDir()
			writeDefaultAWDPackageLayout(t, rootDir, false, tc.extraFiles)
			writeAWDChallengeManifest(t, rootDir, buildAWDManifest(awdManifestOptions{
				Slug:              "awd-defense-workspace",
				Title:             "AWD Defense Workspace",
				Category:          "web",
				Difficulty:        "hard",
				ServiceType:       "web_http",
				Version:           "v2026.04",
				DefenseEntryMode:  "http",
				RuntimeImageBlock: "  image:\n    ref: registry.example.edu/ctf/awd-defense-workspace:v1\n",
				CheckerBlock:      defaultHTTPCheckerYAML,
				AccessConfigBlock: defaultHTTPAccessConfigYAML,
				RuntimeConfigBlock: tc.runtimeConfigBlock,
			}))

			_, err := ParseAWDChallengePackageDir(rootDir)
			if err == nil {
				t.Fatal("expected invalid defense_workspace to be rejected")
			}
			assertAppErrorCauseContains(t, err, tc.wantContains)
		})
	}
}

func TestParseAWDChallengePackageDirRejectsLegacyDockerfileLayout(t *testing.T) {
	rootDir := t.TempDir()
	writeDefaultAWDPackageLayout(t, rootDir, false, map[string]string{
		"docker/Dockerfile": "FROM python:3.12-alpine\n",
	})
	writeAWDChallengeManifest(t, rootDir, buildAWDManifest(awdManifestOptions{
		Slug:              "awd-legacy-dockerfile",
		Title:             "AWD Legacy Dockerfile",
		Category:          "web",
		Difficulty:        "hard",
		ServiceType:       "web_http",
		Version:           "v2026.04",
		DefenseEntryMode:  "http",
		RuntimeImageBlock: "  image:\n    tag: c1\n",
		CheckerBlock:      defaultHTTPCheckerYAML,
		AccessConfigBlock: defaultHTTPAccessConfigYAML,
		RuntimeConfigBlock: joinRuntimeConfigBlocks(
			defaultDefenseWorkspaceYAML,
			defaultDefenseScopeYAML,
		),
	}))

	_, err := ParseAWDChallengePackageDir(rootDir)
	if err == nil {
		t.Fatal("expected legacy Dockerfile layout to be rejected")
	}
	assertAppErrorCauseContains(t, err, "docker/runtime/Dockerfile")
}

func TestParseAWDChallengePackageDirRejectsInvalidScriptCheckerFiles(t *testing.T) {
	cases := []struct {
		name      string
		filesYAML string
	}{
		{name: "absolute", filesYAML: "          - /tmp/check.py\n"},
		{name: "parent", filesYAML: "          - ../check.py\n"},
		{name: "directory", filesYAML: "          - docker/check\n"},
		{name: "missing", filesYAML: "          - docker/check/missing.py\n"},
		{name: "entry not included", filesYAML: "          - docker/check/protocol.py\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rootDir := t.TempDir()
			writeDefaultAWDPackageLayout(t, rootDir, false, map[string]string{
				"docker/check/protocol.py": "OK = True\n",
			})
			writeAWDChallengeManifest(t, rootDir, buildAWDManifest(awdManifestOptions{
				Slug:              "script-checker-files",
				Title:             "Script Checker Files",
				Category:          "web",
				Difficulty:        "hard",
				ServiceType:       "web_http",
				Version:           "v2026.04",
				DefenseEntryMode:  "http",
				RuntimeImageBlock: "  image:\n    ref: registry.example.edu/ctf/script:v1\n",
				CheckerBlock: `      type: script_checker
      config:
        runtime: python3
        entry: docker/check/check.py
        files:
` + tc.filesYAML + `        timeout_sec: 10
        args:
          - "{{TARGET_URL}}"
        output: json
`,
				AccessConfigBlock: "      service_port: 8080\n",
				RuntimeConfigBlock: joinRuntimeConfigBlocks(
					defaultDefenseWorkspaceYAML,
					defaultDefenseScopeYAML,
				),
			}))

			if _, err := ParseAWDChallengePackageDir(rootDir); err == nil {
				t.Fatal("expected invalid script_checker files to be rejected")
			}
		})
	}
}

func TestBuildParsedChallengePackageRejectsAwdModeForJeopardyImport(t *testing.T) {
	rootDir := t.TempDir()
	manifest := &ChallengePackageManifest{
		APIVersion: "v1",
		Kind:       "challenge",
		Meta: ChallengePackageMeta{
			Mode:       "awd",
			Slug:       "awd-bank-portal-01",
			Title:      "Bank Portal AWD",
			Category:   "web",
			Difficulty: "hard",
			Points:     500,
		},
		Content: ChallengePackageContent{
			Statement: "statement.md",
		},
		Flag: ChallengePackageFlag{
			Type:   "dynamic",
			Prefix: "awd",
		},
		Runtime: ChallengePackageRuntime{
			Type: "container",
			Image: ChallengePackageRuntimeImage{
				Ref: "registry.example.edu/ctf/awd-bank-portal:v1",
			},
		},
	}

	if err := os.WriteFile(filepath.Join(rootDir, "statement.md"), []byte("awd statement"), 0o644); err != nil {
		t.Fatalf("write statement.md: %v", err)
	}

	if _, err := buildParsedChallengePackage(rootDir, manifest, ""); err == nil {
		t.Fatal("expected buildParsedChallengePackage() to reject awd mode")
	}
}

func buildAWDManifest(opts awdManifestOptions) string {
	category := opts.Category
	if category == "" {
		category = "web"
	}
	difficulty := opts.Difficulty
	if difficulty == "" {
		difficulty = "hard"
	}
	serviceType := opts.ServiceType
	if serviceType == "" {
		serviceType = "web_http"
	}
	version := opts.Version
	if version == "" {
		version = "v2026.04"
	}
	defenseEntryMode := opts.DefenseEntryMode
	if defenseEntryMode == "" {
		defenseEntryMode = "http"
	}
	return fmt.Sprintf(`api_version: v1
kind: challenge

meta:
  mode: awd
  slug: %s
  title: %s
  category: %s
  difficulty: %s
  points: 500

content:
  statement: statement.md

flag:
  type: dynamic
  prefix: awd

runtime:
  type: container
%sextensions:
  awd:
    service_type: %s
    deployment_mode: single_container
    version: %s
    checker:
%s    flag_policy:
      mode: dynamic_team
      config:
        flag_prefix: awd
        rotate_interval_sec: 120
    defense_entry:
      mode: %s
    access_config:
%s    runtime_config:
      instance_sharing: per_team
      service_port: 8080
%s`,
		opts.Slug,
		opts.Title,
		category,
		difficulty,
		opts.RuntimeImageBlock,
		serviceType,
		version,
		opts.CheckerBlock,
		defenseEntryMode,
		opts.AccessConfigBlock,
		opts.RuntimeConfigBlock,
	)
}

func joinRuntimeConfigBlocks(blocks ...string) string {
	var builder strings.Builder
	for _, block := range blocks {
		builder.WriteString(block)
	}
	return builder.String()
}

func writeAWDChallengeManifest(t *testing.T, rootDir, manifest string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(rootDir, "challenge.yml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write challenge.yml: %v", err)
	}
}

func writeDefaultAWDPackageLayout(t *testing.T, rootDir string, withRuntimeDockerfile bool, extraFiles map[string]string) {
	t.Helper()
	files := map[string]string{
		"statement.md":                     "AWD package statement.",
		"docker/runtime/app.py":           "print('entry')\n",
		"docker/runtime/ctf_runtime.py":   "print('runtime')\n",
		"docker/workspace/src/app.py":     "print('workspace entry')\n",
		"docker/workspace/src/service.py": "print('service logic')\n",
		"docker/workspace/templates/index.html": "<h1>workspace</h1>\n",
		"docker/workspace/static/site.css":      "body { color: black; }\n",
		"docker/workspace/data/seed.txt":        "seed\n",
		"docker/check/check.py":                 "print('check')\n",
	}
	if withRuntimeDockerfile {
		files["docker/runtime/Dockerfile"] = "FROM python:3.12-alpine\nWORKDIR /app\nCOPY runtime /app/runtime\n"
	}
	for path, content := range extraFiles {
		files[path] = content
	}
	writeTestFiles(t, rootDir, files)
}

func writeTestFiles(t *testing.T, rootDir string, files map[string]string) {
	t.Helper()
	for relPath, content := range files {
		fullPath := filepath.Join(rootDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("create parent for %s: %v", relPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", relPath, err)
		}
	}
}

func assertAppErrorCauseContains(t *testing.T, err error, want string) {
	t.Helper()
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T (%v)", err, err)
	}
	if appErr.Cause == nil || !strings.Contains(appErr.Cause.Error(), want) {
		t.Fatalf("expected cause containing %q, got %v", want, appErr.Cause)
	}
}
