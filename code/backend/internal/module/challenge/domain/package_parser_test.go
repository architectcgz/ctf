package domain

import (
	"os"
	"path/filepath"
	"testing"

	"ctf-platform/internal/model"
)

func TestParseChallengePackageDirNormalizesUnknownDifficultyToEasy(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(rootDir, "statement.md"), []byte("demo statement"), 0o644); err != nil {
		t.Fatalf("write statement.md: %v", err)
	}
	manifest := `api_version: v1
kind: challenge
meta:
  slug: demo
  title: Demo
  category: web
  difficulty: hell
  points: 100
content:
  statement: statement.md
flag:
  type: static
  value: flag{demo}
`
	if err := os.WriteFile(filepath.Join(rootDir, "challenge.yml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write challenge.yml: %v", err)
	}

	parsed, err := ParseChallengePackageDir(rootDir)
	if err != nil {
		t.Fatalf("ParseChallengePackageDir() error = %v", err)
	}
	if parsed.Difficulty != model.ChallengeDifficultyEasy {
		t.Fatalf("unexpected difficulty: got %q want %q", parsed.Difficulty, model.ChallengeDifficultyEasy)
	}
}

func TestParseChallengePackageDirRejectsSharedProofFlagType(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(rootDir, "statement.md"), []byte("demo statement"), 0o644); err != nil {
		t.Fatalf("write statement.md: %v", err)
	}
	manifest := `api_version: v1
kind: challenge
meta:
  slug: demo-shared-proof
  title: Demo Shared Proof
  category: crypto
  difficulty: easy
  points: 100
content:
  statement: statement.md
flag:
  type: shared_proof
`
	if err := os.WriteFile(filepath.Join(rootDir, "challenge.yml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write challenge.yml: %v", err)
	}

	_, err := ParseChallengePackageDir(rootDir)
	if err == nil {
		t.Fatal("expected ParseChallengePackageDir() to reject shared_proof flag type")
	}
}
