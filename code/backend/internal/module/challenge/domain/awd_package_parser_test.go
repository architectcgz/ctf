package domain

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseAWDServiceTemplatePackageDir(t *testing.T) {
	rootDir := t.TempDir()

	manifest := `api_version: v1
kind: challenge

meta:
  mode: awd
  slug: awd-bank-portal-01
  title: Bank Portal AWD
  category: web
  difficulty: hard
  points: 500

content:
  statement: statement.md

flag:
  type: dynamic
  prefix: awd

runtime:
  type: container
  image:
    ref: registry.example.edu/ctf/awd-bank-portal:v1

extensions:
  awd:
    service_type: web_http
    deployment_mode: single_container
    version: v2026.04
    checker:
      type: http_standard
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
    flag_policy:
      mode: dynamic_team
      config:
        flag_prefix: awd
        rotate_interval_sec: 120
    defense_entry:
      mode: http
    access_config:
      public_base_url: http://{{TEAM_HOST}}:8080
      service_port: 8080
      exposed_ports:
        - port: 8080
          protocol: tcp
          purpose: http
    runtime_config:
      instance_sharing: per_team
      service_port: 8080
`

	if err := os.WriteFile(filepath.Join(rootDir, "challenge.yml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write challenge.yml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "statement.md"), []byte("银行门户存在越权修改 flag 的逻辑。"), 0o644); err != nil {
		t.Fatalf("write statement.md: %v", err)
	}

	parsed, err := ParseAWDServiceTemplatePackageDir(rootDir)
	if err != nil {
		t.Fatalf("ParseAWDServiceTemplatePackageDir() error = %v", err)
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

	if _, err := buildParsedChallengePackage(rootDir, manifest); err == nil {
		t.Fatal("expected buildParsedChallengePackage() to reject awd mode")
	}
}
