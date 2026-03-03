package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	healthHandler "ctf-platform/internal/handler/health"
	"ctf-platform/internal/middleware"
	healthService "ctf-platform/internal/service/health"
	"ctf-platform/internal/config"
)

type HTTPServer struct {
	engine *gin.Engine
	server *http.Server
}

func NewHTTPServer(cfg *config.Config, log *zap.Logger) *HTTPServer {
	if cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(middleware.RequestID())
	engine.Use(middleware.AccessLog(log))
	engine.Use(middleware.Recovery(log))

	healthSvc := healthService.NewService(cfg)
	health := healthHandler.NewHandler(healthSvc)

	engine.GET("/healthz", health.Get)

	apiV1 := engine.Group("/api/v1")
	apiV1.GET("/healthz", health.Get)

	return &HTTPServer{
		engine: engine,
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
			Handler:      engine,
			ReadTimeout:  cfg.HTTP.ReadTimeout,
			WriteTimeout: cfg.HTTP.WriteTimeout,
			IdleTimeout:  cfg.HTTP.IdleTimeout,
		},
	}
}

func (s *HTTPServer) Start() error {
	return s.server.ListenAndServe()
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *HTTPServer) Engine() *gin.Engine {
	return s.engine
}
