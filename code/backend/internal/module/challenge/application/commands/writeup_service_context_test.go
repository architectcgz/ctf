package commands

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type writeupCommandContextStub struct {
	findByIDFn                                         func(id int64) (*model.Challenge, error)
	findByIDWithContextFn                              func(ctx context.Context, id int64) (*model.Challenge, error)
	findUserByIDFn                                     func(userID int64) (*model.User, error)
	findUserByIDWithContextFn                          func(ctx context.Context, userID int64) (*model.User, error)
	findWriteupByChallengeIDFn                         func(challengeID int64) (*model.ChallengeWriteup, error)
	findWriteupByChallengeIDWithContextFn              func(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error)
	upsertWriteupFn                                    func(writeup *model.ChallengeWriteup) error
	upsertWriteupWithContextFn                         func(ctx context.Context, writeup *model.ChallengeWriteup) error
	deleteWriteupByChallengeIDFn                       func(challengeID int64) error
	deleteWriteupByChallengeIDWithContextFn            func(ctx context.Context, challengeID int64) error
	findReleasedWriteupByChallengeIDFn                 func(challengeID int64, now time.Time) (*model.ChallengeWriteup, error)
	findReleasedWriteupByChallengeIDWithContextFn      func(ctx context.Context, challengeID int64, now time.Time) (*model.ChallengeWriteup, error)
	getSolvedStatusFn                                  func(userID, challengeID int64) (bool, error)
	getSolvedStatusWithContextFn                       func(ctx context.Context, userID, challengeID int64) (bool, error)
	findSubmissionWriteupByUserChallengeFn             func(userID, challengeID int64) (*model.SubmissionWriteup, error)
	findSubmissionWriteupByUserChallengeWithContextFn  func(ctx context.Context, userID, challengeID int64) (*model.SubmissionWriteup, error)
	findSubmissionWriteupByIDFn                        func(id int64) (*model.SubmissionWriteup, error)
	upsertSubmissionWriteupFn                          func(writeup *model.SubmissionWriteup) error
	upsertSubmissionWriteupWithContextFn               func(ctx context.Context, writeup *model.SubmissionWriteup) error
	getTeacherSubmissionWriteupByIDFn                  func(id int64) (*challengeports.TeacherSubmissionWriteupRecord, error)
	getTeacherSubmissionWriteupByIDWithContextFn       func(ctx context.Context, id int64) (*challengeports.TeacherSubmissionWriteupRecord, error)
	listTeacherSubmissionWriteupsFn                    func(query *dto.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error)
	listTeacherSubmissionWriteupsWithContextFn         func(ctx context.Context, query *dto.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error)
	listRecommendedSolutionsByChallengeIDFn            func(challengeID int64, now time.Time) ([]challengeports.RecommendedSolutionRecord, error)
	listRecommendedSolutionsByChallengeIDWithContextFn func(ctx context.Context, challengeID int64, now time.Time) ([]challengeports.RecommendedSolutionRecord, error)
	listCommunitySolutionsByChallengeIDFn              func(challengeID int64, query *dto.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error)
	listCommunitySolutionsByChallengeIDWithContextFn   func(ctx context.Context, challengeID int64, query *dto.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error)
}

func (s *writeupCommandContextStub) FindByID(id int64) (*model.Challenge, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(id)
	}
	return nil, nil
}

func (s *writeupCommandContextStub) FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return s.FindByID(id)
}

func (s *writeupCommandContextStub) FindUserByID(userID int64) (*model.User, error) {
	if s.findUserByIDFn != nil {
		return s.findUserByIDFn(userID)
	}
	return nil, nil
}

func (s *writeupCommandContextStub) FindUserByIDWithContext(ctx context.Context, userID int64) (*model.User, error) {
	if s.findUserByIDWithContextFn != nil {
		return s.findUserByIDWithContextFn(ctx, userID)
	}
	return s.FindUserByID(userID)
}

func (s *writeupCommandContextStub) FindWriteupByChallengeID(challengeID int64) (*model.ChallengeWriteup, error) {
	if s.findWriteupByChallengeIDFn != nil {
		return s.findWriteupByChallengeIDFn(challengeID)
	}
	return nil, nil
}

func (s *writeupCommandContextStub) FindWriteupByChallengeIDWithContext(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error) {
	if s.findWriteupByChallengeIDWithContextFn != nil {
		return s.findWriteupByChallengeIDWithContextFn(ctx, challengeID)
	}
	return s.FindWriteupByChallengeID(challengeID)
}

func (s *writeupCommandContextStub) UpsertWriteup(writeup *model.ChallengeWriteup) error {
	if s.upsertWriteupFn != nil {
		return s.upsertWriteupFn(writeup)
	}
	return nil
}

func (s *writeupCommandContextStub) UpsertWriteupWithContext(ctx context.Context, writeup *model.ChallengeWriteup) error {
	if s.upsertWriteupWithContextFn != nil {
		return s.upsertWriteupWithContextFn(ctx, writeup)
	}
	return s.UpsertWriteup(writeup)
}

func (s *writeupCommandContextStub) DeleteWriteupByChallengeID(challengeID int64) error {
	if s.deleteWriteupByChallengeIDFn != nil {
		return s.deleteWriteupByChallengeIDFn(challengeID)
	}
	return nil
}

func (s *writeupCommandContextStub) DeleteWriteupByChallengeIDWithContext(ctx context.Context, challengeID int64) error {
	if s.deleteWriteupByChallengeIDWithContextFn != nil {
		return s.deleteWriteupByChallengeIDWithContextFn(ctx, challengeID)
	}
	return s.DeleteWriteupByChallengeID(challengeID)
}

func (s *writeupCommandContextStub) FindReleasedWriteupByChallengeID(challengeID int64, now time.Time) (*model.ChallengeWriteup, error) {
	if s.findReleasedWriteupByChallengeIDFn != nil {
		return s.findReleasedWriteupByChallengeIDFn(challengeID, now)
	}
	return nil, nil
}

func (s *writeupCommandContextStub) FindReleasedWriteupByChallengeIDWithContext(ctx context.Context, challengeID int64, now time.Time) (*model.ChallengeWriteup, error) {
	if s.findReleasedWriteupByChallengeIDWithContextFn != nil {
		return s.findReleasedWriteupByChallengeIDWithContextFn(ctx, challengeID, now)
	}
	return s.FindReleasedWriteupByChallengeID(challengeID, now)
}

func (s *writeupCommandContextStub) GetSolvedStatus(userID, challengeID int64) (bool, error) {
	if s.getSolvedStatusFn != nil {
		return s.getSolvedStatusFn(userID, challengeID)
	}
	return false, nil
}

func (s *writeupCommandContextStub) GetSolvedStatusWithContext(ctx context.Context, userID, challengeID int64) (bool, error) {
	if s.getSolvedStatusWithContextFn != nil {
		return s.getSolvedStatusWithContextFn(ctx, userID, challengeID)
	}
	return s.GetSolvedStatus(userID, challengeID)
}

func (s *writeupCommandContextStub) FindSubmissionWriteupByUserChallenge(userID, challengeID int64) (*model.SubmissionWriteup, error) {
	if s.findSubmissionWriteupByUserChallengeFn != nil {
		return s.findSubmissionWriteupByUserChallengeFn(userID, challengeID)
	}
	return nil, nil
}

func (s *writeupCommandContextStub) FindSubmissionWriteupByUserChallengeWithContext(ctx context.Context, userID, challengeID int64) (*model.SubmissionWriteup, error) {
	if s.findSubmissionWriteupByUserChallengeWithContextFn != nil {
		return s.findSubmissionWriteupByUserChallengeWithContextFn(ctx, userID, challengeID)
	}
	return s.FindSubmissionWriteupByUserChallenge(userID, challengeID)
}

func (s *writeupCommandContextStub) FindSubmissionWriteupByID(id int64) (*model.SubmissionWriteup, error) {
	if s.findSubmissionWriteupByIDFn != nil {
		return s.findSubmissionWriteupByIDFn(id)
	}
	return nil, nil
}

func (s *writeupCommandContextStub) UpsertSubmissionWriteup(writeup *model.SubmissionWriteup) error {
	if s.upsertSubmissionWriteupFn != nil {
		return s.upsertSubmissionWriteupFn(writeup)
	}
	return nil
}

func (s *writeupCommandContextStub) UpsertSubmissionWriteupWithContext(ctx context.Context, writeup *model.SubmissionWriteup) error {
	if s.upsertSubmissionWriteupWithContextFn != nil {
		return s.upsertSubmissionWriteupWithContextFn(ctx, writeup)
	}
	return s.UpsertSubmissionWriteup(writeup)
}

func (s *writeupCommandContextStub) GetTeacherSubmissionWriteupByID(id int64) (*challengeports.TeacherSubmissionWriteupRecord, error) {
	if s.getTeacherSubmissionWriteupByIDFn != nil {
		return s.getTeacherSubmissionWriteupByIDFn(id)
	}
	return nil, nil
}

func (s *writeupCommandContextStub) GetTeacherSubmissionWriteupByIDWithContext(ctx context.Context, id int64) (*challengeports.TeacherSubmissionWriteupRecord, error) {
	if s.getTeacherSubmissionWriteupByIDWithContextFn != nil {
		return s.getTeacherSubmissionWriteupByIDWithContextFn(ctx, id)
	}
	return s.GetTeacherSubmissionWriteupByID(id)
}

func (s *writeupCommandContextStub) ListTeacherSubmissionWriteups(query *dto.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error) {
	if s.listTeacherSubmissionWriteupsFn != nil {
		return s.listTeacherSubmissionWriteupsFn(query)
	}
	return nil, 0, nil
}

func (s *writeupCommandContextStub) ListTeacherSubmissionWriteupsWithContext(ctx context.Context, query *dto.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error) {
	if s.listTeacherSubmissionWriteupsWithContextFn != nil {
		return s.listTeacherSubmissionWriteupsWithContextFn(ctx, query)
	}
	return s.ListTeacherSubmissionWriteups(query)
}

func (s *writeupCommandContextStub) ListRecommendedSolutionsByChallengeID(challengeID int64, now time.Time) ([]challengeports.RecommendedSolutionRecord, error) {
	if s.listRecommendedSolutionsByChallengeIDFn != nil {
		return s.listRecommendedSolutionsByChallengeIDFn(challengeID, now)
	}
	return nil, nil
}

func (s *writeupCommandContextStub) ListRecommendedSolutionsByChallengeIDWithContext(ctx context.Context, challengeID int64, now time.Time) ([]challengeports.RecommendedSolutionRecord, error) {
	if s.listRecommendedSolutionsByChallengeIDWithContextFn != nil {
		return s.listRecommendedSolutionsByChallengeIDWithContextFn(ctx, challengeID, now)
	}
	return s.ListRecommendedSolutionsByChallengeID(challengeID, now)
}

func (s *writeupCommandContextStub) ListCommunitySolutionsByChallengeID(challengeID int64, query *dto.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error) {
	if s.listCommunitySolutionsByChallengeIDFn != nil {
		return s.listCommunitySolutionsByChallengeIDFn(challengeID, query)
	}
	return nil, 0, nil
}

func (s *writeupCommandContextStub) ListCommunitySolutionsByChallengeIDWithContext(ctx context.Context, challengeID int64, query *dto.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error) {
	if s.listCommunitySolutionsByChallengeIDWithContextFn != nil {
		return s.listCommunitySolutionsByChallengeIDWithContextFn(ctx, challengeID, query)
	}
	return s.ListCommunitySolutionsByChallengeID(challengeID, query)
}

type writeupCommandContextKey string

func TestWriteupServiceUpsertWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := writeupCommandContextKey("writeup-upsert")
	expectedCtxValue := "ctx-writeup-upsert"
	findChallengeCalled := false
	findExistingCalled := false
	upsertCalled := false
	findUpdatedCalled := false
	repo := &writeupCommandContextStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findChallengeCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id}, nil
		},
		findWriteupByChallengeIDWithContextFn: func(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-writeup ctx value %v, got %v", expectedCtxValue, got)
			}
			if !upsertCalled {
				findExistingCalled = true
				return nil, nil
			}
			findUpdatedCalled = true
			return &model.ChallengeWriteup{ID: 21, ChallengeID: challengeID, Title: "Official", Content: "Walkthrough", Visibility: model.WriteupVisibilityPublic}, nil
		},
		upsertWriteupWithContextFn: func(ctx context.Context, writeup *model.ChallengeWriteup) error {
			upsertCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected upsert-writeup ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.UpsertWithContext(ctx, 11, 7, &dto.UpsertChallengeWriteupReq{Title: "Official", Content: "Walkthrough", Visibility: model.WriteupVisibilityPublic})
	if err != nil {
		t.Fatalf("UpsertWithContext() error = %v", err)
	}
	if !findChallengeCalled || !findExistingCalled || !upsertCalled || !findUpdatedCalled {
		t.Fatalf("expected repository calls, got challenge=%v existing=%v upsert=%v updated=%v", findChallengeCalled, findExistingCalled, upsertCalled, findUpdatedCalled)
	}
	if resp == nil || resp.ID != 21 || resp.ChallengeID != 11 {
		t.Fatalf("unexpected upsert resp: %+v", resp)
	}
}

func TestWriteupServiceDeleteWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := writeupCommandContextKey("writeup-delete")
	expectedCtxValue := "ctx-writeup-delete"
	findChallengeCalled := false
	deleteCalled := false
	repo := &writeupCommandContextStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findChallengeCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id}, nil
		},
		deleteWriteupByChallengeIDWithContextFn: func(ctx context.Context, challengeID int64) error {
			deleteCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected delete-writeup ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.DeleteWithContext(ctx, 11); err != nil {
		t.Fatalf("DeleteWithContext() error = %v", err)
	}
	if !findChallengeCalled || !deleteCalled {
		t.Fatalf("expected repository calls, got challenge=%v delete=%v", findChallengeCalled, deleteCalled)
	}
}

func TestWriteupServiceUpsertSubmissionWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := writeupCommandContextKey("writeup-upsert-submission")
	expectedCtxValue := "ctx-writeup-upsert-submission"
	findChallengeCalled := false
	findExistingCalled := false
	getSolvedCalled := false
	upsertCalled := false
	findUpdatedCalled := false
	repo := &writeupCommandContextStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findChallengeCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id, Status: model.ChallengeStatusPublished}, nil
		},
		findSubmissionWriteupByUserChallengeWithContextFn: func(ctx context.Context, userID, challengeID int64) (*model.SubmissionWriteup, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-submission ctx value %v, got %v", expectedCtxValue, got)
			}
			if !upsertCalled {
				findExistingCalled = true
				return nil, nil
			}
			findUpdatedCalled = true
			return &model.SubmissionWriteup{ID: 31, UserID: userID, ChallengeID: challengeID, SubmissionStatus: model.SubmissionWriteupStatusPublished, VisibilityStatus: model.SubmissionWriteupVisibilityVisible}, nil
		},
		getSolvedStatusWithContextFn: func(ctx context.Context, userID, challengeID int64) (bool, error) {
			getSolvedCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected get-solved ctx value %v, got %v", expectedCtxValue, got)
			}
			return true, nil
		},
		upsertSubmissionWriteupWithContextFn: func(ctx context.Context, writeup *model.SubmissionWriteup) error {
			upsertCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected upsert-submission ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.UpsertSubmissionWithContext(ctx, 11, 7, &dto.UpsertSubmissionWriteupReq{Title: "Published", Content: "Walkthrough", SubmissionStatus: model.SubmissionWriteupStatusPublished})
	if err != nil {
		t.Fatalf("UpsertSubmissionWithContext() error = %v", err)
	}
	if !findChallengeCalled || !findExistingCalled || !getSolvedCalled || !upsertCalled || !findUpdatedCalled {
		t.Fatalf("expected repository calls, got challenge=%v existing=%v solved=%v upsert=%v updated=%v", findChallengeCalled, findExistingCalled, getSolvedCalled, upsertCalled, findUpdatedCalled)
	}
	if resp == nil || resp.ID != 31 || resp.ChallengeID != 11 || resp.UserID != 7 {
		t.Fatalf("unexpected submission resp: %+v", resp)
	}
}

func TestWriteupServiceRecommendOfficialWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := writeupCommandContextKey("writeup-recommend-official")
	expectedCtxValue := "ctx-writeup-recommend-official"
	findChallengeCalled := false
	findWriteupCalled := false
	upsertCalled := false
	findUpdatedCalled := false
	repo := &writeupCommandContextStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findChallengeCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id}, nil
		},
		findWriteupByChallengeIDWithContextFn: func(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-writeup ctx value %v, got %v", expectedCtxValue, got)
			}
			if !upsertCalled {
				findWriteupCalled = true
				return &model.ChallengeWriteup{ID: 41, ChallengeID: challengeID, Title: "Official", Content: "Walkthrough", Visibility: model.WriteupVisibilityPublic}, nil
			}
			findUpdatedCalled = true
			return &model.ChallengeWriteup{ID: 41, ChallengeID: challengeID, Title: "Official", Content: "Walkthrough", Visibility: model.WriteupVisibilityPublic, IsRecommended: true}, nil
		},
		upsertWriteupWithContextFn: func(ctx context.Context, writeup *model.ChallengeWriteup) error {
			upsertCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected upsert-writeup ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.RecommendOfficialWithContext(ctx, 11, 7)
	if err != nil {
		t.Fatalf("RecommendOfficialWithContext() error = %v", err)
	}
	if !findChallengeCalled || !findWriteupCalled || !upsertCalled || !findUpdatedCalled {
		t.Fatalf("expected repository calls, got challenge=%v writeup=%v upsert=%v updated=%v", findChallengeCalled, findWriteupCalled, upsertCalled, findUpdatedCalled)
	}
	if resp == nil || !resp.IsRecommended || resp.ID != 41 {
		t.Fatalf("unexpected recommend official resp: %+v", resp)
	}
}

func TestWriteupServiceUnrecommendOfficialWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := writeupCommandContextKey("writeup-unrecommend-official")
	expectedCtxValue := "ctx-writeup-unrecommend-official"
	findChallengeCalled := false
	findWriteupCalled := false
	upsertCalled := false
	findUpdatedCalled := false
	repo := &writeupCommandContextStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findChallengeCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id}, nil
		},
		findWriteupByChallengeIDWithContextFn: func(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-writeup ctx value %v, got %v", expectedCtxValue, got)
			}
			if !upsertCalled {
				findWriteupCalled = true
				return &model.ChallengeWriteup{ID: 42, ChallengeID: challengeID, Title: "Official", Content: "Walkthrough", Visibility: model.WriteupVisibilityPublic, IsRecommended: true}, nil
			}
			findUpdatedCalled = true
			return &model.ChallengeWriteup{ID: 42, ChallengeID: challengeID, Title: "Official", Content: "Walkthrough", Visibility: model.WriteupVisibilityPublic, IsRecommended: false}, nil
		},
		upsertWriteupWithContextFn: func(ctx context.Context, writeup *model.ChallengeWriteup) error {
			upsertCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected upsert-writeup ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.UnrecommendOfficialWithContext(ctx, 11, 7)
	if err != nil {
		t.Fatalf("UnrecommendOfficialWithContext() error = %v", err)
	}
	if !findChallengeCalled || !findWriteupCalled || !upsertCalled || !findUpdatedCalled {
		t.Fatalf("expected repository calls, got challenge=%v writeup=%v upsert=%v updated=%v", findChallengeCalled, findWriteupCalled, upsertCalled, findUpdatedCalled)
	}
	if resp == nil || resp.IsRecommended || resp.ID != 42 {
		t.Fatalf("unexpected unrecommend official resp: %+v", resp)
	}
}
