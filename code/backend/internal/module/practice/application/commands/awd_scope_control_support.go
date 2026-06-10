package commands

import (
	"context"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"strings"

	"ctf-platform/internal/apperror"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimestate "ctf-platform/internal/module/runtime/contracts"
)

const awdScopeControlReasonLimit = 256

type awdScopeControlState struct {
	TeamRetired       *runtimestate.AWDScopeControl
	ServiceDisabled   *runtimestate.AWDScopeControl
	DesiredSuppressed *runtimestate.AWDScopeControl
}

type awdContestControlIndex struct {
	teamRetired       map[int64]*runtimestate.AWDScopeControl
	serviceDisabled   map[string]*runtimestate.AWDScopeControl
	desiredSuppressed map[string]*runtimestate.AWDScopeControl
}

func buildAWDScopeControlState(rows []*runtimestate.AWDScopeControl) awdScopeControlState {
	state := awdScopeControlState{}
	for _, row := range rows {
		if row == nil {
			continue
		}
		switch row.ControlType {
		case runtimestate.AWDScopeControlTypeRetired:
			state.TeamRetired = row
		case runtimestate.AWDScopeControlTypeServiceDisabled:
			state.ServiceDisabled = row
		case runtimestate.AWDScopeControlTypeDesiredReconcileSuppressed:
			state.DesiredSuppressed = row
		}
	}
	return state
}

func buildAWDContestControlIndex(rows []*runtimestate.AWDScopeControl) awdContestControlIndex {
	index := awdContestControlIndex{
		teamRetired:       make(map[int64]*runtimestate.AWDScopeControl),
		serviceDisabled:   make(map[string]*runtimestate.AWDScopeControl),
		desiredSuppressed: make(map[string]*runtimestate.AWDScopeControl),
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		switch row.ControlType {
		case runtimestate.AWDScopeControlTypeRetired:
			index.teamRetired[row.TeamID] = row
		case runtimestate.AWDScopeControlTypeServiceDisabled:
			index.serviceDisabled[awdDesiredRuntimeScopeKey(row.TeamID, row.ServiceID)] = row
		case runtimestate.AWDScopeControlTypeDesiredReconcileSuppressed:
			index.desiredSuppressed[awdDesiredRuntimeScopeKey(row.TeamID, row.ServiceID)] = row
		}
	}
	return index
}

func (idx awdContestControlIndex) state(teamID, serviceID int64) awdScopeControlState {
	state := awdScopeControlState{}
	if row, ok := idx.teamRetired[teamID]; ok {
		state.TeamRetired = row
	}
	if row, ok := idx.serviceDisabled[awdDesiredRuntimeScopeKey(teamID, serviceID)]; ok {
		state.ServiceDisabled = row
	}
	if row, ok := idx.desiredSuppressed[awdDesiredRuntimeScopeKey(teamID, serviceID)]; ok {
		state.DesiredSuppressed = row
	}
	return state
}

func (s awdScopeControlState) blocksLifecycle() error {
	if s.TeamRetired != nil {
		return contestcontracts.ErrAWDTeamRetired
	}
	if s.ServiceDisabled != nil {
		return contestcontracts.ErrAWDServiceDisabled
	}
	return nil
}

func (s awdScopeControlState) suppressesDesiredReconcile() bool {
	return s.TeamRetired != nil || s.ServiceDisabled != nil || s.DesiredSuppressed != nil
}

func normalizeAWDScopeControlReason(reason string) string {
	trimmed := strings.TrimSpace(reason)
	if len(trimmed) <= awdScopeControlReasonLimit {
		return trimmed
	}
	return trimmed[:awdScopeControlReasonLimit]
}

func (s *Service) listContestAWDScopeControls(ctx context.Context, contestID int64) ([]*runtimestate.AWDScopeControl, error) {
	if s == nil || s.repo == nil || contestID <= 0 {
		return nil, nil
	}
	return s.repo.ListContestAWDScopeControls(ctx, contestID)
}

func (s *Service) loadAWDScopeControlState(ctx context.Context, contestID, teamID, serviceID int64) (awdScopeControlState, error) {
	if s == nil || s.repo == nil || contestID <= 0 || teamID <= 0 {
		return awdScopeControlState{}, nil
	}
	rows, err := s.repo.ListScopeAWDScopeControls(ctx, contestID, teamID, serviceID)
	if err != nil {
		return awdScopeControlState{}, err
	}
	return buildAWDScopeControlState(rows), nil
}

func (s *Service) ensureAWDScopeAllowsLifecycle(ctx context.Context, scope practiceports.InstanceScope) error {
	if scope.ContestID == nil || scope.TeamID == nil || scope.ServiceID == nil {
		return nil
	}
	state, err := s.loadAWDScopeControlState(ctx, *scope.ContestID, *scope.TeamID, *scope.ServiceID)
	if err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	if blocked := state.blocksLifecycle(); blocked != nil {
		return blocked
	}
	return nil
}
