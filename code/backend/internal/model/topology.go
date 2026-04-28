package model

import (
	"encoding/json"
	"time"
)

const (
	TopologyTierPublic        = "public"
	TopologyTierService       = "service"
	TopologyTierInternal      = "internal"
	TopologyDefaultNetworkKey = "default"

	TopologyPolicyActionAllow = "allow"
	TopologyPolicyActionDeny  = "deny"

	TopologyPolicyProtocolTCP = "tcp"
	TopologyPolicyProtocolUDP = "udp"
	TopologyPolicyProtocolAny = "any"

	WriteupVisibilityPrivate = "private"
	WriteupVisibilityPublic  = "public"
)

type ChallengeTopology struct {
	ID           int64      `gorm:"column:id;primaryKey"`
	ChallengeID  int64      `gorm:"column:challenge_id;uniqueIndex"`
	TemplateID   *int64     `gorm:"column:template_id"`
	EntryNodeKey string     `gorm:"column:entry_node_key"`
	Spec         string     `gorm:"column:spec"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
}

func (ChallengeTopology) TableName() string {
	return "challenge_topologies"
}

type EnvironmentTemplate struct {
	ID           int64      `gorm:"column:id;primaryKey"`
	Name         string     `gorm:"column:name"`
	Description  string     `gorm:"column:description"`
	EntryNodeKey string     `gorm:"column:entry_node_key"`
	Spec         string     `gorm:"column:spec"`
	UsageCount   int        `gorm:"column:usage_count"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
}

func (EnvironmentTemplate) TableName() string {
	return "environment_templates"
}

type TopologySpec struct {
	Networks []TopologyNetwork       `json:"networks,omitempty"`
	Nodes    []TopologyNode          `json:"nodes"`
	Links    []TopologyLink          `json:"links,omitempty"`
	Policies []TopologyTrafficPolicy `json:"policies,omitempty"`
}

type TopologyNetwork struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	CIDR     string `json:"cidr,omitempty"`
	Internal bool   `json:"internal,omitempty"`
}

type TopologyNode struct {
	Key             string             `json:"key"`
	Name            string             `json:"name"`
	ImageID         int64              `json:"image_id,omitempty"`
	ServicePort     int                `json:"service_port,omitempty"`
	ServiceProtocol string             `json:"service_protocol,omitempty"`
	InjectFlag      bool               `json:"inject_flag,omitempty"`
	Tier            string             `json:"tier,omitempty"`
	NetworkKeys     []string           `json:"network_keys,omitempty"`
	Env             map[string]string  `json:"env,omitempty"`
	Resources       *TopologyResources `json:"resources,omitempty"`
}

type TopologyResources struct {
	CPUQuota  float64 `json:"cpu_quota,omitempty"`
	MemoryMB  int64   `json:"memory_mb,omitempty"`
	PidsLimit int64   `json:"pids_limit,omitempty"`
}

type TopologyLink struct {
	FromNodeKey string `json:"from_node_key"`
	ToNodeKey   string `json:"to_node_key"`
}

type TopologyTrafficPolicy struct {
	SourceNodeKey string `json:"source_node_key"`
	TargetNodeKey string `json:"target_node_key"`
	Action        string `json:"action"`
	Protocol      string `json:"protocol,omitempty"`
	Ports         []int  `json:"ports,omitempty"`
}

type ChallengeWriteup struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	ChallengeID   int64      `gorm:"column:challenge_id;uniqueIndex"`
	Title         string     `gorm:"column:title"`
	Content       string     `gorm:"column:content"`
	Visibility    string     `gorm:"column:visibility"`
	CreatedBy     *int64     `gorm:"column:created_by"`
	IsRecommended bool       `gorm:"column:is_recommended;index:idx_challenge_writeups_recommended"`
	RecommendedAt *time.Time `gorm:"column:recommended_at"`
	RecommendedBy *int64     `gorm:"column:recommended_by"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (ChallengeWriteup) TableName() string {
	return "challenge_writeups"
}

type InstanceRuntimeDetails struct {
	Networks   []InstanceRuntimeNetwork   `json:"networks,omitempty"`
	Containers []InstanceRuntimeContainer `json:"containers,omitempty"`
	ACLRules   []InstanceRuntimeACLRule   `json:"acl_rules,omitempty"`
}

type InstanceRuntimeNetwork struct {
	Key       string `json:"key,omitempty"`
	Name      string `json:"name,omitempty"`
	NetworkID string `json:"network_id,omitempty"`
	Internal  bool   `json:"internal,omitempty"`
}

type InstanceRuntimeContainer struct {
	NodeKey         string   `json:"node_key,omitempty"`
	ContainerID     string   `json:"container_id"`
	HostPort        int      `json:"host_port,omitempty"`
	ServicePort     int      `json:"service_port,omitempty"`
	ServiceProtocol string   `json:"service_protocol,omitempty"`
	IsEntryPoint    bool     `json:"is_entry_point,omitempty"`
	NetworkKeys     []string `json:"network_keys,omitempty"`
}

type InstanceRuntimeACLRule struct {
	Comment           string `json:"comment,omitempty"`
	SourceNodeKey     string `json:"source_node_key,omitempty"`
	TargetNodeKey     string `json:"target_node_key,omitempty"`
	SourceContainerID string `json:"source_container_id,omitempty"`
	TargetContainerID string `json:"target_container_id,omitempty"`
	SourceIP          string `json:"source_ip,omitempty"`
	TargetIP          string `json:"target_ip,omitempty"`
	Action            string `json:"action,omitempty"`
	Protocol          string `json:"protocol,omitempty"`
	Ports             []int  `json:"ports,omitempty"`
}

func EncodeTopologySpec(spec TopologySpec) (string, error) {
	raw, err := json.Marshal(spec)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func DecodeTopologySpec(raw string) (TopologySpec, error) {
	if raw == "" {
		return TopologySpec{}, nil
	}
	var spec TopologySpec
	if err := json.Unmarshal([]byte(raw), &spec); err != nil {
		return TopologySpec{}, err
	}
	return spec, nil
}

func EncodeInstanceRuntimeDetails(details InstanceRuntimeDetails) (string, error) {
	raw, err := json.Marshal(details)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func DecodeInstanceRuntimeDetails(raw string) (InstanceRuntimeDetails, error) {
	if raw == "" {
		return InstanceRuntimeDetails{}, nil
	}
	var details InstanceRuntimeDetails
	if err := json.Unmarshal([]byte(raw), &details); err != nil {
		return InstanceRuntimeDetails{}, err
	}
	return details, nil
}
