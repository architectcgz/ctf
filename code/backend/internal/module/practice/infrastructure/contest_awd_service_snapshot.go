package infrastructure

import "encoding/json"

type contestAWDServiceSnapshot struct {
	Name             string         `json:"name"`
	Category         string         `json:"category"`
	Difficulty       string         `json:"difficulty"`
	Description      string         `json:"description,omitempty"`
	ServiceType      string         `json:"service_type,omitempty"`
	DeploymentMode   string         `json:"deployment_mode,omitempty"`
	FlagMode         string         `json:"flag_mode,omitempty"`
	FlagConfig       map[string]any `json:"flag_config,omitempty"`
	DefenseEntryMode string         `json:"defense_entry_mode,omitempty"`
	AccessConfig     map[string]any `json:"access_config,omitempty"`
	RuntimeConfig    map[string]any `json:"runtime_config,omitempty"`
}

func decodeContestAWDServiceSnapshot(raw string) (contestAWDServiceSnapshot, error) {
	if raw == "" {
		return contestAWDServiceSnapshot{}, nil
	}
	var snapshot contestAWDServiceSnapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return contestAWDServiceSnapshot{}, err
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
