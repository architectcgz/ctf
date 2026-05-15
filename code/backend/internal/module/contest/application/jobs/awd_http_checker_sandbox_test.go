package jobs

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

func TestAWDHTTPStandardUsesSandboxNetworkForAliasTarget(t *testing.T) {
	runner := &fakeCheckerRunner{
		result: contestports.CheckerRunResult{
			Status: contestports.CheckerRunStatusOK,
			Reason: contestports.CheckerReasonPassed,
			Stdout: `{"status_code":200,"body":"flag{round}"}`,
		},
	}
	updater := NewAWDRoundUpdater(nil, nil, config.ContestAWDConfig{
		CheckerTimeout: time.Second,
		CheckerSandbox: config.CheckerSandboxConfig{
			Timeout:          time.Second,
			CPUQuota:         0.5,
			MemoryBytes:      128 * 1024 * 1024,
			PidsLimit:        64,
			NofileLimit:      128,
			OutputLimitBytes: 32768,
		},
	}, "", nil, nil)
	updater.SetCheckerRunner(runner)

	runtimeDetails, err := model.EncodeInstanceRuntimeDetails(model.InstanceRuntimeDetails{
		Networks: []model.InstanceRuntimeNetwork{
			{Key: model.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-8", Shared: true},
		},
		Containers: []model.InstanceRuntimeContainer{
			{
				ContainerID:    "ctr-awd",
				IsEntryPoint:   true,
				NetworkKeys:    []string{model.TopologyDefaultNetworkKey},
				NetworkAliases: []string{"awd-c8-t15-s21"},
				NetworkIPs:     map[string]string{"ctf-awd-contest-8": "192.168.176.2"},
			},
		},
	})
	if err != nil {
		t.Fatalf("encode runtime details: %v", err)
	}

	result := updater.runAWDHTTPCheckerAction(
		context.Background(),
		"http://awd-c8-t15-s21:8080",
		runtimeDetails,
		awdHTTPCheckerActionConfig{
			Method:         "GET",
			Path:           "/api/flag",
			Headers:        map[string]string{"X-AWD-Checker-Token": "{{CHECKER_TOKEN}}"},
			ExpectedStatus: 200,
		},
		awdHTTPCheckerTemplateData{CheckerToken: "sandbox-checker-token"},
		[]string{"flag{round}"},
	)

	if !result.summary.Healthy {
		t.Fatalf("expected healthy sandbox result, got %+v", result.summary)
	}
	if len(runner.jobs) != 1 {
		t.Fatalf("expected one sandbox job, got %d", len(runner.jobs))
	}
	job := runner.jobs[0]
	if job.NetworkMode != "ctf-awd-contest-8" {
		t.Fatalf("expected contest network mode, got %q", job.NetworkMode)
	}
	if job.Env["AWD_HTTP_URL"] != "http://awd-c8-t15-s21:8080/api/flag" {
		t.Fatalf("unexpected sandbox url: %q", job.Env["AWD_HTTP_URL"])
	}
	var headers map[string]string
	if err := json.Unmarshal([]byte(job.Env["AWD_HTTP_HEADERS"]), &headers); err != nil {
		t.Fatalf("unmarshal sandbox headers: %v", err)
	}
	if headers["X-AWD-Checker-Token"] != "sandbox-checker-token" {
		t.Fatalf("unexpected sandbox headers: %+v", headers)
	}
}
