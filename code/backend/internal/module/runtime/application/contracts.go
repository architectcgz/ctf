package application

import runtimeports "ctf-platform/internal/module/runtime/ports"

// CountRunningRepository 定义运行中实例统计仓储能力。
type CountRunningRepository = runtimeports.CountRunningRepository

// InstanceLookupRepository 定义实例按 ID 查询能力。
type InstanceLookupRepository = runtimeports.InstanceLookupRepository

// InstanceUserLookupRepository 定义实例相关用户查询能力。
type InstanceUserLookupRepository = runtimeports.InstanceUserLookupRepository

// InstanceAccessRepository 定义实例访问校验查询能力。
type InstanceAccessRepository = runtimeports.InstanceAccessRepository

// UserVisibleInstanceRepository 定义用户可见实例列表查询能力。
type UserVisibleInstanceRepository = runtimeports.UserVisibleInstanceRepository

// TeacherInstanceQueryRepository 定义教师端实例列表查询能力。
type TeacherInstanceQueryRepository = runtimeports.TeacherInstanceQueryRepository

// InstanceExtendRepository 定义实例续期能力。
type InstanceExtendRepository = runtimeports.InstanceExtendRepository

// InstanceStatusRepository 定义实例状态更新能力。
type InstanceStatusRepository = runtimeports.InstanceStatusRepository

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
