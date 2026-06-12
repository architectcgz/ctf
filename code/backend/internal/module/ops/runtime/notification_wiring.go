package runtime

import (
	authcontracts "ctf-platform/internal/module/auth/contracts"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	opshttp "ctf-platform/internal/module/ops/api/http"
	opscmd "ctf-platform/internal/module/ops/application/commands"
	opsqry "ctf-platform/internal/module/ops/application/queries"
	opscontracts "ctf-platform/internal/module/ops/contracts"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	platformevents "ctf-platform/internal/platform/events"
)

func (m *Module) BindNotificationHandler(tokenService authcontracts.TokenService) {
	if m == nil || m.notificationBuilder == nil {
		return
	}
	handler, service := m.notificationBuilder(tokenService)
	m.NotificationHandler = handler
	m.notificationCommandService = service
}

func (m *Module) RegisterPracticeOutboxHandlers(registry *platformevents.OutboxHandlerRegistry) {
	if m == nil || m.notificationCommandService == nil || registry == nil {
		return
	}
	registry.Register(practicecontracts.EventFlagAccepted, m.notificationCommandService.HandlePracticeFlagAcceptedOutboxEvent)
}

func buildNotificationHandler(deps moduleDeps, tokenService authcontracts.TokenService) (*opshttp.NotificationHandler, *opscmd.NotificationService) {
	cfg := deps.input.Config
	log := deps.input.Logger

	notificationCommandService := opscmd.NewNotificationService(
		deps.notificationRepo,
		cfg.Pagination,
		deps.webSocketManager,
		log.Named("notification_command_service"),
	)
	registerNotificationOutboxHandlers(deps.outboxHandlers, notificationCommandService)
	notificationQueryService := opsqry.NewNotificationService(
		deps.notificationRepo,
		cfg.Pagination,
		log.Named("notification_query_service"),
	)
	return opshttp.NewNotificationHandler(
		notificationCommandService,
		notificationQueryService,
		tokenService,
		deps.webSocketManager,
		log.Named("notification_handler"),
	), notificationCommandService
}

func registerNotificationOutboxHandlers(registry *platformevents.OutboxHandlerRegistry, service *opscmd.NotificationService) {
	if registry == nil || service == nil {
		return
	}
	registry.Register(challengecontracts.EventPublishCheckFinished, service.HandleChallengePublishCheckFinishedOutboxEvent)
	registry.Register(opscontracts.EventNotificationCreated, service.HandleNotificationFanoutEvent)
	registry.Register(opscontracts.EventNotificationRead, service.HandleNotificationFanoutEvent)
}
