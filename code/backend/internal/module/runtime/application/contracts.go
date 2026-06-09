package application

import runtimeports "ctf-platform/internal/module/runtime/ports"

// CountRunningRepository 定义运行中实例统计仓储能力。
type CountRunningRepository = runtimeports.CountRunningRepository

// ManagedContainerStat 表示 runtime application 层使用的受管容器运行指标快照。
type ManagedContainerStat = runtimeports.ManagedContainerStat

// ManagedContainerStatsReader 定义受管容器指标读取能力。
type ManagedContainerStatsReader = runtimeports.ManagedContainerStatsReader

// ManagedContainer 表示 runtime application 层使用的受管容器元数据。
type ManagedContainer = runtimeports.ManagedContainer

// ContainerFileWriter 定义容器文件写入能力。
type ContainerFileWriter = runtimeports.ContainerFileWriter

// ContainerImageRuntime 定义镜像检查与删除能力。
type ContainerImageRuntime = runtimeports.ContainerImageRuntime

// TopologyCreateNode 定义运行时拓扑节点契约。
type TopologyCreateNode = runtimeports.TopologyCreateNode

// TopologyCreateNetwork 定义运行时拓扑网络契约。
type TopologyCreateNetwork = runtimeports.TopologyCreateNetwork

// TopologyCreateRequest 定义运行时拓扑创建请求。
type TopologyCreateRequest = runtimeports.TopologyCreateRequest

// TopologyCreateResult 定义运行时拓扑创建结果。
type TopologyCreateResult = runtimeports.TopologyCreateResult
