package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

const awdHTTPCheckerSandboxEntry = "http_action.py"

type awdHTTPSandboxResponse struct {
	StatusCode int    `json:"status_code"`
	Body       string `json:"body"`
	Error      string `json:"error"`
}

func (u *AWDRoundUpdater) runAWDHTTPCheckerActionInSandbox(
	ctx context.Context,
	targetURL string,
	accessURL string,
	runtimeDetails string,
	action awdHTTPCheckerActionConfig,
	headers map[string]string,
	bodyValue string,
) (awdHTTPSandboxResponse, bool) {
	networkMode := resolveAWDCheckerNetworkMode(accessURL, runtimeDetails)
	if u.checkerRunner == nil || networkMode == "" {
		return awdHTTPSandboxResponse{}, false
	}

	headersPayload, err := json.Marshal(headers)
	if err != nil {
		return awdHTTPSandboxResponse{Error: sanitizeAWDCheckError(err)}, true
	}
	timeout := normalizedAWDCheckerTimeout(u.cfg.CheckerTimeout)
	sandboxTimeout := u.cfg.CheckerSandbox.Timeout
	if sandboxTimeout <= 0 {
		sandboxTimeout = timeout + 5*time.Second
	}
	job := contestports.CheckerRunJob{
		Runtime:     "python3",
		Entry:       awdHTTPCheckerSandboxEntry,
		NetworkMode: networkMode,
		Timeout:     sandboxTimeout,
		Env: map[string]string{
			"AWD_HTTP_METHOD":  action.Method,
			"AWD_HTTP_URL":     targetURL,
			"AWD_HTTP_HEADERS": string(headersPayload),
			"AWD_HTTP_BODY":    bodyValue,
			"AWD_HTTP_TIMEOUT": strconv.FormatFloat(timeout.Seconds(), 'f', 3, 64),
		},
		Files: []contestports.CheckerRunFile{
			{
				Path:    awdHTTPCheckerSandboxEntry,
				Content: []byte(awdHTTPCheckerSandboxScript),
				Mode:    0o555,
			},
		},
		Limits: contestports.CheckerRunLimits{
			CPUQuota:         u.cfg.CheckerSandbox.CPUQuota,
			MemoryBytes:      u.cfg.CheckerSandbox.MemoryBytes,
			PidsLimit:        u.cfg.CheckerSandbox.PidsLimit,
			NofileLimit:      u.cfg.CheckerSandbox.NofileLimit,
			OutputLimitBytes: u.cfg.CheckerSandbox.OutputLimitBytes,
		},
	}

	runResult, err := u.checkerRunner.RunChecker(ctx, job)
	if err != nil {
		return awdHTTPSandboxResponse{Error: sanitizeAWDCheckError(err)}, true
	}
	if runResult.Status != contestports.CheckerRunStatusOK {
		reason := strings.TrimSpace(runResult.Reason)
		if reason == "" {
			reason = contestports.CheckerReasonFailed
		}
		return awdHTTPSandboxResponse{Error: sanitizeAWDCheckError(fmt.Errorf("%s: %s", reason, runResult.Stderr))}, true
	}

	var response awdHTTPSandboxResponse
	if err := json.Unmarshal([]byte(strings.TrimSpace(runResult.Stdout)), &response); err != nil {
		return awdHTTPSandboxResponse{Error: sanitizeAWDCheckError(err)}, true
	}
	return response, true
}

func resolveAWDCheckerNetworkMode(accessURL, runtimeDetails string) string {
	targetHost, _ := resolveAWDHTTPDialOverride(accessURL, runtimeDetails)
	if !strings.HasPrefix(targetHost, "awd-c") {
		return ""
	}
	details, err := model.DecodeInstanceRuntimeDetails(runtimeDetails)
	if err != nil {
		return ""
	}
	for _, network := range details.Networks {
		name := strings.TrimSpace(network.Name)
		if name != "" && network.Shared {
			return name
		}
	}
	for _, network := range details.Networks {
		if name := strings.TrimSpace(network.Name); name != "" {
			return name
		}
	}
	return ""
}

var awdHTTPCheckerSandboxScript = strings.TrimSpace(`
import json
import os
import urllib.error
import urllib.request

method = os.environ.get("AWD_HTTP_METHOD", "GET").upper()
url = os.environ["AWD_HTTP_URL"]
body = os.environ.get("AWD_HTTP_BODY", "")
timeout = float(os.environ.get("AWD_HTTP_TIMEOUT", "3"))
headers = json.loads(os.environ.get("AWD_HTTP_HEADERS", "{}"))
data = body.encode("utf-8") if method not in ("GET", "HEAD") or body else None
req = urllib.request.Request(url, data=data, headers=headers, method=method)

try:
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        payload = resp.read(65536).decode("utf-8", "replace")
        print(json.dumps({"status_code": resp.status, "body": payload}))
except urllib.error.HTTPError as exc:
    payload = exc.read(65536).decode("utf-8", "replace")
    print(json.dumps({"status_code": exc.code, "body": payload}))
except Exception as exc:
    print(json.dumps({"error": str(exc)}))
`)
