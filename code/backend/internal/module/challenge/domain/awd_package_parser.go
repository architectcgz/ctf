package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
	"gopkg.in/yaml.v3"
)

func ParseAWDChallengePackageDir(rootDir string) (*ParsedAWDChallengePackage, error) {
	manifestPath := filepath.Join(rootDir, "challenge.yml")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read challenge.yml %s: %w", manifestPath, err)
	}

	var manifest ChallengePackageManifest
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("parse challenge.yml %s: %w", manifestPath, err)
	}

	return buildParsedAWDChallengePackage(rootDir, &manifest)
}

func buildParsedAWDChallengePackage(
	rootDir string,
	manifest *ChallengePackageManifest,
) (*ParsedAWDChallengePackage, error) {
	if manifest == nil {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("challenge.yml 不能为空"))
	}
	if strings.TrimSpace(manifest.APIVersion) != "v1" {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("challenge.yml api_version 仅支持 v1"))
	}
	if strings.TrimSpace(manifest.Kind) != "challenge" {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("challenge.yml kind 必须为 challenge"))
	}
	if !strings.EqualFold(strings.TrimSpace(manifest.Meta.Mode), "awd") {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("AWD 题目包必须声明 meta.mode = awd"))
	}

	slug := strings.TrimSpace(manifest.Meta.Slug)
	if slug == "" {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("challenge.yml meta.slug 不能为空"))
	}
	title := strings.TrimSpace(manifest.Meta.Title)
	if title == "" {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("challenge.yml meta.title 不能为空"))
	}
	if err := validateAWDPackageDockerfileLayout(rootDir); err != nil {
		return nil, err
	}

	statementFile := strings.TrimSpace(manifest.Content.Statement)
	if statementFile == "" {
		statementFile = "statement.md"
	}
	statementPath, err := safePackageJoin(rootDir, statementFile)
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("题面路径非法: %w", err))
	}
	statementBytes, err := os.ReadFile(statementPath)
	if err != nil {
		return nil, fmt.Errorf("read statement %s: %w", statementPath, err)
	}
	description := strings.TrimSpace(string(statementBytes))
	if description == "" {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("题面内容不能为空"))
	}

	runtimeImageRef := resolvePackageRuntimeImageRef(manifest.Runtime)
	imageSourceType, suggestedImageTag, dockerfilePath, buildContextPath := resolveAWDPackageImageSource(rootDir, manifest.Runtime)
	if strings.TrimSpace(manifest.Runtime.Type) != "container" {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("AWD 题目包必须提供 runtime.type=container"))
	}
	if runtimeImageRef == "" && dockerfilePath == "" {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("AWD 题目包必须提供 runtime.image.ref 或 docker/runtime/Dockerfile"))
	}

	awd := manifest.Extensions.AWD
	serviceType := strings.TrimSpace(awd.ServiceType)
	switch serviceType {
	case string(model.AWDServiceTypeWebHTTP), string(model.AWDServiceTypeBinaryTCP), string(model.AWDServiceTypeMultiContainer):
	default:
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("extensions.awd.service_type 仅支持 web_http、binary_tcp、multi_container"))
	}

	deploymentMode := strings.TrimSpace(awd.DeploymentMode)
	switch deploymentMode {
	case string(model.AWDDeploymentModeSingleContainer), string(model.AWDDeploymentModeTopology):
	default:
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("extensions.awd.deployment_mode 仅支持 single_container、topology"))
	}

	checkerType := strings.TrimSpace(awd.Checker.Type)
	switch checkerType {
	case string(model.AWDCheckerTypeLegacyProbe), string(model.AWDCheckerTypeHTTPStandard), string(model.AWDCheckerTypeTCPStandard), string(model.AWDCheckerTypeScript):
	default:
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("extensions.awd.checker.type 仅支持 legacy_probe、http_standard、tcp_standard、script_checker"))
	}

	flagMode := strings.TrimSpace(awd.FlagPolicy.Mode)
	if flagMode == "" {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("extensions.awd.flag_policy.mode 不能为空"))
	}

	defenseEntryMode := strings.TrimSpace(awd.DefenseEntry.Mode)
	if defenseEntryMode == "" {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("extensions.awd.defense_entry.mode 不能为空"))
	}

	checkerConfig := normalizePackageConfigMap(awd.Checker.Config)
	checkerEntryPath, checkerEntryAbs, checkerFiles, err := resolveAWDPackageCheckerFiles(rootDir, checkerType, checkerConfig)
	if err != nil {
		return nil, err
	}
	flagConfig := normalizePackageConfigMap(awd.FlagPolicy.Config)
	accessConfig := normalizePackageConfigMap(awd.AccessConfig)
	runtimeConfig := normalizePackageConfigMap(awd.RuntimeConfig)
	if err := validateAWDPackageDefenseWorkspace(rootDir, runtimeConfig); err != nil {
		return nil, err
	}
	if err := validateAWDPackageDefenseScope(rootDir, runtimeConfig); err != nil {
		return nil, err
	}

	if len(accessConfig) == 0 {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("extensions.awd.access_config 不能为空"))
	}

	warnings := make([]string, 0, 1)
	if manifest.Meta.Points > 0 {
		warnings = append(warnings, "meta.points 仅作为建议分值，不会直接写入 AWD 题目。")
	}

	version := strings.TrimSpace(awd.Version)
	if version == "" {
		version = "v1"
	}

	return &ParsedAWDChallengePackage{
		Manifest:          *manifest,
		RootDir:           rootDir,
		Slug:              slug,
		Title:             title,
		Description:       description,
		Category:          normalizePackageCategory(manifest.Meta.Category),
		Difficulty:        normalizePackageDifficulty(manifest.Meta.Difficulty),
		SuggestedPoints:   manifest.Meta.Points,
		RuntimeImageRef:   runtimeImageRef,
		ImageSourceType:   imageSourceType,
		SuggestedImageTag: suggestedImageTag,
		DockerfilePath:    dockerfilePath,
		BuildContextPath:  buildContextPath,
		ServiceType:       serviceType,
		DeploymentMode:    deploymentMode,
		Version:           version,
		CheckerType:       checkerType,
		CheckerConfig:     checkerConfig,
		CheckerEntryPath:  checkerEntryPath,
		CheckerEntryAbs:   checkerEntryAbs,
		CheckerFiles:      checkerFiles,
		FlagMode:          flagMode,
		FlagConfig:        flagConfig,
		DefenseEntryMode:  defenseEntryMode,
		AccessConfig:      accessConfig,
		RuntimeConfig:     runtimeConfig,
		Warnings:          warnings,
	}, nil
}

func resolveAWDPackageCheckerFiles(rootDir, checkerType string, checkerConfig map[string]any) (string, string, []ParsedAWDCheckerFile, error) {
	if checkerType != string(model.AWDCheckerTypeScript) {
		return "", "", nil, nil
	}
	rawEntry, ok := checkerConfig["entry"].(string)
	entry := strings.TrimSpace(rawEntry)
	if !ok || entry == "" {
		return "", "", nil, errcode.ErrInvalidParams.WithCause(errors.New("script_checker config.entry 不能为空"))
	}
	entryPath, entryAbs, err := resolveAWDPackageCheckerFile(rootDir, entry, "script_checker config.entry")
	if err != nil {
		return "", "", nil, err
	}

	fileValues := []string{entryPath}
	if rawFiles, ok := checkerConfig["files"]; ok {
		values, err := readAWDPackageCheckerFileList(rawFiles)
		if err != nil {
			return "", "", nil, err
		}
		fileValues = values
	}

	seen := map[string]bool{}
	files := make([]ParsedAWDCheckerFile, 0, len(fileValues))
	entryIncluded := false
	for _, value := range fileValues {
		pathValue, absValue, err := resolveAWDPackageCheckerFile(rootDir, value, "script_checker config.files")
		if err != nil {
			return "", "", nil, err
		}
		if seen[pathValue] {
			continue
		}
		seen[pathValue] = true
		if pathValue == entryPath {
			entryIncluded = true
		}
		files = append(files, ParsedAWDCheckerFile{Path: pathValue, Abs: absValue})
	}
	if !entryIncluded {
		return "", "", nil, errcode.ErrInvalidParams.WithCause(errors.New("script_checker config.entry 必须包含在 config.files 中"))
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return entryPath, entryAbs, files, nil
}

type awdPackageRuntimeMount struct {
	Source string
	Target string
	Mode   string
}

func validateAWDPackageDockerfileLayout(rootDir string) error {
	files, err := listChallengePackageFiles(rootDir)
	if err != nil {
		return fmt.Errorf("list package files: %w", err)
	}
	for _, item := range files {
		if filepath.Base(item.Path) != "Dockerfile" {
			continue
		}
		if item.Path != "docker/runtime/Dockerfile" {
			return errcode.ErrInvalidParams.WithCause(
				fmt.Errorf("AWD 题目 Dockerfile 必须位于 docker/runtime/Dockerfile，当前为 %s", item.Path),
			)
		}
	}
	return nil
}

func resolveAWDPackageImageSource(rootDir string, runtime ChallengePackageRuntime) (string, string, string, string) {
	runtimeImageRef := resolvePackageRuntimeImageRef(runtime)
	suggestedTag := ExtractImageTagSuggestion(runtimeImageRef, runtime.Image.Tag)
	if suggestedTag == "" && strings.TrimSpace(runtimeImageRef) != "" {
		suggestedTag = "latest"
	}

	dockerfilePath := filepath.Join(rootDir, "docker", "runtime", "Dockerfile")
	info, err := os.Stat(dockerfilePath)
	if err == nil && !info.IsDir() {
		return ImageSourceTypePlatformBuild, suggestedTag, dockerfilePath, filepath.Join(rootDir, "docker")
	}
	if strings.TrimSpace(runtimeImageRef) != "" {
		return ImageSourceTypeExternalRef, suggestedTag, "", ""
	}
	return "", suggestedTag, "", ""
}

func validateAWDPackageDefenseWorkspace(rootDir string, runtimeConfig map[string]any) error {
	raw, ok := runtimeConfig["defense_workspace"]
	if !ok {
		return errcode.ErrInvalidParams.WithCause(errors.New("extensions.awd.runtime_config.defense_workspace 不能为空"))
	}
	payload, ok := raw.(map[string]any)
	if !ok {
		return errcode.ErrInvalidParams.WithCause(errors.New("extensions.awd.runtime_config.defense_workspace 必须是对象"))
	}

	entryMode, err := readAWDPackageRequiredString(payload, "entry_mode", "defense_workspace.entry_mode")
	if err != nil {
		return err
	}
	if entryMode != "ssh" {
		return errcode.ErrInvalidParams.WithCause(errors.New("defense_workspace.entry_mode 当前仅支持 ssh"))
	}

	seedRoot, err := readAWDPackageDefenseWorkspaceDir(rootDir, payload["seed_root"], "defense_workspace.seed_root")
	if err != nil {
		return err
	}
	if seedRoot == "docker" {
		return errcode.ErrInvalidParams.WithCause(errors.New("defense_workspace.seed_root 不能指向 docker 根目录"))
	}

	workspaceRoots, err := readAWDPackageDefenseWorkspaceDirList(rootDir, payload["workspace_roots"], "defense_workspace.workspace_roots")
	if err != nil {
		return err
	}
	writableRoots, err := readAWDPackageDefenseWorkspaceDirList(rootDir, payload["writable_roots"], "defense_workspace.writable_roots")
	if err != nil {
		return err
	}
	readonlyRoots, err := readAWDPackageDefenseWorkspaceDirList(rootDir, payload["readonly_roots"], "defense_workspace.readonly_roots")
	if err != nil {
		return err
	}

	workspaceRootSet := make(map[string]struct{}, len(workspaceRoots))
	for _, root := range workspaceRoots {
		if err := ensureAWDPackageDefenseWorkspaceInsideSeedRoot(seedRoot, root, "defense_workspace.workspace_roots"); err != nil {
			return err
		}
		workspaceRootSet[root] = struct{}{}
	}

	writableRootSet := make(map[string]struct{}, len(writableRoots))
	for _, root := range writableRoots {
		if err := ensureAWDPackageDefenseWorkspaceInsideSeedRoot(seedRoot, root, "defense_workspace.writable_roots"); err != nil {
			return err
		}
		if _, ok := workspaceRootSet[root]; !ok {
			return errcode.ErrInvalidParams.WithCause(
				fmt.Errorf("defense_workspace.workspace_roots 必须覆盖 writable_roots 与 readonly_roots: %s", root),
			)
		}
		writableRootSet[root] = struct{}{}
	}
	readonlyRootSet := make(map[string]struct{}, len(readonlyRoots))
	for _, root := range readonlyRoots {
		if err := ensureAWDPackageDefenseWorkspaceInsideSeedRoot(seedRoot, root, "defense_workspace.readonly_roots"); err != nil {
			return err
		}
		if _, ok := workspaceRootSet[root]; !ok {
			return errcode.ErrInvalidParams.WithCause(
				fmt.Errorf("defense_workspace.workspace_roots 必须覆盖 writable_roots 与 readonly_roots: %s", root),
			)
		}
		if _, exists := writableRootSet[root]; exists {
			return errcode.ErrInvalidParams.WithCause(
				fmt.Errorf("defense_workspace.writable_roots 与 readonly_roots 不能重叠: %s", root),
			)
		}
		readonlyRootSet[root] = struct{}{}
	}

	runtimeMounts, err := readAWDPackageRuntimeMounts(rootDir, payload["runtime_mounts"])
	if err != nil {
		return err
	}
	for _, mount := range runtimeMounts {
		if _, ok := workspaceRootSet[mount.Source]; !ok {
			return errcode.ErrInvalidParams.WithCause(
				fmt.Errorf("defense_workspace.runtime_mounts.source 必须来自 workspace_roots: %s", mount.Source),
			)
		}
	}
	return nil
}

func validateAWDPackageDefenseScope(rootDir string, runtimeConfig map[string]any) error {
	raw, ok := runtimeConfig["defense_scope"]
	if !ok {
		return nil
	}
	scope, ok := raw.(map[string]any)
	if !ok {
		return errcode.ErrInvalidParams.WithCause(errors.New("extensions.awd.runtime_config.defense_scope 必须是对象"))
	}
	if _, exists := scope["editable_paths"]; exists {
		return errcode.ErrInvalidParams.WithCause(errors.New("defense_scope.editable_paths 已废弃，请改用 defense_workspace"))
	}

	protectedPaths, err := readAWDPackageDefenseScopePathList(scope["protected_paths"], "protected_paths")
	if err != nil {
		return err
	}
	if _, err := readAWDPackageDefenseScopeStringList(scope["service_contracts"], "service_contracts"); err != nil {
		return err
	}

	protectedSet := make(map[string]struct{}, len(protectedPaths))
	for _, item := range protectedPaths {
		if err := validateAWDPackageExistingPath(rootDir, item, "defense_scope.protected_paths"); err != nil {
			return err
		}
		protectedSet[item] = struct{}{}
	}

	requiredProtected := []string{"docker/runtime/app.py", "docker/runtime/ctf_runtime.py", "docker/check/check.py", "challenge.yml"}
	for _, required := range requiredProtected {
		if _, exists := protectedSet[required]; !exists {
			return errcode.ErrInvalidParams.WithCause(fmt.Errorf("defense_scope.protected_paths 必须包含 %s", required))
		}
	}
	return nil
}

func readAWDPackageRequiredString(payload map[string]any, key, field string) (string, error) {
	value, ok := payload[key].(string)
	if !ok || strings.TrimSpace(value) == "" {
		return "", errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 不能为空", field))
	}
	return strings.TrimSpace(value), nil
}

func readAWDPackageDefenseWorkspaceDirList(rootDir string, raw any, field string) ([]string, error) {
	items, err := readAWDPackageStringList(raw, field)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		clean, err := readAWDPackageDefenseWorkspaceDir(rootDir, item, field)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[clean]; exists {
			return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 存在重复路径: %s", field, clean))
		}
		seen[clean] = struct{}{}
		paths = append(paths, clean)
	}
	return paths, nil
}

func readAWDPackageDefenseWorkspaceDir(rootDir string, raw any, field string) (string, error) {
	value, ok := raw.(string)
	if !ok || strings.TrimSpace(value) == "" {
		return "", errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 不能为空", field))
	}
	clean, err := normalizeAWDPackageRelativePath(strings.TrimSpace(value), field)
	if err != nil {
		return "", err
	}
	if isAWDDefenseWorkspaceProtectedPath(clean) {
		return "", errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 不能包含受保护路径: %s", field, clean))
	}
	abs, err := safePackageJoin(rootDir, clean)
	if err != nil {
		return "", errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 路径非法: %w", field, err))
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 指向的目录不存在: %s", field, clean))
	}
	if !info.IsDir() {
		return "", errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 必须指向目录: %s", field, clean))
	}
	return clean, nil
}

func ensureAWDPackageDefenseWorkspaceInsideSeedRoot(seedRoot, value, field string) error {
	if value == seedRoot || strings.HasPrefix(value, seedRoot+"/") {
		return nil
	}
	return errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 必须位于 seed_root 内: %s", field, value))
}

func readAWDPackageRuntimeMounts(rootDir string, raw any) ([]awdPackageRuntimeMount, error) {
	rawList, ok := raw.([]any)
	if !ok || len(rawList) == 0 {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("defense_workspace.runtime_mounts 必须是非空数组"))
	}
	mounts := make([]awdPackageRuntimeMount, 0, len(rawList))
	seenSources := make(map[string]struct{}, len(rawList))
	for _, item := range rawList {
		payload, ok := item.(map[string]any)
		if !ok {
			return nil, errcode.ErrInvalidParams.WithCause(errors.New("defense_workspace.runtime_mounts 必须是对象数组"))
		}
		source, err := readAWDPackageDefenseWorkspaceDir(rootDir, payload["source"], "defense_workspace.runtime_mounts.source")
		if err != nil {
			return nil, err
		}
		target, err := readAWDPackageRequiredString(payload, "target", "defense_workspace.runtime_mounts.target")
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(target, "/") {
			return nil, errcode.ErrInvalidParams.WithCause(
				fmt.Errorf("defense_workspace.runtime_mounts.target 必须是绝对路径: %s", target),
			)
		}
		mode, err := readAWDPackageRequiredString(payload, "mode", "defense_workspace.runtime_mounts.mode")
		if err != nil {
			return nil, err
		}
		if mode != "rw" && mode != "ro" {
			return nil, errcode.ErrInvalidParams.WithCause(
				fmt.Errorf("defense_workspace.runtime_mounts.mode 仅支持 rw 或 ro: %s", mode),
			)
		}
		if _, exists := seenSources[source]; exists {
			return nil, errcode.ErrInvalidParams.WithCause(
				fmt.Errorf("defense_workspace.runtime_mounts.source 存在重复路径: %s", source),
			)
		}
		seenSources[source] = struct{}{}
		mounts = append(mounts, awdPackageRuntimeMount{
			Source: source,
			Target: target,
			Mode:   mode,
		})
	}
	return mounts, nil
}

func readAWDPackageDefenseScopePathList(raw any, field string) ([]string, error) {
	items, err := readAWDPackageStringList(raw, "defense_scope."+field)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		clean, err := normalizeAWDPackageRelativePath(item, "defense_scope."+field)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[clean]; exists {
			return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("defense_scope.%s 存在重复路径: %s", field, clean))
		}
		seen[clean] = struct{}{}
		paths = append(paths, clean)
	}
	return paths, nil
}

func readAWDPackageDefenseScopeStringList(raw any, field string) ([]string, error) {
	return readAWDPackageStringList(raw, "defense_scope."+field)
}

func readAWDPackageStringList(raw any, field string) ([]string, error) {
	rawList, ok := raw.([]any)
	if !ok {
		return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 必须是非空字符串数组", field))
	}
	values := make([]string, 0, len(rawList))
	for _, item := range rawList {
		value, ok := item.(string)
		if !ok || strings.TrimSpace(value) == "" {
			return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 必须是非空字符串数组", field))
		}
		values = append(values, strings.TrimSpace(value))
	}
	if len(values) == 0 {
		return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 不能为空", field))
	}
	return values, nil
}

func normalizeAWDPackageRelativePath(value, field string) (string, error) {
	clean := filepath.ToSlash(filepath.Clean(value))
	if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 必须是题目包内相对路径", field))
	}
	return clean, nil
}

func validateAWDPackageExistingPath(rootDir, value, field string) error {
	abs, err := safePackageJoin(rootDir, value)
	if err != nil {
		return errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 路径非法: %w", field, err))
	}
	if _, err := os.Stat(abs); err != nil {
		return errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 指向的路径不存在: %s", field, value))
	}
	return nil
}

func isAWDDefenseWorkspaceProtectedPath(value string) bool {
	switch value {
	case "challenge.yml", "docker/runtime", "docker/check":
		return true
	}
	return strings.HasPrefix(value, "docker/runtime/") || strings.HasPrefix(value, "docker/check/")
}

func readAWDPackageCheckerFileList(raw any) ([]string, error) {
	rawList, ok := raw.([]any)
	if !ok {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("script_checker config.files 必须是字符串数组"))
	}
	values := make([]string, 0, len(rawList))
	for _, item := range rawList {
		value, ok := item.(string)
		if !ok || strings.TrimSpace(value) == "" {
			return nil, errcode.ErrInvalidParams.WithCause(errors.New("script_checker config.files 必须是非空字符串数组"))
		}
		values = append(values, value)
	}
	return values, nil
}

func resolveAWDPackageCheckerFile(rootDir, value, fieldName string) (string, string, error) {
	clean := filepath.Clean(strings.TrimSpace(value))
	if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", "", errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 必须是题目包内相对文件路径", fieldName))
	}
	abs, err := safePackageJoin(rootDir, clean)
	if err != nil {
		return "", "", errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 路径非法: %w", fieldName, err))
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", "", fmt.Errorf("read %s %s: %w", fieldName, abs, err)
	}
	if info.IsDir() {
		return "", "", errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 不能是目录", fieldName))
	}
	return filepath.ToSlash(clean), abs, nil
}

func normalizePackageConfigMap(raw map[string]any) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return map[string]any{}
	}
	var normalized map[string]any
	if err := json.Unmarshal(encoded, &normalized); err != nil {
		return map[string]any{}
	}
	if normalized == nil {
		return map[string]any{}
	}
	return normalized
}
