package container

import (
	"context"
	"fmt"
	"testing"

	"ctf-platform/internal/model"
)

// TestNetworkManagement 测试网络管理功能
func TestNetworkManagement(t *testing.T) {
	engine, err := NewEngine("")
	if err != nil {
		t.Fatalf("创建引擎失败: %v", err)
	}
	defer engine.Close()

	ctx := context.Background()

	// 测试创建网络
	networkName := "ctf-test-1-1"
	networkID, err := engine.CreateNetwork(ctx, networkName)
	if err != nil {
		t.Fatalf("创建网络失败: %v", err)
	}
	t.Logf("网络创建成功: %s", networkID)

	// 测试删除网络
	err = engine.RemoveNetwork(ctx, networkID)
	if err != nil {
		t.Fatalf("删除网络失败: %v", err)
	}
	t.Logf("网络删除成功")
}

// TestPortAllocation 测试端口分配
func TestPortAllocation(t *testing.T) {
	pm := NewPortManager(30000, 30100)

	// 分配端口
	port1, err := pm.AllocatePort()
	if err != nil {
		t.Fatalf("分配端口失败: %v", err)
	}
	t.Logf("分配端口: %d", port1)

	// 分配第二个端口
	port2, err := pm.AllocatePort()
	if err != nil {
		t.Fatalf("分配第二个端口失败: %v", err)
	}
	t.Logf("分配第二个端口: %d", port2)

	if port1 == port2 {
		t.Fatalf("端口冲突: %d == %d", port1, port2)
	}

	// 释放端口
	pm.ReleasePort(port1)
	t.Logf("释放端口: %d", port1)
}

// TestContainerWithNetwork 测试容器网络隔离
func TestContainerWithNetwork(t *testing.T) {
	engine, err := NewEngine("")
	if err != nil {
		t.Fatalf("创建引擎失败: %v", err)
	}
	defer engine.Close()

	ctx := context.Background()
	pm := NewPortManager(30000, 30100)

	// 创建网络
	networkName := fmt.Sprintf("ctf-test-%d", 1)
	networkID, err := engine.CreateNetwork(ctx, networkName)
	if err != nil {
		t.Fatalf("创建网络失败: %v", err)
	}
	defer engine.RemoveNetwork(ctx, networkID)

	// 分配端口
	hostPort, err := pm.AllocatePort()
	if err != nil {
		t.Fatalf("分配端口失败: %v", err)
	}
	defer pm.ReleasePort(hostPort)

	// 创建容器配置
	cfg := &model.ContainerConfig{
		Image: "alpine:latest",
		Env:   []string{"FLAG=flag{test}"},
		Ports: map[string]string{
			"80": fmt.Sprintf("%d", hostPort),
		},
		Network: networkID,
		Resources: &model.ResourceLimits{
			CPUQuota:  50000,
			Memory:    256 * 1024 * 1024,
			PidsLimit: 100,
		},
		Security: DefaultSecurityConfig(),
	}

	// 创建容器
	containerID, err := engine.CreateContainer(ctx, cfg)
	if err != nil {
		t.Fatalf("创建容器失败: %v", err)
	}
	defer engine.RemoveContainer(ctx, containerID)

	t.Logf("容器创建成功: %s, 网络: %s, 端口映射: 80->%d", containerID[:12], networkID[:12], hostPort)
}
