package runtime_test

import (
	"context"
	"testing"

	runtimecmd "ctf-platform/internal/module/container_runtime/application/commands"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
)

func TestServiceCleanupRuntimeRemovesACLByHandle(t *testing.T) {
	t.Parallel()

	engine := &fakeRuntimeEngine{}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, nil, nil)

	details, err := runtimecontracts.EncodeInstanceRuntimeDetails(runtimecontracts.InstanceRuntimeDetails{
		ACL: &runtimecontracts.InstanceRuntimeACLHandle{Chain: "CTF-INS-123"},
		ACLRules: []runtimecontracts.InstanceRuntimeACLRule{
			{SourceIP: "172.30.0.2", TargetIP: "172.30.0.3", Action: "allow", Protocol: "tcp", Ports: []int{3306}},
		},
	})
	if err != nil {
		t.Fatalf("encode runtime details: %v", err)
	}

	if err := cleanupService.CleanupRuntime(context.Background(), runtimeCleanupTarget(&instanceentity.Instance{RuntimeDetails: details})); err != nil {
		t.Fatalf("CleanupRuntime() error = %v", err)
	}
	if engine.removedACLHandle == nil || engine.removedACLHandle.Chain != "CTF-INS-123" {
		t.Fatalf("expected acl removed by handle, got handle=%+v rules=%v", engine.removedACLHandle, engine.removedACLRules)
	}
	if len(engine.removedACLRules) != 0 {
		t.Fatalf("expected no legacy acl rule removal, got %v", engine.removedACLRules)
	}
}

func TestServiceCleanupRuntimeSkipsLegacyACLRulesWithoutHandle(t *testing.T) {
	t.Parallel()

	engine := &fakeRuntimeEngine{}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, nil, nil)

	details, err := runtimecontracts.EncodeInstanceRuntimeDetails(runtimecontracts.InstanceRuntimeDetails{
		ACLRules: []runtimecontracts.InstanceRuntimeACLRule{
			{SourceIP: "172.30.0.2", TargetIP: "172.30.0.3", Action: "allow", Protocol: "tcp", Ports: []int{3306}},
		},
	})
	if err != nil {
		t.Fatalf("encode runtime details: %v", err)
	}

	if err := cleanupService.CleanupRuntime(context.Background(), runtimeCleanupTarget(&instanceentity.Instance{RuntimeDetails: details})); err != nil {
		t.Fatalf("CleanupRuntime() error = %v", err)
	}
	if engine.removedACLHandle != nil {
		t.Fatalf("expected no acl handle removal, got handle=%+v", engine.removedACLHandle)
	}
	if len(engine.removedACLRules) != 0 {
		t.Fatalf("expected no legacy acl rule removal, got %v", engine.removedACLRules)
	}
}

func TestServiceCleanupRuntimeACLHandleUnaffectedByPollutedRules(t *testing.T) {
	t.Parallel()

	engine := &fakeRuntimeEngine{}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, nil, nil)

	details, err := runtimecontracts.EncodeInstanceRuntimeDetails(runtimecontracts.InstanceRuntimeDetails{
		ACL: &runtimecontracts.InstanceRuntimeACLHandle{Chain: "CTF-INS-456"},
		// 脏数据：含有不合法的 action，不应影响按 handle 清理。
		ACLRules: []runtimecontracts.InstanceRuntimeACLRule{
			{SourceIP: "172.30.0.2", TargetIP: "172.30.0.3", Action: "DROP", Protocol: "icmp"},
		},
	})
	if err != nil {
		t.Fatalf("encode runtime details: %v", err)
	}

	if err := cleanupService.CleanupRuntime(context.Background(), runtimeCleanupTarget(&instanceentity.Instance{RuntimeDetails: details})); err != nil {
		t.Fatalf("CleanupRuntime() error = %v", err)
	}
	if engine.removedACLHandle == nil || engine.removedACLHandle.Chain != "CTF-INS-456" {
		t.Fatalf("expected acl removed by handle despite polluted rules, got handle=%+v rules=%v", engine.removedACLHandle, engine.removedACLRules)
	}
}
