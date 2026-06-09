package contracts

import (
	"encoding/json"
	"net"
	"net/url"
	"strings"
)

func ResolveInstanceAliasAccessURL(accessURL, rawRuntimeDetails string) string {
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

func ResolveInstancePublicAccessURL(accessURL, publicHost, accessHost string) string {
	internalHost := strings.TrimSpace(accessHost)
	if internalHost == "" {
		return accessURL
	}
	return rewriteAccessURLHost(accessURL, internalHost, publicHost)
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

func ExtractInstanceRuntimeContainerIDs(rawRuntimeDetails string) ([]string, error) {
	details, err := decodeInstanceRuntimeDetails(rawRuntimeDetails)
	if err != nil {
		return nil, err
	}
	containerIDs := make([]string, 0, len(details.Containers))
	for _, container := range details.Containers {
		containerIDs = appendUniqueRuntimeContainerID(containerIDs, container.ContainerID)
	}
	return containerIDs, nil
}

func resolveRuntimeEntryNetworkIP(rawRuntimeDetails string) string {
	details, err := decodeInstanceRuntimeDetails(rawRuntimeDetails)
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

func resolveContainerNetworkIP(container instanceRuntimeContainer, networkNameByKey map[string]string) string {
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

func decodeInstanceRuntimeDetails(raw string) (instanceRuntimeDetails, error) {
	if raw == "" {
		return instanceRuntimeDetails{}, nil
	}
	var details instanceRuntimeDetails
	if err := json.Unmarshal([]byte(raw), &details); err != nil {
		return instanceRuntimeDetails{}, err
	}
	return details, nil
}

func appendUniqueRuntimeContainerID(ids []string, containerID string) []string {
	if containerID == "" {
		return ids
	}
	for _, existing := range ids {
		if existing == containerID {
			return ids
		}
	}
	return append(ids, containerID)
}

type instanceRuntimeDetails struct {
	Networks   []instanceRuntimeNetwork   `json:"networks,omitempty"`
	Containers []instanceRuntimeContainer `json:"containers,omitempty"`
}

type instanceRuntimeNetwork struct {
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
}

type instanceRuntimeContainer struct {
	ContainerID  string            `json:"container_id"`
	IsEntryPoint bool              `json:"is_entry_point,omitempty"`
	NetworkKeys  []string          `json:"network_keys,omitempty"`
	NetworkIPs   map[string]string `json:"network_ips,omitempty"`
}
