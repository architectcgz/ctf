package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type participationRegistrationSourceStub struct {
	findRegistrationFn     func(context.Context, int64, int64) (*model.ContestRegistration, error)
	findRegistrationByIDFn func(context.Context, int64, int64) (*model.ContestRegistration, error)
}

func (s participationRegistrationSourceStub) FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	if s.findRegistrationFn != nil {
		return s.findRegistrationFn(ctx, contestID, userID)
	}
	return &model.ContestRegistration{}, nil
}

func (s participationRegistrationSourceStub) FindRegistrationByID(ctx context.Context, contestID, registrationID int64) (*model.ContestRegistration, error) {
	if s.findRegistrationByIDFn != nil {
		return s.findRegistrationByIDFn(ctx, contestID, registrationID)
	}
	return &model.ContestRegistration{}, nil
}

func (s participationRegistrationSourceStub) CreateRegistration(context.Context, *model.ContestRegistration) error {
	return nil
}

func (s participationRegistrationSourceStub) SaveRegistration(context.Context, *model.ContestRegistration) error {
	return nil
}

func (s participationRegistrationSourceStub) ListRegistrations(context.Context, int64, *string, int, int) ([]*contestports.ContestParticipationRegistrationRow, int64, error) {
	return nil, 0, nil
}

func (s participationRegistrationSourceStub) FindUserByID(context.Context, int64) (*model.User, error) {
	return &model.User{}, nil
}

func (s participationRegistrationSourceStub) ListAnnouncements(context.Context, int64) ([]*model.ContestAnnouncement, error) {
	return nil, nil
}

func (s participationRegistrationSourceStub) CreateAnnouncement(context.Context, *model.ContestAnnouncement) error {
	return nil
}

func (s participationRegistrationSourceStub) DeleteAnnouncement(context.Context, int64, int64) (bool, error) {
	return true, nil
}

func (s participationRegistrationSourceStub) ListSolvedProgress(context.Context, int64, int64) ([]*contestports.ContestParticipationSolvedProgressRow, error) {
	return nil, nil
}

func TestParticipationRegistrationRepositoryMapsFindRegistrationNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewParticipationRegistrationRepository(participationRegistrationSourceStub{
		findRegistrationFn: func(context.Context, int64, int64) (*model.ContestRegistration, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindRegistration(context.Background(), 1, 2); !errors.Is(err, contestports.ErrContestParticipationRegistrationNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestParticipationRegistrationNotFound)
	}
}

func TestParticipationRegistrationRepositoryMapsFindRegistrationByIDNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewParticipationRegistrationRepository(participationRegistrationSourceStub{
		findRegistrationByIDFn: func(context.Context, int64, int64) (*model.ContestRegistration, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindRegistrationByID(context.Background(), 1, 2); !errors.Is(err, contestports.ErrContestParticipationRegistrationNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestParticipationRegistrationNotFound)
	}
}
