package contracts

import (
	"encoding/json"
	"strings"
)

// InstanceRuntimeACLHandle 表示实例级 ACL 资源句柄，是 cleanup 的 authority source。
type InstanceRuntimeACLHandle struct {
	Chain string `json:"chain"`
}

type InstanceRuntimeDetails struct {
	Networks   []InstanceRuntimeNetwork   `json:"networks,omitempty"`
	Containers []InstanceRuntimeContainer `json:"containers,omitempty"`
	ACL        *InstanceRuntimeACLHandle  `json:"acl,omitempty"`
	ACLRules   []InstanceRuntimeACLRule   `json:"acl_rules,omitempty"`
	AWD        *InstanceAWDRuntimeDetails `json:"awd,omitempty"`
}

type InstanceRuntimeNetwork struct {
	Key       string `json:"key,omitempty"`
	Name      string `json:"name,omitempty"`
	NetworkID string `json:"network_id,omitempty"`
	Subnet    string `json:"subnet,omitempty"`
	Internal  bool   `json:"internal,omitempty"`
	Shared    bool   `json:"shared,omitempty"`
}

type InstanceRuntimeContainer struct {
	NodeKey         string            `json:"node_key,omitempty"`
	ContainerID     string            `json:"container_id"`
	HostPort        int               `json:"host_port,omitempty"`
	ServicePort     int               `json:"service_port,omitempty"`
	ServiceProtocol string            `json:"service_protocol,omitempty"`
	IsEntryPoint    bool              `json:"is_entry_point,omitempty"`
	NetworkKeys     []string          `json:"network_keys,omitempty"`
	NetworkAliases  []string          `json:"network_aliases,omitempty"`
	NetworkIPs      map[string]string `json:"network_ips,omitempty"`
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

type InstanceAWDRuntimeDetails struct {
	Checker *InstanceAWDCheckerRuntimeDetails `json:"checker,omitempty"`
}

type InstanceAWDCheckerRuntimeDetails struct {
	TokenEnv string `json:"token_env,omitempty"`
	Token    string `json:"token,omitempty"`
}

func (d *InstanceRuntimeDetails) SetAWDCheckerToken(tokenEnv, token string) {
	if d == nil {
		return
	}
	tokenEnv = strings.TrimSpace(tokenEnv)
	token = strings.TrimSpace(token)
	if tokenEnv == "" || token == "" {
		return
	}
	if d.AWD == nil {
		d.AWD = &InstanceAWDRuntimeDetails{}
	}
	d.AWD.Checker = &InstanceAWDCheckerRuntimeDetails{
		TokenEnv: tokenEnv,
		Token:    token,
	}
}

func (d InstanceRuntimeDetails) FindAWDCheckerToken(expectedEnv string) string {
	expectedEnv = strings.TrimSpace(expectedEnv)
	if d.AWD == nil || d.AWD.Checker == nil {
		return ""
	}
	token := strings.TrimSpace(d.AWD.Checker.Token)
	if token == "" {
		return ""
	}
	tokenEnv := strings.TrimSpace(d.AWD.Checker.TokenEnv)
	if expectedEnv != "" && !strings.EqualFold(tokenEnv, expectedEnv) {
		return ""
	}
	return token
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
