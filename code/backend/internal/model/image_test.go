package model

import "testing"

func TestBuildRuntimeImageRefPrefersDigest(t *testing.T) {
	image := &Image{
		Name:   "registry.example.edu/ctf/awd/demo",
		Tag:    "latest",
		Digest: "sha256:demo",
	}

	if got := BuildRuntimeImageRef(image); got != "registry.example.edu/ctf/awd/demo@sha256:demo" {
		t.Fatalf("BuildRuntimeImageRef() = %q", got)
	}
}

func TestBuildRuntimeImageRefFallsBackToTag(t *testing.T) {
	image := &Image{
		Name: "ctf/web-demo",
		Tag:  "v1",
	}

	if got := BuildRuntimeImageRef(image); got != "ctf/web-demo:v1" {
		t.Fatalf("BuildRuntimeImageRef() = %q", got)
	}
}
