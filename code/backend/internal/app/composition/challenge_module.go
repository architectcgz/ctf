package composition

import (
	"context"
	"fmt"

	"ctf-platform/internal/model"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengeruntime "ctf-platform/internal/module/challenge/runtime"
	opscmd "ctf-platform/internal/module/ops/application/commands"
	opsinfra "ctf-platform/internal/module/ops/infrastructure"
)

type BackgroundTaskCloser = challengeruntime.BackgroundTaskCloser
type ChallengeModule = challengeruntime.Module

type challengePublishNotificationSender struct {
	service *opscmd.NotificationService
}

func (s *challengePublishNotificationSender) SendChallengePublishCheckResult(ctx context.Context, userID, challengeID int64, challengeTitle string, passed bool, failureSummary string) error {
	if s == nil || s.service == nil {
		return nil
	}

	title := "题目发布失败"
	content := fmt.Sprintf("《%s》未通过平台自检。", challengeTitle)
	if passed {
		title = "题目发布成功"
		content = fmt.Sprintf("《%s》已通过平台自检并自动发布。", challengeTitle)
	} else if failureSummary != "" {
		content = fmt.Sprintf("《%s》未通过平台自检：%s", challengeTitle, failureSummary)
	}

	link := fmt.Sprintf("/admin/challenges/%d", challengeID)
	return s.service.SendNotification(ctx, userID, opscmd.SendNotificationInput{
		Type:    model.NotificationTypeChallenge,
		Title:   title,
		Content: content,
		Link:    &link,
	})
}

func BuildChallengeModule(root *Root, runtime *ContainerRuntimeModule, ops *OpsModule) (*ChallengeModule, error) {
	module, err := challengeruntime.Build(challengeruntime.Deps{
		AppContext:    root.Context(),
		Config:        root.Config(),
		Logger:        root.Logger(),
		DB:            root.DB(),
		Cache:         root.Cache(),
		ImageRuntime:  runtime.ChallengeImageRuntime,
		RuntimeProbe:  runtime.ChallengeRuntimeProbe,
		Notifications: buildChallengeNotificationSender(root, ops),
	})
	if err != nil {
		return nil, err
	}

	for _, job := range module.BackgroundJobs {
		root.RegisterBackgroundJob(NewLoopBackgroundJob(job.Name, job.Run))
	}
	return module, nil
}

func buildChallengeNotificationSender(root *Root, ops *OpsModule) challengecmd.ChallengeNotificationSender {
	if ops == nil {
		return nil
	}

	return &challengePublishNotificationSender{
		service: opscmd.NewNotificationService(
			opsinfra.NewNotificationRepository(root.DB()),
			root.Config().Pagination,
			ops.WebSocketManager,
			root.Logger().Named("challenge_publish_notification_service"),
		),
	}
}
