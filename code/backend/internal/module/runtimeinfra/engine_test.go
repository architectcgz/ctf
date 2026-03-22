package runtimeinfra

import (
	"reflect"
	"testing"

	"github.com/docker/go-connections/nat"

	"ctf-platform/internal/config"
)

func TestBuildSecurityOpts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		seccomp string
		want    []string
	}{
		{
			name:    "empty uses docker default seccomp",
			seccomp: "",
			want:    []string{"no-new-privileges:true"},
		},
		{
			name:    "default uses docker default seccomp",
			seccomp: "default",
			want:    []string{"no-new-privileges:true"},
		},
		{
			name:    "trimmed default uses docker default seccomp",
			seccomp: " default ",
			want:    []string{"no-new-privileges:true"},
		},
		{
			name:    "unconfined is passed through",
			seccomp: "unconfined",
			want:    []string{"seccomp=unconfined", "no-new-privileges:true"},
		},
		{
			name:    "custom profile is passed through",
			seccomp: "/etc/docker/seccomp/ctf.json",
			want:    []string{"seccomp=/etc/docker/seccomp/ctf.json", "no-new-privileges:true"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := buildSecurityOpts(tt.seccomp)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("buildSecurityOpts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultSecurityConfigUsesNormalizedSecurityOpts(t *testing.T) {
	t.Parallel()

	cfg := &config.ContainerConfig{
		ReadonlyRootfs:      true,
		AllowedCapabilities: []string{"CHOWN"},
		RunAsUser:           "1000:1000",
		Seccomp:             "default",
	}

	got := DefaultSecurityConfig(cfg)
	want := []string{"no-new-privileges:true"}
	if !reflect.DeepEqual(got.SecurityOpt, want) {
		t.Fatalf("DefaultSecurityConfig().SecurityOpt = %v, want %v", got.SecurityOpt, want)
	}
}

func TestSelectServicePort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		exposedPorts  nat.PortSet
		preferredPort int
		want          int
	}{
		{
			name:          "preferred port wins when exposed",
			exposedPorts:  nat.PortSet{"8080/tcp": struct{}{}, "80/tcp": struct{}{}},
			preferredPort: 8080,
			want:          8080,
		},
		{
			name:          "single exposed port is used",
			exposedPorts:  nat.PortSet{"80/tcp": struct{}{}},
			preferredPort: 8080,
			want:          80,
		},
		{
			name:          "web port preferred over arbitrary lowest port",
			exposedPorts:  nat.PortSet{"80/tcp": struct{}{}, "3306/tcp": struct{}{}},
			preferredPort: 8080,
			want:          80,
		},
		{
			name:          "preferred port returned when image exposes nothing",
			exposedPorts:  nat.PortSet{},
			preferredPort: 8080,
			want:          8080,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := selectServicePort(tt.exposedPorts, tt.preferredPort)
			if got != tt.want {
				t.Fatalf("selectServicePort() = %d, want %d", got, tt.want)
			}
		})
	}
}
