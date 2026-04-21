package model

import "encoding/json"

type ContestAWDServiceSnapshot struct {
	Name             string            `json:"name"`
	Category         string            `json:"category"`
	Difficulty       string            `json:"difficulty"`
	Description      string            `json:"description,omitempty"`
	ServiceType      AWDServiceType    `json:"service_type,omitempty"`
	DeploymentMode   AWDDeploymentMode `json:"deployment_mode,omitempty"`
	FlagMode         string            `json:"flag_mode,omitempty"`
	FlagConfig       map[string]any    `json:"flag_config,omitempty"`
	DefenseEntryMode string            `json:"defense_entry_mode,omitempty"`
	AccessConfig     map[string]any    `json:"access_config,omitempty"`
	RuntimeConfig    map[string]any    `json:"runtime_config,omitempty"`
}

func EncodeContestAWDServiceSnapshot(snapshot ContestAWDServiceSnapshot) (string, error) {
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func DecodeContestAWDServiceSnapshot(raw string) (ContestAWDServiceSnapshot, error) {
	if raw == "" {
		return ContestAWDServiceSnapshot{}, nil
	}
	var snapshot ContestAWDServiceSnapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return ContestAWDServiceSnapshot{}, err
	}
	if snapshot.FlagConfig == nil {
		snapshot.FlagConfig = map[string]any{}
	}
	if snapshot.AccessConfig == nil {
		snapshot.AccessConfig = map[string]any{}
	}
	if snapshot.RuntimeConfig == nil {
		snapshot.RuntimeConfig = map[string]any{}
	}
	return snapshot, nil
}
