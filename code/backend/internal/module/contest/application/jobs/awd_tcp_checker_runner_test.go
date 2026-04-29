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
	contestports "ctf-platform/internal/module/contest/ports"
)

func TestAWDRoundUpdaterPreviewTCPStandardRunsTCPSteps(t *testing.T) {
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
			case strings.HasPrefix(line, "SET_FLAG "):
				storedFlag = strings.TrimSpace(strings.TrimPrefix(line, "SET_FLAG "))
				_, _ = conn.Write([]byte("OK\n"))
			case line == "GET_FLAG\n":
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
				{"send_template": "SET_FLAG {{FLAG}}\n", "expect_contains": "OK"},
				{"send": "GET_FLAG\n", "expect_contains": "{{FLAG}}"}
			]
		}`,
		AccessURL:   "tcp://" + listener.Addr().String(),
		PreviewFlag: "flag{preview}",
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
