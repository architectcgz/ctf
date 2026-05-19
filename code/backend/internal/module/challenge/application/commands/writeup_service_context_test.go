package commands

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/model"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"ctf-platform/pkg/errcode"
)

type writeupCommandContextStub struct {
	findByIDFn                                         func(id int64) (*model.Challenge, error)
	findByIDWithContextFn                              func(ctx context.Context, id int64) (*model.Challenge, error)
	findUserByIDFn                                     func(userID int64) (*identitycontracts.User, error)
	findUserByIDWithContextFn                          func(ctx context.Context, userID int64) (*identitycontracts.User, error)
	findWriteupByChallengeIDFn                         func(challengeID int64) (*challengeentity.ChallengeWriteup, error)
	findWriteupByChallengeIDWithContextFn              func(ctx context.Context, challengeID int64) (*challengeentity.ChallengeWriteup, error)
	upsertWriteupFn                                    func(writeup *challengeentity.ChallengeWriteup) error
	upsertWriteupWithContextFn                         func(ctx context.Context, writeup *challengeentity.ChallengeWriteup) error
	deleteWriteupByChallengeIDFn                       func(challengeID int64) error
	deleteWriteupByChallengeIDWithContextFn            func(ctx context.Context, challengeID int64) error
	findReleasedWriteupByChallengeIDFn                 func(challengeID int64, now time.Time) (*challengeentity.ChallengeWriteup, error)
	findReleasedWriteupByChallengeIDWithContextFn      func(ctx context.Context, challengeID int64, now time.Time) (*challengeentity.ChallengeWriteup, error)
	getSolvedStatusFn                                  func(userID, challengeID int64) (bool, error)
	getSolvedStatusWithContextFn                       func(ctx context.Context, userID, challengeID int64) (bool, error)
	findSubmissionWriteupByUserChallengeFn             func(userID, challengeID int64) (*challengeentity.SubmissionWriteup, error)
	findSubmissionWriteupByUserChallengeWithContextFn  func(ctx context.Context, userID, challengeID int64) (*challengeentity.SubmissionWriteup, error)
	findSubmissionWriteupByIDFn                        func(id int64) (*challengeentity.SubmissionWriteup, error)
	findSubmissionWriteupByIDWithContextFn             func(ctx context.Context, id int64) (*challengeentity.SubmissionWriteup, error)
	upsertSubmissionWriteupFn                          func(writeup *challengeentity.SubmissionWriteup) error
	upsertSubmissionWriteupWithContextFn               func(ctx context.Context, writeup *challengeentity.SubmissionWriteup) error
	getTeacherSubmissionWriteupByIDFn                  func(id int64) (*challengeports.TeacherSubmissionWriteupRecord, error)
	getTeacherSubmissionWriteupByIDWithContextFn       func(ctx context.Context, id int64) (*challengeports.TeacherSubmissionWriteupRecord, error)
	listTeacherSubmissionWriteupsFn                    func(query *challengecontracts.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error)
	listTeacherSubmissionWriteupsWithContextFn         func(ctx context.Context, query *challengecontracts.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error)
	listRecommendedSolutionsByChallengeIDFn            func(challengeID int64, now time.Time) ([]challengeports.RecommendedSolutionRecord, error)
	listRecommendedSolutionsByChallengeIDWithContextFn func(ctx context.Context, challengeID int64, now time.Time) ([]challengeports.RecommendedSolutionRecord, error)
	listCommunitySolutionsByChallengeIDFn              func(challengeID int64, query *challengecontracts.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error)
	listCommunitySolutionsByChallengeIDWithContextFn   func(ctx context.Context, challengeID int64, query *challengecontracts.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error)
}

func (s *writeupCommandContextStub) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *writeupCommandContextStub) FindUserByID(ctx context.Context, userID int64) (*identitycontracts.User, error) {
	if s.findUserByIDWithContextFn != nil {
		return s.findUserByIDWithContextFn(ctx, userID)
	}
	return nil, nil
}

func (s *writeupCommandContextStub) FindWriteupByChallengeID(ctx context.Context, challengeID int64) (*challengeentity.ChallengeWriteup, error) {
	if s.findWriteupByChallengeIDWithContextFn != nil {
		return s.findWriteupByChallengeIDWithContextFn(ctx, challengeID)
	}
	return nil, nil
}

func (s *writeupCommandContextStub) UpsertWriteup(ctx context.Context, writeup *challengeentity.ChallengeWriteup) error {
	if s.upsertWriteupWithContextFn != nil {
		return s.upsertWriteupWithContextFn(ctx, writeup)
	}
	return nil
}

func (s *writeupCommandContextStub) DeleteWriteupByChallengeID(ctx context.Context, challengeID int64) error {
	if s.deleteWriteupByChallengeIDWithContextFn != nil {
		return s.deleteWriteupByChallengeIDWithContextFn(ctx, challengeID)
	}
	return nil
}

func (s *writeupCommandContextStub) FindReleasedWriteupByChallengeID(ctx context.Context, challengeID int64, now time.Time) (*challengeentity.ChallengeWriteup, error) {
	if s.findReleasedWriteupByChallengeIDWithContextFn != nil {
		return s.findReleasedWriteupByChallengeIDWithContextFn(ctx, challengeID, now)
	}
	return nil, nil
}

func (s *writeupCommandContextStub) GetSolvedStatus(ctx context.Context, userID, challengeID int64) (bool, error) {
	if s.getSolvedStatusWithContextFn != nil {
		return s.getSolvedStatusWithContextFn(ctx, userID, challengeID)
	}
	return false, nil
}

func (s *writeupCommandContextStub) FindSubmissionWriteupByUserChallenge(ctx context.Context, userID, challengeID int64) (*challengeentity.SubmissionWriteup, error) {
	if s.findSubmissionWriteupByUserChallengeWithContextFn != nil {
		return s.findSubmissionWriteupByUserChallengeWithContextFn(ctx, userID, challengeID)
	}
	return nil, nil
}

func (s *writeupCommandContextStub) FindSubmissionWriteupByID(ctx context.Context, id int64) (*challengeentity.SubmissionWriteup, error) {
	if s.findSubmissionWriteupByIDWithContextFn != nil {
		return s.findSubmissionWriteupByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *writeupCommandContextStub) UpsertSubmissionWriteup(ctx context.Context, writeup *challengeentity.SubmissionWriteup) error {
	if s.upsertSubmissionWriteupWithContextFn != nil {
		return s.upsertSubmissionWriteupWithContextFn(ctx, writeup)
	}
	return nil
}

func (s *writeupCommandContextStub) GetTeacherSubmissionWriteupByID(ctx context.Context, id int64) (*challengeports.TeacherSubmissionWriteupRecord, error) {
	if s.getTeacherSubmissionWriteupByIDWithContextFn != nil {
		return s.getTeacherSubmissionWriteupByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *writeupCommandContextStub) ListTeacherSubmissionWriteups(ctx context.Context, query *challengecontracts.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error) {
	if s.listTeacherSubmissionWriteupsWithContextFn != nil {
		return s.listTeacherSubmissionWriteupsWithContextFn(ctx, query)
	}
	return nil, 0, nil
}

func (s *writeupCommandContextStub) ListRecommendedSolutionsByChallengeID(ctx context.Context, challengeID int64, now time.Time) ([]challengeports.RecommendedSolutionRecord, error) {
	if s.listRecommendedSolutionsByChallengeIDWithContextFn != nil {
		return s.listRecommendedSolutionsByChallengeIDWithContextFn(ctx, challengeID, now)
	}
	return nil, nil
}

func (s *writeupCommandContextStub) ListCommunitySolutionsByChallengeID(ctx context.Context, challengeID int64, query *challengecontracts.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error) {
	if s.listCommunitySolutionsByChallengeIDWithContextFn != nil {
		return s.listCommunitySolutionsByChallengeIDWithContextFn(ctx, challengeID, query)
	}
	return nil, 0, nil
}

type writeupCommandContextKey string

func TestWriteupServiceUpsertPropagatesContextToRepository(t *testing.T) {
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
		findWriteupByChallengeIDWithContextFn: func(ctx context.Context, challengeID int64) (*challengeentity.ChallengeWriteup, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-writeup ctx value %v, got %v", expectedCtxValue, got)
			}
			if !upsertCalled {
				findExistingCalled = true
				return nil, nil
			}
			findUpdatedCalled = true
			return &challengeentity.ChallengeWriteup{ID: 21, ChallengeID: challengeID, Title: "Official", Content: "Walkthrough", Visibility: challengeentity.WriteupVisibilityPublic}, nil
		},
		upsertWriteupWithContextFn: func(ctx context.Context, writeup *challengeentity.ChallengeWriteup) error {
			upsertCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected upsert-writeup ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.Upsert(ctx, 11, 7, UpsertOfficialWriteupInput{Title: "Official", Content: "Walkthrough", Visibility: challengeentity.WriteupVisibilityPublic})
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	if !findChallengeCalled || !findExistingCalled || !upsertCalled || !findUpdatedCalled {
		t.Fatalf("expected repository calls, got challenge=%v existing=%v upsert=%v updated=%v", findChallengeCalled, findExistingCalled, upsertCalled, findUpdatedCalled)
	}
	if resp == nil || resp.ID != 21 || resp.ChallengeID != 11 {
		t.Fatalf("unexpected upsert resp: %+v", resp)
	}
}

func TestWriteupServiceDeletePropagatesContextToRepository(t *testing.T) {
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
	if err := service.Delete(ctx, 11); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if !findChallengeCalled || !deleteCalled {
		t.Fatalf("expected repository calls, got challenge=%v delete=%v", findChallengeCalled, deleteCalled)
	}
}

func TestWriteupServiceUpsertSubmissionPropagatesContextToRepository(t *testing.T) {
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
		findSubmissionWriteupByUserChallengeWithContextFn: func(ctx context.Context, userID, challengeID int64) (*challengeentity.SubmissionWriteup, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-submission ctx value %v, got %v", expectedCtxValue, got)
			}
			if !upsertCalled {
				findExistingCalled = true
				return nil, nil
			}
			findUpdatedCalled = true
			return &challengeentity.SubmissionWriteup{ID: 31, UserID: userID, ChallengeID: challengeID, SubmissionStatus: challengeentity.SubmissionWriteupStatusPublished, VisibilityStatus: challengeentity.SubmissionWriteupVisibilityVisible}, nil
		},
		getSolvedStatusWithContextFn: func(ctx context.Context, userID, challengeID int64) (bool, error) {
			getSolvedCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected get-solved ctx value %v, got %v", expectedCtxValue, got)
			}
			return true, nil
		},
		upsertSubmissionWriteupWithContextFn: func(ctx context.Context, writeup *challengeentity.SubmissionWriteup) error {
			upsertCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected upsert-submission ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.UpsertSubmission(ctx, 11, 7, UpsertSubmissionWriteupInput{Title: "Published", Content: "Walkthrough", SubmissionStatus: challengeentity.SubmissionWriteupStatusPublished})
	if err != nil {
		t.Fatalf("UpsertSubmission() error = %v", err)
	}
	if !findChallengeCalled || !findExistingCalled || !getSolvedCalled || !upsertCalled || !findUpdatedCalled {
		t.Fatalf("expected repository calls, got challenge=%v existing=%v solved=%v upsert=%v updated=%v", findChallengeCalled, findExistingCalled, getSolvedCalled, upsertCalled, findUpdatedCalled)
	}
	if resp == nil || resp.ID != 31 || resp.ChallengeID != 11 || resp.UserID != 7 {
		t.Fatalf("unexpected submission resp: %+v", resp)
	}
}

func TestWriteupServiceRecommendOfficialPropagatesContextToRepository(t *testing.T) {
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
		findWriteupByChallengeIDWithContextFn: func(ctx context.Context, challengeID int64) (*challengeentity.ChallengeWriteup, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-writeup ctx value %v, got %v", expectedCtxValue, got)
			}
			if !upsertCalled {
				findWriteupCalled = true
				return &challengeentity.ChallengeWriteup{ID: 41, ChallengeID: challengeID, Title: "Official", Content: "Walkthrough", Visibility: challengeentity.WriteupVisibilityPublic}, nil
			}
			findUpdatedCalled = true
			return &challengeentity.ChallengeWriteup{ID: 41, ChallengeID: challengeID, Title: "Official", Content: "Walkthrough", Visibility: challengeentity.WriteupVisibilityPublic, IsRecommended: true}, nil
		},
		upsertWriteupWithContextFn: func(ctx context.Context, writeup *challengeentity.ChallengeWriteup) error {
			upsertCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected upsert-writeup ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.RecommendOfficial(ctx, 11, 7)
	if err != nil {
		t.Fatalf("RecommendOfficial() error = %v", err)
	}
	if !findChallengeCalled || !findWriteupCalled || !upsertCalled || !findUpdatedCalled {
		t.Fatalf("expected repository calls, got challenge=%v writeup=%v upsert=%v updated=%v", findChallengeCalled, findWriteupCalled, upsertCalled, findUpdatedCalled)
	}
	if resp == nil || !resp.IsRecommended || resp.ID != 41 {
		t.Fatalf("unexpected recommend official resp: %+v", resp)
	}
}

func TestWriteupServiceUpsertTreatsChallengeNotFoundSentinelAsChallengeNotFound(t *testing.T) {
	t.Parallel()

	service := NewWriteupService(&writeupCommandContextStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			return nil, challengeports.ErrChallengeWriteupChallengeNotFound
		},
	})

	_, err := service.Upsert(context.Background(), 11, 7, UpsertOfficialWriteupInput{
		Title:      "Official",
		Content:    "Walkthrough",
		Visibility: challengeentity.WriteupVisibilityPublic,
	})
	if err == nil || err.Error() != errcode.ErrChallengeNotFound.Error() {
		t.Fatalf("expected challenge not found, got %v", err)
	}
}

func TestWriteupServiceUpsertTreatsOfficialWriteupNotFoundSentinelAsCreateFlow(t *testing.T) {
	t.Parallel()

	upsertCalled := false
	service := NewWriteupService(&writeupCommandContextStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			return &model.Challenge{ID: id}, nil
		},
		findWriteupByChallengeIDWithContextFn: func(ctx context.Context, challengeID int64) (*challengeentity.ChallengeWriteup, error) {
			if upsertCalled {
				return &challengeentity.ChallengeWriteup{
					ID:          21,
					ChallengeID: challengeID,
					Title:       "Official",
					Content:     "Walkthrough",
					Visibility:  challengeentity.WriteupVisibilityPublic,
				}, nil
			}
			return nil, challengeports.ErrChallengeOfficialWriteupNotFound
		},
		upsertWriteupWithContextFn: func(ctx context.Context, writeup *challengeentity.ChallengeWriteup) error {
			upsertCalled = true
			return nil
		},
	})

	resp, err := service.Upsert(context.Background(), 11, 7, UpsertOfficialWriteupInput{
		Title:      "Official",
		Content:    "Walkthrough",
		Visibility: challengeentity.WriteupVisibilityPublic,
	})
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	if !upsertCalled {
		t.Fatal("expected upsert to be called")
	}
	if resp == nil || resp.ID != 21 || resp.ChallengeID != 11 {
		t.Fatalf("unexpected upsert resp: %+v", resp)
	}
}

func TestWriteupServiceRecommendCommunityTreatsRequesterNotFoundSentinelAsUnauthorized(t *testing.T) {
	t.Parallel()

	now := time.Now()
	service := NewWriteupService(&writeupCommandContextStub{
		getTeacherSubmissionWriteupByIDWithContextFn: func(ctx context.Context, id int64) (*challengeports.TeacherSubmissionWriteupRecord, error) {
			return &challengeports.TeacherSubmissionWriteupRecord{
				Submission:      challengeentity.SubmissionWriteup{ID: id, ChallengeID: 11, UserID: 88, CreatedAt: now, UpdatedAt: now},
				ClassName:       "Class A",
				StudentUsername: "student",
				ChallengeTitle:  "challenge",
			}, nil
		},
		findUserByIDWithContextFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			return nil, challengeports.ErrChallengeWriteupRequesterNotFound
		},
	})

	_, err := service.RecommendCommunity(context.Background(), 91, 1001, identitycontracts.RoleTeacher)
	if err == nil || err.Error() != errcode.ErrUnauthorized.Error() {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestWriteupServiceUnrecommendOfficialPropagatesContextToRepository(t *testing.T) {
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
		findWriteupByChallengeIDWithContextFn: func(ctx context.Context, challengeID int64) (*challengeentity.ChallengeWriteup, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-writeup ctx value %v, got %v", expectedCtxValue, got)
			}
			if !upsertCalled {
				findWriteupCalled = true
				return &challengeentity.ChallengeWriteup{ID: 42, ChallengeID: challengeID, Title: "Official", Content: "Walkthrough", Visibility: challengeentity.WriteupVisibilityPublic, IsRecommended: true}, nil
			}
			findUpdatedCalled = true
			return &challengeentity.ChallengeWriteup{ID: 42, ChallengeID: challengeID, Title: "Official", Content: "Walkthrough", Visibility: challengeentity.WriteupVisibilityPublic, IsRecommended: false}, nil
		},
		upsertWriteupWithContextFn: func(ctx context.Context, writeup *challengeentity.ChallengeWriteup) error {
			upsertCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected upsert-writeup ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.UnrecommendOfficial(ctx, 11, 7)
	if err != nil {
		t.Fatalf("UnrecommendOfficial() error = %v", err)
	}
	if !findChallengeCalled || !findWriteupCalled || !upsertCalled || !findUpdatedCalled {
		t.Fatalf("expected repository calls, got challenge=%v writeup=%v upsert=%v updated=%v", findChallengeCalled, findWriteupCalled, upsertCalled, findUpdatedCalled)
	}
	if resp == nil || resp.IsRecommended || resp.ID != 42 {
		t.Fatalf("unexpected unrecommend official resp: %+v", resp)
	}
}

func TestWriteupServiceRecommendCommunityPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := writeupCommandContextKey("writeup-recommend-community")
	expectedCtxValue := "ctx-writeup-recommend-community"
	getTeacherRecordCalled := false
	findRequesterCalled := false
	findSubmissionCalled := false
	upsertCalled := false
	findUpdatedCalled := false
	repo := &writeupCommandContextStub{
		getTeacherSubmissionWriteupByIDWithContextFn: func(ctx context.Context, id int64) (*challengeports.TeacherSubmissionWriteupRecord, error) {
			getTeacherRecordCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected get-teacher-record ctx value %v, got %v", expectedCtxValue, got)
			}
			return &challengeports.TeacherSubmissionWriteupRecord{
				Submission:      challengeentity.SubmissionWriteup{ID: id, UserID: 88, ChallengeID: 11, VisibilityStatus: challengeentity.SubmissionWriteupVisibilityVisible},
				StudentUsername: "student88",
				ClassName:       "Class A",
			}, nil
		},
		findUserByIDWithContextFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			findRequesterCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-user ctx value %v, got %v", expectedCtxValue, got)
			}
			return &identitycontracts.User{ID: userID, Role: identitycontracts.RoleTeacher, ClassName: "Class A"}, nil
		},
		findSubmissionWriteupByIDWithContextFn: func(ctx context.Context, id int64) (*challengeentity.SubmissionWriteup, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-submission ctx value %v, got %v", expectedCtxValue, got)
			}
			if !upsertCalled {
				findSubmissionCalled = true
				return &challengeentity.SubmissionWriteup{ID: id, UserID: 88, ChallengeID: 11, VisibilityStatus: challengeentity.SubmissionWriteupVisibilityVisible}, nil
			}
			findUpdatedCalled = true
			return &challengeentity.SubmissionWriteup{ID: id, UserID: 88, ChallengeID: 11, VisibilityStatus: challengeentity.SubmissionWriteupVisibilityVisible, IsRecommended: true}, nil
		},
		upsertSubmissionWriteupWithContextFn: func(ctx context.Context, writeup *challengeentity.SubmissionWriteup) error {
			upsertCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected upsert-submission ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.RecommendCommunity(ctx, 31, 1001, identitycontracts.RoleTeacher)
	if err != nil {
		t.Fatalf("RecommendCommunity() error = %v", err)
	}
	if !getTeacherRecordCalled || !findRequesterCalled || !findSubmissionCalled || !upsertCalled || !findUpdatedCalled {
		t.Fatalf("expected repository calls, got record=%v requester=%v submission=%v upsert=%v updated=%v", getTeacherRecordCalled, findRequesterCalled, findSubmissionCalled, upsertCalled, findUpdatedCalled)
	}
	if resp == nil || !resp.IsRecommended || resp.ID != 31 {
		t.Fatalf("unexpected recommend community resp: %+v", resp)
	}
}

func TestWriteupServiceUnrecommendCommunityPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := writeupCommandContextKey("writeup-unrecommend-community")
	expectedCtxValue := "ctx-writeup-unrecommend-community"
	getTeacherRecordCalled := false
	findRequesterCalled := false
	findSubmissionCalled := false
	upsertCalled := false
	findUpdatedCalled := false
	repo := &writeupCommandContextStub{
		getTeacherSubmissionWriteupByIDWithContextFn: func(ctx context.Context, id int64) (*challengeports.TeacherSubmissionWriteupRecord, error) {
			getTeacherRecordCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected get-teacher-record ctx value %v, got %v", expectedCtxValue, got)
			}
			return &challengeports.TeacherSubmissionWriteupRecord{
				Submission:      challengeentity.SubmissionWriteup{ID: id, UserID: 88, ChallengeID: 11, VisibilityStatus: challengeentity.SubmissionWriteupVisibilityVisible, IsRecommended: true},
				StudentUsername: "student88",
				ClassName:       "Class A",
			}, nil
		},
		findUserByIDWithContextFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			findRequesterCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-user ctx value %v, got %v", expectedCtxValue, got)
			}
			return &identitycontracts.User{ID: userID, Role: identitycontracts.RoleTeacher, ClassName: "Class A"}, nil
		},
		findSubmissionWriteupByIDWithContextFn: func(ctx context.Context, id int64) (*challengeentity.SubmissionWriteup, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-submission ctx value %v, got %v", expectedCtxValue, got)
			}
			if !upsertCalled {
				findSubmissionCalled = true
				return &challengeentity.SubmissionWriteup{ID: id, UserID: 88, ChallengeID: 11, VisibilityStatus: challengeentity.SubmissionWriteupVisibilityVisible, IsRecommended: true}, nil
			}
			findUpdatedCalled = true
			return &challengeentity.SubmissionWriteup{ID: id, UserID: 88, ChallengeID: 11, VisibilityStatus: challengeentity.SubmissionWriteupVisibilityVisible, IsRecommended: false}, nil
		},
		upsertSubmissionWriteupWithContextFn: func(ctx context.Context, writeup *challengeentity.SubmissionWriteup) error {
			upsertCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected upsert-submission ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.UnrecommendCommunity(ctx, 31, 1001, identitycontracts.RoleTeacher)
	if err != nil {
		t.Fatalf("UnrecommendCommunity() error = %v", err)
	}
	if !getTeacherRecordCalled || !findRequesterCalled || !findSubmissionCalled || !upsertCalled || !findUpdatedCalled {
		t.Fatalf("expected repository calls, got record=%v requester=%v submission=%v upsert=%v updated=%v", getTeacherRecordCalled, findRequesterCalled, findSubmissionCalled, upsertCalled, findUpdatedCalled)
	}
	if resp == nil || resp.IsRecommended || resp.ID != 31 {
		t.Fatalf("unexpected unrecommend community resp: %+v", resp)
	}
}

func TestWriteupServiceHideCommunityPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := writeupCommandContextKey("writeup-hide-community")
	expectedCtxValue := "ctx-writeup-hide-community"
	getTeacherRecordCalled := false
	findRequesterCalled := false
	findSubmissionCalled := false
	upsertCalled := false
	findUpdatedCalled := false
	repo := &writeupCommandContextStub{
		getTeacherSubmissionWriteupByIDWithContextFn: func(ctx context.Context, id int64) (*challengeports.TeacherSubmissionWriteupRecord, error) {
			getTeacherRecordCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected get-teacher-record ctx value %v, got %v", expectedCtxValue, got)
			}
			return &challengeports.TeacherSubmissionWriteupRecord{
				Submission:      challengeentity.SubmissionWriteup{ID: id, UserID: 88, ChallengeID: 11, VisibilityStatus: challengeentity.SubmissionWriteupVisibilityVisible, IsRecommended: true},
				StudentUsername: "student88",
				ClassName:       "Class A",
			}, nil
		},
		findUserByIDWithContextFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			findRequesterCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-user ctx value %v, got %v", expectedCtxValue, got)
			}
			return &identitycontracts.User{ID: userID, Role: identitycontracts.RoleTeacher, ClassName: "Class A"}, nil
		},
		findSubmissionWriteupByIDWithContextFn: func(ctx context.Context, id int64) (*challengeentity.SubmissionWriteup, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-submission ctx value %v, got %v", expectedCtxValue, got)
			}
			if !upsertCalled {
				findSubmissionCalled = true
				return &challengeentity.SubmissionWriteup{ID: id, UserID: 88, ChallengeID: 11, VisibilityStatus: challengeentity.SubmissionWriteupVisibilityVisible, IsRecommended: true}, nil
			}
			findUpdatedCalled = true
			return &challengeentity.SubmissionWriteup{ID: id, UserID: 88, ChallengeID: 11, VisibilityStatus: challengeentity.SubmissionWriteupVisibilityHidden, IsRecommended: false}, nil
		},
		upsertSubmissionWriteupWithContextFn: func(ctx context.Context, writeup *challengeentity.SubmissionWriteup) error {
			upsertCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected upsert-submission ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.HideCommunity(ctx, 31, 1001, identitycontracts.RoleTeacher)
	if err != nil {
		t.Fatalf("HideCommunity() error = %v", err)
	}
	if !getTeacherRecordCalled || !findRequesterCalled || !findSubmissionCalled || !upsertCalled || !findUpdatedCalled {
		t.Fatalf("expected repository calls, got record=%v requester=%v submission=%v upsert=%v updated=%v", getTeacherRecordCalled, findRequesterCalled, findSubmissionCalled, upsertCalled, findUpdatedCalled)
	}
	if resp == nil || resp.VisibilityStatus != challengeentity.SubmissionWriteupVisibilityHidden || resp.IsRecommended {
		t.Fatalf("unexpected hide community resp: %+v", resp)
	}
}

func TestWriteupServiceRestoreCommunityPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := writeupCommandContextKey("writeup-restore-community")
	expectedCtxValue := "ctx-writeup-restore-community"
	getTeacherRecordCalled := false
	findRequesterCalled := false
	findSubmissionCalled := false
	upsertCalled := false
	findUpdatedCalled := false
	repo := &writeupCommandContextStub{
		getTeacherSubmissionWriteupByIDWithContextFn: func(ctx context.Context, id int64) (*challengeports.TeacherSubmissionWriteupRecord, error) {
			getTeacherRecordCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected get-teacher-record ctx value %v, got %v", expectedCtxValue, got)
			}
			return &challengeports.TeacherSubmissionWriteupRecord{
				Submission:      challengeentity.SubmissionWriteup{ID: id, UserID: 88, ChallengeID: 11, VisibilityStatus: challengeentity.SubmissionWriteupVisibilityHidden},
				StudentUsername: "student88",
				ClassName:       "Class A",
			}, nil
		},
		findUserByIDWithContextFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			findRequesterCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-user ctx value %v, got %v", expectedCtxValue, got)
			}
			return &identitycontracts.User{ID: userID, Role: identitycontracts.RoleTeacher, ClassName: "Class A"}, nil
		},
		findSubmissionWriteupByIDWithContextFn: func(ctx context.Context, id int64) (*challengeentity.SubmissionWriteup, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-submission ctx value %v, got %v", expectedCtxValue, got)
			}
			if !upsertCalled {
				findSubmissionCalled = true
				return &challengeentity.SubmissionWriteup{ID: id, UserID: 88, ChallengeID: 11, VisibilityStatus: challengeentity.SubmissionWriteupVisibilityHidden}, nil
			}
			findUpdatedCalled = true
			return &challengeentity.SubmissionWriteup{ID: id, UserID: 88, ChallengeID: 11, VisibilityStatus: challengeentity.SubmissionWriteupVisibilityVisible}, nil
		},
		upsertSubmissionWriteupWithContextFn: func(ctx context.Context, writeup *challengeentity.SubmissionWriteup) error {
			upsertCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected upsert-submission ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.RestoreCommunity(ctx, 31, 1001, identitycontracts.RoleTeacher)
	if err != nil {
		t.Fatalf("RestoreCommunity() error = %v", err)
	}
	if !getTeacherRecordCalled || !findRequesterCalled || !findSubmissionCalled || !upsertCalled || !findUpdatedCalled {
		t.Fatalf("expected repository calls, got record=%v requester=%v submission=%v upsert=%v updated=%v", getTeacherRecordCalled, findRequesterCalled, findSubmissionCalled, upsertCalled, findUpdatedCalled)
	}
	if resp == nil || resp.VisibilityStatus != challengeentity.SubmissionWriteupVisibilityVisible {
		t.Fatalf("unexpected restore community resp: %+v", resp)
	}
}
