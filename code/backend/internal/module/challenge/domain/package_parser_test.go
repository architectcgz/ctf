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

func TestParseChallengePackageDirLoadsTopologyAndPackageFiles(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(rootDir, "docker"), 0o755); err != nil {
		t.Fatalf("mkdir docker: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "statement.md"), []byte("demo statement"), 0o644); err != nil {
		t.Fatalf("write statement.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "docker", "Dockerfile"), []byte("FROM nginx:1.27"), 0o644); err != nil {
		t.Fatalf("write Dockerfile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "docker", "app.py"), []byte("print('demo')"), 0o644); err != nil {
		t.Fatalf("write app.py: %v", err)
	}
	topology := `api_version: v1
kind: topology
entry_node_key: web
networks:
  - key: public
    name: Public
nodes:
  - key: web
    name: Web
    tier: public
    image:
      ref: ctf/demo-topology:web
      dockerfile: docker/Dockerfile
      context: .
    service_port: 8080
    inject_flag: true
    network_keys: [public]
`
	if err := os.WriteFile(filepath.Join(rootDir, "docker", "topology.yml"), []byte(topology), 0o644); err != nil {
		t.Fatalf("write topology.yml: %v", err)
	}
	manifest := `api_version: v1
kind: challenge
meta:
  slug: demo-topology
  title: Demo Topology
  category: web
  difficulty: medium
  points: 300
content:
  statement: statement.md
flag:
  type: dynamic
  prefix: flag
runtime:
  type: container
  image:
    ref: ctf/demo-topology:web
extensions:
  topology:
    enabled: true
    source: docker/topology.yml
`
	if err := os.WriteFile(filepath.Join(rootDir, "challenge.yml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write challenge.yml: %v", err)
	}

	parsed, err := ParseChallengePackageDir(rootDir)
	if err != nil {
		t.Fatalf("ParseChallengePackageDir() error = %v", err)
	}
	if parsed.Topology == nil {
		t.Fatal("expected parsed topology")
	}
	if parsed.Topology.EntryNodeKey != "web" {
		t.Fatalf("unexpected entry node key: %q", parsed.Topology.EntryNodeKey)
	}
	if len(parsed.Topology.Nodes) != 1 {
		t.Fatalf("expected 1 topology node, got %d", len(parsed.Topology.Nodes))
	}
	if parsed.Topology.Nodes[0].Image.Ref != "ctf/demo-topology:web" {
		t.Fatalf("unexpected node image ref: %q", parsed.Topology.Nodes[0].Image.Ref)
	}
	if len(parsed.PackageFiles) == 0 {
		t.Fatal("expected package file tree")
	}
	foundTopology := false
	foundDockerfile := false
	for _, item := range parsed.PackageFiles {
		if item.Path == "docker/topology.yml" {
			foundTopology = true
		}
		if item.Path == "docker/Dockerfile" {
			foundDockerfile = true
		}
	}
	if !foundTopology || !foundDockerfile {
		t.Fatalf("expected package file tree to include topology and dockerfile, got %+v", parsed.PackageFiles)
	}
}

func TestParseChallengePackageDirParsesRuntimeServiceTCP(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(rootDir, "statement.md"), []byte("tcp statement"), 0o644); err != nil {
		t.Fatalf("write statement.md: %v", err)
	}
	manifest := `api_version: v1
kind: challenge
meta:
  slug: pwn-tcp-demo
  title: Pwn TCP Demo
  category: pwn
  difficulty: beginner
  points: 100
content:
  statement: statement.md
flag:
  type: static
  value: flag{tcp}
runtime:
  type: container
  image:
    ref: 127.0.0.1:5000/ctf/pwn-tcp-demo:v1
  service:
    protocol: tcp
    port: 31337
`
	if err := os.WriteFile(filepath.Join(rootDir, "challenge.yml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write challenge.yml: %v", err)
	}

	parsed, err := ParseChallengePackageDir(rootDir)
	if err != nil {
		t.Fatalf("ParseChallengePackageDir() error = %v", err)
	}
	if parsed.RuntimeProtocol != model.ChallengeTargetProtocolTCP {
		t.Fatalf("unexpected runtime protocol: got %q", parsed.RuntimeProtocol)
	}
	if parsed.RuntimePort != 31337 {
		t.Fatalf("unexpected runtime port: got %d", parsed.RuntimePort)
	}
}
