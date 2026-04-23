package queries

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type stubChallengeWriteupRepository struct {
	findByIDFn                                        func(id int64) (*model.Challenge, error)
	findByIDWithContextFn                             func(ctx context.Context, id int64) (*model.Challenge, error)
	findUserByIDFn                                    func(userID int64) (*model.User, error)
	findWriteupByChallengeIDFn                        func(challengeID int64) (*model.ChallengeWriteup, error)
	findWriteupByChallengeIDWithContextFn             func(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error)
	upsertWriteupFn                                   func(writeup *model.ChallengeWriteup) error
	deleteWriteupByChallengeIDFn                      func(challengeID int64) error
	findReleasedWriteupByChallengeIDFn                func(challengeID int64, now time.Time) (*model.ChallengeWriteup, error)
	findReleasedWriteupByChallengeIDWithContextFn     func(ctx context.Context, challengeID int64, now time.Time) (*model.ChallengeWriteup, error)
	getSolvedStatusFn                                 func(userID, challengeID int64) (bool, error)
	getSolvedStatusWithContextFn                      func(ctx context.Context, userID, challengeID int64) (bool, error)
	findSubmissionWriteupByUserChallengeFn            func(userID, challengeID int64) (*model.SubmissionWriteup, error)
	findSubmissionWriteupByUserChallengeWithContextFn func(ctx context.Context, userID, challengeID int64) (*model.SubmissionWriteup, error)
	findSubmissionWriteupByIDFn                       func(id int64) (*model.SubmissionWriteup, error)
	upsertSubmissionWriteupFn                         func(writeup *model.SubmissionWriteup) error
	getTeacherSubmissionWriteupByIDFn                 func(id int64) (*challengeports.TeacherSubmissionWriteupRecord, error)
	listTeacherSubmissionWriteupsFn                   func(query *dto.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error)
	listRecommendedSolutionsByChallengeIDFn           func(challengeID int64, now time.Time) ([]challengeports.RecommendedSolutionRecord, error)
	listCommunitySolutionsByChallengeIDFn             func(challengeID int64, query *dto.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error)
}

func (s *stubChallengeWriteupRepository) FindByID(id int64) (*model.Challenge, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(id)
	}
	return nil, nil
}

func (s *stubChallengeWriteupRepository) FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return s.FindByID(id)
}

func (s *stubChallengeWriteupRepository) FindUserByID(userID int64) (*model.User, error) {
	if s.findUserByIDFn != nil {
		return s.findUserByIDFn(userID)
	}
	return nil, nil
}

func (s *stubChallengeWriteupRepository) FindWriteupByChallengeID(challengeID int64) (*model.ChallengeWriteup, error) {
	if s.findWriteupByChallengeIDFn != nil {
		return s.findWriteupByChallengeIDFn(challengeID)
	}
	return nil, nil
}

func (s *stubChallengeWriteupRepository) FindWriteupByChallengeIDWithContext(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error) {
	if s.findWriteupByChallengeIDWithContextFn != nil {
		return s.findWriteupByChallengeIDWithContextFn(ctx, challengeID)
	}
	return s.FindWriteupByChallengeID(challengeID)
}

func (s *stubChallengeWriteupRepository) UpsertWriteup(writeup *model.ChallengeWriteup) error {
	if s.upsertWriteupFn != nil {
		return s.upsertWriteupFn(writeup)
	}
	return nil
}

func (s *stubChallengeWriteupRepository) DeleteWriteupByChallengeID(challengeID int64) error {
	if s.deleteWriteupByChallengeIDFn != nil {
		return s.deleteWriteupByChallengeIDFn(challengeID)
	}
	return nil
}

func (s *stubChallengeWriteupRepository) FindReleasedWriteupByChallengeID(challengeID int64, now time.Time) (*model.ChallengeWriteup, error) {
	if s.findReleasedWriteupByChallengeIDFn != nil {
		return s.findReleasedWriteupByChallengeIDFn(challengeID, now)
	}
	return nil, nil
}

func (s *stubChallengeWriteupRepository) FindReleasedWriteupByChallengeIDWithContext(ctx context.Context, challengeID int64, now time.Time) (*model.ChallengeWriteup, error) {
	if s.findReleasedWriteupByChallengeIDWithContextFn != nil {
		return s.findReleasedWriteupByChallengeIDWithContextFn(ctx, challengeID, now)
	}
	return s.FindReleasedWriteupByChallengeID(challengeID, now)
}

func (s *stubChallengeWriteupRepository) GetSolvedStatus(userID, challengeID int64) (bool, error) {
	if s.getSolvedStatusFn != nil {
		return s.getSolvedStatusFn(userID, challengeID)
	}
	return false, nil
}

func (s *stubChallengeWriteupRepository) GetSolvedStatusWithContext(ctx context.Context, userID, challengeID int64) (bool, error) {
	if s.getSolvedStatusWithContextFn != nil {
		return s.getSolvedStatusWithContextFn(ctx, userID, challengeID)
	}
	return s.GetSolvedStatus(userID, challengeID)
}

func (s *stubChallengeWriteupRepository) FindSubmissionWriteupByUserChallenge(userID, challengeID int64) (*model.SubmissionWriteup, error) {
	if s.findSubmissionWriteupByUserChallengeFn != nil {
		return s.findSubmissionWriteupByUserChallengeFn(userID, challengeID)
	}
	return nil, nil
}

func (s *stubChallengeWriteupRepository) FindSubmissionWriteupByUserChallengeWithContext(ctx context.Context, userID, challengeID int64) (*model.SubmissionWriteup, error) {
	if s.findSubmissionWriteupByUserChallengeWithContextFn != nil {
		return s.findSubmissionWriteupByUserChallengeWithContextFn(ctx, userID, challengeID)
	}
	return s.FindSubmissionWriteupByUserChallenge(userID, challengeID)
}

func (s *stubChallengeWriteupRepository) FindSubmissionWriteupByID(id int64) (*model.SubmissionWriteup, error) {
	if s.findSubmissionWriteupByIDFn != nil {
		return s.findSubmissionWriteupByIDFn(id)
	}
	return nil, nil
}

func (s *stubChallengeWriteupRepository) UpsertSubmissionWriteup(writeup *model.SubmissionWriteup) error {
	if s.upsertSubmissionWriteupFn != nil {
		return s.upsertSubmissionWriteupFn(writeup)
	}
	return nil
}

func (s *stubChallengeWriteupRepository) GetTeacherSubmissionWriteupByID(id int64) (*challengeports.TeacherSubmissionWriteupRecord, error) {
	if s.getTeacherSubmissionWriteupByIDFn != nil {
		return s.getTeacherSubmissionWriteupByIDFn(id)
	}
	return nil, nil
}

func (s *stubChallengeWriteupRepository) ListTeacherSubmissionWriteups(query *dto.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error) {
	if s.listTeacherSubmissionWriteupsFn != nil {
		return s.listTeacherSubmissionWriteupsFn(query)
	}
	return nil, 0, nil
}

func (s *stubChallengeWriteupRepository) ListRecommendedSolutionsByChallengeID(challengeID int64, now time.Time) ([]challengeports.RecommendedSolutionRecord, error) {
	if s.listRecommendedSolutionsByChallengeIDFn != nil {
		return s.listRecommendedSolutionsByChallengeIDFn(challengeID, now)
	}
	return nil, nil
}

func (s *stubChallengeWriteupRepository) ListCommunitySolutionsByChallengeID(challengeID int64, query *dto.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error) {
	if s.listCommunitySolutionsByChallengeIDFn != nil {
		return s.listCommunitySolutionsByChallengeIDFn(challengeID, query)
	}
	return nil, 0, nil
}

type challengeWriteupContextKey string

func TestWriteupServiceGetAdminWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := challengeWriteupContextKey("writeup-admin")
	expectedCtxValue := "ctx-writeup-admin"
	findChallengeCalled := false
	findWriteupCalled := false
	repo := &stubChallengeWriteupRepository{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findChallengeCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id}, nil
		},
		findWriteupByChallengeIDWithContextFn: func(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error) {
			findWriteupCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-writeup ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.ChallengeWriteup{ID: 101, ChallengeID: challengeID, Title: "Official", Content: "Walkthrough", Visibility: model.WriteupVisibilityPrivate}, nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.GetAdminWithContext(ctx, 11)
	if err != nil {
		t.Fatalf("GetAdminWithContext() error = %v", err)
	}
	if !findChallengeCalled || !findWriteupCalled {
		t.Fatalf("expected repository calls, got challenge=%v writeup=%v", findChallengeCalled, findWriteupCalled)
	}
	if resp == nil || resp.ID != 101 || resp.ChallengeID != 11 || resp.Title != "Official" {
		t.Fatalf("unexpected writeup resp: %+v", resp)
	}
}

func TestWriteupServiceGetMySubmissionWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := challengeWriteupContextKey("writeup-my-submission")
	expectedCtxValue := "ctx-writeup-my-submission"
	findChallengeCalled := false
	findSubmissionCalled := false
	repo := &stubChallengeWriteupRepository{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findChallengeCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id, Status: model.ChallengeStatusPublished}, nil
		},
		findSubmissionWriteupByUserChallengeWithContextFn: func(ctx context.Context, userID, challengeID int64) (*model.SubmissionWriteup, error) {
			findSubmissionCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-submission ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.SubmissionWriteup{ID: 201, UserID: userID, ChallengeID: challengeID, Content: "my writeup"}, nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.GetMySubmissionWithContext(ctx, 7, 11)
	if err != nil {
		t.Fatalf("GetMySubmissionWithContext() error = %v", err)
	}
	if !findChallengeCalled || !findSubmissionCalled {
		t.Fatalf("expected repository calls, got challenge=%v submission=%v", findChallengeCalled, findSubmissionCalled)
	}
	if resp == nil || resp.ID != 201 || resp.UserID != 7 || resp.ChallengeID != 11 {
		t.Fatalf("unexpected submission resp: %+v", resp)
	}
}

func TestWriteupServiceGetPublishedWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := challengeWriteupContextKey("writeup-published")
	expectedCtxValue := "ctx-writeup-published"
	findChallengeCalled := false
	findReleasedCalled := false
	getSolvedCalled := false
	repo := &stubChallengeWriteupRepository{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findChallengeCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id, Status: model.ChallengeStatusPublished}, nil
		},
		findReleasedWriteupByChallengeIDWithContextFn: func(ctx context.Context, challengeID int64, now time.Time) (*model.ChallengeWriteup, error) {
			findReleasedCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-released ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.ChallengeWriteup{ID: 301, ChallengeID: challengeID, Title: "Published", Content: "walkthrough", Visibility: model.WriteupVisibilityPublic}, nil
		},
		getSolvedStatusWithContextFn: func(ctx context.Context, userID, challengeID int64) (bool, error) {
			getSolvedCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected get-solved ctx value %v, got %v", expectedCtxValue, got)
			}
			if userID != 7 || challengeID != 11 {
				t.Fatalf("unexpected get solved args: user=%d challenge=%d", userID, challengeID)
			}
			return true, nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.GetPublishedWithContext(ctx, 7, 11)
	if err != nil {
		t.Fatalf("GetPublishedWithContext() error = %v", err)
	}
	if !findChallengeCalled || !findReleasedCalled || !getSolvedCalled {
		t.Fatalf("expected repository calls, got challenge=%v released=%v solved=%v", findChallengeCalled, findReleasedCalled, getSolvedCalled)
	}
	if resp == nil || resp.ID != 301 || resp.ChallengeID != 11 || resp.Title != "Published" || resp.RequiresSpoilerWarning {
		t.Fatalf("unexpected published resp: %+v", resp)
	}
}
