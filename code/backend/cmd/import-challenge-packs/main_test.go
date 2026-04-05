package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

func TestImportOnePackSupportsChallengeYAML(t *testing.T) {
	db := setupImportTestDB(t)

	packDir := t.TempDir()
	mustWriteFile(t, filepath.Join(packDir, "challenge.yml"), []byte(`
api_version: v1
kind: challenge
meta:
  slug: web-sqli-101
  title: SQL Injection 101
  category: web
  difficulty: easy
  points: 100
content:
  statement: statement.md
  attachments:
    - path: attachments/web-sqli-101.zip
      name: web-sqli-101.zip
flag:
  type: static
  value: flag{sqli_101}
  prefix: flag
hints:
  - level: 1
    title: Hint 1
    content: 从登录参数开始看
runtime:
  type: container
  image:
    ref: ctf/web-sqli-101:latest
extensions:
  topology:
    source: docker/topology.yml
    enabled: false
`))
	mustWriteFile(t, filepath.Join(packDir, "statement.md"), []byte("# SQLi 101\n\nFind the bypass."))
	mustWriteFile(t, filepath.Join(packDir, "attachments", "web-sqli-101.zip"), []byte("zip-bytes"))

	created, published, err := importOnePack(db, packDir, false, false)
	if err != nil {
		t.Fatalf("importOnePack() error = %v", err)
	}
	if !created {
		t.Fatalf("expected challenge to be created")
	}
	if published {
		t.Fatalf("expected challenge to remain draft when publish=false")
	}

	var challenge model.Challenge
	if err := db.Where("title = ?", "SQL Injection 101").First(&challenge).Error; err != nil {
		t.Fatalf("find imported challenge: %v", err)
	}
	if challenge.Description != "# SQLi 101\n\nFind the bypass." {
		t.Fatalf("unexpected description: %q", challenge.Description)
	}
	if challenge.Category != "web" {
		t.Fatalf("unexpected category: %s", challenge.Category)
	}
	if challenge.AttachmentURL == "" {
		t.Fatal("expected attachment URL to be populated")
	}
	if challenge.ImageID == 0 {
		t.Fatal("expected runtime image to be resolved")
	}
	if challenge.FlagType != model.FlagTypeStatic {
		t.Fatalf("unexpected flag type: %s", challenge.FlagType)
	}

	var hints []model.ChallengeHint
	if err := db.Where("challenge_id = ?", challenge.ID).Order("level ASC").Find(&hints).Error; err != nil {
		t.Fatalf("list hints: %v", err)
	}
	if len(hints) != 1 {
		t.Fatalf("expected 1 hint, got %d", len(hints))
	}
	if hints[0].Content != "从登录参数开始看" {
		t.Fatalf("unexpected hint content: %q", hints[0].Content)
	}
}

func TestImportOnePackUpsertsByPackageSlug(t *testing.T) {
	db := setupImportTestDB(t)

	packDir := t.TempDir()
	writeChallengePackFixture(t, packDir, "web-sqli-101", "SQL Injection 101", 100)

	created, published, err := importOnePack(db, packDir, false, false)
	if err != nil {
		t.Fatalf("first importOnePack() error = %v", err)
	}
	if !created {
		t.Fatalf("expected first import to create challenge")
	}
	if published {
		t.Fatalf("expected draft challenge on first import")
	}

	writeChallengePackFixture(t, packDir, "web-sqli-101", "SQL Injection 102", 200)
	created, published, err = importOnePack(db, packDir, false, false)
	if err != nil {
		t.Fatalf("second importOnePack() error = %v", err)
	}
	if created {
		t.Fatalf("expected second import to update existing challenge")
	}
	if published {
		t.Fatalf("expected draft challenge on second import")
	}

	var count int64
	if err := db.Model(&model.Challenge{}).Count(&count).Error; err != nil {
		t.Fatalf("count challenges: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 challenge after slug upsert, got %d", count)
	}

	var challenge model.Challenge
	if err := db.First(&challenge).Error; err != nil {
		t.Fatalf("find challenge: %v", err)
	}
	if challenge.Title != "SQL Injection 102" {
		t.Fatalf("expected updated title, got %q", challenge.Title)
	}
	if challenge.Points != 200 {
		t.Fatalf("expected updated points, got %d", challenge.Points)
	}

	var packageSlug string
	if err := db.Raw("SELECT package_slug FROM challenges WHERE id = ?", challenge.ID).Scan(&packageSlug).Error; err != nil {
		t.Fatalf("query package_slug: %v", err)
	}
	if packageSlug != "web-sqli-101" {
		t.Fatalf("expected persisted package_slug, got %q", packageSlug)
	}
}

func setupImportTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Image{}, &model.Challenge{}, &model.ChallengeHint{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func mustWriteFile(t *testing.T, path string, content []byte) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func writeChallengePackFixture(t *testing.T, packDir, slug, title string, points int) {
	t.Helper()

	mustWriteFile(t, filepath.Join(packDir, "challenge.yml"), []byte(`
api_version: v1
kind: challenge
meta:
  slug: `+slug+`
  title: `+title+`
  category: web
  difficulty: easy
  points: `+itoa(points)+`
content:
  statement: statement.md
  attachments:
    - path: attachments/web-sqli-101.zip
      name: web-sqli-101.zip
flag:
  type: static
  value: flag{sqli_101}
  prefix: flag
hints:
  - level: 1
    title: Hint 1
    content: 从登录参数开始看
runtime:
  type: container
  image:
    ref: ctf/web-sqli-101:latest
extensions:
  topology:
    source: docker/topology.yml
    enabled: false
`))
	mustWriteFile(t, filepath.Join(packDir, "statement.md"), []byte("# SQLi 101\n\nFind the bypass."))
	mustWriteFile(t, filepath.Join(packDir, "attachments", "web-sqli-101.zip"), []byte("zip-bytes"))
}

func itoa(value int) string {
	return fmt.Sprintf("%d", value)
}
