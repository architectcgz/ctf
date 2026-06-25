package domain

import (
	"testing"
	"time"

	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

func TestInstanceRespFromModelIncludesProvisioningProgress(t *testing.T) {
	t.Parallel()

	resp := InstanceRespFromModel(&instancecontracts.Instance{
		ID:                  81001,
		UserID:              7,
		ChallengeID:         41,
		Status:              instancecontracts.InstanceStatusCreating,
		ProvisioningStage:   instancecontracts.ProvisioningStageAllocatingPort,
		ProvisioningAttempt: 3,
		ExpiresAt:           time.Now().UTC().Add(time.Hour),
		CreatedAt:           time.Now().UTC(),
	}, "", "")

	if resp == nil {
		t.Fatal("expected response")
	}
	if resp.ProvisioningStage != instancecontracts.ProvisioningStageAllocatingPort {
		t.Fatalf("provisioning stage = %q, want %q", resp.ProvisioningStage, instancecontracts.ProvisioningStageAllocatingPort)
	}
	if resp.ProvisioningMessage != "正在分配访问端口" {
		t.Fatalf("provisioning message = %q, want allocating port label", resp.ProvisioningMessage)
	}
	if resp.ProvisioningAttempt != 3 {
		t.Fatalf("provisioning attempt = %d, want 3", resp.ProvisioningAttempt)
	}
}
