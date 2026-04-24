package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

const defaultAWDPreviewFlag = "flag{preview}"

func normalizeAWDPreviewFlag(value string) string {
	if strings.TrimSpace(value) == "" {
		return defaultAWDPreviewFlag
	}
	return strings.TrimSpace(value)
}

func (s *AWDService) prepareCheckerPreviewAccessURL(
	ctx context.Context,
	previewService *model.ContestAWDService,
	previewChallengeID int64,
	explicitAccessURL string,
	previewFlag string,
) (string, func(context.Context) error, error) {
	if strings.TrimSpace(explicitAccessURL) != "" {
		return strings.TrimSpace(explicitAccessURL), nil, nil
	}
	if s.runtimeProbe == nil {
		return "", nil, errcode.ErrInvalidParams.WithCause(errors.New("当前模板无法自动拉起试跑实例，请手动填写目标访问地址"))
	}

	deploymentMode, runtimeConfig, err := s.loadPreviewRuntimeDefinition(ctx, previewService, previewChallengeID)
	if err != nil {
		return "", nil, err
	}
	if deploymentMode != "" && deploymentMode != model.AWDDeploymentModeSingleContainer {
		return "", nil, errcode.ErrInvalidParams.WithCause(errors.New("当前模板尚不支持自动拉起该部署模式的试跑实例，请手动填写目标访问地址"))
	}

	imageRef, err := s.resolvePreviewImageRef(ctx, runtimeConfig)
	if err != nil {
		return "", nil, err
	}

	accessURL, details, err := s.runtimeProbe.CreateContainer(ctx, imageRef, map[string]string{
		"FLAG": normalizeAWDPreviewFlag(previewFlag),
	})
	if err != nil {
		return "", nil, errcode.ErrInternal.WithCause(err)
	}

	return accessURL, func(cleanupCtx context.Context) error {
		return s.runtimeProbe.CleanupRuntimeDetails(cleanupCtx, details)
	}, nil
}

func (s *AWDService) loadPreviewRuntimeDefinition(
	ctx context.Context,
	previewService *model.ContestAWDService,
	previewChallengeID int64,
) (model.AWDDeploymentMode, map[string]any, error) {
	if previewService != nil {
		snapshot, err := model.DecodeContestAWDServiceSnapshot(previewService.ServiceSnapshot)
		if err != nil {
			return "", nil, errcode.ErrInternal.WithCause(err)
		}
		if len(snapshot.RuntimeConfig) > 0 {
			return snapshot.DeploymentMode, snapshot.RuntimeConfig, nil
		}
	}
	if previewChallengeID <= 0 || s.templateRepo == nil {
		return "", nil, errcode.ErrInvalidParams.WithCause(errors.New("当前模板缺少可用的运行配置，请手动填写目标访问地址"))
	}

	template, err := s.templateRepo.FindAWDServiceTemplateByIDWithContext(ctx, previewChallengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errcode.ErrNotFound
		}
		return "", nil, errcode.ErrInternal.WithCause(err)
	}
	return template.DeploymentMode, parseContestAWDServiceJSONMap(template.RuntimeConfig), nil
}

func (s *AWDService) resolvePreviewImageRef(ctx context.Context, runtimeConfig map[string]any) (string, error) {
	if imageRef := strings.TrimSpace(readStringFromAny(runtimeConfig["image_ref"])); imageRef != "" {
		if imageID := readInt64FromAny(runtimeConfig["image_id"]); imageID > 0 && s.imageRepo != nil {
			if resolved, err := s.resolvePreviewImageRefByID(ctx, imageID); err == nil {
				return resolved, nil
			}
			return imageRef, nil
		}
		return imageRef, nil
	}

	imageID := readInt64FromAny(runtimeConfig["image_id"])
	if imageID <= 0 {
		return "", errcode.ErrInvalidParams.WithCause(errors.New("当前模板未配置可拉起的镜像，请手动填写目标访问地址"))
	}
	if s.imageRepo == nil {
		return "", errcode.ErrInvalidParams.WithCause(errors.New("当前模板无法解析镜像配置，请手动填写目标访问地址"))
	}
	return s.resolvePreviewImageRefByID(ctx, imageID)
}

func (s *AWDService) resolvePreviewImageRefByID(ctx context.Context, imageID int64) (string, error) {
	if imageID <= 0 {
		return "", errcode.ErrInvalidParams.WithCause(errors.New("invalid preview image id"))
	}
	imageItem, err := s.imageRepo.FindByIDWithContext(ctx, imageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errcode.ErrNotFound
		}
		return "", errcode.ErrInternal.WithCause(err)
	}
	if imageItem.Status != model.ImageStatusAvailable {
		return "", errcode.ErrInvalidParams.WithCause(fmt.Errorf("preview image %d status=%s", imageItem.ID, imageItem.Status))
	}
	return fmt.Sprintf("%s:%s", imageItem.Name, imageItem.Tag), nil
}

func readStringFromAny(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}

func readInt64FromAny(value any) int64 {
	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int32:
		return int64(typed)
	case int64:
		return typed
	case float64:
		return int64(typed)
	default:
		return 0
	}
}

func (s *AWDService) cleanupCheckerPreviewRuntime(ctx context.Context, cleanup func(context.Context) error, previewErr error) error {
	if cleanup == nil {
		return nil
	}
	if err := cleanup(ctx); err != nil {
		if previewErr != nil {
			s.log.Warn("cleanup_checker_preview_runtime_failed", zap.Error(err))
			return nil
		}
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}
