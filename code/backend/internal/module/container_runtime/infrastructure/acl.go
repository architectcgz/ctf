package infrastructure

import (
	"context"
	"fmt"
	"net/netip"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
)

const (
	validACLActionAllow = "allow"
	validACLActionDeny  = "deny"

	validACLProtocolAny = "any"
	validACLProtocolTCP = "tcp"
	validACLProtocolUDP = "udp"

	maxMultiportCount = 15
)

const dockerUserChain = "DOCKER-USER"

var (
	iptablesLookPath = exec.LookPath
	runACLCommand    = runACLRuleCommand
)

func applyACLRules(ctx context.Context, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	if len(rules) == 0 {
		return nil
	}
	if _, err := iptablesLookPath("iptables"); err != nil {
		return fmt.Errorf("iptables not found: %w", err)
	}

	validated, err := validateACLRules(rules)
	if err != nil {
		return err
	}

	applied := make([]runtimecontracts.InstanceRuntimeACLRule, 0, len(validated))
	for idx := len(validated) - 1; idx >= 0; idx-- {
		rule := validated[idx]
		if err := runACLRuleCommand(ctx, insertACLCommand(rule)); err != nil {
			_ = removeACLRules(ctx, applied)
			return err
		}
		applied = append(applied, rule)
	}
	return nil
}

func removeACLRules(ctx context.Context, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	validated, err := validateACLRulesForDelete(rules)
	if err != nil {
		return err
	}

	var firstErr error
	for _, rule := range validated {
		if err := runACLCommand(ctx, deleteACLCommand(rule)); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func insertACLCommand(rule runtimecontracts.InstanceRuntimeACLRule) []string {
	return buildACLCommand("-I", rule)
}

func deleteACLCommand(rule runtimecontracts.InstanceRuntimeACLRule) []string {
	return buildACLCommand("-D", rule)
}

func buildACLCommand(operation string, rule runtimecontracts.InstanceRuntimeACLRule) []string {
	args := []string{operation, dockerUserChain, "-s", rule.SourceIP, "-d", rule.TargetIP}
	protocol := strings.TrimSpace(rule.Protocol)
	if protocol != "" && protocol != runtimecontracts.TopologyPolicyProtocolAny {
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

func applyInstanceACL(ctx context.Context, handle *runtimecontracts.InstanceRuntimeACLHandle, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	if handle == nil || strings.TrimSpace(handle.Chain) == "" {
		return fmt.Errorf("acl handle is required")
	}
	if len(rules) == 0 {
		return nil
	}
	if _, err := iptablesLookPath("iptables"); err != nil {
		return fmt.Errorf("iptables not found: %w", err)
	}

	chain := strings.TrimSpace(handle.Chain)

	validated, err := validateACLRules(rules)
	if err != nil {
		return err
	}

	// 确保实例级 chain 存在（新建，或 flush 后重用）。
	if err := ensureInstanceChain(ctx, chain); err != nil {
		return err
	}
	success := false
	defer func() {
		if success {
			return
		}
		_ = removeInstanceACL(ctx, &runtimecontracts.InstanceRuntimeACLHandle{Chain: chain})
	}()

	// 向 chain 写入规则。
	for _, rule := range validated {
		args := buildInstanceChainRuleArgs(chain, rule)
		if err := runACLCommand(ctx, args); err != nil {
			return err
		}
	}

	// 挂接 DOCKER-USER -> chain jump。
	if err := ensureInstanceChainJump(ctx, chain); err != nil {
		return err
	}
	success = true
	return nil
}

func removeInstanceACL(ctx context.Context, handle *runtimecontracts.InstanceRuntimeACLHandle) error {
	if handle == nil || strings.TrimSpace(handle.Chain) == "" {
		return nil
	}
	if _, err := iptablesLookPath("iptables"); err != nil {
		return fmt.Errorf("iptables not found: %w", err)
	}

	chain := strings.TrimSpace(handle.Chain)
	var firstErr error

	// 从 DOCKER-USER 删除 jump。
	if err := removeInstanceChainJump(ctx, chain); err != nil && firstErr == nil {
		firstErr = err
	}

	// flush chain。
	if err := flushInstanceChain(ctx, chain); err != nil && firstErr == nil {
		firstErr = err
	}

	// delete chain。
	if err := deleteInstanceChain(ctx, chain); err != nil && firstErr == nil {
		firstErr = err
	}

	return firstErr
}

func ensureInstanceChain(ctx context.Context, chain string) error {
	// 尝试创建 chain；如果已存在则 flush 重用。
	createArgs := []string{"-N", chain}
	if err := runACLCommand(ctx, createArgs); err != nil {
		// chain 已存在，flush 后继续
		return flushInstanceChain(ctx, chain)
	}
	return nil
}

func ensureInstanceChainJump(ctx context.Context, chain string) error {
	// 检查 jump 是否已存在，避免重复插入。
	checkArgs := []string{"-C", dockerUserChain, "-j", chain}
	if runACLCommand(ctx, checkArgs) == nil {
		return nil
	}
	jumpArgs := []string{"-I", dockerUserChain, "1", "-j", chain}
	return runACLCommand(ctx, jumpArgs)
}

func removeInstanceChainJump(ctx context.Context, chain string) error {
	jumpArgs := []string{"-D", dockerUserChain, "-j", chain}
	if err := runACLCommand(ctx, jumpArgs); err != nil {
		// jump 不存在不是错误
		if strings.Contains(err.Error(), "does a matching rule exist") {
			return nil
		}
		return err
	}
	return nil
}

func flushInstanceChain(ctx context.Context, chain string) error {
	flushArgs := []string{"-F", chain}
	if err := runACLCommand(ctx, flushArgs); err != nil {
		// chain 不存在不是错误
		if strings.Contains(err.Error(), "No chain/target/match by that name") {
			return nil
		}
		return err
	}
	return nil
}

func deleteInstanceChain(ctx context.Context, chain string) error {
	deleteArgs := []string{"-X", chain}
	if err := runACLCommand(ctx, deleteArgs); err != nil {
		// chain 不存在不是错误
		if strings.Contains(err.Error(), "No chain/target/match by that name") {
			return nil
		}
		return err
	}
	return nil
}

func buildInstanceChainRuleArgs(chain string, rule runtimecontracts.InstanceRuntimeACLRule) []string {
	args := []string{"-A", chain, "-s", rule.SourceIP, "-d", rule.TargetIP}
	protocol := strings.TrimSpace(rule.Protocol)
	if protocol != "" && protocol != runtimecontracts.TopologyPolicyProtocolAny {
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

// validateACLRules 对所有规则做白名单校验与 canonicalize，fail fast 模式。
func validateACLRules(rules []runtimecontracts.InstanceRuntimeACLRule) ([]runtimecontracts.InstanceRuntimeACLRule, error) {
	if len(rules) == 0 {
		return nil, nil
	}
	validated := make([]runtimecontracts.InstanceRuntimeACLRule, 0, len(rules))
	for i, rule := range rules {
		v, err := validateAndCanonicalizeACLRule(rule)
		if err != nil {
			return nil, fmt.Errorf("acl rule[%d] validation failed: %w", i, err)
		}
		validated = append(validated, v)
	}
	return validated, nil
}

func validateACLRulesForDelete(rules []runtimecontracts.InstanceRuntimeACLRule) ([]runtimecontracts.InstanceRuntimeACLRule, error) {
	if len(rules) == 0 {
		return nil, nil
	}
	validated := make([]runtimecontracts.InstanceRuntimeACLRule, 0, len(rules))
	for i, rule := range rules {
		v, err := validateACLRuleForDelete(rule)
		if err != nil {
			return nil, fmt.Errorf("acl rule[%d] delete validation failed: %w", i, err)
		}
		validated = append(validated, v)
	}
	return validated, nil
}

func validateAndCanonicalizeACLRule(rule runtimecontracts.InstanceRuntimeACLRule) (runtimecontracts.InstanceRuntimeACLRule, error) {
	return validateACLRule(rule, false)
}

func validateACLRuleForDelete(rule runtimecontracts.InstanceRuntimeACLRule) (runtimecontracts.InstanceRuntimeACLRule, error) {
	return validateACLRule(rule, true)
}

func validateACLRule(rule runtimecontracts.InstanceRuntimeACLRule, preserveComment bool) (runtimecontracts.InstanceRuntimeACLRule, error) {
	// SourceIP: 必须是单 IP。
	srcIP := strings.TrimSpace(rule.SourceIP)
	if srcIP == "" {
		return rule, fmt.Errorf("source ip is empty")
	}
	srcAddr, err := netip.ParseAddr(srcIP)
	if err != nil {
		return rule, fmt.Errorf("invalid source ip %q: %w", srcIP, err)
	}
	if !srcAddr.Is4() {
		return rule, fmt.Errorf("source ip %q is not ipv4", srcIP)
	}
	rule.SourceIP = srcAddr.String()

	// TargetIP: 必须是单 IP。
	dstIP := strings.TrimSpace(rule.TargetIP)
	if dstIP == "" {
		return rule, fmt.Errorf("target ip is empty")
	}
	dstAddr, err := netip.ParseAddr(dstIP)
	if err != nil {
		return rule, fmt.Errorf("invalid target ip %q: %w", dstIP, err)
	}
	if !dstAddr.Is4() {
		return rule, fmt.Errorf("target ip %q is not ipv4", dstIP)
	}
	rule.TargetIP = dstAddr.String()

	// Action: 白名单。
	action := strings.ToLower(strings.TrimSpace(rule.Action))
	switch action {
	case validACLActionAllow, validACLActionDeny:
		rule.Action = action
	default:
		return rule, fmt.Errorf("unsupported action %q", rule.Action)
	}

	// Protocol: 白名单。
	protocol := strings.ToLower(strings.TrimSpace(rule.Protocol))
	switch protocol {
	case "", validACLProtocolAny:
		protocol = validACLProtocolAny
		rule.Protocol = validACLProtocolAny
	case validACLProtocolTCP, validACLProtocolUDP:
		rule.Protocol = protocol
	default:
		return rule, fmt.Errorf("unsupported protocol %q", rule.Protocol)
	}

	// Ports: 1-65535, 去重排序, multiport 上限 15, protocol=any 禁止端口。
	if protocol == validACLProtocolAny && len(rule.Ports) > 0 {
		return rule, fmt.Errorf("protocol=any cannot carry ports")
	}

	if len(rule.Ports) > 0 {
		for _, port := range rule.Ports {
			if port < 1 || port > 65535 {
				return rule, fmt.Errorf("invalid port %d", port)
			}
		}
		deduped := deduplicateAndSortPorts(rule.Ports)
		if len(deduped) > maxMultiportCount {
			return rule, fmt.Errorf("port count %d exceeds multiport limit %d", len(deduped), maxMultiportCount)
		}
		rule.Ports = deduped
	}

	if preserveComment {
		rule.Comment = strings.TrimSpace(rule.Comment)
		if rule.Comment == "" {
			rule.Comment = systemACLComment(rule)
		}
	} else {
		// Comment: 统一系统重建，不信任持久化值。
		rule.Comment = systemACLComment(rule)
	}
	if err := validateACLComment(rule.Comment); err != nil {
		return rule, err
	}

	return rule, nil
}

func deduplicateAndSortPorts(ports []int) []int {
	if len(ports) == 0 {
		return nil
	}
	seen := make(map[int]struct{}, len(ports))
	unique := make([]int, 0, len(ports))
	for _, port := range ports {
		if _, exists := seen[port]; exists {
			continue
		}
		seen[port] = struct{}{}
		unique = append(unique, port)
	}
	sort.Ints(unique)
	return unique
}

func systemACLComment(rule runtimecontracts.InstanceRuntimeACLRule) string {
	parts := []string{
		"ctf:acl",
		rule.SourceIP,
		rule.TargetIP,
		rule.Action,
		rule.Protocol,
	}
	if len(rule.Ports) > 0 {
		ports := make([]string, 0, len(rule.Ports))
		for _, p := range rule.Ports {
			ports = append(ports, strconv.Itoa(p))
		}
		parts = append(parts, strings.Join(ports, ","))
	}
	return strings.Join(parts, ":")
}

func validateACLComment(comment string) error {
	if comment == "" {
		return nil
	}
	if !strings.HasPrefix(comment, "ctf:acl:") {
		return fmt.Errorf("unsupported comment prefix")
	}
	for _, r := range comment {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == ':', r == '.', r == ',', r == '-', r == '_':
		default:
			return fmt.Errorf("comment contains unsupported character %q", r)
		}
	}
	return nil
}
