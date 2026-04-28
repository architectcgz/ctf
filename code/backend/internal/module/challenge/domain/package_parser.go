package domain

import (
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

func ParseChallengePackageDir(rootDir string) (*ParsedChallengePackage, error) {
	manifestPath := filepath.Join(rootDir, "challenge.yml")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read challenge.yml %s: %w", manifestPath, err)
	}

	var manifest ChallengePackageManifest
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("parse challenge.yml %s: %w", manifestPath, err)
	}

	return buildParsedChallengePackage(rootDir, &manifest, string(content))
}

func buildParsedChallengePackage(rootDir string, manifest *ChallengePackageManifest, manifestRaw string) (*ParsedChallengePackage, error) {
	if manifest == nil {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("challenge.yml 不能为空"))
	}
	if strings.TrimSpace(manifest.APIVersion) != "v1" {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("challenge.yml api_version 仅支持 v1"))
	}
	if strings.TrimSpace(manifest.Kind) != "challenge" {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("challenge.yml kind 必须为 challenge"))
	}
	if strings.EqualFold(strings.TrimSpace(manifest.Meta.Mode), "awd") {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("AWD 题目包请使用 AWD 服务模板导入入口"))
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

	attachments, err := resolvePackageAttachments(rootDir, manifest.Content.Attachments)
	if err != nil {
		return nil, err
	}
	hints, err := resolvePackageHints(manifest.Hints)
	if err != nil {
		return nil, err
	}

	flagType := strings.ToLower(strings.TrimSpace(manifest.Flag.Type))
	switch flagType {
	case model.FlagTypeStatic, model.FlagTypeDynamic, model.FlagTypeRegex, model.FlagTypeManualReview:
	default:
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("flag.type 仅支持 static、dynamic、regex 或 manual_review"))
	}

	flagPrefix := strings.TrimSpace(manifest.Flag.Prefix)
	if flagPrefix == "" {
		flagPrefix = "flag"
	}

	flagValue := strings.TrimSpace(manifest.Flag.Value)
	if (flagType == model.FlagTypeStatic || flagType == model.FlagTypeRegex) && flagValue == "" {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("static/regex 题目必须提供 flag.value"))
	}

	points := manifest.Meta.Points
	if points <= 0 {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("meta.points 必须大于 0"))
	}

	topology, err := parseChallengePackageTopology(rootDir, manifest.Extensions.Topology)
	if err != nil {
		return nil, err
	}
	packageFiles, err := listChallengePackageFiles(rootDir)
	if err != nil {
		return nil, fmt.Errorf("list package files: %w", err)
	}

	parsed := &ParsedChallengePackage{
		Manifest:        *manifest,
		ManifestRaw:     manifestRaw,
		RootDir:         rootDir,
		Slug:            slug,
		Title:           title,
		Description:     description,
		Category:        normalizePackageCategory(manifest.Meta.Category),
		Difficulty:      normalizePackageDifficulty(manifest.Meta.Difficulty),
		Points:          points,
		FlagType:        flagType,
		FlagValue:       flagValue,
		FlagPrefix:      flagPrefix,
		RuntimeImageRef: resolvePackageRuntimeImageRef(manifest.Runtime),
		RuntimeProtocol: normalizePackageRuntimeProtocol(manifest.Runtime.Service.Protocol),
		RuntimePort:     normalizePackageRuntimePort(manifest.Runtime.Service.Port),
		Attachments:     attachments,
		Hints:           hints,
		Topology:        topology,
		PackageFiles:    packageFiles,
	}
	return parsed, nil
}

func resolvePackageAttachments(rootDir string, attachments []ChallengePackageAttachment) ([]ParsedChallengePackageAttachment, error) {
	if len(attachments) == 0 {
		return nil, nil
	}

	parsed := make([]ParsedChallengePackageAttachment, 0, len(attachments))
	for _, attachment := range attachments {
		relPath := strings.TrimSpace(attachment.Path)
		if relPath == "" {
			return nil, errcode.ErrInvalidParams.WithCause(errors.New("content.attachments.path 不能为空"))
		}
		absolutePath, err := safePackageJoin(rootDir, relPath)
		if err != nil {
			return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("附件路径非法: %w", err))
		}
		info, err := os.Stat(absolutePath)
		if err != nil {
			return nil, fmt.Errorf("attachment not found %s: %w", relPath, err)
		}
		if info.IsDir() {
			return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("附件必须是文件: %s", relPath))
		}

		name := strings.TrimSpace(attachment.Name)
		if name == "" {
			name = filepath.Base(absolutePath)
		}

		parsed = append(parsed, ParsedChallengePackageAttachment{
			Path:         filepath.ToSlash(filepath.Clean(relPath)),
			Name:         name,
			AbsolutePath: absolutePath,
		})
	}
	return parsed, nil
}

func resolvePackageHints(hints []ChallengePackageHint) ([]ParsedChallengePackageHint, error) {
	if len(hints) == 0 {
		return nil, nil
	}

	levels := make(map[int]struct{}, len(hints))
	parsed := make([]ParsedChallengePackageHint, 0, len(hints))
	for index, hint := range hints {
		content := strings.TrimSpace(hint.Content)
		if content == "" {
			return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("第 %d 个提示内容不能为空", index+1))
		}
		level := hint.Level
		if level <= 0 {
			level = len(parsed) + 1
		}
		if _, exists := levels[level]; exists {
			return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("提示级别重复: %d", level))
		}
		levels[level] = struct{}{}

		title := strings.TrimSpace(hint.Title)
		if title == "" {
			title = fmt.Sprintf("Hint %d", level)
		}
		parsed = append(parsed, ParsedChallengePackageHint{
			Level:   level,
			Title:   title,
			Content: content,
		})
	}

	sort.Slice(parsed, func(i, j int) bool {
		return parsed[i].Level < parsed[j].Level
	})
	return parsed, nil
}

func resolvePackageRuntimeImageRef(runtime ChallengePackageRuntime) string {
	if strings.TrimSpace(runtime.Type) != "container" {
		return ""
	}

	if ref := strings.TrimSpace(runtime.Image.Ref); ref != "" {
		return ref
	}

	name := strings.TrimSpace(runtime.Image.Name)
	if name == "" {
		return ""
	}
	tag := strings.TrimSpace(runtime.Image.Tag)
	if tag == "" {
		tag = "latest"
	}
	return fmt.Sprintf("%s:%s", name, tag)
}

func normalizePackageCategory(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "web", "pwn", "reverse", "crypto", "misc", "forensics":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return "misc"
	}
}

func normalizePackageDifficulty(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case model.ChallengeDifficultyBeginner:
		return model.ChallengeDifficultyBeginner
	case model.ChallengeDifficultyEasy:
		return model.ChallengeDifficultyEasy
	case model.ChallengeDifficultyMedium:
		return model.ChallengeDifficultyMedium
	case model.ChallengeDifficultyHard:
		return model.ChallengeDifficultyHard
	case model.ChallengeDifficultyInsane:
		return model.ChallengeDifficultyInsane
	default:
		return model.ChallengeDifficultyEasy
	}
}

func normalizePackageRuntimeProtocol(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case model.ChallengeTargetProtocolTCP:
		return model.ChallengeTargetProtocolTCP
	default:
		return model.ChallengeTargetProtocolHTTP
	}
}

func normalizePackageRuntimePort(port int) int {
	if port <= 0 || port > 65535 {
		return 0
	}
	return port
}

func safePackageJoin(baseDir, rel string) (string, error) {
	if strings.TrimSpace(rel) == "" {
		return "", fmt.Errorf("relative path is empty")
	}

	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}

	target := filepath.Clean(filepath.Join(baseAbs, rel))
	prefix := baseAbs + string(os.PathSeparator)
	if target != baseAbs && !strings.HasPrefix(target, prefix) {
		return "", fmt.Errorf("path escapes package root: %s", rel)
	}
	return target, nil
}
