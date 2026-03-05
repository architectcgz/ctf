package container

import (
	"context"
	"time"

	"ctf-platform/internal/model"
)

// Example 使用示例（仅供参考，实际使用时删除）
func Example() error {
	// 1. 初始化 Docker 引擎（Unix Socket）
	engine, err := NewEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	// 2. 健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := engine.HealthCheck(ctx); err != nil {
		return err
	}

	// 3. 创建容器配置
	config := &model.ContainerConfig{
		Image: "nginx:alpine",
		Env:   []string{"FLAG=flag{test}"},
		Ports: map[string]string{
			"80": "8080",
		},
		Resources: &model.ResourceLimits{
			CPUQuota:  50000,  // 0.5 核
			Memory:    256 * 1024 * 1024, // 256MB
			PidsLimit: 100,
		},
		Network: "bridge",
	}

	// 4. 创建并启动容器
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	containerID, err := engine.CreateContainer(ctx, config)
	if err != nil {
		return err
	}

	if err := engine.StartContainer(ctx, containerID); err != nil {
		return err
	}

	// 5. 获取容器状态
	status, err := engine.GetContainerStatus(ctx, containerID)
	if err != nil {
		return err
	}
	_ = status

	// 6. 停止并删除容器
	if err := engine.StopContainer(ctx, containerID, 10*time.Second); err != nil {
		return err
	}

	return engine.RemoveContainer(ctx, containerID)
}
