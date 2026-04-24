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
	findSubmissionWriteupByIDWithContextFn             func(ctx context.Context, id int64) (*model.SubmissionWriteup, error)
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

func (s *stubChallengeWriteupRepository) FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *stubChallengeWriteupRepository) FindUserByIDWithContext(ctx context.Context, userID int64) (*model.User, error) {
	if s.findUserByIDWithContextFn != nil {
		return s.findUserByIDWithContextFn(ctx, userID)
	}
	return nil, nil
}

func (s *stubChallengeWriteupRepository) FindWriteupByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error) {
	if s.findWriteupByChallengeIDWithContextFn != nil {
		return s.findWriteupByChallengeIDWithContextFn(ctx, challengeID)
	}
	return nil, nil
}

func (s *stubChallengeWriteupRepository) UpsertWriteup(ctx context.Context, writeup *model.ChallengeWriteup) error {
	if s.upsertWriteupWithContextFn != nil {
		return s.upsertWriteupWithContextFn(ctx, writeup)
	}
	return nil
}

func (s *stubChallengeWriteupRepository) DeleteWriteupByChallengeID(ctx context.Context, challengeID int64) error {
	if s.deleteWriteupByChallengeIDWithContextFn != nil {
		return s.deleteWriteupByChallengeIDWithContextFn(ctx, challengeID)
	}
	return nil
}

func (s *stubChallengeWriteupRepository) FindReleasedWriteupByChallengeID(ctx context.Context, challengeID int64, now time.Time) (*model.ChallengeWriteup, error) {
	if s.findReleasedWriteupByChallengeIDWithContextFn != nil {
		return s.findReleasedWriteupByChallengeIDWithContextFn(ctx, challengeID, now)
	}
	return nil, nil
}

func (s *stubChallengeWriteupRepository) GetSolvedStatusWithContext(ctx context.Context, userID, challengeID int64) (bool, error) {
	if s.getSolvedStatusWithContextFn != nil {
		return s.getSolvedStatusWithContextFn(ctx, userID, challengeID)
	}
	return false, nil
}

func (s *stubChallengeWriteupRepository) FindSubmissionWriteupByUserChallenge(ctx context.Context, userID, challengeID int64) (*model.SubmissionWriteup, error) {
	if s.findSubmissionWriteupByUserChallengeWithContextFn != nil {
		return s.findSubmissionWriteupByUserChallengeWithContextFn(ctx, userID, challengeID)
	}
	return nil, nil
}

func (s *stubChallengeWriteupRepository) FindSubmissionWriteupByID(ctx context.Context, id int64) (*model.SubmissionWriteup, error) {
	if s.findSubmissionWriteupByIDWithContextFn != nil {
		return s.findSubmissionWriteupByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *stubChallengeWriteupRepository) UpsertSubmissionWriteup(ctx context.Context, writeup *model.SubmissionWriteup) error {
	if s.upsertSubmissionWriteupWithContextFn != nil {
		return s.upsertSubmissionWriteupWithContextFn(ctx, writeup)
	}
	return nil
}

func (s *stubChallengeWriteupRepository) GetTeacherSubmissionWriteupByID(ctx context.Context, id int64) (*challengeports.TeacherSubmissionWriteupRecord, error) {
	if s.getTeacherSubmissionWriteupByIDWithContextFn != nil {
		return s.getTeacherSubmissionWriteupByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *stubChallengeWriteupRepository) ListTeacherSubmissionWriteups(ctx context.Context, query *dto.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error) {
	if s.listTeacherSubmissionWriteupsWithContextFn != nil {
		return s.listTeacherSubmissionWriteupsWithContextFn(ctx, query)
	}
	return nil, 0, nil
}

func (s *stubChallengeWriteupRepository) ListRecommendedSolutionsByChallengeID(ctx context.Context, challengeID int64, now time.Time) ([]challengeports.RecommendedSolutionRecord, error) {
	if s.listRecommendedSolutionsByChallengeIDWithContextFn != nil {
		return s.listRecommendedSolutionsByChallengeIDWithContextFn(ctx, challengeID, now)
	}
	return nil, nil
}

func (s *stubChallengeWriteupRepository) ListCommunitySolutionsByChallengeID(ctx context.Context, challengeID int64, query *dto.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error) {
	if s.listCommunitySolutionsByChallengeIDWithContextFn != nil {
		return s.listCommunitySolutionsByChallengeIDWithContextFn(ctx, challengeID, query)
	}
	return nil, 0, nil
}

type challengeWriteupContextKey string

func TestWriteupServiceGetAdminPropagatesContextToRepository(t *testing.T) {
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
	resp, err := service.GetAdmin(ctx, 11)
	if err != nil {
		t.Fatalf("GetAdmin() error = %v", err)
	}
	if !findChallengeCalled || !findWriteupCalled {
		t.Fatalf("expected repository calls, got challenge=%v writeup=%v", findChallengeCalled, findWriteupCalled)
	}
	if resp == nil || resp.ID != 101 || resp.ChallengeID != 11 || resp.Title != "Official" {
		t.Fatalf("unexpected writeup resp: %+v", resp)
	}
}

func TestWriteupServiceGetMySubmissionPropagatesContextToRepository(t *testing.T) {
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
	resp, err := service.GetMySubmission(ctx, 7, 11)
	if err != nil {
		t.Fatalf("GetMySubmission() error = %v", err)
	}
	if !findChallengeCalled || !findSubmissionCalled {
		t.Fatalf("expected repository calls, got challenge=%v submission=%v", findChallengeCalled, findSubmissionCalled)
	}
	if resp == nil || resp.ID != 201 || resp.UserID != 7 || resp.ChallengeID != 11 {
		t.Fatalf("unexpected submission resp: %+v", resp)
	}
}

func TestWriteupServiceGetPublishedPropagatesContextToRepository(t *testing.T) {
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
	resp, err := service.GetPublished(ctx, 7, 11)
	if err != nil {
		t.Fatalf("GetPublished() error = %v", err)
	}
	if !findChallengeCalled || !findReleasedCalled || !getSolvedCalled {
		t.Fatalf("expected repository calls, got challenge=%v released=%v solved=%v", findChallengeCalled, findReleasedCalled, getSolvedCalled)
	}
	if resp == nil || resp.ID != 301 || resp.ChallengeID != 11 || resp.Title != "Published" || resp.RequiresSpoilerWarning {
		t.Fatalf("unexpected published resp: %+v", resp)
	}
}

func TestWriteupServiceListTeacherSubmissionsPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := challengeWriteupContextKey("writeup-list-teacher-submissions")
	expectedCtxValue := "ctx-writeup-list-teacher-submissions"
	listCalled := false
	repo := &stubChallengeWriteupRepository{
		findUserByIDWithContextFn: func(ctx context.Context, userID int64) (*model.User, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-user ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.User{ID: userID, Role: model.RoleTeacher, ClassName: "Class A"}, nil
		},
		listTeacherSubmissionWriteupsWithContextFn: func(ctx context.Context, query *dto.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error) {
			listCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected list-teacher-submissions ctx value %v, got %v", expectedCtxValue, got)
			}
			if query.ClassName != "Class A" {
				t.Fatalf("expected normalized class name, got %+v", query)
			}
			return []challengeports.TeacherSubmissionWriteupRecord{}, 0, nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.ListTeacherSubmissions(ctx, 1001, model.RoleTeacher, &dto.TeacherSubmissionWriteupQuery{}); err != nil {
		t.Fatalf("ListTeacherSubmissions() error = %v", err)
	}
	if !listCalled {
		t.Fatal("expected list teacher submissions repository to be called")
	}
}

func TestWriteupServiceGetTeacherSubmissionPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := challengeWriteupContextKey("writeup-get-teacher-submission")
	expectedCtxValue := "ctx-writeup-get-teacher-submission"
	now := time.Now()
	getCalled := false
	findRequesterCalled := false
	repo := &stubChallengeWriteupRepository{
		getTeacherSubmissionWriteupByIDWithContextFn: func(ctx context.Context, id int64) (*challengeports.TeacherSubmissionWriteupRecord, error) {
			getCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected get-teacher-submission ctx value %v, got %v", expectedCtxValue, got)
			}
			return &challengeports.TeacherSubmissionWriteupRecord{
				Submission:      model.SubmissionWriteup{ID: id, UserID: 88, ChallengeID: 11, SubmissionStatus: model.SubmissionWriteupStatusDraft, VisibilityStatus: model.SubmissionWriteupVisibilityHidden, CreatedAt: now, UpdatedAt: now},
				StudentUsername: "student88",
				StudentName:     "Student 88",
				ClassName:       "Class A",
				ChallengeTitle:  "writeup challenge",
			}, nil
		},
		findUserByIDWithContextFn: func(ctx context.Context, userID int64) (*model.User, error) {
			findRequesterCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-user ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.User{ID: userID, Role: model.RoleTeacher, ClassName: "Class A"}, nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.GetTeacherSubmission(ctx, 91, 1001, model.RoleTeacher); err != nil {
		t.Fatalf("GetTeacherSubmission() error = %v", err)
	}
	if !getCalled {
		t.Fatal("expected get teacher submission repository to be called")
	}
	if !findRequesterCalled {
		t.Fatal("expected requester repository to be called")
	}
}

func TestWriteupServiceListRecommendedSolutionsPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := challengeWriteupContextKey("writeup-list-recommended")
	expectedCtxValue := "ctx-writeup-list-recommended"
	findChallengeCalled := false
	getSolvedCalled := false
	listRecommendedCalled := false
	repo := &stubChallengeWriteupRepository{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findChallengeCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id, Status: model.ChallengeStatusPublished}, nil
		},
		getSolvedStatusWithContextFn: func(ctx context.Context, userID, challengeID int64) (bool, error) {
			getSolvedCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected get-solved ctx value %v, got %v", expectedCtxValue, got)
			}
			return true, nil
		},
		listRecommendedSolutionsByChallengeIDWithContextFn: func(ctx context.Context, challengeID int64, now time.Time) ([]challengeports.RecommendedSolutionRecord, error) {
			listRecommendedCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected list-recommended ctx value %v, got %v", expectedCtxValue, got)
			}
			return []challengeports.RecommendedSolutionRecord{{SourceType: "official", SourceID: 1, ChallengeID: challengeID, Title: "Recommended", Content: "walkthrough", AuthorName: "teacher"}}, nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.ListRecommendedSolutions(ctx, 7, 11)
	if err != nil {
		t.Fatalf("ListRecommendedSolutions() error = %v", err)
	}
	if !findChallengeCalled || !getSolvedCalled || !listRecommendedCalled {
		t.Fatalf("expected repository calls, got challenge=%v solved=%v recommended=%v", findChallengeCalled, getSolvedCalled, listRecommendedCalled)
	}
	if resp == nil || resp.Total != 1 || resp.Size != 1 {
		t.Fatalf("unexpected recommended resp: %+v", resp)
	}
}

func TestWriteupServiceListCommunitySolutionsPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := challengeWriteupContextKey("writeup-list-community")
	expectedCtxValue := "ctx-writeup-list-community"
	findChallengeCalled := false
	getSolvedCalled := false
	listCommunityCalled := false
	repo := &stubChallengeWriteupRepository{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findChallengeCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id, Status: model.ChallengeStatusPublished}, nil
		},
		getSolvedStatusWithContextFn: func(ctx context.Context, userID, challengeID int64) (bool, error) {
			getSolvedCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected get-solved ctx value %v, got %v", expectedCtxValue, got)
			}
			return true, nil
		},
		listCommunitySolutionsByChallengeIDWithContextFn: func(ctx context.Context, challengeID int64, query *dto.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error) {
			listCommunityCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected list-community ctx value %v, got %v", expectedCtxValue, got)
			}
			if query.Page != 1 || query.Size != 20 {
				t.Fatalf("expected normalized query, got %+v", query)
			}
			return []challengeports.CommunitySolutionRecord{{Submission: model.SubmissionWriteup{ID: 5, UserID: 9, ChallengeID: challengeID, Title: "Community", Content: "notes"}, AuthorName: "student", ChallengeID: challengeID, ChallengeTitle: "challenge"}}, 1, nil
		},
	}
	service := NewWriteupService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.ListCommunitySolutions(ctx, 7, 11, &dto.CommunityChallengeSolutionQuery{})
	if err != nil {
		t.Fatalf("ListCommunitySolutions() error = %v", err)
	}
	if !findChallengeCalled || !getSolvedCalled || !listCommunityCalled {
		t.Fatalf("expected repository calls, got challenge=%v solved=%v community=%v", findChallengeCalled, getSolvedCalled, listCommunityCalled)
	}
	if resp == nil || resp.Total != 1 || resp.Page != 1 || resp.Size != 20 {
		t.Fatalf("unexpected community resp: %+v", resp)
	}
}
