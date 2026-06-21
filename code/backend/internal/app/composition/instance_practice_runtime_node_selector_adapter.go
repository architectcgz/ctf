package composition

import (
	"context"
	"strings"

	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type practiceRuntimeNodeSelectorAdapter struct {
	selector       runtimeports.RuntimeNodeSelector
	placementStore practiceports.AWDRuntimePlacementStore
	healthLookup   practiceports.RuntimeNodeHealthLookup
}

func newPracticeRuntimeNodeSelectorAdapter(
	selector runtimeports.RuntimeNodeSelector,
	placementStore practiceports.AWDRuntimePlacementStore,
	healthLookup practiceports.RuntimeNodeHealthLookup,
) practiceports.RuntimeNodeSelector {
	if selector == nil {
		return nil
	}
	return &practiceRuntimeNodeSelectorAdapter{
		selector:       selector,
		placementStore: placementStore,
		healthLookup:   healthLookup,
	}
}

func (a *practiceRuntimeNodeSelectorAdapter) SelectRuntimeNode(ctx context.Context, scope practiceports.InstanceScope) (*practiceports.RuntimeNodeBinding, error) {
	if a == nil || a.selector == nil {
		return nil, nil
	}
	if a.isAWDPlacementScope(scope) {
		return a.selectAWDRuntimeNode(ctx, scope)
	}
	return a.selectDefaultRuntimeNode(ctx)
}

func (a *practiceRuntimeNodeSelectorAdapter) isAWDPlacementScope(scope practiceports.InstanceScope) bool {
	return scope.ContestID != nil && *scope.ContestID > 0 && strings.EqualFold(strings.TrimSpace(scope.ContestMode), "awd")
}

func (a *practiceRuntimeNodeSelectorAdapter) selectAWDRuntimeNode(ctx context.Context, scope practiceports.InstanceScope) (*practiceports.RuntimeNodeBinding, error) {
	if a.placementStore == nil || a.healthLookup == nil {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}
	contestID := *scope.ContestID
	if placement, exists, err := a.placementStore.FindActiveContestRuntimePlacement(ctx, contestID); err != nil || exists {
		if err != nil || placement == nil {
			return placement, err
		}
		return a.healthLookup.FindHealthyRuntimeNodeByID(ctx, placement.RuntimeNodeID)
	}

	selected, err := a.selectDefaultRuntimeNode(ctx)
	if err != nil || selected == nil {
		return nil, err
	}
	placement, err := a.placementStore.EnsureActiveContestRuntimePlacement(ctx, contestID, selected.RuntimeNodeID)
	if err != nil || placement == nil {
		return placement, err
	}
	return a.healthLookup.FindHealthyRuntimeNodeByID(ctx, placement.RuntimeNodeID)
}

func (a *practiceRuntimeNodeSelectorAdapter) selectDefaultRuntimeNode(ctx context.Context) (*practiceports.RuntimeNodeBinding, error) {
	binding, err := a.selector.SelectDefaultNode(ctx)
	if err != nil || binding == nil {
		return nil, err
	}
	return &practiceports.RuntimeNodeBinding{
		RuntimeNodeID: binding.NodeID,
		NodeName:      binding.NodeName,
	}, nil
}
