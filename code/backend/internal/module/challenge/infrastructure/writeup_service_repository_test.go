package infrastructure

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type writeupServiceRepositoryStub struct {
	findByIDFn                              func(ctx context.Context, id int64) (*model.Challenge, error)
	findUserByIDFn                          func(ctx context.Context, userID int64) (*model.User, error)
	findWriteupByChallengeIDFn              func(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error)
	upsertWriteupFn                         func(ctx context.Context, writeup *model.ChallengeWriteup) error
	deleteWriteupByChallengeIDFn            func(ctx context.Context, challengeID int64) error
	findReleasedWriteupByChallengeIDFn      func(ctx context.Context, challengeID int64, now time.Time) (*model.ChallengeWriteup, error)
	getSolvedStatusFn                       func(ctx context.Context, userID, challengeID int64) (bool, error)
	findSubmissionWriteupByUserChallengeFn  func(ctx context.Context, userID, challengeID int64) (*model.SubmissionWriteup, error)
	findSubmissionWriteupByIDFn             func(ctx context.Context, id int64) (*model.SubmissionWriteup, error)
	upsertSubmissionWriteupFn               func(ctx context.Context, writeup *model.SubmissionWriteup) error
	getTeacherSubmissionWriteupByIDFn       func(ctx context.Context, id int64) (*challengeports.TeacherSubmissionWriteupRecord, error)
	listTeacherSubmissionWriteupsFn         func(ctx context.Context, query *dto.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error)
	listRecommendedSolutionsByChallengeIDFn func(ctx context.Context, challengeID int64, now time.Time) ([]challengeports.RecommendedSolutionRecord, error)
	listCommunitySolutionsByChallengeIDFn   func(ctx context.Context, challengeID int64, query *dto.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error)
}

func (s *writeupServiceRepositoryStub) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	return s.findByIDFn(ctx, id)
}

func (s *writeupServiceRepositoryStub) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	return s.findUserByIDFn(ctx, userID)
}

func (s *writeupServiceRepositoryStub) FindWriteupByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error) {
	return s.findWriteupByChallengeIDFn(ctx, challengeID)
}

func (s *writeupServiceRepositoryStub) UpsertWriteup(ctx context.Context, writeup *model.ChallengeWriteup) error {
	return s.upsertWriteupFn(ctx, writeup)
}

func (s *writeupServiceRepositoryStub) DeleteWriteupByChallengeID(ctx context.Context, challengeID int64) error {
	return s.deleteWriteupByChallengeIDFn(ctx, challengeID)
}

func (s *writeupServiceRepositoryStub) FindReleasedWriteupByChallengeID(ctx context.Context, challengeID int64, now time.Time) (*model.ChallengeWriteup, error) {
	return s.findReleasedWriteupByChallengeIDFn(ctx, challengeID, now)
}

func (s *writeupServiceRepositoryStub) GetSolvedStatus(ctx context.Context, userID, challengeID int64) (bool, error) {
	return s.getSolvedStatusFn(ctx, userID, challengeID)
}

func (s *writeupServiceRepositoryStub) FindSubmissionWriteupByUserChallenge(ctx context.Context, userID, challengeID int64) (*model.SubmissionWriteup, error) {
	return s.findSubmissionWriteupByUserChallengeFn(ctx, userID, challengeID)
}

func (s *writeupServiceRepositoryStub) FindSubmissionWriteupByID(ctx context.Context, id int64) (*model.SubmissionWriteup, error) {
	return s.findSubmissionWriteupByIDFn(ctx, id)
}

func (s *writeupServiceRepositoryStub) UpsertSubmissionWriteup(ctx context.Context, writeup *model.SubmissionWriteup) error {
	return s.upsertSubmissionWriteupFn(ctx, writeup)
}

func (s *writeupServiceRepositoryStub) GetTeacherSubmissionWriteupByID(ctx context.Context, id int64) (*challengeports.TeacherSubmissionWriteupRecord, error) {
	return s.getTeacherSubmissionWriteupByIDFn(ctx, id)
}

func (s *writeupServiceRepositoryStub) ListTeacherSubmissionWriteups(ctx context.Context, query *dto.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error) {
	return s.listTeacherSubmissionWriteupsFn(ctx, query)
}

func (s *writeupServiceRepositoryStub) ListRecommendedSolutionsByChallengeID(ctx context.Context, challengeID int64, now time.Time) ([]challengeports.RecommendedSolutionRecord, error) {
	return s.listRecommendedSolutionsByChallengeIDFn(ctx, challengeID, now)
}

func (s *writeupServiceRepositoryStub) ListCommunitySolutionsByChallengeID(ctx context.Context, challengeID int64, query *dto.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error) {
	return s.listCommunitySolutionsByChallengeIDFn(ctx, challengeID, query)
}

func TestWriteupRepositoryMapsRawNotFoundToPortsSentinels(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewWriteupServiceRepository(&writeupServiceRepositoryStub{
		findByIDFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findUserByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findWriteupByChallengeIDFn: func(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error) {
			return nil, gorm.ErrRecordNotFound
		},
		upsertWriteupFn: func(ctx context.Context, writeup *model.ChallengeWriteup) error {
			return nil
		},
		deleteWriteupByChallengeIDFn: func(ctx context.Context, challengeID int64) error {
			return nil
		},
		findReleasedWriteupByChallengeIDFn: func(ctx context.Context, challengeID int64, now time.Time) (*model.ChallengeWriteup, error) {
			return nil, gorm.ErrRecordNotFound
		},
		getSolvedStatusFn: func(ctx context.Context, userID, challengeID int64) (bool, error) {
			return false, nil
		},
		findSubmissionWriteupByUserChallengeFn: func(ctx context.Context, userID, challengeID int64) (*model.SubmissionWriteup, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findSubmissionWriteupByIDFn: func(ctx context.Context, id int64) (*model.SubmissionWriteup, error) {
			return nil, gorm.ErrRecordNotFound
		},
		upsertSubmissionWriteupFn: func(ctx context.Context, writeup *model.SubmissionWriteup) error {
			return nil
		},
		getTeacherSubmissionWriteupByIDFn: func(ctx context.Context, id int64) (*challengeports.TeacherSubmissionWriteupRecord, error) {
			return nil, gorm.ErrRecordNotFound
		},
		listTeacherSubmissionWriteupsFn: func(ctx context.Context, query *dto.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error) {
			return nil, 0, nil
		},
		listRecommendedSolutionsByChallengeIDFn: func(ctx context.Context, challengeID int64, now time.Time) ([]challengeports.RecommendedSolutionRecord, error) {
			return nil, nil
		},
		listCommunitySolutionsByChallengeIDFn: func(ctx context.Context, challengeID int64, query *dto.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error) {
			return nil, 0, nil
		},
	})

	cases := []struct {
		name string
		run  func() error
		want error
	}{
		{
			name: "challenge lookup",
			run: func() error {
				_, err := repo.FindByID(ctx, 11)
				return err
			},
			want: challengeports.ErrChallengeWriteupChallengeNotFound,
		},
		{
			name: "requester lookup",
			run: func() error {
				_, err := repo.FindUserByID(ctx, 7)
				return err
			},
			want: challengeports.ErrChallengeWriteupRequesterNotFound,
		},
		{
			name: "official writeup lookup",
			run: func() error {
				_, err := repo.FindWriteupByChallengeID(ctx, 11)
				return err
			},
			want: challengeports.ErrChallengeOfficialWriteupNotFound,
		},
		{
			name: "released writeup lookup",
			run: func() error {
				_, err := repo.FindReleasedWriteupByChallengeID(ctx, 11, time.Now())
				return err
			},
			want: challengeports.ErrChallengeReleasedWriteupNotFound,
		},
		{
			name: "submission by user challenge lookup",
			run: func() error {
				_, err := repo.FindSubmissionWriteupByUserChallenge(ctx, 7, 11)
				return err
			},
			want: challengeports.ErrChallengeSubmissionWriteupNotFound,
		},
		{
			name: "submission by id lookup",
			run: func() error {
				_, err := repo.FindSubmissionWriteupByID(ctx, 91)
				return err
			},
			want: challengeports.ErrChallengeSubmissionWriteupDetailNotFound,
		},
		{
			name: "teacher submission lookup",
			run: func() error {
				_, err := repo.GetTeacherSubmissionWriteupByID(ctx, 91)
				return err
			},
			want: challengeports.ErrChallengeTeacherSubmissionWriteupNotFound,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := tc.run(); !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
}
