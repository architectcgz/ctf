package ports

import (
	"context"

	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
)

type RuntimeNodeSelector interface {
	SelectDefaultNode(ctx context.Context) (*runtimecontracts.RuntimeNodeBinding, error)
}
