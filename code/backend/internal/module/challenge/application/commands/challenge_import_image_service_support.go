package commands

import (
	"errors"

	"go.uber.org/zap"

	"ctf-platform/internal/module/challenge/domain"
	"ctf-platform/pkg/errcode"
)

const challengeImportImageBuildServiceUnavailableBase = "当前后端未启用题包镜像构建/校验服务，请检查 registry 配置"

func challengeImportRequiresImageBuildService(sourceType string) bool {
	switch sourceType {
	case domain.ImageSourceTypePlatformBuild, domain.ImageSourceTypeExternalRef:
		return true
	default:
		return false
	}
}

func challengeImportImageBuildServiceUnavailableMessage(sourceType string) string {
	switch sourceType {
	case domain.ImageSourceTypePlatformBuild:
		return challengeImportImageBuildServiceUnavailableBase + "；该题包依赖平台镜像构建，当前无法提交导入。"
	case domain.ImageSourceTypeExternalRef:
		return challengeImportImageBuildServiceUnavailableBase + "；该题包依赖外部镜像校验，当前无法提交导入。"
	default:
		return challengeImportImageBuildServiceUnavailableBase
	}
}

func challengeImportImageBuildServiceUnavailableError(sourceType string) error {
	return errcode.New(
		errcode.ErrServiceUnavailable.Code,
		challengeImportImageBuildServiceUnavailableMessage(sourceType),
		errcode.ErrServiceUnavailable.HTTPStatus,
	).WithCause(errors.New("image build service is not configured"))
}

func challengeImportMissingImageBuildService(imageBuild *ImageBuildService, sourceType string) bool {
	return challengeImportRequiresImageBuildService(sourceType) && imageBuild == nil
}

func appendChallengeImportImageBuildWarning(warnings []string, sourceType string) []string {
	if !challengeImportRequiresImageBuildService(sourceType) {
		return warnings
	}

	warning := challengeImportImageBuildServiceUnavailableMessage(sourceType)
	for _, existing := range warnings {
		if existing == warning {
			return warnings
		}
	}

	result := append([]string(nil), warnings...)
	result = append(result, warning)
	return result
}

func warnChallengeImportImageBuildServiceUnavailable(
	logger *zap.Logger,
	packageSlug string,
	sourceType string,
	action string,
) {
	if logger == nil {
		return
	}
	logger.Warn(
		challengeImportImageBuildServiceUnavailableMessage(sourceType),
		zap.String("package_slug", packageSlug),
		zap.String("image_source_type", sourceType),
		zap.String("action", action),
	)
}
