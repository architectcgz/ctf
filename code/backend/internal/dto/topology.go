package dto

import "time"

type TopologyResourcesReq struct {
	CPUQuota  float64 `json:"cpu_quota" binding:"omitempty,gt=0,lte=16"`
	MemoryMB  int64   `json:"memory_mb" binding:"omitempty,min=64,max=16384"`
	PidsLimit int64   `json:"pids_limit" binding:"omitempty,min=1,max=10000"`
}

type TopologyNetworkReq struct {
	Key      string `json:"key" binding:"required,max=64"`
	Name     string `json:"name" binding:"required,max=128"`
	CIDR     string `json:"cidr" binding:"omitempty,max=64"`
	Internal bool   `json:"internal"`
}

type TopologyNodeReq struct {
	Key         string                `json:"key" binding:"required,max=64"`
	Name        string                `json:"name" binding:"required,max=128"`
	ImageID     int64                 `json:"image_id" binding:"omitempty,min=1"`
	ServicePort int                   `json:"service_port" binding:"omitempty,min=1,max=65535"`
	InjectFlag  bool                  `json:"inject_flag"`
	Tier        string                `json:"tier" binding:"omitempty,oneof=public service internal"`
	NetworkKeys []string              `json:"network_keys" binding:"omitempty,min=1,dive,required,max=64"`
	Env         map[string]string     `json:"env"`
	Resources   *TopologyResourcesReq `json:"resources"`
}

type TopologyLinkReq struct {
	FromNodeKey string `json:"from_node_key" binding:"required,max=64"`
	ToNodeKey   string `json:"to_node_key" binding:"required,max=64"`
}

type TopologyTrafficPolicyReq struct {
	SourceNodeKey string `json:"source_node_key" binding:"required,max=64"`
	TargetNodeKey string `json:"target_node_key" binding:"required,max=64"`
	Action        string `json:"action" binding:"required,oneof=allow deny"`
	Protocol      string `json:"protocol" binding:"omitempty,oneof=tcp udp any"`
	Ports         []int  `json:"ports" binding:"omitempty,dive,min=1,max=65535"`
}

type SaveChallengeTopologyReq struct {
	TemplateID   *int64                     `json:"template_id" binding:"omitempty,min=1"`
	EntryNodeKey string                     `json:"entry_node_key" binding:"omitempty,max=64"`
	Networks     []TopologyNetworkReq       `json:"networks" binding:"omitempty,dive"`
	Nodes        []TopologyNodeReq          `json:"nodes" binding:"omitempty,dive"`
	Links        []TopologyLinkReq          `json:"links" binding:"omitempty,dive"`
	Policies     []TopologyTrafficPolicyReq `json:"policies" binding:"omitempty,dive"`
}

type TopologyResourcesResp struct {
	CPUQuota  float64 `json:"cpu_quota,omitempty"`
	MemoryMB  int64   `json:"memory_mb,omitempty"`
	PidsLimit int64   `json:"pids_limit,omitempty"`
}

type TopologyNetworkResp struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	CIDR     string `json:"cidr,omitempty"`
	Internal bool   `json:"internal,omitempty"`
}

type TopologyNodeResp struct {
	Key         string                 `json:"key"`
	Name        string                 `json:"name"`
	ImageID     int64                  `json:"image_id,omitempty"`
	ServicePort int                    `json:"service_port,omitempty"`
	InjectFlag  bool                   `json:"inject_flag,omitempty"`
	Tier        string                 `json:"tier,omitempty"`
	NetworkKeys []string               `json:"network_keys,omitempty"`
	Env         map[string]string      `json:"env,omitempty"`
	Resources   *TopologyResourcesResp `json:"resources,omitempty"`
}

type TopologyLinkResp struct {
	FromNodeKey string `json:"from_node_key"`
	ToNodeKey   string `json:"to_node_key"`
}

type TopologyTrafficPolicyResp struct {
	SourceNodeKey string `json:"source_node_key"`
	TargetNodeKey string `json:"target_node_key"`
	Action        string `json:"action"`
	Protocol      string `json:"protocol,omitempty"`
	Ports         []int  `json:"ports,omitempty"`
}

type ChallengeTopologyResp struct {
	ID           int64                       `json:"id"`
	ChallengeID  int64                       `json:"challenge_id"`
	TemplateID   *int64                      `json:"template_id,omitempty"`
	EntryNodeKey string                      `json:"entry_node_key"`
	Networks     []TopologyNetworkResp       `json:"networks,omitempty"`
	Nodes        []TopologyNodeResp          `json:"nodes"`
	Links        []TopologyLinkResp          `json:"links,omitempty"`
	Policies     []TopologyTrafficPolicyResp `json:"policies,omitempty"`
	CreatedAt    time.Time                   `json:"created_at"`
	UpdatedAt    time.Time                   `json:"updated_at"`
}

type UpsertEnvironmentTemplateReq struct {
	Name         string                     `json:"name" binding:"required,max=128"`
	Description  string                     `json:"description" binding:"omitempty,max=2000"`
	EntryNodeKey string                     `json:"entry_node_key" binding:"required,max=64"`
	Networks     []TopologyNetworkReq       `json:"networks" binding:"omitempty,dive"`
	Nodes        []TopologyNodeReq          `json:"nodes" binding:"required,min=1,dive"`
	Links        []TopologyLinkReq          `json:"links" binding:"omitempty,dive"`
	Policies     []TopologyTrafficPolicyReq `json:"policies" binding:"omitempty,dive"`
}

type EnvironmentTemplateResp struct {
	ID           int64                       `json:"id"`
	Name         string                      `json:"name"`
	Description  string                      `json:"description"`
	EntryNodeKey string                      `json:"entry_node_key"`
	Networks     []TopologyNetworkResp       `json:"networks,omitempty"`
	Nodes        []TopologyNodeResp          `json:"nodes"`
	Links        []TopologyLinkResp          `json:"links,omitempty"`
	Policies     []TopologyTrafficPolicyResp `json:"policies,omitempty"`
	UsageCount   int                         `json:"usage_count"`
	CreatedAt    time.Time                   `json:"created_at"`
	UpdatedAt    time.Time                   `json:"updated_at"`
}
