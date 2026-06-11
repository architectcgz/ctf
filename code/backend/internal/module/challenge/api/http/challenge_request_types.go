package http

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type ChallengeHintReq struct {
	Level   int    `json:"level" binding:"required,min=1"`
	Title   string `json:"title" binding:"omitempty,max=128"`
	Content string `json:"content" binding:"required"`
}

type OptionalImageIDField struct {
	Set   bool
	Value *int64
}

func (f *OptionalImageIDField) UnmarshalJSON(data []byte) error {
	f.Set = true
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		f.Value = nil
		return nil
	}

	var value int64
	if err := json.Unmarshal(trimmed, &value); err != nil {
		return err
	}
	if value <= 0 {
		return fmt.Errorf("image_id must be greater than 0")
	}
	f.Value = &value
	return nil
}

type CreateChallengeReq struct {
	Title           string                `json:"title" binding:"required"`
	Description     string                `json:"description" binding:"required"`
	Category        string                `json:"category" binding:"required"`
	Difficulty      string                `json:"difficulty" binding:"required,oneof=beginner easy medium hard insane"`
	Points          int                   `json:"points" binding:"required,min=1"`
	ImageID         *int64                `json:"image_id" binding:"omitempty,min=1"`
	AttachmentURL   string                `json:"attachment_url" binding:"omitempty,max=2048"`
	InstanceSharing string                `json:"instance_sharing" binding:"omitempty,oneof=per_user per_team shared"`
	Hints           []ChallengeHintReq    `json:"hints" binding:"omitempty,dive"`
}

type UpdateChallengeReq struct {
	Title           string                `json:"title"`
	Description     string                `json:"description"`
	Category        string                `json:"category"`
	Difficulty      string                `json:"difficulty" binding:"omitempty,oneof=beginner easy medium hard insane"`
	Points          int                   `json:"points" binding:"omitempty,min=1"`
	ImageID         OptionalImageIDField  `json:"image_id"`
	AttachmentURL   *string               `json:"attachment_url" binding:"omitempty,max=2048"`
	InstanceSharing string                `json:"instance_sharing" binding:"omitempty,oneof=per_user per_team shared"`
	Hints           []ChallengeHintReq    `json:"hints" binding:"omitempty,dive"`
}

type ChallengeQuery struct {
	Category   string `form:"category"`
	Difficulty string `form:"difficulty"`
	Status     string `form:"status"`
	CreatedBy  *int64 `form:"created_by"`
	Keyword    string `form:"keyword"`
	SortBy     string `form:"sort_by" binding:"omitempty,oneof=created_at difficulty"`
	Page       int    `form:"page" binding:"omitempty,min=1"`
	Size       int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type CreateTagReq struct {
	Name        string `json:"name" binding:"required,min=2,max=64"`
	Type        string `json:"type" binding:"required,oneof=vulnerability tech_stack knowledge"`
	Description string `json:"description" binding:"max=500"`
}

type AttachTagsReq struct {
	TagIDs []int64 `json:"tag_ids" binding:"required,min=1"`
}

type TagQuery struct {
	Type string `form:"type" binding:"omitempty,oneof=vulnerability tech_stack knowledge"`
}

type CreateImageReq struct {
	Name        string `json:"name" binding:"required,ctf_image_name"`
	Tag         string `json:"tag" binding:"required,ctf_image_tag"`
	Description string `json:"description"`
}

type UpdateImageReq struct {
	Description *string `json:"description"`
	Status      string  `json:"status" binding:"omitempty,oneof=pending building available failed"`
}

type ImageQuery struct {
	Name   string `form:"name"`
	Status string `form:"status"`
	Page   int    `form:"page" binding:"omitempty,min=1"`
	Size   int    `form:"page_size" binding:"omitempty,min=1"`
}

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
	Key             string                `json:"key" binding:"required,max=64"`
	Name            string                `json:"name" binding:"required,max=128"`
	ImageID         int64                 `json:"image_id" binding:"omitempty,min=1"`
	ServicePort     int                   `json:"service_port" binding:"omitempty,min=1,max=65535"`
	ServiceProtocol string                `json:"service_protocol" binding:"omitempty,oneof=http tcp"`
	InjectFlag      bool                  `json:"inject_flag"`
	Tier            string                `json:"tier" binding:"omitempty,oneof=public service internal"`
	NetworkKeys     []string              `json:"network_keys" binding:"omitempty,min=1,dive,required,max=64"`
	Env             map[string]string     `json:"env"`
	Resources       *TopologyResourcesReq `json:"resources"`
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

type UpsertEnvironmentTemplateReq struct {
	Name         string                     `json:"name" binding:"required,max=128"`
	Description  string                     `json:"description" binding:"omitempty,max=2000"`
	EntryNodeKey string                     `json:"entry_node_key" binding:"required,max=64"`
	Networks     []TopologyNetworkReq       `json:"networks" binding:"omitempty,dive"`
	Nodes        []TopologyNodeReq          `json:"nodes" binding:"required,min=1,dive"`
	Links        []TopologyLinkReq          `json:"links" binding:"omitempty,dive"`
	Policies     []TopologyTrafficPolicyReq `json:"policies" binding:"omitempty,dive"`
}

type ConfigureFlagReq struct {
	FlagType   string `json:"flag_type" binding:"required,oneof=static dynamic regex manual_review"`
	Flag       string `json:"flag" binding:"required_if=FlagType static"`
	FlagRegex  string `json:"flag_regex" binding:"required_if=FlagType regex"`
	FlagPrefix string `json:"flag_prefix" binding:"omitempty,max=32"`
}

type AWDChallengeQuery struct {
	Keyword     string `form:"keyword"`
	ServiceType string `form:"service_type"`
	Status      string `form:"status"`
	Page        int    `form:"page" binding:"omitempty,min=1"`
	Size        int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type CreateAWDChallengeReq struct {
	Name           string `json:"name" binding:"required"`
	Slug           string `json:"slug" binding:"required"`
	Category       string `json:"category" binding:"required,oneof=web pwn reverse crypto misc forensics"`
	Difficulty     string `json:"difficulty" binding:"required,oneof=beginner easy medium hard insane"`
	Description    string `json:"description"`
	ServiceType    string `json:"service_type" binding:"required,oneof=web_http binary_tcp multi_container"`
	DeploymentMode string `json:"deployment_mode" binding:"required,oneof=single_container topology"`
}

type UpdateAWDChallengeReq struct {
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	Category       string `json:"category" binding:"omitempty,oneof=web pwn reverse crypto misc forensics"`
	Difficulty     string `json:"difficulty" binding:"omitempty,oneof=beginner easy medium hard insane"`
	Description    string `json:"description"`
	ServiceType    string `json:"service_type" binding:"omitempty,oneof=web_http binary_tcp multi_container"`
	DeploymentMode string `json:"deployment_mode" binding:"omitempty,oneof=single_container topology"`
	Status         string `json:"status" binding:"omitempty,oneof=draft published archived"`
}
