package container

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"ctf-platform/internal/model"
)

const dockerUserChain = "DOCKER-USER"

func applyACLRules(ctx context.Context, rules []model.InstanceRuntimeACLRule) error {
	if len(rules) == 0 {
		return nil
	}
	if _, err := exec.LookPath("iptables"); err != nil {
		return fmt.Errorf("iptables not found: %w", err)
	}

	applied := make([]model.InstanceRuntimeACLRule, 0, len(rules))
	for idx := len(rules) - 1; idx >= 0; idx-- {
		rule := rules[idx]
		if err := runACLRuleCommand(ctx, insertACLCommand(rule)); err != nil {
			_ = removeACLRules(ctx, applied)
			return err
		}
		applied = append(applied, rule)
	}
	return nil
}

func removeACLRules(ctx context.Context, rules []model.InstanceRuntimeACLRule) error {
	var firstErr error
	for _, rule := range rules {
		if err := runACLRuleCommand(ctx, deleteACLCommand(rule)); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func insertACLCommand(rule model.InstanceRuntimeACLRule) []string {
	return buildACLCommand("-I", rule)
}

func deleteACLCommand(rule model.InstanceRuntimeACLRule) []string {
	return buildACLCommand("-D", rule)
}

func buildACLCommand(operation string, rule model.InstanceRuntimeACLRule) []string {
	args := []string{operation, dockerUserChain, "-s", rule.SourceIP, "-d", rule.TargetIP}
	protocol := strings.TrimSpace(rule.Protocol)
	if protocol != "" && protocol != model.TopologyPolicyProtocolAny {
		args = append(args, "-p", protocol)
	}
	if len(rule.Ports) == 1 {
		args = append(args, "--dport", strconv.Itoa(rule.Ports[0]))
	} else if len(rule.Ports) > 1 {
		ports := make([]string, 0, len(rule.Ports))
		for _, port := range rule.Ports {
			ports = append(ports, strconv.Itoa(port))
		}
		args = append(args, "-m", "multiport", "--dports", strings.Join(ports, ","))
	}
	args = append(args, "-j", strings.ToUpper(rule.Action))
	if strings.TrimSpace(rule.Comment) != "" {
		args = append(args, "-m", "comment", "--comment", rule.Comment)
	}
	return args
}

func runACLRuleCommand(ctx context.Context, args []string) error {
	cmd := exec.CommandContext(ctx, "iptables", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("iptables %s failed: %w", strings.Join(args, " "), fmt.Errorf("%s", strings.TrimSpace(string(output))))
	}
	return nil
}
