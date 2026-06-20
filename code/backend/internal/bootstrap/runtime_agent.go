package bootstrap

import (
	"errors"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"ctf-platform/internal/config"
	"ctf-platform/internal/infrastructure/logger"
	runtimeinfra "ctf-platform/internal/module/container_runtime/infrastructure"
	"ctf-platform/internal/module/container_runtime/infrastructure/agentserver"

	"go.uber.org/zap"
)

func RunRuntimeAgent() {
	env := os.Getenv("APP_ENV")
	cfg, err := config.LoadRuntimeAgent(env)
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg.Log)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = log.Sync()
	}()

	if !cfg.RuntimeAgent.Server.Enabled {
		log.Fatal("runtime_agent_server_disabled")
	}

	hostExecutor, err := runtimeinfra.NewLocalHostExecutor(&cfg.Container)
	if err != nil {
		log.Fatal("runtime_agent_executor_init_failed", zap.Error(err))
	}
	sandboxExecutor, err := runtimeinfra.NewDockerSandboxExecutor(cfg.Contest.AWD.CheckerSandbox)
	if err != nil {
		log.Fatal("runtime_agent_sandbox_executor_init_failed", zap.Error(err))
	}
	tlsConfig, err := agentserver.LoadServerTLSConfig(
		cfg.RuntimeAgent.Server.CertFile,
		cfg.RuntimeAgent.Server.KeyFile,
		cfg.RuntimeAgent.Server.ClientCAFile,
	)
	if err != nil {
		log.Fatal("runtime_agent_tls_init_failed", zap.Error(err))
	}

	address := net.JoinHostPort(cfg.RuntimeAgent.Server.Host, strconv.Itoa(cfg.RuntimeAgent.Server.Port))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("runtime_agent_listen_failed", zap.Error(err), zap.String("addr", address))
	}
	defer func() {
		_ = listener.Close()
	}()

	server := agentserver.NewGRPCServer(tlsConfig, agentserver.NewService(hostExecutor, sandboxExecutor), cfg.RuntimeAgent.Server)
	go func() {
		log.Info("runtime_agent_starting",
			zap.String("env", cfg.App.Env),
			zap.String("addr", cfg.RuntimeAgent.Server.Host),
			zap.Int("port", cfg.RuntimeAgent.Server.Port),
		)
		if serveErr := server.Serve(listener); serveErr != nil && !errors.Is(serveErr, net.ErrClosed) {
			log.Fatal("runtime_agent_serve_failed", zap.Error(serveErr))
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)
	<-signals

	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	timeout := cfg.RuntimeAgent.Server.ShutdownTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	select {
	case <-done:
		log.Info("runtime_agent_stopped")
	case <-time.After(timeout):
		server.Stop()
		log.Warn("runtime_agent_forced_stop", zap.Duration("timeout", timeout))
	}
}
