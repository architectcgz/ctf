package commands

import (
	"context"
	"errors"
	"testing"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type teamCommandRepoStub struct {
	findContestRegistrationFn func(context.Context, int64, int64) (*model.ContestRegistration, error)
	findByIDFn                func(context.Context, int64) (*model.Team, error)
	findUserTeamInContestFn   func(context.Context, int64, int64) (*model.Team, error)
	createWithMemberFn        func(context.Context, *model.Team, int64) error
	addMemberWithLockFn       func(context.Context, int64, int64, int64) error
	getMemberCountFn          func(context.Context, int64) (int64, error)
	getMembersFn              func(context.Context, int64) ([]*model.TeamMember, error)
	removeMemberFn            func(context.Context, int64, int64) error
	deleteWithMembersFn       func(context.Context, int64) error
}

func (s teamCommandRepoStub) FindContestRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	if s.findContestRegistrationFn != nil {
		return s.findContestRegistrationFn(ctx, contestID, userID)
	}
	return &model.ContestRegistration{ContestID: contestID, UserID: userID, Status: model.ContestRegistrationStatusApproved}, nil
}

func (s teamCommandRepoStub) FindByID(ctx context.Context, teamID int64) (*model.Team, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, teamID)
	}
	return &model.Team{ID: teamID, ContestID: 10, CaptainID: 2001, MaxMembers: 4}, nil
}

func (s teamCommandRepoStub) FindUserTeamInContest(ctx context.Context, userID, contestID int64) (*model.Team, error) {
	if s.findUserTeamInContestFn != nil {
		return s.findUserTeamInContestFn(ctx, userID, contestID)
	}
	return nil, contestports.ErrContestUserTeamNotFound
}

func (s teamCommandRepoStub) CreateWithMember(ctx context.Context, team *model.Team, captainID int64) error {
	if s.createWithMemberFn != nil {
		return s.createWithMemberFn(ctx, team, captainID)
	}
	team.ID = 88
	return nil
}

func (s teamCommandRepoStub) DeleteWithMembers(ctx context.Context, teamID int64) error {
	if s.deleteWithMembersFn != nil {
		return s.deleteWithMembersFn(ctx, teamID)
	}
	return nil
}

func (s teamCommandRepoStub) IsUniqueViolation(error, string) bool { return false }

func (s teamCommandRepoStub) AddMemberWithLock(ctx context.Context, contestID, teamID, userID int64) error {
	if s.addMemberWithLockFn != nil {
		return s.addMemberWithLockFn(ctx, contestID, teamID, userID)
	}
	return nil
}

func (s teamCommandRepoStub) RemoveMember(ctx context.Context, teamID, userID int64) error {
	if s.removeMemberFn != nil {
		return s.removeMemberFn(ctx, teamID, userID)
	}
	return nil
}

func (s teamCommandRepoStub) GetMembers(ctx context.Context, teamID int64) ([]*model.TeamMember, error) {
	if s.getMembersFn != nil {
		return s.getMembersFn(ctx, teamID)
	}
	return []*model.TeamMember{{TeamID: teamID, UserID: 2002}}, nil
}

func (s teamCommandRepoStub) GetMemberCount(ctx context.Context, teamID int64) (int64, error) {
	if s.getMemberCountFn != nil {
		return s.getMemberCountFn(ctx, teamID)
	}
	return 2, nil
}

func (s teamCommandRepoStub) GetMemberCountBatch(context.Context, []int64) (map[int64]int, error) {
	return map[int64]int{}, nil
}

func TestTeamServiceCreateTeamTreatsRegistrationNotFoundAsNotRegistered(t *testing.T) {
	t.Parallel()

	service := NewTeamService(teamCommandRepoStub{
		findContestRegistrationFn: func(context.Context, int64, int64) (*model.ContestRegistration, error) {
			return nil, contestports.ErrContestParticipationRegistrationNotFound
		},
	}, participationContestLookupStub{})

	_, err := service.CreateTeam(context.Background(), 10, 2001, CreateTeamInput{Name: "alpha"})
	if !errors.Is(err, errcode.ErrNotRegistered) {
		t.Fatalf("expected ErrNotRegistered, got %v", err)
	}
}

func TestTeamServiceCreateTeamTreatsUnexpectedCurrentTeamLookupErrorAsInternal(t *testing.T) {
	t.Parallel()

	lookupErr := errors.New("team lookup exploded")
	createCalled := false
	service := NewTeamService(teamCommandRepoStub{
		findUserTeamInContestFn: func(context.Context, int64, int64) (*model.Team, error) {
			return nil, lookupErr
		},
		createWithMemberFn: func(context.Context, *model.Team, int64) error {
			createCalled = true
			return nil
		},
	}, participationContestLookupStub{})

	_, err := service.CreateTeam(context.Background(), 10, 2001, CreateTeamInput{Name: "alpha"})
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrInternal.Code {
		t.Fatalf("expected ErrInternal, got %v", err)
	}
	if createCalled {
		t.Fatal("expected create path to stop on unexpected current-team lookup error")
	}
}

func TestTeamServiceCreateTeamTreatsCreateWithMemberRegistrationMissingAsNotRegistered(t *testing.T) {
	t.Parallel()

	service := NewTeamService(teamCommandRepoStub{
		createWithMemberFn: func(context.Context, *model.Team, int64) error {
			return contestports.ErrContestParticipationRegistrationNotFound
		},
	}, participationContestLookupStub{})

	_, err := service.CreateTeam(context.Background(), 10, 2001, CreateTeamInput{Name: "alpha"})
	if !errors.Is(err, errcode.ErrNotRegistered) {
		t.Fatalf("expected ErrNotRegistered, got %v", err)
	}
}

func TestTeamServiceJoinTeamTreatsTeamNotFoundAsErrTeamNotFound(t *testing.T) {
	t.Parallel()

	service := NewTeamService(teamCommandRepoStub{
		findByIDFn: func(context.Context, int64) (*model.Team, error) {
			return nil, contestports.ErrContestTeamNotFound
		},
	}, participationContestLookupStub{})

	_, err := service.JoinTeam(context.Background(), 10, 2002, 404)
	if !errors.Is(err, errcode.ErrTeamNotFound) {
		t.Fatalf("expected ErrTeamNotFound, got %v", err)
	}
}

func TestTeamServiceJoinTeamTreatsUnexpectedCurrentTeamLookupErrorAsInternal(t *testing.T) {
	t.Parallel()

	lookupErr := errors.New("team lookup exploded")
	addCalled := false
	service := NewTeamService(teamCommandRepoStub{
		findByIDFn: func(context.Context, int64) (*model.Team, error) {
			return &model.Team{ID: 33, ContestID: 10, CaptainID: 2001, MaxMembers: 4}, nil
		},
		findUserTeamInContestFn: func(context.Context, int64, int64) (*model.Team, error) {
			return nil, lookupErr
		},
		addMemberWithLockFn: func(context.Context, int64, int64, int64) error {
			addCalled = true
			return nil
		},
	}, participationContestLookupStub{})

	_, err := service.JoinTeam(context.Background(), 10, 2002, 33)
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrInternal.Code {
		t.Fatalf("expected ErrInternal, got %v", err)
	}
	if addCalled {
		t.Fatal("expected join path to stop on unexpected current-team lookup error")
	}
}

func TestTeamServiceJoinTeamTreatsMembershipRegistrationMissingAsNotRegistered(t *testing.T) {
	t.Parallel()

	service := NewTeamService(teamCommandRepoStub{
		findByIDFn: func(context.Context, int64) (*model.Team, error) {
			return &model.Team{ID: 33, ContestID: 10, CaptainID: 2001, MaxMembers: 4}, nil
		},
		addMemberWithLockFn: func(context.Context, int64, int64, int64) error {
			return contestports.ErrContestParticipationRegistrationNotFound
		},
	}, participationContestLookupStub{})

	_, err := service.JoinTeam(context.Background(), 10, 2002, 33)
	if !errors.Is(err, errcode.ErrNotRegistered) {
		t.Fatalf("expected ErrNotRegistered, got %v", err)
	}
}

func TestTeamServiceLeaveTeamTreatsTeamNotFoundAsErrTeamNotFound(t *testing.T) {
	t.Parallel()

	service := NewTeamService(teamCommandRepoStub{
		findByIDFn: func(context.Context, int64) (*model.Team, error) {
			return nil, contestports.ErrContestTeamNotFound
		},
	}, participationContestLookupStub{})

	err := service.LeaveTeam(context.Background(), 10, 2002, 404)
	if !errors.Is(err, errcode.ErrTeamNotFound) {
		t.Fatalf("expected ErrTeamNotFound, got %v", err)
	}
}

func TestTeamServiceDismissTeamTreatsTeamNotFoundAsErrTeamNotFound(t *testing.T) {
	t.Parallel()

	service := NewTeamService(teamCommandRepoStub{
		findByIDFn: func(context.Context, int64) (*model.Team, error) {
			return nil, contestports.ErrContestTeamNotFound
		},
	}, participationContestLookupStub{})

	err := service.DismissTeam(context.Background(), 10, 2001, 404)
	if !errors.Is(err, errcode.ErrTeamNotFound) {
		t.Fatalf("expected ErrTeamNotFound, got %v", err)
	}
}

func TestTeamServiceKickMemberTreatsTeamNotFoundAsErrTeamNotFound(t *testing.T) {
	t.Parallel()

	service := NewTeamService(teamCommandRepoStub{
		findByIDFn: func(context.Context, int64) (*model.Team, error) {
			return nil, contestports.ErrContestTeamNotFound
		},
	}, participationContestLookupStub{})

	err := service.KickMember(context.Background(), 10, 2001, 404, 2002)
	if !errors.Is(err, errcode.ErrTeamNotFound) {
		t.Fatalf("expected ErrTeamNotFound, got %v", err)
	}
}
