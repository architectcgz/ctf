package infrastructure

import (
	"context"
	"errors"
	"strings"
	"testing"

	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
)

func TestValidateACLRuleCanonicalIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sourceIP  string
		targetIP  string
		wantErr   bool
		errSubstr string
	}{
		{name: "valid ipv4", sourceIP: "172.30.0.2", targetIP: "172.30.0.3"},
		{name: "valid ipv4 with spaces", sourceIP: " 172.30.0.2 ", targetIP: " 172.30.0.3 "},
		{name: "ipv6 rejected", sourceIP: "::1", targetIP: "172.30.0.3", wantErr: true, errSubstr: "not ipv4"},
		{name: "cidr rejected", sourceIP: "172.30.0.0/24", targetIP: "172.30.0.3", wantErr: true, errSubstr: "invalid source ip"},
		{name: "hostname rejected", sourceIP: "foo.bar", targetIP: "172.30.0.3", wantErr: true, errSubstr: "invalid source ip"},
		{name: "empty source rejected", sourceIP: "", targetIP: "172.30.0.3", wantErr: true, errSubstr: "source ip is empty"},
		{name: "empty target rejected", sourceIP: "172.30.0.2", targetIP: "", wantErr: true, errSubstr: "target ip is empty"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rule := runtimecontracts.InstanceRuntimeACLRule{
				SourceIP: tt.sourceIP,
				TargetIP: tt.targetIP,
				Action:   runtimecontracts.TopologyPolicyActionAllow,
				Protocol: runtimecontracts.TopologyPolicyProtocolAny,
			}
			validated, err := validateAndCanonicalizeACLRule(rule)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("expected error to contain %q, got %v", tt.errSubstr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if validated.SourceIP != strings.TrimSpace(tt.sourceIP) {
				t.Fatalf("source ip not canonicalized: got %q want %q", validated.SourceIP, strings.TrimSpace(tt.sourceIP))
			}
		})
	}
}

func TestValidateACLRuleActionWhitelist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		action  string
		wantErr bool
	}{
		{name: "allow", action: "allow"},
		{name: "deny", action: "deny"},
		{name: "allow uppercase", action: "ALLOW"},
		{name: "drop rejected", action: "DROP", wantErr: true},
		{name: "accept rejected", action: "ACCEPT", wantErr: true},
		{name: "empty rejected", action: "", wantErr: true},
		{name: "random rejected", action: "forward", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rule := runtimecontracts.InstanceRuntimeACLRule{
				SourceIP: "172.30.0.2",
				TargetIP: "172.30.0.3",
				Action:   tt.action,
				Protocol: runtimecontracts.TopologyPolicyProtocolAny,
			}
			_, err := validateAndCanonicalizeACLRule(rule)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateACLRuleProtocolWhitelist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		protocol string
		wantErr  bool
	}{
		{name: "tcp", protocol: "tcp"},
		{name: "udp", protocol: "udp"},
		{name: "any", protocol: "any"},
		{name: "empty defaults to any", protocol: ""},
		{name: "icmp rejected", protocol: "icmp", wantErr: true},
		{name: "all rejected", protocol: "all", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rule := runtimecontracts.InstanceRuntimeACLRule{
				SourceIP: "172.30.0.2",
				TargetIP: "172.30.0.3",
				Action:   runtimecontracts.TopologyPolicyActionAllow,
				Protocol: tt.protocol,
			}
			_, err := validateAndCanonicalizeACLRule(rule)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateACLRulePortConstraints(t *testing.T) {
	t.Parallel()

	t.Run("protocol any with ports rejected", func(t *testing.T) {
		t.Parallel()

		rule := runtimecontracts.InstanceRuntimeACLRule{
			SourceIP: "172.30.0.2",
			TargetIP: "172.30.0.3",
			Action:   runtimecontracts.TopologyPolicyActionAllow,
			Protocol: runtimecontracts.TopologyPolicyProtocolAny,
			Ports:    []int{80},
		}
		_, err := validateAndCanonicalizeACLRule(rule)
		if err == nil {
			t.Fatal("expected error for protocol=any with ports")
		}
	})

	t.Run("out of range port", func(t *testing.T) {
		t.Parallel()

		rule := runtimecontracts.InstanceRuntimeACLRule{
			SourceIP: "172.30.0.2",
			TargetIP: "172.30.0.3",
			Action:   runtimecontracts.TopologyPolicyActionAllow,
			Protocol: runtimecontracts.TopologyPolicyProtocolTCP,
			Ports:    []int{70000},
		}
		_, err := validateAndCanonicalizeACLRule(rule)
		if err == nil {
			t.Fatal("expected error for out of range port")
		}
	})

	t.Run("dedupes and sorts ports", func(t *testing.T) {
		t.Parallel()

		rule := runtimecontracts.InstanceRuntimeACLRule{
			SourceIP: "172.30.0.2",
			TargetIP: "172.30.0.3",
			Action:   runtimecontracts.TopologyPolicyActionAllow,
			Protocol: runtimecontracts.TopologyPolicyProtocolTCP,
			Ports:    []int{8080, 80, 443, 80, 22},
		}
		validated, err := validateAndCanonicalizeACLRule(rule)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(validated.Ports) != 4 {
			t.Fatalf("expected 4 deduped ports, got %v", validated.Ports)
		}
		for i := 1; i < len(validated.Ports); i++ {
			if validated.Ports[i] <= validated.Ports[i-1] {
				t.Fatalf("ports not sorted: %v", validated.Ports)
			}
		}
	})

	t.Run("multiport limit exceeded", func(t *testing.T) {
		t.Parallel()

		ports := make([]int, 16)
		for i := range ports {
			ports[i] = i + 1
		}
		rule := runtimecontracts.InstanceRuntimeACLRule{
			SourceIP: "172.30.0.2",
			TargetIP: "172.30.0.3",
			Action:   runtimecontracts.TopologyPolicyActionAllow,
			Protocol: runtimecontracts.TopologyPolicyProtocolTCP,
			Ports:    ports,
		}
		_, err := validateAndCanonicalizeACLRule(rule)
		if err == nil {
			t.Fatal("expected error for exceeding multiport limit")
		}
	})
}

func TestValidateACLRuleCommentIsRebuilt(t *testing.T) {
	t.Parallel()

	rule := runtimecontracts.InstanceRuntimeACLRule{
		SourceIP: "172.30.0.2",
		TargetIP: "172.30.0.3",
		Action:   runtimecontracts.TopologyPolicyActionAllow,
		Protocol: runtimecontracts.TopologyPolicyProtocolTCP,
		Ports:    []int{3306},
		Comment:  "; rm -rf /",
	}
	validated, err := validateAndCanonicalizeACLRule(rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(validated.Comment, "rm") {
		t.Fatalf("comment should be rebuilt, got %q", validated.Comment)
	}
	if !strings.HasPrefix(validated.Comment, "ctf:acl:") {
		t.Fatalf("expected system comment prefix, got %q", validated.Comment)
	}
}

func TestValidateACLRulesBatch(t *testing.T) {
	t.Parallel()

	validRule := runtimecontracts.InstanceRuntimeACLRule{
		SourceIP: "172.30.0.2",
		TargetIP: "172.30.0.3",
		Action:   runtimecontracts.TopologyPolicyActionAllow,
		Protocol: runtimecontracts.TopologyPolicyProtocolTCP,
		Ports:    []int{3306},
	}

	t.Run("all valid", func(t *testing.T) {
		t.Parallel()

		rules := []runtimecontracts.InstanceRuntimeACLRule{validRule, validRule}
		validated, err := validateACLRules(rules)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(validated) != 2 {
			t.Fatalf("expected 2 validated rules, got %d", len(validated))
		}
	})

	t.Run("one invalid fails fast", func(t *testing.T) {
		t.Parallel()

		rules := []runtimecontracts.InstanceRuntimeACLRule{
			validRule,
			{SourceIP: "", TargetIP: "172.30.0.3", Action: "allow", Protocol: "any"},
		}
		_, err := validateACLRules(rules)
		if err == nil {
			t.Fatal("expected error for invalid rule in batch")
		}
	})
}

func TestDeduplicateAndSortPorts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  []int
		expect []int
	}{
		{name: "empty", input: nil, expect: nil},
		{name: "single", input: []int{80}, expect: []int{80}},
		{name: "duplicates", input: []int{80, 443, 80, 22}, expect: []int{22, 80, 443}},
		{name: "sorted", input: []int{22, 80, 443}, expect: []int{22, 80, 443}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := deduplicateAndSortPorts(tt.input)
			if len(got) != len(tt.expect) {
				t.Fatalf("expected %v, got %v", tt.expect, got)
			}
			for i, v := range got {
				if tt.expect[i] != v {
					t.Fatalf("expected %v, got %v", tt.expect, got)
				}
			}
		})
	}
}

func TestRemoveACLRulesRebuildsCanonicalComment(t *testing.T) {
	originalLookPath := iptablesLookPath
	originalRun := runACLCommand
	t.Cleanup(func() {
		iptablesLookPath = originalLookPath
		runACLCommand = originalRun
	})

	var captured [][]string
	iptablesLookPath = func(file string) (string, error) {
		return "/usr/sbin/iptables", nil
	}
	runACLCommand = func(_ context.Context, args []string) error {
		captured = append(captured, append([]string(nil), args...))
		return nil
	}

	rules := []runtimecontracts.InstanceRuntimeACLRule{
		{
			SourceIP: "172.30.0.2",
			TargetIP: "172.30.0.3",
			Action:   runtimecontracts.TopologyPolicyActionAllow,
			Protocol: runtimecontracts.TopologyPolicyProtocolTCP,
			Ports:    []int{3306},
			Comment:  "ctf:acl:4a51f4cb6b8c2f17",
		},
	}

	if err := removeACLRules(context.Background(), rules); err != nil {
		t.Fatalf("removeACLRules() error = %v", err)
	}
	if len(captured) != 1 {
		t.Fatalf("expected 1 command, got %d (%v)", len(captured), captured)
	}
	joined := strings.Join(captured[0], " ")
	if !strings.Contains(joined, "--comment ctf:acl:172.30.0.2:172.30.0.3:allow:tcp:3306") {
		t.Fatalf("expected canonical comment to be rebuilt, command = %q", joined)
	}
}

func TestRemoveACLRulesCanonicalizesMultiportOrder(t *testing.T) {
	originalLookPath := iptablesLookPath
	originalRun := runACLCommand
	t.Cleanup(func() {
		iptablesLookPath = originalLookPath
		runACLCommand = originalRun
	})

	var captured [][]string
	iptablesLookPath = func(file string) (string, error) {
		return "/usr/sbin/iptables", nil
	}
	runACLCommand = func(_ context.Context, args []string) error {
		captured = append(captured, append([]string(nil), args...))
		return nil
	}

	rules := []runtimecontracts.InstanceRuntimeACLRule{
		{
			SourceIP: "172.30.0.2",
			TargetIP: "172.30.0.3",
			Action:   runtimecontracts.TopologyPolicyActionAllow,
			Protocol: runtimecontracts.TopologyPolicyProtocolTCP,
			Ports:    []int{8080, 3306},
			Comment:  "ctf:acl:4a51f4cb6b8c2f17",
		},
	}

	if err := removeACLRules(context.Background(), rules); err != nil {
		t.Fatalf("removeACLRules() error = %v", err)
	}
	if len(captured) != 1 {
		t.Fatalf("expected 1 command, got %d (%v)", len(captured), captured)
	}
	joined := strings.Join(captured[0], " ")
	if !strings.Contains(joined, "--dports 3306,8080") {
		t.Fatalf("expected canonical multiport order, command = %q", joined)
	}
	if strings.Contains(joined, "--dports 8080,3306") {
		t.Fatalf("expected command not to preserve non-canonical multiport order, command = %q", joined)
	}
}

func TestApplyInstanceACLRollsBackChainOnJumpFailure(t *testing.T) {
	t.Parallel()

	originalLookPath := iptablesLookPath
	originalRun := runACLCommand
	t.Cleanup(func() {
		iptablesLookPath = originalLookPath
		runACLCommand = originalRun
	})

	var commands [][]string
	iptablesLookPath = func(file string) (string, error) {
		return "/usr/sbin/iptables", nil
	}
	runACLCommand = func(_ context.Context, args []string) error {
		commands = append(commands, append([]string(nil), args...))
		joined := strings.Join(args, " ")
		switch {
		case strings.HasPrefix(joined, "-C "):
			return errors.New("check failed")
		case strings.HasPrefix(joined, "-I DOCKER-USER 1 -j CTF-INS-77"):
			return errors.New("jump insert failed")
		case strings.HasPrefix(joined, "-D DOCKER-USER -j CTF-INS-77"):
			return errors.New("does a matching rule exist")
		default:
			return nil
		}
	}

	handle := &runtimecontracts.InstanceRuntimeACLHandle{Chain: "CTF-INS-77"}
	rules := []runtimecontracts.InstanceRuntimeACLRule{
		{
			SourceIP: "172.30.0.2",
			TargetIP: "172.30.0.3",
			Action:   runtimecontracts.TopologyPolicyActionAllow,
			Protocol: runtimecontracts.TopologyPolicyProtocolTCP,
			Ports:    []int{3306},
		},
	}

	err := applyInstanceACL(context.Background(), handle, rules)
	if err == nil {
		t.Fatal("expected applyInstanceACL() to fail on jump insertion")
	}

	foundFlush := false
	foundDelete := false
	for _, cmd := range commands {
		joined := strings.Join(cmd, " ")
		if strings.HasPrefix(joined, "-F CTF-INS-77") {
			foundFlush = true
		}
		if strings.HasPrefix(joined, "-X CTF-INS-77") {
			foundDelete = true
		}
	}
	if !foundFlush || !foundDelete {
		t.Fatalf("expected rollback to flush and delete chain, got commands %v", commands)
	}
}
