package commands

import (
	"context"
	"strings"

	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/pkg/errcode"
)

const awdScopeControlReasonLimit = 256

type awdScopeControlState struct {
	TeamRetired      *model.AWDScopeControl
	ServiceDisabled  *model.AWDScopeControl
	DesiredSuppressed *model.AWDScopeControl
}

type awdContestControlIndex struct {
	teamRetired       map[int64]*model.AWDScopeControl
	serviceDisabled   map[string]*model.AWDScopeControl
	desiredSuppressed map[string]*model.AWDScopeControl
}

func buildAWDScopeControlState(rows []*model.AWDScopeControl) awdScopeControlState {
	state := awdScopeControlState{}
	for _, row := range rows {
		if row == nil {
			continue
		}
		switch row.ControlType {
		case model.AWDScopeControlTypeRetired:
			state.TeamRetired = row
		case model.AWDScopeControlTypeServiceDisabled:
			state.ServiceDisabled = row
		case model.AWDScopeControlTypeDesiredReconcileSuppressed:
			state.DesiredSuppressed = row
		}
	}
	return state
}

func buildAWDContestControlIndex(rows []*model.AWDScopeControl) awdContestControlIndex {
	index := awdContestControlIndex{
		teamRetired:       make(map[int64]*model.AWDScopeControl),
		serviceDisabled:   make(map[string]*model.AWDScopeControl),
		desiredSuppressed: make(map[string]*model.AWDScopeControl),
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		switch row.ControlType {
		case model.AWDScopeControlTypeRetired:
			index.teamRetired[row.TeamID] = row
		case model.AWDScopeControlTypeServiceDisabled:
			index.serviceDisabled[awdDesiredRuntimeScopeKey(row.TeamID, row.ServiceID)] = row
		case model.AWDScopeControlTypeDesiredReconcileSuppressed:
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
		return errcode.ErrAWDTeamRetired
	}
	if s.ServiceDisabled != nil {
		return errcode.ErrAWDServiceDisabled
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

func (s *Service) listContestAWDScopeControls(ctx context.Context, contestID int64) ([]*model.AWDScopeControl, error) {
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
		return errcode.ErrInternal.WithCause(err)
	}
	if blocked := state.blocksLifecycle(); blocked != nil {
		return blocked
	}
	return nil
}
