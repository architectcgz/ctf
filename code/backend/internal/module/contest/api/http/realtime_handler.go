package http

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	xws "golang.org/x/net/websocket"

	"ctf-platform/internal/authctx"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
)

type contestRealtimeAuthContextKey struct{}

type contestRealtimeManager interface {
	ServeChannels(user authctx.CurrentUser, conn *xws.Conn, channels ...string)
}

type RealtimeHandler struct {
	tokenService authcontracts.TokenService
	manager      contestRealtimeManager
	logger       *zap.Logger
}

func NewRealtimeHandler(tokenService authcontracts.TokenService, manager contestRealtimeManager, logger *zap.Logger) *RealtimeHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &RealtimeHandler{
		tokenService: tokenService,
		manager:      manager,
		logger:       logger,
	}
}

func (h *RealtimeHandler) ServeAnnouncementWS(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	h.serveContestWS(c, contestcontracts.AnnouncementChannel(contestID))
}

func (h *RealtimeHandler) ServeScoreboardWS(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	h.serveContestWS(c, contestcontracts.ScoreboardChannel(contestID))
}

func (h *RealtimeHandler) ServeAWDPreviewWS(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	h.serveContestUserWS(c)
}

func (h *RealtimeHandler) serveContestWS(c *gin.Context, channel string) {
	server := xws.Server{
		Handshake: h.handshake,
		Handler: func(conn *xws.Conn) {
			claims, _ := conn.Request().Context().Value(contestRealtimeAuthContextKey{}).(*authctx.CurrentUser)
			if claims == nil {
				_ = conn.Close()
				return
			}
			h.manager.ServeChannels(*claims, conn, channel)
		},
	}
	server.ServeHTTP(c.Writer, c.Request)
}

func (h *RealtimeHandler) serveContestUserWS(c *gin.Context) {
	server := xws.Server{
		Handshake: h.handshake,
		Handler: func(conn *xws.Conn) {
			claims, _ := conn.Request().Context().Value(contestRealtimeAuthContextKey{}).(*authctx.CurrentUser)
			if claims == nil {
				_ = conn.Close()
				return
			}
			h.manager.ServeChannels(*claims, conn)
		},
	}
	server.ServeHTTP(c.Writer, c.Request)
}

func (h *RealtimeHandler) handshake(_ *xws.Config, req *http.Request) error {
	ticket := strings.TrimSpace(req.URL.Query().Get("ticket"))
	claims, err := h.tokenService.ConsumeWSTicket(req.Context(), ticket)
	if err != nil {
		h.logger.Warn("contest_ws_handshake_failed", zap.Error(err))
		return err
	}

	*req = *req.WithContext(context.WithValue(req.Context(), contestRealtimeAuthContextKey{}, claims))
	return nil
}
