package http

import instancecontracts "ctf-platform/internal/module/instance/contracts"

type InstanceAccessResp struct {
	AccessURL string                                `json:"access_url"`
	Access    *instancecontracts.InstanceAccessInfo `json:"access,omitempty"`
}

type AWDDefenseSSHAccessResp struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Command   string `json:"command"`
	ExpiresAt string `json:"expires_at"`
}
