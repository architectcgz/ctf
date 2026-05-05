package domain

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"ctf-platform/internal/model"
)

const (
	ChallengePackageModeJeopardy = "jeopardy"
	ChallengePackageModeAWD      = "awd"

	ImageSourceTypeManual        = model.ImageSourceTypeManual
	ImageSourceTypePlatformBuild = model.ImageSourceTypePlatformBuild
	ImageSourceTypeExternalRef   = model.ImageSourceTypeExternalRef
)

var packageImageSlugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

func BuildPlatformImageRef(registry, mode, slug, tag string) (string, error) {
	mode = strings.ToLower(strings.TrimSpace(mode))
	switch mode {
	case ChallengePackageModeJeopardy, ChallengePackageModeAWD:
	default:
		return "", fmt.Errorf("unsupported challenge mode %q", mode)
	}

	slug = strings.TrimSpace(slug)
	if !packageImageSlugPattern.MatchString(slug) {
		return "", fmt.Errorf("invalid package slug %q", slug)
	}

	tag = NormalizeImageTag(tag)
	if tag == "" {
		return "", fmt.Errorf("image tag is required")
	}

	repository := path.Join(mode, slug)
	registry = strings.Trim(strings.TrimSpace(registry), "/")
	if registry == "" {
		return fmt.Sprintf("%s:%s", repository, tag), nil
	}
	return fmt.Sprintf("%s/%s:%s", registry, repository, tag), nil
}

func ExtractImageTagSuggestion(ref, tag string) string {
	if normalized := NormalizeImageTag(tag); normalized != "" {
		return normalized
	}
	_, parsedTag, err := SplitImageRef(ref)
	if err != nil {
		return ""
	}
	return NormalizeImageTag(parsedTag)
}

func NormalizeImageTag(tag string) string {
	return strings.TrimSpace(tag)
}

func SplitImageRef(ref string) (string, string, error) {
	trimmed := strings.TrimSpace(ref)
	if trimmed == "" {
		return "", "", fmt.Errorf("empty image ref")
	}

	lastSlash := strings.LastIndex(trimmed, "/")
	lastColon := strings.LastIndex(trimmed, ":")
	if lastColon > lastSlash {
		name := strings.TrimSpace(trimmed[:lastColon])
		tag := strings.TrimSpace(trimmed[lastColon+1:])
		if name == "" || tag == "" {
			return "", "", fmt.Errorf("invalid image ref %q", ref)
		}
		return name, tag, nil
	}
	return trimmed, "latest", nil
}

func resolvePackageImageSource(rootDir string, runtime ChallengePackageRuntime) (string, string, string, string) {
	runtimeImageRef := resolvePackageRuntimeImageRef(runtime)
	suggestedTag := ExtractImageTagSuggestion(runtimeImageRef, runtime.Image.Tag)
	if suggestedTag == "" && strings.TrimSpace(runtimeImageRef) != "" {
		suggestedTag = "latest"
	}

	dockerfilePath, contextPath := packageDockerBuildPaths(rootDir)
	if dockerfilePath != "" {
		return ImageSourceTypePlatformBuild, suggestedTag, dockerfilePath, contextPath
	}
	if strings.TrimSpace(runtimeImageRef) != "" {
		return ImageSourceTypeExternalRef, suggestedTag, "", ""
	}
	return "", suggestedTag, "", ""
}
