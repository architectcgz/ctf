package commands

import (
	"context"
	"errors"
	"testing"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type participationContestLookupStub struct {
	findByIDFn func(context.Context, int64) (*model.Contest, error)
}

func (s participationContestLookupStub) FindByID(ctx context.Context, id int64) (*model.Contest, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return &model.Contest{ID: id, Status: model.ContestStatusRegistration}, nil
}

func (s participationContestLookupStub) List(context.Context, *string, int, int) ([]*model.Contest, int64, error) {
	return nil, 0, nil
}

type participationCommandRepoStub struct {
	findRegistrationFn     func(context.Context, int64, int64) (*model.ContestRegistration, error)
	findRegistrationByIDFn func(context.Context, int64, int64) (*model.ContestRegistration, error)
	createRegistrationFn   func(context.Context, *model.ContestRegistration) error
	saveRegistrationFn     func(context.Context, *model.ContestRegistration) error
	findUserByIDFn         func(context.Context, int64) (*model.User, error)
	createAnnouncementFn   func(context.Context, *model.ContestAnnouncement) error
	deleteAnnouncementFn   func(context.Context, int64, int64) (bool, error)
}

func (s participationCommandRepoStub) FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	if s.findRegistrationFn != nil {
		return s.findRegistrationFn(ctx, contestID, userID)
	}
	return nil, contestports.ErrContestParticipationRegistrationNotFound
}

func (s participationCommandRepoStub) FindRegistrationByID(ctx context.Context, contestID, registrationID int64) (*model.ContestRegistration, error) {
	if s.findRegistrationByIDFn != nil {
		return s.findRegistrationByIDFn(ctx, contestID, registrationID)
	}
	return nil, contestports.ErrContestParticipationRegistrationNotFound
}

func (s participationCommandRepoStub) CreateRegistration(ctx context.Context, registration *model.ContestRegistration) error {
	if s.createRegistrationFn != nil {
		return s.createRegistrationFn(ctx, registration)
	}
	return nil
}

func (s participationCommandRepoStub) SaveRegistration(ctx context.Context, registration *model.ContestRegistration) error {
	if s.saveRegistrationFn != nil {
		return s.saveRegistrationFn(ctx, registration)
	}
	return nil
}

func (s participationCommandRepoStub) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	if s.findUserByIDFn != nil {
		return s.findUserByIDFn(ctx, userID)
	}
	return &model.User{ID: userID, Username: "reviewer-target"}, nil
}

func (s participationCommandRepoStub) CreateAnnouncement(ctx context.Context, announcement *model.ContestAnnouncement) error {
	if s.createAnnouncementFn != nil {
		return s.createAnnouncementFn(ctx, announcement)
	}
	return nil
}

func (s participationCommandRepoStub) DeleteAnnouncement(ctx context.Context, contestID, announcementID int64) (bool, error) {
	if s.deleteAnnouncementFn != nil {
		return s.deleteAnnouncementFn(ctx, contestID, announcementID)
	}
	return true, nil
}

type participationTeamFinderStub struct {
	findUserTeamInContestFn func(context.Context, int64, int64) (*model.Team, error)
}

func (s participationTeamFinderStub) FindUserTeamInContest(ctx context.Context, userID, contestID int64) (*model.Team, error) {
	if s.findUserTeamInContestFn != nil {
		return s.findUserTeamInContestFn(ctx, userID, contestID)
	}
	return nil, contestports.ErrContestUserTeamNotFound
}

type submissionRepositoryStub struct {
	findRegistrationFn func(context.Context, int64, int64) (*model.ContestRegistration, error)
}

func (s submissionRepositoryStub) WithinScoringTransaction(ctx context.Context, fn func(repo contestports.ContestSubmissionScoringTxRepository) error) error {
	return errors.New("unexpected WithinScoringTransaction call")
}

func (s submissionRepositoryStub) FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	if s.findRegistrationFn != nil {
		return s.findRegistrationFn(ctx, contestID, userID)
	}
	return nil, contestports.ErrContestParticipationRegistrationNotFound
}

func (s submissionRepositoryStub) FindContestChallenge(context.Context, int64, int64) (*model.ContestChallenge, error) {
	return nil, errors.New("unexpected FindContestChallenge call")
}

func (s submissionRepositoryStub) FindChallengeByID(context.Context, int64) (*model.Challenge, error) {
	return nil, errors.New("unexpected FindChallengeByID call")
}

func (s submissionRepositoryStub) CreateSubmission(context.Context, *model.Submission) error {
	return errors.New("unexpected CreateSubmission call")
}

func TestParticipationServiceRegisterContestTreatsSentinelNotFoundAsPendingCreate(t *testing.T) {
	t.Parallel()

	createCalled := false
	service := NewParticipationService(
		participationContestLookupStub{},
		participationCommandRepoStub{
			createRegistrationFn: func(_ context.Context, registration *model.ContestRegistration) error {
				createCalled = true
				if registration.Status != model.ContestRegistrationStatusPending {
					t.Fatalf("unexpected status: %s", registration.Status)
				}
				if registration.TeamID != nil {
					t.Fatalf("expected nil team id, got %+v", registration.TeamID)
				}
				return nil
			},
		},
		participationTeamFinderStub{},
	)

	if err := service.RegisterContest(context.Background(), 10, 1001); err != nil {
		t.Fatalf("RegisterContest() error = %v", err)
	}
	if !createCalled {
		t.Fatal("expected registration create path")
	}
}

func TestParticipationServiceReviewRegistrationTreatsRegistrationNotFoundAsContestRegistrationNotFound(t *testing.T) {
	t.Parallel()

	service := NewParticipationService(
		participationContestLookupStub{},
		participationCommandRepoStub{
			findRegistrationByIDFn: func(context.Context, int64, int64) (*model.ContestRegistration, error) {
				return nil, contestports.ErrContestParticipationRegistrationNotFound
			},
		},
		participationTeamFinderStub{},
	)

	_, err := service.ReviewRegistration(context.Background(), 10, 404, 9001, ReviewRegistrationInput{
		Status: model.ContestRegistrationStatusApproved,
	})
	if err == nil {
		t.Fatal("expected registration not found")
	}
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrContestRegistrationNotFound.Code {
		t.Fatalf("expected errcode.ErrContestRegistrationNotFound, got %v", err)
	}
}

func TestSubmissionServiceResolveTeamIDTreatsMissingRegistrationAndTeamAsNotRegistered(t *testing.T) {
	t.Parallel()

	service := NewSubmissionService(
		participationContestLookupStub{
			findByIDFn: func(context.Context, int64) (*model.Contest, error) {
				return nil, contestdomain.ErrContestNotFound
			},
		},
		submissionRepositoryStub{
			findRegistrationFn: func(context.Context, int64, int64) (*model.ContestRegistration, error) {
				return nil, contestports.ErrContestParticipationRegistrationNotFound
			},
		},
		nil,
		nil,
		participationTeamFinderStub{
			findUserTeamInContestFn: func(context.Context, int64, int64) (*model.Team, error) {
				return nil, contestports.ErrContestUserTeamNotFound
			},
		},
		nil,
		nil,
	)

	teamID, err := service.resolveTeamID(context.Background(), 1001, 10)
	if !errors.Is(err, errcode.ErrNotRegistered) {
		t.Fatalf("expected ErrNotRegistered, got %v", err)
	}
	if teamID != nil {
		t.Fatalf("expected nil team id, got %v", *teamID)
	}
}
