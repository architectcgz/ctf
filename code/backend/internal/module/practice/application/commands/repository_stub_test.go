package commands

import (
	"context"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type stubPracticeRepository struct {
	withinTransactionFn                  func(ctx context.Context, fn func(txRepo practiceports.PracticeRepository) error) error
	findContestByIDWithContextFn         func(ctx context.Context, contestID int64) (*model.Contest, error)
	findContestChallengeWithContextFn    func(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error)
	findContestRegistrationWithContextFn func(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error)
	lockInstanceScopeFn                  func(userID int64, scope practiceports.InstanceScope) error
	findScopedExistingInstanceFn         func(userID, challengeID int64, scope practiceports.InstanceScope) (*model.Instance, error)
	countScopedRunningInstancesFn        func(userID int64, scope practiceports.InstanceScope) (int, error)
	createInstanceFn                     func(instance *model.Instance) error
	reserveAvailablePortFn               func(start, end int) (int, error)
	bindReservedPortFn                   func(port int, instanceID int64) error
	createSubmissionFn                   func(submission *model.Submission) error
	findCorrectSubmissionFn              func(userID, challengeID int64) (*model.Submission, error)
	isUniqueViolationFn                  func(err error) bool
	findChallengeScoreWithContextFn      func(ctx context.Context, challengeID int64) (*model.Challenge, error)
	findChallengesScoresWithContextFn    func(ctx context.Context, challengeIDs []int64) ([]model.Challenge, error)
	listSolvedChallengeIDsWithContextFn  func(ctx context.Context, userID int64) ([]int64, error)
	upsertUserScoreWithContextFn         func(ctx context.Context, userScore *model.UserScore) error
	findUserScoreWithContextFn           func(ctx context.Context, userID int64) (*model.UserScore, error)
	listTopUserScoresWithContextFn       func(ctx context.Context, limit int) ([]model.UserScore, error)
	findUsersByIDsWithContextFn          func(ctx context.Context, userIDs []int64) ([]model.User, error)
	countRecentSubmissionsFn             func(userID, challengeID int64, since time.Time) (int64, error)
}

func (s *stubPracticeRepository) WithinTransaction(ctx context.Context, fn func(txRepo practiceports.PracticeRepository) error) error {
	if s.withinTransactionFn != nil {
		return s.withinTransactionFn(ctx, fn)
	}
	return fn(s)
}

func (s *stubPracticeRepository) FindContestByIDWithContext(ctx context.Context, contestID int64) (*model.Contest, error) {
	if s.findContestByIDWithContextFn != nil {
		return s.findContestByIDWithContextFn(ctx, contestID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) FindContestChallengeWithContext(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error) {
	if s.findContestChallengeWithContextFn != nil {
		return s.findContestChallengeWithContextFn(ctx, contestID, challengeID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) FindContestRegistrationWithContext(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	if s.findContestRegistrationWithContextFn != nil {
		return s.findContestRegistrationWithContextFn(ctx, contestID, userID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) LockInstanceScope(userID int64, scope practiceports.InstanceScope) error {
	if s.lockInstanceScopeFn != nil {
		return s.lockInstanceScopeFn(userID, scope)
	}
	return nil
}

func (s *stubPracticeRepository) FindScopedExistingInstance(userID, challengeID int64, scope practiceports.InstanceScope) (*model.Instance, error) {
	if s.findScopedExistingInstanceFn != nil {
		return s.findScopedExistingInstanceFn(userID, challengeID, scope)
	}
	return nil, nil
}

func (s *stubPracticeRepository) CountScopedRunningInstances(userID int64, scope practiceports.InstanceScope) (int, error) {
	if s.countScopedRunningInstancesFn != nil {
		return s.countScopedRunningInstancesFn(userID, scope)
	}
	return 0, nil
}

func (s *stubPracticeRepository) CreateInstance(instance *model.Instance) error {
	if s.createInstanceFn != nil {
		return s.createInstanceFn(instance)
	}
	return nil
}

func (s *stubPracticeRepository) ReserveAvailablePort(start, end int) (int, error) {
	if s.reserveAvailablePortFn != nil {
		return s.reserveAvailablePortFn(start, end)
	}
	return start, nil
}

func (s *stubPracticeRepository) BindReservedPort(port int, instanceID int64) error {
	if s.bindReservedPortFn != nil {
		return s.bindReservedPortFn(port, instanceID)
	}
	return nil
}

func (s *stubPracticeRepository) CreateSubmission(submission *model.Submission) error {
	if s.createSubmissionFn != nil {
		return s.createSubmissionFn(submission)
	}
	return nil
}

func (s *stubPracticeRepository) FindCorrectSubmission(userID, challengeID int64) (*model.Submission, error) {
	if s.findCorrectSubmissionFn != nil {
		return s.findCorrectSubmissionFn(userID, challengeID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) IsUniqueViolation(err error) bool {
	if s.isUniqueViolationFn != nil {
		return s.isUniqueViolationFn(err)
	}
	return false
}

func (s *stubPracticeRepository) FindChallengeScoreWithContext(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	if s.findChallengeScoreWithContextFn != nil {
		return s.findChallengeScoreWithContextFn(ctx, challengeID)
	}
	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) FindChallengesScoresWithContext(ctx context.Context, challengeIDs []int64) ([]model.Challenge, error) {
	if s.findChallengesScoresWithContextFn != nil {
		return s.findChallengesScoresWithContextFn(ctx, challengeIDs)
	}
	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return nil, nil
}

func (s *stubPracticeRepository) ListSolvedChallengeIDsWithContext(ctx context.Context, userID int64) ([]int64, error) {
	if s.listSolvedChallengeIDsWithContextFn != nil {
		return s.listSolvedChallengeIDsWithContextFn(ctx, userID)
	}
	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return nil, nil
}

func (s *stubPracticeRepository) UpsertUserScoreWithContext(ctx context.Context, userScore *model.UserScore) error {
	if s.upsertUserScoreWithContextFn != nil {
		return s.upsertUserScoreWithContextFn(ctx, userScore)
	}
	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}

func (s *stubPracticeRepository) FindUserScoreWithContext(ctx context.Context, userID int64) (*model.UserScore, error) {
	if s.findUserScoreWithContextFn != nil {
		return s.findUserScoreWithContextFn(ctx, userID)
	}
	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) ListTopUserScoresWithContext(ctx context.Context, limit int) ([]model.UserScore, error) {
	if s.listTopUserScoresWithContextFn != nil {
		return s.listTopUserScoresWithContextFn(ctx, limit)
	}
	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return nil, nil
}

func (s *stubPracticeRepository) FindUsersByIDsWithContext(ctx context.Context, userIDs []int64) ([]model.User, error) {
	if s.findUsersByIDsWithContextFn != nil {
		return s.findUsersByIDsWithContextFn(ctx, userIDs)
	}
	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return nil, nil
}

func (s *stubPracticeRepository) CountRecentSubmissions(userID, challengeID int64, since time.Time) (int64, error) {
	if s.countRecentSubmissionsFn != nil {
		return s.countRecentSubmissionsFn(userID, challengeID, since)
	}
	return 0, nil
}
