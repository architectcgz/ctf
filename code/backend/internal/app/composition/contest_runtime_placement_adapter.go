package composition

import (
	"context"
	"time"

	containerruntimeinfra "ctf-platform/internal/module/container_runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type contestRuntimePlacementStoreAdapter struct {
	repo *contestinfra.ContestRuntimePlacementRepository
}

func newContestRuntimePlacementStoreAdapter(repo *contestinfra.ContestRuntimePlacementRepository) *contestRuntimePlacementStoreAdapter {
	if repo == nil {
		return nil
	}
	return &contestRuntimePlacementStoreAdapter{repo: repo}
}

func (a *contestRuntimePlacementStoreAdapter) FindActiveContestRuntimePlacement(ctx context.Context, contestID int64) (*practiceports.RuntimeNodeBinding, bool, error) {
	placement, exists, err := a.repo.FindActiveContestRuntimePlacement(ctx, contestID)
	if err != nil || !exists {
		return nil, exists, err
	}
	return runtimeNodeBindingFromContestPlacement(placement), true, nil
}

func (a *contestRuntimePlacementStoreAdapter) EnsureActiveContestRuntimePlacement(ctx context.Context, contestID, runtimeNodeID int64) (*practiceports.RuntimeNodeBinding, error) {
	placement, err := a.repo.EnsureActiveContestRuntimePlacement(ctx, contestID, runtimeNodeID)
	if err != nil {
		return nil, err
	}
	return runtimeNodeBindingFromContestPlacement(placement), nil
}

func runtimeNodeBindingFromContestPlacement(placement *contestentity.ContestRuntimePlacement) *practiceports.RuntimeNodeBinding {
	if placement == nil || placement.RuntimeNodeID <= 0 {
		return nil
	}
	return &practiceports.RuntimeNodeBinding{RuntimeNodeID: placement.RuntimeNodeID}
}

type runtimeNodeHealthLookupAdapter struct {
	repo           *containerruntimeinfra.RuntimeNodeRepository
	staleThreshold time.Duration
}

func newRuntimeNodeHealthLookupAdapter(repo *containerruntimeinfra.RuntimeNodeRepository, staleThreshold time.Duration) *runtimeNodeHealthLookupAdapter {
	if repo == nil {
		return nil
	}
	return &runtimeNodeHealthLookupAdapter{repo: repo, staleThreshold: staleThreshold}
}

func (a *runtimeNodeHealthLookupAdapter) FindHealthyRuntimeNodeByID(ctx context.Context, runtimeNodeID int64) (*practiceports.RuntimeNodeBinding, error) {
	if a == nil || a.repo == nil || runtimeNodeID <= 0 {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}
	var (
		nodeName string
		err      error
	)
	if a.staleThreshold > 0 {
		node, findErr := a.repo.FindHealthyByID(ctx, runtimeNodeID, a.staleThreshold, time.Now().UTC())
		if findErr != nil {
			return nil, findErr
		}
		nodeName = node.Name
	} else {
		node, findErr := a.repo.FindByID(ctx, runtimeNodeID)
		if findErr != nil {
			err = findErr
		} else {
			nodeName = node.Name
		}
	}
	if err != nil {
		return nil, err
	}
	return &practiceports.RuntimeNodeBinding{RuntimeNodeID: runtimeNodeID, NodeName: nodeName}, nil
}
