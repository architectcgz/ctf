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
	if err := validatePackageDockerfileLayout(rootDir); err != nil {
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
	imageSourceType, suggestedImageTag, dockerfilePath, buildContextPath := resolvePackageImageSource(rootDir, manifest.Runtime)
	if strings.TrimSpace(manifest.Runtime.Type) != "container" {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("AWD 题目包必须提供 runtime.type=container"))
	}
	if runtimeImageRef == "" && dockerfilePath == "" {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("AWD 题目包必须提供 runtime.image.ref 或 docker/Dockerfile"))
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

func validateAWDPackageDefenseScope(rootDir string, runtimeConfig map[string]any) error {
	raw, ok := runtimeConfig["defense_scope"]
	if !ok {
		return errcode.ErrInvalidParams.WithCause(errors.New("extensions.awd.runtime_config.defense_scope 不能为空"))
	}
	scope, ok := raw.(map[string]any)
	if !ok {
		return errcode.ErrInvalidParams.WithCause(errors.New("extensions.awd.runtime_config.defense_scope 必须是对象"))
	}

	editablePaths, err := readAWDPackageDefenseScopePathList(scope["editable_paths"], "editable_paths")
	if err != nil {
		return err
	}
	protectedPaths, err := readAWDPackageDefenseScopePathList(scope["protected_paths"], "protected_paths")
	if err != nil {
		return err
	}
	if _, err := readAWDPackageDefenseScopeStringList(scope["service_contracts"], "service_contracts"); err != nil {
		return err
	}

	if len(editablePaths) == 0 {
		return errcode.ErrInvalidParams.WithCause(errors.New("defense_scope.editable_paths 不能为空"))
	}
	if len(protectedPaths) == 0 {
		return errcode.ErrInvalidParams.WithCause(errors.New("defense_scope.protected_paths 不能为空"))
	}

	editableSet := make(map[string]struct{}, len(editablePaths))
	for _, item := range editablePaths {
		if err := validateAWDPackageDefenseScopeFile(rootDir, item, "defense_scope.editable_paths"); err != nil {
			return err
		}
		editableSet[item] = struct{}{}
	}
	protectedSet := make(map[string]struct{}, len(protectedPaths))
	for _, item := range protectedPaths {
		if err := validateAWDPackageDefenseScopeFile(rootDir, item, "defense_scope.protected_paths"); err != nil {
			return err
		}
		if _, exists := editableSet[item]; exists {
			return errcode.ErrInvalidParams.WithCause(fmt.Errorf("defense_scope 路径不能同时出现在 editable_paths 和 protected_paths: %s", item))
		}
		protectedSet[item] = struct{}{}
	}

	requiredProtected := []string{"docker/app.py", "docker/ctf_runtime.py", "docker/check/check.py", "challenge.yml"}
	for _, required := range requiredProtected {
		if _, exists := protectedSet[required]; !exists {
			return errcode.ErrInvalidParams.WithCause(fmt.Errorf("defense_scope.protected_paths 必须包含 %s", required))
		}
	}
	return nil
}

func readAWDPackageDefenseScopePathList(raw any, field string) ([]string, error) {
	items, err := readAWDPackageDefenseScopeStringList(raw, field)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		clean := filepath.ToSlash(filepath.Clean(item))
		if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
			return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("defense_scope.%s 必须是题目包内相对文件路径", field))
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
	rawList, ok := raw.([]any)
	if !ok {
		return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("defense_scope.%s 必须是非空字符串数组", field))
	}
	values := make([]string, 0, len(rawList))
	for _, item := range rawList {
		value, ok := item.(string)
		if !ok || strings.TrimSpace(value) == "" {
			return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("defense_scope.%s 必须是非空字符串数组", field))
		}
		values = append(values, strings.TrimSpace(value))
	}
	if len(values) == 0 {
		return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("defense_scope.%s 不能为空", field))
	}
	return values, nil
}

func validateAWDPackageDefenseScopeFile(rootDir, value, field string) error {
	abs, err := safePackageJoin(rootDir, value)
	if err != nil {
		return errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 路径非法: %w", field, err))
	}
	info, err := os.Stat(abs)
	if err != nil {
		return errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 指向的文件不存在: %s", field, value))
	}
	if info.IsDir() {
		return errcode.ErrInvalidParams.WithCause(fmt.Errorf("%s 不能指向目录: %s", field, value))
	}
	return nil
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
