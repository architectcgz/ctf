package contracts

import "strings"

const (
	ProvisioningStageQueued            = "queued"
	ProvisioningStageSelectingNode     = "selecting_node"
	ProvisioningStageAllocatingPort    = "allocating_port"
	ProvisioningStageAllocatingNetwork = "allocating_network"
	ProvisioningStageCreatingNetwork   = "creating_network"
	ProvisioningStageCreatingContainer = "creating_container"
	ProvisioningStageStartingContainer = "starting_container"
	ProvisioningStageProbingReadiness  = "probing_readiness"
	ProvisioningStageCleaningPrevious  = "cleaning_previous"
	ProvisioningStageRescheduling      = "rescheduling"
	ProvisioningStageFailed            = "failed"
)

type ProvisioningProgress struct {
	InstanceID            int64
	Attempt               int
	Stage                 string
	Message               string
	Severity              string
	RuntimeNodeID         *int64
	Detail                string
	LastProvisioningError string
}

func ProvisioningStageMessage(stage string) string {
	switch strings.TrimSpace(stage) {
	case ProvisioningStageQueued:
		return "排队中"
	case ProvisioningStageSelectingNode:
		return "正在选择运行节点"
	case ProvisioningStageAllocatingPort:
		return "正在分配访问端口"
	case ProvisioningStageAllocatingNetwork:
		return "正在分配隔离网络"
	case ProvisioningStageCreatingNetwork:
		return "正在创建隔离网络"
	case ProvisioningStageCreatingContainer:
		return "正在创建靶机容器"
	case ProvisioningStageStartingContainer:
		return "正在启动靶机"
	case ProvisioningStageProbingReadiness:
		return "正在检查服务可用性"
	case ProvisioningStageCleaningPrevious:
		return "正在清理上一次启动残留"
	case ProvisioningStageRescheduling:
		return "正在重新调度"
	case ProvisioningStageFailed:
		return "启动失败"
	default:
		return ""
	}
}

func ResolveProvisioningMessage(stage, message string) string {
	if trimmed := strings.TrimSpace(message); trimmed != "" {
		return trimmed
	}
	return ProvisioningStageMessage(stage)
}
