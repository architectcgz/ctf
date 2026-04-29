package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	if strings.TrimSpace(manifest.Runtime.Type) != "container" || runtimeImageRef == "" {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("AWD 题目包必须提供 runtime.type=container 和 runtime.image.ref"))
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
	case string(model.AWDCheckerTypeLegacyProbe), string(model.AWDCheckerTypeHTTPStandard):
	default:
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("extensions.awd.checker.type 仅支持 legacy_probe、http_standard"))
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
	flagConfig := normalizePackageConfigMap(awd.FlagPolicy.Config)
	accessConfig := normalizePackageConfigMap(awd.AccessConfig)
	runtimeConfig := normalizePackageConfigMap(awd.RuntimeConfig)

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
		Manifest:         *manifest,
		RootDir:          rootDir,
		Slug:             slug,
		Title:            title,
		Description:      description,
		Category:         normalizePackageCategory(manifest.Meta.Category),
		Difficulty:       normalizePackageDifficulty(manifest.Meta.Difficulty),
		SuggestedPoints:  manifest.Meta.Points,
		RuntimeImageRef:  runtimeImageRef,
		ServiceType:      serviceType,
		DeploymentMode:   deploymentMode,
		Version:          version,
		CheckerType:      checkerType,
		CheckerConfig:    checkerConfig,
		FlagMode:         flagMode,
		FlagConfig:       flagConfig,
		DefenseEntryMode: defenseEntryMode,
		AccessConfig:     accessConfig,
		RuntimeConfig:    runtimeConfig,
		Warnings:         warnings,
	}, nil
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
