package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type teamCommandAdapterSourceStub struct {
	findByIDFn                func(context.Context, int64) (*model.Team, error)
	findContestRegistrationFn func(context.Context, int64, int64) (*model.ContestRegistration, error)
	findUserTeamInContestFn   func(context.Context, int64, int64) (*model.Team, error)
	createWithMemberFn        func(context.Context, *model.Team, int64) error
	addMemberWithLockFn       func(context.Context, int64, int64, int64) error
}

func (s teamCommandAdapterSourceStub) FindByID(ctx context.Context, id int64) (*model.Team, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return &model.Team{}, nil
}

func (s teamCommandAdapterSourceStub) FindContestRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	if s.findContestRegistrationFn != nil {
		return s.findContestRegistrationFn(ctx, contestID, userID)
	}
	return &model.ContestRegistration{}, nil
}

func (s teamCommandAdapterSourceStub) FindUserTeamInContest(ctx context.Context, userID, contestID int64) (*model.Team, error) {
	if s.findUserTeamInContestFn != nil {
		return s.findUserTeamInContestFn(ctx, userID, contestID)
	}
	return &model.Team{}, nil
}

func (s teamCommandAdapterSourceStub) CreateWithMember(ctx context.Context, team *model.Team, captainID int64) error {
	if s.createWithMemberFn != nil {
		return s.createWithMemberFn(ctx, team, captainID)
	}
	return nil
}

func (s teamCommandAdapterSourceStub) DeleteWithMembers(context.Context, int64) error { return nil }
func (s teamCommandAdapterSourceStub) IsUniqueViolation(error, string) bool           { return false }

func (s teamCommandAdapterSourceStub) AddMemberWithLock(ctx context.Context, contestID, teamID, userID int64) error {
	if s.addMemberWithLockFn != nil {
		return s.addMemberWithLockFn(ctx, contestID, teamID, userID)
	}
	return nil
}

func (s teamCommandAdapterSourceStub) RemoveMember(context.Context, int64, int64) error { return nil }
func (s teamCommandAdapterSourceStub) GetMembers(context.Context, int64) ([]*model.TeamMember, error) {
	return nil, nil
}
func (s teamCommandAdapterSourceStub) GetMemberCount(context.Context, int64) (int64, error) {
	return 0, nil
}
func (s teamCommandAdapterSourceStub) GetMemberCountBatch(context.Context, []int64) (map[int64]int, error) {
	return map[int64]int{}, nil
}

func TestTeamCommandAdapterMapsFindByIDNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewTeamCommandAdapter(teamCommandAdapterSourceStub{
		findByIDFn: func(context.Context, int64) (*model.Team, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindByID(context.Background(), 1); !errors.Is(err, contestports.ErrContestTeamNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestTeamNotFound)
	}
}

func TestTeamCommandAdapterMapsFindContestRegistrationNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewTeamCommandAdapter(teamCommandAdapterSourceStub{
		findContestRegistrationFn: func(context.Context, int64, int64) (*model.ContestRegistration, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindContestRegistration(context.Background(), 1, 2); !errors.Is(err, contestports.ErrContestParticipationRegistrationNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestParticipationRegistrationNotFound)
	}
}

func TestTeamCommandAdapterMapsFindUserTeamInContestNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewTeamCommandAdapter(teamCommandAdapterSourceStub{
		findUserTeamInContestFn: func(context.Context, int64, int64) (*model.Team, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindUserTeamInContest(context.Background(), 1, 2); !errors.Is(err, contestports.ErrContestUserTeamNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestUserTeamNotFound)
	}
}

func TestTeamCommandAdapterMapsCreateWithMemberRegistrationBindingNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewTeamCommandAdapter(teamCommandAdapterSourceStub{
		createWithMemberFn: func(context.Context, *model.Team, int64) error {
			return gorm.ErrRecordNotFound
		},
	})

	err := repo.CreateWithMember(context.Background(), &model.Team{}, 1)
	if !errors.Is(err, contestports.ErrContestParticipationRegistrationNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestParticipationRegistrationNotFound)
	}
}

func TestTeamCommandAdapterMapsAddMemberWithLockRegistrationBindingNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewTeamCommandAdapter(teamCommandAdapterSourceStub{
		addMemberWithLockFn: func(context.Context, int64, int64, int64) error {
			return gorm.ErrRecordNotFound
		},
	})

	err := repo.AddMemberWithLock(context.Background(), 1, 2, 3)
	if !errors.Is(err, contestports.ErrContestParticipationRegistrationNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestParticipationRegistrationNotFound)
	}
}
