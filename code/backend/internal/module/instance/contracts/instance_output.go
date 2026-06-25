package contracts

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type InstanceAccessInfo struct {
	Protocol string `json:"protocol"`
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Command  string `json:"command,omitempty"`
}

type InstanceResp struct {
	ID                  int64               `json:"id"`
	ChallengeID         int64               `json:"challenge_id"`
	Status              string              `json:"status"`
	ProvisioningStage   string              `json:"provisioning_stage,omitempty"`
	ProvisioningMessage string              `json:"provisioning_message,omitempty"`
	ProvisioningAttempt int                 `json:"provisioning_attempt,omitempty"`
	ShareScope          ShareScope          `json:"share_scope"`
	AccessURL           string              `json:"access_url"`
	Access              *InstanceAccessInfo `json:"access,omitempty"`
	ExpiresAt           time.Time           `json:"expires_at"`
	ExtendCount         int                 `json:"extend_count"`
	MaxExtends          int                 `json:"max_extends"`
	RemainingExtends    int                 `json:"remaining_extends"`
	CreatedAt           time.Time           `json:"created_at"`
}

type InstanceInfo struct {
	ID                  int64               `json:"id"`
	ContestMode         string              `json:"contest_mode,omitempty"`
	ChallengeID         int64               `json:"challenge_id"`
	ChallengeTitle      string              `json:"challenge_title,omitempty"`
	Category            string              `json:"category,omitempty"`
	Difficulty          string              `json:"difficulty,omitempty"`
	FlagType            string              `json:"flag_type,omitempty"`
	Status              string              `json:"status"`
	ProvisioningStage   string              `json:"provisioning_stage,omitempty"`
	ProvisioningMessage string              `json:"provisioning_message,omitempty"`
	ProvisioningAttempt int                 `json:"provisioning_attempt,omitempty"`
	ShareScope          ShareScope          `json:"share_scope"`
	AccessURL           string              `json:"access_url"`
	Access              *InstanceAccessInfo `json:"access,omitempty"`
	ExpiresAt           time.Time           `json:"expires_at"`
	RemainingTime       int64               `json:"remaining_time"`
	ExtendCount         int                 `json:"extend_count"`
	MaxExtends          int                 `json:"max_extends"`
	RemainingExtends    int                 `json:"remaining_extends"`
	CreatedAt           time.Time           `json:"created_at"`
}

const challengeTargetProtocolTCP = "tcp"

func BuildInstanceAccessInfo(accessURL string) *InstanceAccessInfo {
	trimmed := strings.TrimSpace(accessURL)
	if trimmed == "" {
		return nil
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil
	}
	protocol := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if protocol == "" {
		return nil
	}

	info := &InstanceAccessInfo{Protocol: protocol}
	host := parsed.Hostname()
	if host != "" {
		info.Host = host
	}
	if portValue := parsed.Port(); portValue != "" {
		if port, parseErr := strconv.Atoi(portValue); parseErr == nil {
			info.Port = port
		}
	}
	if protocol == challengeTargetProtocolTCP && info.Host != "" && info.Port > 0 {
		info.Command = fmt.Sprintf("nc %s %d", formatCommandHost(info.Host), info.Port)
	}
	return info
}

func formatCommandHost(host string) string {
	if strings.Contains(host, ":") {
		return "[" + host + "]"
	}
	return host
}
