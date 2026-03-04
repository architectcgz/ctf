package model

type HealthStatus struct {
	Status       string            `json:"status"`
	Service      string            `json:"service"`
	Environment  string            `json:"environment"`
	Dependencies map[string]string `json:"dependencies"`
	Version      string            `json:"version"`
}
