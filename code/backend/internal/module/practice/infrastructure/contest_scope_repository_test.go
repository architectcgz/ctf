package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestentity "ctf-platform/internal/module/contest/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type contestScopeSourceStub struct {
	findContestByIDFn                     func(context.Context, int64) (*contestentity.Contest, error)
	findContestChallengeFn                func(context.Context, int64, int64) (*contestentity.ContestChallenge, error)
	findContestAWDServiceFn               func(context.Context, int64, int64) (*contestentity.ContestAWDService, error)
	findContestAWDServiceRuntimeSubjectFn func(context.Context, int64, int64) (*practiceports.ContestAWDServiceRuntimeSubject, error)
	listContestAWDServicesFn              func(context.Context, int64) ([]*contestentity.ContestAWDService, error)
	listContestAWDInstancesFn             func(context.Context, int64) ([]*model.Instance, error)
	findContestTeamFn                     func(context.Context, int64, int64) (*contestentity.Team, error)
	listContestTeamsFn                    func(context.Context, int64) ([]*contestentity.Team, error)
	findContestRegistrationFn             func(context.Context, int64, int64) (*practiceports.ContestParticipation, error)
}

func (s contestScopeSourceStub) FindContestByID(ctx context.Context, contestID int64) (*contestentity.Contest, error) {
	return s.findContestByIDFn(ctx, contestID)
}

func (s contestScopeSourceStub) FindContestChallenge(ctx context.Context, contestID, challengeID int64) (*contestentity.ContestChallenge, error) {
	return s.findContestChallengeFn(ctx, contestID, challengeID)
}

func (s contestScopeSourceStub) FindContestAWDService(ctx context.Context, contestID, serviceID int64) (*contestentity.ContestAWDService, error) {
	return s.findContestAWDServiceFn(ctx, contestID, serviceID)
}

func (s contestScopeSourceStub) FindContestAWDServiceRuntimeSubject(ctx context.Context, contestID, serviceID int64) (*practiceports.ContestAWDServiceRuntimeSubject, error) {
	if s.findContestAWDServiceRuntimeSubjectFn != nil {
		return s.findContestAWDServiceRuntimeSubjectFn(ctx, contestID, serviceID)
	}
	service, err := s.FindContestAWDService(ctx, contestID, serviceID)
	if err != nil || service == nil {
		return nil, err
	}
	return &practiceports.ContestAWDServiceRuntimeSubject{
		ServiceID:   service.ID,
		ChallengeID: service.AWDChallengeID,
		Visible:     service.IsVisible,
	}, nil
}

func (s contestScopeSourceStub) ListContestAWDServices(ctx context.Context, contestID int64) ([]*contestentity.ContestAWDService, error) {
	if s.listContestAWDServicesFn == nil {
		return nil, nil
	}
	return s.listContestAWDServicesFn(ctx, contestID)
}

func (s contestScopeSourceStub) ListContestAWDInstances(ctx context.Context, contestID int64) ([]*model.Instance, error) {
	if s.listContestAWDInstancesFn == nil {
		return nil, nil
	}
	return s.listContestAWDInstancesFn(ctx, contestID)
}

func (s contestScopeSourceStub) FindContestTeam(ctx context.Context, contestID, teamID int64) (*contestentity.Team, error) {
	return s.findContestTeamFn(ctx, contestID, teamID)
}

func (s contestScopeSourceStub) ListContestTeams(ctx context.Context, contestID int64) ([]*contestentity.Team, error) {
	if s.listContestTeamsFn == nil {
		return nil, nil
	}
	return s.listContestTeamsFn(ctx, contestID)
}

func (s contestScopeSourceStub) FindContestRegistration(ctx context.Context, contestID, userID int64) (*practiceports.ContestParticipation, error) {
	return s.findContestRegistrationFn(ctx, contestID, userID)
}

func TestContestScopeRepositoryMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewContestScopeRepository(contestScopeSourceStub{
		findContestByIDFn: func(context.Context, int64) (*contestentity.Contest, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findContestChallengeFn: func(context.Context, int64, int64) (*contestentity.ContestChallenge, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findContestAWDServiceFn: func(context.Context, int64, int64) (*contestentity.ContestAWDService, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findContestTeamFn: func(context.Context, int64, int64) (*contestentity.Team, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findContestRegistrationFn: func(context.Context, int64, int64) (*practiceports.ContestParticipation, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	tests := []struct {
		name string
		run  func() error
		want error
	}{
		{
			name: "contest",
			run: func() error {
				_, err := repo.FindContestByID(context.Background(), 1)
				return err
			},
			want: practiceports.ErrPracticeContestNotFound,
		},
		{
			name: "contest challenge",
			run: func() error {
				_, err := repo.FindContestChallenge(context.Background(), 1, 2)
				return err
			},
			want: practiceports.ErrPracticeContestChallengeNotFound,
		},
		{
			name: "awd service",
			run: func() error {
				_, err := repo.FindContestAWDService(context.Background(), 1, 2)
				return err
			},
			want: practiceports.ErrPracticeContestAWDServiceNotFound,
		},
		{
			name: "awd runtime subject",
			run: func() error {
				_, err := repo.FindContestAWDServiceRuntimeSubject(context.Background(), 1, 2)
				return err
			},
			want: practiceports.ErrPracticeContestAWDServiceNotFound,
		},
		{
			name: "team",
			run: func() error {
				_, err := repo.FindContestTeam(context.Background(), 1, 2)
				return err
			},
			want: practiceports.ErrPracticeContestTeamNotFound,
		},
		{
			name: "registration",
			run: func() error {
				_, err := repo.FindContestRegistration(context.Background(), 1, 2)
				return err
			},
			want: practiceports.ErrPracticeContestRegistrationNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.run(); !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestContestScopeRepositoryPassesThroughNonNotFoundErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	repo := NewContestScopeRepository(contestScopeSourceStub{
		findContestByIDFn: func(context.Context, int64) (*contestentity.Contest, error) {
			return nil, expectedErr
		},
		findContestChallengeFn: func(context.Context, int64, int64) (*contestentity.ContestChallenge, error) {
			return &contestentity.ContestChallenge{ContestID: 1, ChallengeID: 2}, nil
		},
		findContestAWDServiceFn: func(context.Context, int64, int64) (*contestentity.ContestAWDService, error) {
			return &contestentity.ContestAWDService{ID: 2}, nil
		},
		findContestTeamFn: func(context.Context, int64, int64) (*contestentity.Team, error) {
			return &contestentity.Team{ID: 3}, nil
		},
		findContestRegistrationFn: func(context.Context, int64, int64) (*practiceports.ContestParticipation, error) {
			return &practiceports.ContestParticipation{}, nil
		},
	})

	_, err := repo.FindContestByID(context.Background(), 1)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
}
