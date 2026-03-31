package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	xws "golang.org/x/net/websocket"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	"ctf-platform/pkg/response"
	ctfws "ctf-platform/pkg/websocket"
)

type notificationAuthContextKey struct{}

type notificationCommandService interface {
	MarkAsRead(ctx context.Context, userID, notificationID int64) error
	PublishAdminNotification(ctx context.Context, actorUserID int64, req *dto.AdminNotificationPublishReq) (*dto.AdminNotificationPublishResp, error)
}

type notificationQueryService interface {
	GetNotifications(ctx context.Context, userID int64, query *dto.NotificationQuery) ([]dto.NotificationInfo, int64, int, int, error)
}

type NotificationHandler struct {
	commands     notificationCommandService
	queries      notificationQueryService
	tokenService authcontracts.TokenService
	manager      *ctfws.Manager
	logger       *zap.Logger
}

func NewNotificationHandler(commands notificationCommandService, queries notificationQueryService, tokenService authcontracts.TokenService, manager *ctfws.Manager, logger *zap.Logger) *NotificationHandler {
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
	var query dto.NotificationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	items, total, page, pageSize, err := h.queries.GetNotifications(c.Request.Context(), authUser.UserID, &query)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Page(c, items, total, page, pageSize)
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
	var req dto.AdminNotificationPublishReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.commands.PublishAdminNotification(c.Request.Context(), authUser.UserID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, result)
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
