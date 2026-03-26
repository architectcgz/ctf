package application

import runtimeports "ctf-platform/internal/module/runtime/ports"

// CountRunningRepository 定义运行中实例统计仓储能力。
type CountRunningRepository = runtimeports.CountRunningRepository

// InstanceRepository 定义实例 HTTP 用例所需的仓储能力。
type InstanceRepository = runtimeports.InstanceRepository

// RuntimeCleaner 定义实例销毁时的运行时资源清理能力。
type RuntimeCleaner = runtimeports.RuntimeCleaner

// ManagedContainerStat 表示 runtime application 层使用的受管容器运行指标快照。
type ManagedContainerStat = runtimeports.ManagedContainerStat

// ManagedContainer 表示 runtime application 层使用的受管容器元数据。
type ManagedContainer = runtimeports.ManagedContainer

// TopologyCreateNode 定义运行时拓扑节点契约。
type TopologyCreateNode = runtimeports.TopologyCreateNode

// TopologyCreateNetwork 定义运行时拓扑网络契约。
type TopologyCreateNetwork = runtimeports.TopologyCreateNetwork

// TopologyCreateRequest 定义运行时拓扑创建请求。
type TopologyCreateRequest = runtimeports.TopologyCreateRequest

// TopologyCreateResult 定义运行时拓扑创建结果。
type TopologyCreateResult = runtimeports.TopologyCreateResult

// TeacherInstanceFilter 定义教师端实例列表筛选条件。
type TeacherInstanceFilter = runtimeports.TeacherInstanceFilter

// UserVisibleInstanceRow 表示用户可见实例列表行模型。
type UserVisibleInstanceRow = runtimeports.UserVisibleInstanceRow

// TeacherInstanceRow 表示教师端实例列表行模型。
type TeacherInstanceRow = runtimeports.TeacherInstanceRow
