package application

import (
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

// ManagedContainerStat 表示 runtime application 层使用的受管容器运行指标快照。
type ManagedContainerStat = runtimecontracts.ManagedContainerStat

// ManagedContainerStatsReader 定义受管容器指标读取能力。
type ManagedContainerStatsReader = runtimeports.ManagedContainerStatsReader

// ManagedContainer 表示 runtime application 层使用的受管容器元数据。
type ManagedContainer = runtimecontracts.ManagedContainer

// ContainerFileWriter 定义容器文件写入能力。
type ContainerFileWriter = runtimeports.ContainerFileWriter

// ContainerImageRuntime 定义镜像检查与删除能力。
type ContainerImageRuntime = runtimeports.ContainerImageRuntime

// TopologyCreateNode 定义运行时拓扑节点契约。
type TopologyCreateNode = runtimecontracts.TopologyCreateNode

// TopologyCreateNetwork 定义运行时拓扑网络契约。
type TopologyCreateNetwork = runtimecontracts.TopologyCreateNetwork

// TopologyCreateRequest 定义运行时拓扑创建请求。
type TopologyCreateRequest = runtimecontracts.TopologyCreateRequest

// TopologyCreateResult 定义运行时拓扑创建结果。
type TopologyCreateResult = runtimecontracts.TopologyCreateResult
