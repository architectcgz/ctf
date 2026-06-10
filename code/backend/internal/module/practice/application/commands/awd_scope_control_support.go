package commands

import (
	"context"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"strings"

	"ctf-platform/internal/apperror"
	practiceports "ctf-platform/internal/module/practice/ports"
)

const awdScopeControlReasonLimit = 256

type awdScopeControlState struct {
	TeamRetired       *contestcontracts.AWDScopeControl
	ServiceDisabled   *contestcontracts.AWDScopeControl
	DesiredSuppressed *contestcontracts.AWDScopeControl
}

type awdContestControlIndex struct {
	teamRetired       map[int64]*contestcontracts.AWDScopeControl
	serviceDisabled   map[string]*contestcontracts.AWDScopeControl
	desiredSuppressed map[string]*contestcontracts.AWDScopeControl
}

func buildAWDScopeControlState(rows []*contestcontracts.AWDScopeControl) awdScopeControlState {
	state := awdScopeControlState{}
	for _, row := range rows {
		if row == nil {
			continue
		}
		switch row.ControlType {
		case contestcontracts.AWDScopeControlTypeRetired:
			state.TeamRetired = row
		case contestcontracts.AWDScopeControlTypeServiceDisabled:
			state.ServiceDisabled = row
		case contestcontracts.AWDScopeControlTypeDesiredReconcileSuppressed:
			state.DesiredSuppressed = row
		}
	}
	return state
}

func buildAWDContestControlIndex(rows []*contestcontracts.AWDScopeControl) awdContestControlIndex {
	index := awdContestControlIndex{
		teamRetired:       make(map[int64]*contestcontracts.AWDScopeControl),
		serviceDisabled:   make(map[string]*contestcontracts.AWDScopeControl),
		desiredSuppressed: make(map[string]*contestcontracts.AWDScopeControl),
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		switch row.ControlType {
		case contestcontracts.AWDScopeControlTypeRetired:
			index.teamRetired[row.TeamID] = row
		case contestcontracts.AWDScopeControlTypeServiceDisabled:
			index.serviceDisabled[awdDesiredRuntimeScopeKey(row.TeamID, row.ServiceID)] = row
		case contestcontracts.AWDScopeControlTypeDesiredReconcileSuppressed:
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

func (s *Service) listContestAWDScopeControls(ctx context.Context, contestID int64) ([]*contestcontracts.AWDScopeControl, error) {
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
