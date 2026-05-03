package commands

import (
	"encoding/json"
	"strings"

	"ctf-platform/internal/model"
)

func buildContestAWDServiceVirtualChallenge(service *model.ContestAWDService, snapshot model.ContestAWDServiceSnapshot) *model.Challenge {
	chal := &model.Challenge{
		ID:              service.AWDChallengeID,
		Title:           firstRuntimeValue(service.DisplayName, snapshot.Name),
		Category:        snapshot.Category,
		Difficulty:      snapshot.Difficulty,
		Points:          parseContestAWDServiceSnapshotPoints(service.ScoreConfig),
		Status:          model.ChallengeStatusPublished,
		ImageID:         parseContestAWDServiceSnapshotImageID(snapshot.RuntimeConfig),
		FlagType:        parseContestAWDServiceSnapshotFlagType(snapshot.FlagConfig),
		FlagPrefix:      parseContestAWDServiceSnapshotFlagPrefix(snapshot.FlagConfig),
		InstanceSharing: parseContestAWDServiceSnapshotInstanceSharing(snapshot.RuntimeConfig),
	}
	if chal.FlagPrefix == "" {
		chal.FlagPrefix = "flag"
	}
	return chal
}

func buildContestAWDServiceVirtualTopology(service *model.ContestAWDService, snapshot model.ContestAWDServiceSnapshot) (*model.ChallengeTopology, error) {
	topologyPayload, ok := snapshot.RuntimeConfig["topology"]
	if !ok {
		return nil, nil
	}
	topologyMap, ok := topologyPayload.(map[string]any)
	if !ok {
		return nil, nil
	}
	entryNodeKey, _ := topologyMap["entry_node_key"].(string)
	specPayload, ok := topologyMap["spec"]
	if !ok {
		return nil, nil
	}
	specRaw, err := json.Marshal(specPayload)
	if err != nil {
		return nil, err
	}
	return &model.ChallengeTopology{
		ChallengeID:  service.AWDChallengeID,
		EntryNodeKey: strings.TrimSpace(entryNodeKey),
		Spec:         string(specRaw),
	}, nil
}

func parseContestAWDServiceSnapshotPoints(scoreConfig string) int {
	if scoreConfig == "" {
		return 0
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(scoreConfig), &payload); err != nil {
		return 0
	}
	return parseContestAWDServiceSnapshotInt(payload["points"])
}

func parseContestAWDServiceSnapshotImageID(runtimeConfig map[string]any) int64 {
	if runtimeConfig == nil {
		return 0
	}
	value := parseContestAWDServiceSnapshotInt(runtimeConfig["image_id"])
	if value <= 0 {
		return 0
	}
	return int64(value)
}

func parseContestAWDServiceSnapshotInstanceSharing(runtimeConfig map[string]any) model.InstanceSharing {
	if runtimeConfig == nil {
		return model.InstanceSharingPerTeam
	}
	value, _ := runtimeConfig["instance_sharing"].(string)
	switch model.InstanceSharing(value) {
	case model.InstanceSharingShared:
		return model.InstanceSharingShared
	case model.InstanceSharingPerUser:
		return model.InstanceSharingPerUser
	case model.InstanceSharingPerTeam:
		return model.InstanceSharingPerTeam
	default:
		return model.InstanceSharingPerTeam
	}
}

func parseContestAWDServiceSnapshotFlagType(flagConfig map[string]any) string {
	if flagConfig == nil {
		return model.FlagTypeDynamic
	}
	value, _ := flagConfig["flag_type"].(string)
	if strings.TrimSpace(value) == "" {
		return model.FlagTypeDynamic
	}
	return value
}

func parseContestAWDServiceSnapshotFlagPrefix(flagConfig map[string]any) string {
	if flagConfig == nil {
		return "flag"
	}
	value, _ := flagConfig["flag_prefix"].(string)
	if strings.TrimSpace(value) == "" {
		return "flag"
	}
	return value
}

func parseContestAWDServiceSnapshotInt(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		next, err := typed.Int64()
		if err != nil {
			return 0
		}
		return int(next)
	default:
		return 0
	}
}

func firstRuntimeValue(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
