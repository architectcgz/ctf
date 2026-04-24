package commands

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"path/filepath"
	"testing"

	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/testsupport"
)

func TestAWDServiceTemplateImportFlowPreviewAndCommit(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := NewAWDServiceTemplateImportService(db, repo)

	previewDir := filepath.Join(t.TempDir(), "awd-imports")
	t.Setenv("AWD_SERVICE_TEMPLATE_IMPORT_PREVIEW_DIR", previewDir)

	preview, err := service.PreviewImport(
		context.Background(),
		2001,
		"awd-bank-portal-01.zip",
		bytes.NewReader(buildAWDServiceTemplateImportArchive(t)),
	)
	if err != nil {
		t.Fatalf("PreviewImport() error = %v", err)
	}

	if preview.ID == "" || preview.Slug != "awd-bank-portal-01" {
		t.Fatalf("unexpected preview: %+v", preview)
	}
	if preview.ServiceType != "web_http" || preview.CheckerType != "http_standard" {
		t.Fatalf("unexpected preview awd fields: %+v", preview)
	}
	if preview.FlagMode != "dynamic_team" || preview.DefenseEntryMode != "http" {
		t.Fatalf("unexpected preview imported strategy: %+v", preview)
	}

	committed, err := service.CommitImport(context.Background(), 2001, preview.ID)
	if err != nil {
		t.Fatalf("CommitImport() error = %v", err)
	}

	if committed.ID == 0 || committed.Slug != "awd-bank-portal-01" {
		t.Fatalf("unexpected committed template: %+v", committed)
	}
	if committed.Status != "published" {
		t.Fatalf("expected published imported template, got %+v", committed)
	}
	if committed.RuntimeConfig["image_id"] == nil {
		t.Fatalf("expected runtime_config.image_id in committed template, got %+v", committed.RuntimeConfig)
	}

	stored, err := repo.FindAWDServiceTemplateByIDWithContext(context.Background(), committed.ID)
	if err != nil {
		t.Fatalf("FindAWDServiceTemplateByIDWithContext() error = %v", err)
	}
	if string(stored.Status) != "published" {
		t.Fatalf("unexpected stored status: %+v", stored)
	}

	var accessConfig map[string]any
	if err := json.Unmarshal([]byte(stored.AccessConfig), &accessConfig); err != nil {
		t.Fatalf("unmarshal access_config: %v", err)
	}
	if accessConfig["service_port"] != float64(8080) {
		t.Fatalf("unexpected stored access_config: %+v", accessConfig)
	}

	var runtimeConfig map[string]any
	if err := json.Unmarshal([]byte(stored.RuntimeConfig), &runtimeConfig); err != nil {
		t.Fatalf("unmarshal runtime_config: %v", err)
	}
	if runtimeConfig["image_ref"] != "registry.example.edu/ctf/awd-bank-portal:v1" {
		t.Fatalf("unexpected stored runtime_config: %+v", runtimeConfig)
	}
}

func buildAWDServiceTemplateImportArchive(t *testing.T) []byte {
	t.Helper()

	files := map[string]string{
		"awd-bank-portal-01/challenge.yml": `api_version: v1
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
`,
		"awd-bank-portal-01/statement.md": "银行门户存在越权修改 flag 的逻辑。",
	}

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range files {
		fileWriter, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create(%s) error = %v", name, err)
		}
		if _, err := io.WriteString(fileWriter, content); err != nil {
			t.Fatalf("WriteString(%s) error = %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	return buffer.Bytes()
}
