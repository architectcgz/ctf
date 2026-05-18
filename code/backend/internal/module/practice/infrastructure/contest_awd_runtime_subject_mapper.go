package infrastructure

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"ctf-platform/internal/model"
	contestentity "ctf-platform/internal/module/contest/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
)

func buildContestAWDServiceRuntimeSubject(service *contestentity.ContestAWDService) (*practiceports.ContestAWDServiceRuntimeSubject, error) {
	if service == nil {
		return nil, nil
	}

	snapshot, err := contestentity.DecodeContestAWDServiceSnapshot(service.ServiceSnapshot)
	if err != nil {
		return nil, err
	}

	topology, err := buildContestAWDServiceRuntimeTopology(service, snapshot)
	if err != nil {
		return nil, err
	}

	return &practiceports.ContestAWDServiceRuntimeSubject{
		ServiceID:        service.ID,
		ChallengeID:      service.AWDChallengeID,
		Visible:          service.IsVisible,
		SeedSignature:    buildContestAWDServiceSeedSignature(service.ServiceSnapshot),
		RuntimeChallenge: buildContestAWDServiceRuntimeChallenge(service, snapshot),
		RuntimeTopology:  topology,
		WorkspaceConfig:  parseContestAWDDefenseWorkspaceConfig(snapshot.RuntimeConfig),
	}, nil
}

func buildContestAWDServiceRuntimeChallenge(service *contestentity.ContestAWDService, snapshot contestentity.ContestAWDServiceSnapshot) *model.Challenge {
	chal := &model.Challenge{
		ID:              service.AWDChallengeID,
		Title:           firstNonEmptyRuntimeValue(service.DisplayName, snapshot.Name),
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

func buildContestAWDServiceRuntimeTopology(service *contestentity.ContestAWDService, snapshot contestentity.ContestAWDServiceSnapshot) (*model.ChallengeTopology, error) {
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

func firstNonEmptyRuntimeValue(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func parseContestAWDDefenseWorkspaceConfig(runtimeConfig map[string]any) *practiceports.ContestAWDDefenseWorkspaceConfig {
	if runtimeConfig == nil {
		return nil
	}
	raw, ok := runtimeConfig["defense_workspace"]
	if !ok {
		return nil
	}
	payload, ok := raw.(map[string]any)
	if !ok {
		return nil
	}

	seedRoot := strings.TrimSpace(readRuntimeString(payload["seed_root"]))
	if seedRoot == "" {
		return nil
	}

	workspaceRoots := readRuntimeStringList(payload["workspace_roots"])
	if len(workspaceRoots) == 0 {
		return nil
	}
	writableRootSet := make(map[string]struct{}, len(workspaceRoots))
	for _, root := range readRuntimeStringList(payload["writable_roots"]) {
		writableRootSet[root] = struct{}{}
	}

	roots := make([]practiceports.ContestAWDDefenseWorkspaceRoot, 0, len(workspaceRoots))
	for _, root := range workspaceRoots {
		_, writable := writableRootSet[root]
		roots = append(roots, practiceports.ContestAWDDefenseWorkspaceRoot{
			Source:   root,
			ReadOnly: !writable,
		})
	}

	runtimeMounts := parseContestAWDDefenseRuntimeMounts(payload["runtime_mounts"])
	if len(runtimeMounts) == 0 {
		return nil
	}

	return &practiceports.ContestAWDDefenseWorkspaceConfig{
		SeedRoot:        seedRoot,
		WorkspaceRoots:  roots,
		RuntimeMounts:   runtimeMounts,
		CheckerTokenEnv: strings.TrimSpace(readRuntimeString(runtimeConfig["checker_token_env"])),
	}
}

func parseContestAWDDefenseRuntimeMounts(raw any) []practiceports.ContestAWDDefenseRuntimeMount {
	items, ok := raw.([]any)
	if !ok || len(items) == 0 {
		return nil
	}

	result := make([]practiceports.ContestAWDDefenseRuntimeMount, 0, len(items))
	for _, item := range items {
		payload, ok := item.(map[string]any)
		if !ok {
			return nil
		}
		source := strings.TrimSpace(readRuntimeString(payload["source"]))
		target := strings.TrimSpace(readRuntimeString(payload["target"]))
		mode := strings.ToLower(strings.TrimSpace(readRuntimeString(payload["mode"])))
		if source == "" || target == "" || mode == "" {
			return nil
		}
		result = append(result, practiceports.ContestAWDDefenseRuntimeMount{
			Source:   source,
			Target:   target,
			ReadOnly: mode == "ro",
		})
	}
	return result
}

func readRuntimeString(raw any) string {
	value, _ := raw.(string)
	return value
}

func readRuntimeStringList(raw any) []string {
	switch typed := raw.(type) {
	case []string:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			value := strings.TrimSpace(item)
			if value != "" {
				items = append(items, value)
			}
		}
		return items
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			value := strings.TrimSpace(readRuntimeString(item))
			if value != "" {
				items = append(items, value)
			}
		}
		return items
	default:
		return nil
	}
}

func buildContestAWDServiceSeedSignature(raw string) string {
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}
