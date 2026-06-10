package composition

import (
	"context"

	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type practiceRuntimeNodeSelectorAdapter struct {
	selector runtimeports.RuntimeNodeSelector
}

func newPracticeRuntimeNodeSelectorAdapter(selector runtimeports.RuntimeNodeSelector) practiceports.RuntimeNodeSelector {
	if selector == nil {
		return nil
	}
	return &practiceRuntimeNodeSelectorAdapter{selector: selector}
}

func (a *practiceRuntimeNodeSelectorAdapter) SelectRuntimeNode(ctx context.Context, scope practiceports.InstanceScope) (*practiceports.RuntimeNodeBinding, error) {
	_ = scope
	if a == nil || a.selector == nil {
		return nil, nil
	}
	binding, err := a.selector.SelectDefaultNode(ctx)
	if err != nil || binding == nil {
		return nil, err
	}
	return &practiceports.RuntimeNodeBinding{
		NodeID:   binding.NodeID,
		NodeName: binding.NodeName,
	}, nil
}
