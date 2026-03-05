package container

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
)

// PortManager 端口管理器
type PortManager struct {
	rangeStart int
	rangeEnd   int
	usedPorts  map[int]bool
	mu         sync.Mutex
}

// NewPortManager 创建端口管理器
func NewPortManager(start, end int) *PortManager {
	return &PortManager{
		rangeStart: start,
		rangeEnd:   end,
		usedPorts:  make(map[int]bool),
	}
}

// AllocatePort 分配可用端口
func (pm *PortManager) AllocatePort() (int, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 尝试 100 次随机分配
	for i := 0; i < 100; i++ {
		port := pm.rangeStart + rand.Intn(pm.rangeEnd-pm.rangeStart)

		// 检查是否已被使用
		if pm.usedPorts[port] {
			continue
		}

		// 检查端口是否真实可用
		if pm.isPortAvailable(port) {
			pm.usedPorts[port] = true
			return port, nil
		}
	}

	return 0, fmt.Errorf("无法分配可用端口")
}

// ReleasePort 释放端口
func (pm *PortManager) ReleasePort(port int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.usedPorts, port)
}

// isPortAvailable 检查端口是否可用
func (pm *PortManager) isPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}
