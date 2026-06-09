package ports

import (
	"context"

	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
)

type RuntimeNodeSelector interface {
	SelectDefaultNode(ctx context.Context) (*runtimecontracts.RuntimeNodeBinding, error)
}
