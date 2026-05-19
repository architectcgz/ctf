package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	xws "golang.org/x/net/websocket"

	"ctf-platform/internal/authctx"
	response "ctf-platform/internal/httpresponse"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	opscmd "ctf-platform/internal/module/ops/application/commands"
	opsqry "ctf-platform/internal/module/ops/application/queries"
)

type notificationAuthContextKey struct{}

type notificationCommandService interface {
	MarkAsRead(ctx context.Context, userID, notificationID int64) error
	PublishAdminNotification(ctx context.Context, actorUserID int64, req opscmd.PublishAdminNotificationInput) (*opscmd.AdminNotificationPublishResp, error)
}

type notificationQueryService interface {
	GetNotifications(ctx context.Context, userID int64, query *opsqry.NotificationQuery) ([]opsqry.NotificationInfo, int64, int, int, error)
}

type notificationSocketManager interface {
	Serve(user authctx.CurrentUser, conn *xws.Conn)
}

type NotificationHandler struct {
	commands     notificationCommandService
	queries      notificationQueryService
	tokenService authcontracts.TokenService
	manager      notificationSocketManager
	logger       *zap.Logger
}

func NewNotificationHandler(commands notificationCommandService, queries notificationQueryService, tokenService authcontracts.TokenService, manager notificationSocketManager, logger *zap.Logger) *NotificationHandler {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &NotificationHandler{
		commands:     commands,
		queries:      queries,
		tokenService: tokenService,
		manager:      manager,
		logger:       logger,
	}
}

func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	authUser := authctx.MustCurrentUser(c)
	var query opsqry.NotificationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	items, total, page, pageSize, err := h.queries.GetNotifications(c.Request.Context(), authUser.UserID, &query)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Page(c, toNotificationInfos(items), total, page, pageSize)
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	authUser := authctx.MustCurrentUser(c)
	notificationID := c.GetInt64("id")
	if err := h.commands.MarkAsRead(c.Request.Context(), authUser.UserID, notificationID); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *NotificationHandler) PublishAdminNotification(c *gin.Context) {
	authUser := authctx.MustCurrentUser(c)
	var req AdminNotificationPublishReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.commands.PublishAdminNotification(c.Request.Context(), authUser.UserID, opsRequestMapper.ToPublishAdminNotificationInput(req))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, toAdminNotificationPublishResp(result))
}

func (h *NotificationHandler) ServeWS(c *gin.Context) {
	server := xws.Server{
		Handshake: h.handshake,
		Handler: func(conn *xws.Conn) {
			claims, _ := conn.Request().Context().Value(notificationAuthContextKey{}).(*authctx.CurrentUser)
			if claims == nil {
				_ = conn.Close()
				return
			}

			h.manager.Serve(*claims, conn)
		},
	}
	server.ServeHTTP(c.Writer, c.Request)
}

func (h *NotificationHandler) handshake(_ *xws.Config, req *http.Request) error {
	ticket := strings.TrimSpace(req.URL.Query().Get("ticket"))
	claims, err := h.tokenService.ConsumeWSTicket(req.Context(), ticket)
	if err != nil {
		h.logger.Warn("notification_ws_handshake_failed", zap.Error(err))
		return err
	}

	*req = *req.WithContext(context.WithValue(req.Context(), notificationAuthContextKey{}, claims))
	return nil
}
