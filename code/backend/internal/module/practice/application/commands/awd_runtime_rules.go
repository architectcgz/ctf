package commands

import (
	"fmt"
	"net/url"
	"strings"

	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
)

const runtimeContainerNamePrefix = "ctf-instance-"

func isAWDInstance(instance *model.Instance) bool {
	return instance != nil && instance.ContestID != nil && instance.ServiceID != nil
}

func buildAWDContestNetworkName(instance *model.Instance) string {
	if instance == nil || instance.ContestID == nil {
		return ""
	}
	return fmt.Sprintf("ctf-awd-contest-%d", *instance.ContestID)
}

func buildAWDServiceAlias(instance *model.Instance) string {
	if instance == nil || instance.ContestID == nil || instance.TeamID == nil || instance.ServiceID == nil {
		return ""
	}
	return fmt.Sprintf("awd-c%d-t%d-s%d", *instance.ContestID, *instance.TeamID, *instance.ServiceID)
}

func applyAWDStableNetworkToTopologyRequest(instance *model.Instance, chal *model.Challenge, request *practiceports.TopologyCreateRequest) {
	if !isAWDInstance(instance) || request == nil {
		return
	}
	request.ContainerName = buildRuntimeContainerName(chal, instance)
	entryIndex := -1
	for idx := range request.Nodes {
		if request.Nodes[idx].IsEntryPoint {
			entryIndex = idx
			break
		}
	}
	if entryIndex < 0 {
		return
	}

	networkKey := model.TopologyDefaultNetworkKey
	if len(request.Nodes[entryIndex].NetworkKeys) > 0 && strings.TrimSpace(request.Nodes[entryIndex].NetworkKeys[0]) != "" {
		networkKey = request.Nodes[entryIndex].NetworkKeys[0]
	} else {
		request.Nodes[entryIndex].NetworkKeys = []string{networkKey}
	}

	networkName := buildAWDContestNetworkName(instance)
	networkFound := false
	for idx := range request.Networks {
		if request.Networks[idx].Key != networkKey {
			continue
		}
		request.Networks[idx].Name = networkName
		request.Networks[idx].Shared = true
		networkFound = true
		break
	}
	if !networkFound {
		request.Networks = append(request.Networks, practiceports.TopologyCreateNetwork{
			Key:    networkKey,
			Name:   networkName,
			Shared: true,
		})
	}

	alias := buildAWDServiceAlias(instance)
	if alias != "" {
		request.Nodes[entryIndex].NetworkAliases = appendUniqueString(request.Nodes[entryIndex].NetworkAliases, alias)
	}
}

func appendUniqueString(items []string, item string) []string {
	item = strings.TrimSpace(item)
	if item == "" {
		return items
	}
	for _, existing := range items {
		if strings.TrimSpace(existing) == item {
			return items
		}
	}
	return append(items, item)
}

func usesAWDStableNetworkAlias(instance *model.Instance) bool {
	if !isAWDInstance(instance) {
		return false
	}
	parsed, err := url.Parse(strings.TrimSpace(instance.AccessURL))
	if err != nil {
		return false
	}
	host := parsed.Hostname()
	return strings.HasPrefix(host, "awd-c")
}

func buildRuntimeContainerName(chal *model.Challenge, instance *model.Instance) string {
	if !isAWDInstance(instance) || instance == nil || instance.ContestID == nil || instance.TeamID == nil {
		return ""
	}
	challengeSegment := sanitizeRuntimeContainerSegment(resolveRuntimeChallengeName(chal))
	if challengeSegment == "" {
		challengeSegment = "challenge"
	}
	return fmt.Sprintf("%s%s-c%d-t%d", runtimeContainerNamePrefix, challengeSegment, *instance.ContestID, *instance.TeamID)
}

func resolveRuntimeChallengeName(chal *model.Challenge) string {
	if chal == nil {
		return ""
	}
	if chal.PackageSlug != nil && strings.TrimSpace(*chal.PackageSlug) != "" {
		return strings.TrimSpace(*chal.PackageSlug)
	}
	return strings.TrimSpace(chal.Title)
}

func sanitizeRuntimeContainerSegment(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		isAlphaNum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isAlphaNum {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	result := strings.Trim(b.String(), "-")
	if len(result) > 48 {
		result = strings.Trim(result[:48], "-")
	}
	return result
}
