package commands

import (
	"context"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type stubPracticeRepository struct {
	withinTransactionFn                  func(ctx context.Context, fn func(txRepo practiceports.PracticeCommandTxRepository) error) error
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
}

func (s *stubPracticeRepository) WithinTransaction(ctx context.Context, fn func(txRepo practiceports.PracticeCommandTxRepository) error) error {
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
