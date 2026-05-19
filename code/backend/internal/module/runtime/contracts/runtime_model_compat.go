package contracts

import (
	"encoding/json"
	"net"
	"net/url"
	"strings"
)

const (
	ChallengeTargetProtocolHTTP = "http"
	ChallengeTargetProtocolTCP  = "tcp"

	TopologyDefaultNetworkKey = "default"

	TopologyPolicyActionAllow = "allow"
	TopologyPolicyActionDeny  = "deny"

	TopologyPolicyProtocolTCP = "tcp"
	TopologyPolicyProtocolUDP = "udp"
	TopologyPolicyProtocolAny = "any"
)

type ContainerConfig struct {
	Image          string
	Name           string
	Env            []string
	Command        []string
	WorkingDir     string
	Ports          map[string]string
	Mounts         []ContainerMount
	Labels         map[string]string
	Resources      *ResourceLimits
	Security       *SecurityConfig
	Network        string
	NetworkAliases []string
}

type ContainerMount struct {
	Source   string
	Target   string
	ReadOnly bool
}

type ResourceLimits struct {
	CPUQuota  float64
	Memory    int64
	PidsLimit int64
}

type SecurityConfig struct {
	ReadonlyRootfs bool
	CapDrop        []string
	CapAdd         []string
	SecurityOpt    []string
	User           string
}

type TopologyTrafficPolicy struct {
	SourceNodeKey string `json:"source_node_key"`
	TargetNodeKey string `json:"target_node_key"`
	Action        string `json:"action"`
	Protocol      string `json:"protocol,omitempty"`
	Ports         []int  `json:"ports,omitempty"`
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

func ResolveRuntimeAliasAccessURL(accessURL, rawRuntimeDetails string) string {
	trimmed := strings.TrimSpace(accessURL)
	if trimmed == "" {
		return accessURL
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return accessURL
	}
	if !strings.HasPrefix(parsed.Hostname(), "awd-c") {
		return accessURL
	}
	ip := resolveRuntimeEntryNetworkIP(rawRuntimeDetails)
	if ip == "" {
		return accessURL
	}
	if port := parsed.Port(); port != "" {
		parsed.Host = net.JoinHostPort(ip, port)
	} else {
		parsed.Host = ip
	}
	return parsed.String()
}

func ResolveRuntimeInternalAccessURL(accessURL, publicHost, accessHost string) string {
	internalHost := strings.TrimSpace(accessHost)
	if internalHost == "" {
		return accessURL
	}
	return rewriteAccessURLHost(accessURL, publicHost, internalHost)
}

func ResolveRuntimePublicAccessURL(accessURL, publicHost, accessHost string) string {
	internalHost := strings.TrimSpace(accessHost)
	if internalHost == "" {
		return accessURL
	}
	return rewriteAccessURLHost(accessURL, internalHost, publicHost)
}

func ResolveRuntimePublishedAccessHost(publicHost, accessHost string) string {
	if trimmed := strings.TrimSpace(accessHost); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(publicHost)
}

func IsBroadTopologyPolicy(policy TopologyTrafficPolicy) bool {
	if len(policy.Ports) > 0 {
		return false
	}
	protocol := strings.TrimSpace(policy.Protocol)
	return protocol == "" || protocol == TopologyPolicyProtocolAny
}

func rewriteAccessURLHost(accessURL, fromHost, toHost string) string {
	trimmed := strings.TrimSpace(accessURL)
	fromHost = strings.TrimSpace(fromHost)
	toHost = strings.TrimSpace(toHost)
	if trimmed == "" || fromHost == "" || toHost == "" {
		return accessURL
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return accessURL
	}
	if !strings.EqualFold(parsed.Hostname(), fromHost) {
		return accessURL
	}
	if port := parsed.Port(); port != "" {
		parsed.Host = net.JoinHostPort(toHost, port)
	} else {
		parsed.Host = toHost
	}
	return parsed.String()
}

func resolveRuntimeEntryNetworkIP(rawRuntimeDetails string) string {
	details, err := DecodeInstanceRuntimeDetails(rawRuntimeDetails)
	if err != nil {
		return ""
	}
	networkNameByKey := make(map[string]string, len(details.Networks))
	for _, network := range details.Networks {
		if network.Key == "" || network.Name == "" {
			continue
		}
		networkNameByKey[network.Key] = network.Name
	}
	for _, container := range details.Containers {
		if !container.IsEntryPoint {
			continue
		}
		if ip := resolveContainerNetworkIP(container, networkNameByKey); ip != "" {
			return ip
		}
	}
	for _, container := range details.Containers {
		if ip := resolveContainerNetworkIP(container, networkNameByKey); ip != "" {
			return ip
		}
	}
	return ""
}

func resolveContainerNetworkIP(container InstanceRuntimeContainer, networkNameByKey map[string]string) string {
	for _, networkKey := range container.NetworkKeys {
		networkName := networkNameByKey[networkKey]
		if networkName == "" {
			continue
		}
		if ip := strings.TrimSpace(container.NetworkIPs[networkName]); ip != "" {
			return ip
		}
	}
	for _, ip := range container.NetworkIPs {
		if trimmed := strings.TrimSpace(ip); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
