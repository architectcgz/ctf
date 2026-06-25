package contracts

import (
	"net"
	"net/url"
	"strings"
)

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

func ResolveRuntimeNodePublicHost(nodePublicHost, globalPublicHost string) string {
	if trimmed := strings.TrimSpace(nodePublicHost); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(globalPublicHost)
}

func ResolveRuntimeNodeAccessHost(nodePublicHost, nodeAccessHost, globalPublicHost, globalAccessHost string) string {
	if trimmed := strings.TrimSpace(nodeAccessHost); trimmed != "" {
		return trimmed
	}
	if trimmed := strings.TrimSpace(nodePublicHost); trimmed != "" {
		return trimmed
	}
	return ResolveRuntimePublishedAccessHost(globalPublicHost, globalAccessHost)
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
