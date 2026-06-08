package composition

import (
	"strings"

	"go.uber.org/zap"

	instanceqry "ctf-platform/internal/module/instance/application/queries"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
)

func BuildAWDDefenseSSHGateway(root *Root, runtime *ContainerRuntimeModule) *AWDDefenseSSHGateway {
	if root == nil || runtime == nil || runtime.runtime == nil {
		return nil
	}

	cfg := root.Config()
	if cfg == nil || !cfg.Container.DefenseSSHEnabled {
		return nil
	}

	module := runtime.runtime
	executor := module.InteractiveExecutor
	if runtime.nodeRouter != nil {
		executor = runtime.nodeRouter
	}
	if executor == nil {
		return nil
	}

	log := root.Logger()
	if log == nil {
		log = zap.NewNop()
	}

	repo := runtimeinfra.NewRepository(root.DB())
	if repo == nil {
		return nil
	}

	proxyTicketService := buildRuntimeProxyTicketService(root, repo)
	if proxyTicketService == nil {
		return nil
	}

	hostKeyPath := strings.TrimSpace(cfg.Container.DefenseSSHHostKeyPath)
	if hostKeyPath == "" {
		return nil
	}

	return NewAWDDefenseSSHGateway(
		proxyTicketService,
		repo,
		executor,
		hostKeyPath,
		cfg.Container.DefenseSSHPort,
		log.Named("awd_defense_ssh_gateway"),
	)
}

func buildRuntimeProxyTicketService(root *Root, repo *runtimeinfra.Repository) *instanceqry.ProxyTicketService {
	if root == nil || repo == nil {
		return nil
	}
	cfg := root.Config()
	if cfg == nil {
		return nil
	}
	return instanceqry.NewProxyTicketService(
		runtimeinfra.NewProxyTicketStore(root.Cache()),
		repo,
		cfg.Container.ProxyTicketTTL,
	)
}
