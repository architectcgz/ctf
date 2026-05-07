package jobs

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
)

func TestAWDRoundUpdaterPreviewTCPStandardRunsTCPSteps(t *testing.T) {
	const checkerToken = "preview-checker-token"
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	t.Cleanup(func() {
		_ = listener.Close()
	})
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		storedFlag := ""
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			switch {
			case line == "PING\n":
				_, _ = conn.Write([]byte("PONG\n"))
			case strings.HasPrefix(line, "SET_FLAG "+checkerToken+" "):
				storedFlag = strings.TrimSpace(strings.TrimPrefix(line, "SET_FLAG "+checkerToken+" "))
				_, _ = conn.Write([]byte("OK\n"))
			case line == "GET_FLAG "+checkerToken+"\n":
				_, _ = conn.Write([]byte(storedFlag + "\n"))
				return
			}
		}
	}()

	updater := NewAWDRoundUpdater(nil, nil, config.ContestAWDConfig{CheckerTimeout: time.Second}, "", nil, nil)
	resp, err := updater.PreviewServiceCheck(context.Background(), contestports.AWDServicePreviewRequest{
		ServiceID:      2001,
		AWDChallengeID: 3001,
		CheckerType:    model.AWDCheckerTypeTCPStandard,
		CheckerConfig: `{
			"steps": [
				{"send": "PING\n", "expect_contains": "PONG"},
				{"send_template": "SET_FLAG {{CHECKER_TOKEN}} {{FLAG}}\n", "expect_contains": "OK"},
				{"send_template": "GET_FLAG {{CHECKER_TOKEN}}\n", "expect_contains": "{{FLAG}}"}
			]
		}`,
		CheckerTokenEnv: "CHECKER_TOKEN",
		CheckerToken:    checkerToken,
		AccessURL:       "tcp://" + listener.Addr().String(),
		PreviewFlag:     "flag{preview}",
	})
	if err != nil {
		t.Fatalf("PreviewServiceCheck() error = %v", err)
	}
	if resp.ServiceStatus != model.AWDServiceStatusUp {
		t.Fatalf("ServiceStatus = %s, want up; result=%s", resp.ServiceStatus, resp.CheckResult)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("tcp checker did not connect to fixture")
	}
}

func TestAWDRoundUpdaterTCPStandardDerivesCheckerTokenForRuntimeChecks(t *testing.T) {
	const secret = "runtime-secret-12345678901234567890"
	const contestID int64 = 71
	const teamID int64 = 81
	const serviceID int64 = 2001
	const challengeID int64 = 3001
	expectedToken := contestdomain.BuildAWDCheckerToken(contestID, teamID, serviceID, challengeID, secret)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	t.Cleanup(func() {
		_ = listener.Close()
	})
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		storedFlag := ""
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			switch {
			case line == "PING\n":
				_, _ = conn.Write([]byte("PONG\n"))
			case strings.HasPrefix(line, "SET_FLAG "+expectedToken+" "):
				storedFlag = strings.TrimSpace(strings.TrimPrefix(line, "SET_FLAG "+expectedToken+" "))
				_, _ = conn.Write([]byte("OK\n"))
			case line == "GET_FLAG "+expectedToken+"\n":
				_, _ = conn.Write([]byte(storedFlag + "\n"))
				return
			}
		}
	}()

	updater := NewAWDRoundUpdater(nil, nil, config.ContestAWDConfig{CheckerTimeout: time.Second}, secret, nil, nil)
	outcome, err := updater.buildAWDCheckOutcomeFromTCPStandard(
		context.Background(),
		contestID,
		nil,
		teamID,
		contestports.AWDServiceDefinition{
			ServiceID:       serviceID,
			AWDChallengeID:  challengeID,
			CheckerType:     model.AWDCheckerTypeTCPStandard,
			CheckerTokenEnv: "CHECKER_TOKEN",
			CheckerConfig: `{
				"steps": [
					{"send": "PING\n", "expect_contains": "PONG"},
					{"send_template": "SET_FLAG {{CHECKER_TOKEN}} {{FLAG}}\n", "expect_contains": "OK"},
					{"send_template": "GET_FLAG {{CHECKER_TOKEN}}\n", "expect_contains": "{{FLAG}}"}
				]
			}`,
		},
		[]contestports.AWDServiceInstance{
			{
				ServiceID:      serviceID,
				AWDChallengeID: challengeID,
				AccessURL:      "tcp://" + listener.Addr().String(),
			},
		},
		"manual",
		"flag{round}",
	)
	if err != nil {
		t.Fatalf("buildAWDCheckOutcomeFromTCPStandard() error = %v", err)
	}
	if outcome.serviceStatus != model.AWDServiceStatusUp {
		t.Fatalf("unexpected outcome: %+v", outcome)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("tcp runtime checker did not connect to fixture")
	}
}

func TestAWDRoundUpdaterTCPStandardRedactsFlagInErrors(t *testing.T) {
	updater := NewAWDRoundUpdater(nil, nil, config.ContestAWDConfig{CheckerTimeout: 10 * time.Millisecond}, "", nil, nil)
	resp, err := updater.PreviewServiceCheck(context.Background(), contestports.AWDServicePreviewRequest{
		ServiceID:      2001,
		AWDChallengeID: 3001,
		CheckerType:    model.AWDCheckerTypeTCPStandard,
		CheckerConfig: `{
			"connect": {"host": "{{FLAG}}", "port": 1},
			"steps": [{"send": "PING\n", "expect_contains": "PONG"}]
		}`,
		AccessURL:   "tcp://127.0.0.1:1",
		PreviewFlag: "flag{preview}",
	})
	if err != nil {
		t.Fatalf("PreviewServiceCheck() error = %v", err)
	}
	if strings.Contains(resp.CheckResult, "flag{preview}") {
		t.Fatalf("CheckResult leaked flag: %s", resp.CheckResult)
	}
	if !strings.Contains(resp.CheckResult, "[redacted]") {
		t.Fatalf("CheckResult does not show redaction marker: %s", resp.CheckResult)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(resp.CheckResult), &result); err != nil {
		t.Fatalf("unmarshal check result: %v", err)
	}
	targets, ok := result["targets"].([]any)
	if !ok || len(targets) != 1 {
		t.Fatalf("unexpected targets: %#v", result["targets"])
	}
	target, ok := targets[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected target: %#v", targets[0])
	}
	audit, ok := target["audit"].(map[string]any)
	if !ok {
		t.Fatalf("missing audit: %#v", target)
	}
	if audit["checker_type"] != string(model.AWDCheckerTypeTCPStandard) || audit["service_id"] != float64(2001) || audit["error_code"] != "tcp_connect_failed" {
		t.Fatalf("unexpected audit: %#v", audit)
	}
}
