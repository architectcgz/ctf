package dto

type CheatDetectionSummary struct {
	SubmitBurstUsers int `json:"submit_burst_users"`
	SharedIPGroups   int `json:"shared_ip_groups"`
	AffectedUsers    int `json:"affected_users"`
}

type CheatDetectionUser struct {
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	SubmitCount int    `json:"submit_count"`
	LastSeenAt  string `json:"last_seen_at"`
	Reason      string `json:"reason"`
}

type CheatDetectionIPGroup struct {
	IP        string   `json:"ip"`
	UserCount int      `json:"user_count"`
	Usernames []string `json:"usernames"`
}

type CheatDetectionResp struct {
	GeneratedAt string                  `json:"generated_at"`
	Summary     CheatDetectionSummary   `json:"summary"`
	Suspects    []CheatDetectionUser    `json:"suspects"`
	SharedIPs   []CheatDetectionIPGroup `json:"shared_ips"`
}
