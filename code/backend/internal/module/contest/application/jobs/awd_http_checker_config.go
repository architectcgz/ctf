package jobs

import (
	"encoding/json"
	"strings"
)

type awdHTTPCheckerActionConfig struct {
	Method            string            `json:"method"`
	Path              string            `json:"path"`
	Headers           map[string]string `json:"headers"`
	BodyTemplate      string            `json:"body_template"`
	ExpectedStatus    int               `json:"expected_status"`
	ExpectedSubstring string            `json:"expected_substring"`
}

type awdHTTPCheckerConfig struct {
	PutFlag awdHTTPCheckerActionConfig `json:"put_flag"`
	GetFlag awdHTTPCheckerActionConfig `json:"get_flag"`
	Havoc   awdHTTPCheckerActionConfig `json:"havoc"`
}

func parseAWDHTTPCheckerConfig(value string) (awdHTTPCheckerConfig, error) {
	if strings.TrimSpace(value) == "" {
		return awdHTTPCheckerConfig{}, nil
	}

	var config awdHTTPCheckerConfig
	if err := json.Unmarshal([]byte(value), &config); err != nil {
		return awdHTTPCheckerConfig{}, err
	}

	config.PutFlag = normalizeAWDHTTPCheckerActionConfig(config.PutFlag, "PUT", false)
	config.GetFlag = normalizeAWDHTTPCheckerActionConfig(config.GetFlag, "GET", true)
	config.Havoc = normalizeAWDHTTPCheckerActionConfig(config.Havoc, "GET", false)
	return config, nil
}

func normalizeAWDHTTPCheckerActionConfig(action awdHTTPCheckerActionConfig, defaultMethod string, allowFallbackPath bool) awdHTTPCheckerActionConfig {
	action.Method = strings.ToUpper(strings.TrimSpace(action.Method))
	if action.Method == "" {
		action.Method = defaultMethod
	}
	action.Path = strings.TrimSpace(action.Path)
	if action.Path != "" {
		action.Path = normalizedAWDCheckerHealthPath(action.Path)
	} else if !allowFallbackPath {
		action.Path = ""
	}
	if action.Headers == nil {
		action.Headers = map[string]string{}
	}
	if action.ExpectedStatus <= 0 {
		action.ExpectedStatus = 200
	}
	return action
}

func awdHTTPCheckerActionEnabled(action awdHTTPCheckerActionConfig) bool {
	return strings.TrimSpace(action.Path) != ""
}
