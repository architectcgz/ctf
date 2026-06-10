package composition

import (
	"strings"

	"go.uber.org/zap"

	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	instanceqry "ctf-platform/internal/module/instance/application/queries"
	instanceinfra "ctf-platform/internal/module/instance/infrastructure"
	instanceports "ctf-platform/internal/module/instance/ports"
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

	repo := instanceinfra.NewRepository(root.DB())
	if repo == nil {
		return nil
	}
	proxyTicketReader := newInstanceProxyTicketReader(repo, contestinfra.NewAWDRepository(root.DB()))

	proxyTicketService := buildRuntimeProxyTicketService(root, proxyTicketReader)
	if proxyTicketService == nil {
		return nil
	}

	hostKeyPath := strings.TrimSpace(cfg.Container.DefenseSSHHostKeyPath)
	if hostKeyPath == "" {
		return nil
	}

	return NewAWDDefenseSSHGateway(
		proxyTicketService,
		proxyTicketReader,
		executor,
		hostKeyPath,
		cfg.Container.DefenseSSHPort,
		log.Named("awd_defense_ssh_gateway"),
	)
}

func buildRuntimeProxyTicketService(root *Root, repo instanceports.ProxyTicketInstanceReader) *instanceqry.ProxyTicketService {
	if root == nil || repo == nil {
		return nil
	}
	cfg := root.Config()
	if cfg == nil {
		return nil
	}
	return instanceqry.NewProxyTicketService(
		instanceinfra.NewProxyTicketStore(root.Cache()),
		repo,
		cfg.Container.ProxyTicketTTL,
	)
}
