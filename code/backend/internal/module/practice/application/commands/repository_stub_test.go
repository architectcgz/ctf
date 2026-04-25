package commands

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type stubPracticeRepository struct {
	withinTransactionFn                               func(ctx context.Context, fn func(txRepo practiceports.PracticeCommandTxRepository) error) error
	findContestByIDFn                                 func(ctx context.Context, contestID int64) (*model.Contest, error)
	findContestChallengeFn                            func(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error)
	findContestAWDServiceFn                           func(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error)
	findContestRegistrationFn                         func(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error)
	lockInstanceScopeFn                               func(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) error
	findScopedExistingInstanceFn                      func(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*model.Instance, error)
	countScopedRunningInstancesFn                     func(ctx context.Context, userID int64, scope practiceports.InstanceScope) (int, error)
	refreshInstanceExpiryFn                           func(instanceID int64, expiresAt time.Time) error
	refreshInstanceExpiryWithContextFn                func(ctx context.Context, instanceID int64, expiresAt time.Time) error
	createInstanceFn                                  func(ctx context.Context, instance *model.Instance) error
	reserveAvailablePortFn                            func(ctx context.Context, start, end int) (int, error)
	bindReservedPortFn                                func(ctx context.Context, port int, instanceID int64) error
	createSubmissionFn                                func(submission *model.Submission) error
	createSubmissionWithContextFn                     func(ctx context.Context, submission *model.Submission) error
	findCorrectSubmissionFn                           func(userID, challengeID int64) (*model.Submission, error)
	findCorrectSubmissionWithContextFn                func(ctx context.Context, userID, challengeID int64) (*model.Submission, error)
	listChallengeSubmissionsFn                        func(userID, challengeID int64, limit int) ([]model.Submission, error)
	listChallengeSubmissionsWithContextFn             func(ctx context.Context, userID, challengeID int64, limit int) ([]model.Submission, error)
	updateSubmissionFn                                func(submission *model.Submission) error
	updateSubmissionWithContextFn                     func(ctx context.Context, submission *model.Submission) error
	findUserByIDFn                                    func(userID int64) (*model.User, error)
	findUserByIDWithContextFn                         func(ctx context.Context, userID int64) (*model.User, error)
	listTeacherManualReviewSubmissionsFn              func(query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error)
	listTeacherManualReviewSubmissionsWithContextFn   func(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error)
	getTeacherManualReviewSubmissionByIDFn            func(id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error)
	getTeacherManualReviewSubmissionByIDWithContextFn func(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error)
	isUniqueViolationFn                               func(err error) bool
}

func (s *stubPracticeRepository) WithinTransaction(ctx context.Context, fn func(txRepo practiceports.PracticeCommandTxRepository) error) error {
	if s.withinTransactionFn != nil {
		return s.withinTransactionFn(ctx, fn)
	}
	return fn(s)
}

func (s *stubPracticeRepository) FindContestByID(ctx context.Context, contestID int64) (*model.Contest, error) {
	if s.findContestByIDFn != nil {
		return s.findContestByIDFn(ctx, contestID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) FindContestChallenge(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error) {
	if s.findContestChallengeFn != nil {
		return s.findContestChallengeFn(ctx, contestID, challengeID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) FindContestAWDService(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error) {
	if s.findContestAWDServiceFn != nil {
		return s.findContestAWDServiceFn(ctx, contestID, serviceID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) FindContestRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	if s.findContestRegistrationFn != nil {
		return s.findContestRegistrationFn(ctx, contestID, userID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) LockInstanceScope(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) error {
	if s.lockInstanceScopeFn != nil {
		return s.lockInstanceScopeFn(ctx, userID, challengeID, scope)
	}
	return nil
}

func (s *stubPracticeRepository) FindScopedExistingInstance(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*model.Instance, error) {
	if s.findScopedExistingInstanceFn != nil {
		return s.findScopedExistingInstanceFn(ctx, userID, challengeID, scope)
	}
	return nil, nil
}

func (s *stubPracticeRepository) CountScopedRunningInstances(ctx context.Context, userID int64, scope practiceports.InstanceScope) (int, error) {
	if s.countScopedRunningInstancesFn != nil {
		return s.countScopedRunningInstancesFn(ctx, userID, scope)
	}
	return 0, nil
}

func (s *stubPracticeRepository) RefreshInstanceExpiry(ctx context.Context, instanceID int64, expiresAt time.Time) error {
	if s.refreshInstanceExpiryWithContextFn != nil {
		return s.refreshInstanceExpiryWithContextFn(ctx, instanceID, expiresAt)
	}
	return nil
}

func (s *stubPracticeRepository) CreateInstance(ctx context.Context, instance *model.Instance) error {
	if s.createInstanceFn != nil {
		return s.createInstanceFn(ctx, instance)
	}
	return nil
}

func (s *stubPracticeRepository) ReserveAvailablePort(ctx context.Context, start, end int) (int, error) {
	if s.reserveAvailablePortFn != nil {
		return s.reserveAvailablePortFn(ctx, start, end)
	}
	return start, nil
}

func (s *stubPracticeRepository) BindReservedPort(ctx context.Context, port int, instanceID int64) error {
	if s.bindReservedPortFn != nil {
		return s.bindReservedPortFn(ctx, port, instanceID)
	}
	return nil
}

func (s *stubPracticeRepository) CreateSubmission(submission *model.Submission) error {
	if s.createSubmissionFn != nil {
		return s.createSubmissionFn(submission)
	}
	return nil
}

func (s *stubPracticeRepository) CreateSubmissionWithContext(ctx context.Context, submission *model.Submission) error {
	if s.createSubmissionWithContextFn != nil {
		return s.createSubmissionWithContextFn(ctx, submission)
	}
	return s.CreateSubmission(submission)
}

func (s *stubPracticeRepository) FindCorrectSubmission(userID, challengeID int64) (*model.Submission, error) {
	if s.findCorrectSubmissionFn != nil {
		return s.findCorrectSubmissionFn(userID, challengeID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) FindCorrectSubmissionWithContext(ctx context.Context, userID, challengeID int64) (*model.Submission, error) {
	if s.findCorrectSubmissionWithContextFn != nil {
		return s.findCorrectSubmissionWithContextFn(ctx, userID, challengeID)
	}
	return s.FindCorrectSubmission(userID, challengeID)
}

func (s *stubPracticeRepository) ListChallengeSubmissions(userID, challengeID int64, limit int) ([]model.Submission, error) {
	if s.listChallengeSubmissionsFn != nil {
		return s.listChallengeSubmissionsFn(userID, challengeID, limit)
	}
	return nil, nil
}

func (s *stubPracticeRepository) ListChallengeSubmissionsWithContext(ctx context.Context, userID, challengeID int64, limit int) ([]model.Submission, error) {
	if s.listChallengeSubmissionsWithContextFn != nil {
		return s.listChallengeSubmissionsWithContextFn(ctx, userID, challengeID, limit)
	}
	return s.ListChallengeSubmissions(userID, challengeID, limit)
}

func (s *stubPracticeRepository) UpdateSubmission(submission *model.Submission) error {
	if s.updateSubmissionFn != nil {
		return s.updateSubmissionFn(submission)
	}
	return nil
}

func (s *stubPracticeRepository) UpdateSubmissionWithContext(ctx context.Context, submission *model.Submission) error {
	if s.updateSubmissionWithContextFn != nil {
		return s.updateSubmissionWithContextFn(ctx, submission)
	}
	return s.UpdateSubmission(submission)
}

func (s *stubPracticeRepository) FindUserByID(userID int64) (*model.User, error) {
	if s.findUserByIDFn != nil {
		return s.findUserByIDFn(userID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) FindUserByIDWithContext(ctx context.Context, userID int64) (*model.User, error) {
	if s.findUserByIDWithContextFn != nil {
		return s.findUserByIDWithContextFn(ctx, userID)
	}
	return s.FindUserByID(userID)
}

func (s *stubPracticeRepository) ListTeacherManualReviewSubmissions(query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
	if s.listTeacherManualReviewSubmissionsFn != nil {
		return s.listTeacherManualReviewSubmissionsFn(query)
	}
	return nil, 0, nil
}

func (s *stubPracticeRepository) ListTeacherManualReviewSubmissionsWithContext(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
	if s.listTeacherManualReviewSubmissionsWithContextFn != nil {
		return s.listTeacherManualReviewSubmissionsWithContextFn(ctx, query)
	}
	return s.ListTeacherManualReviewSubmissions(query)
}

func (s *stubPracticeRepository) GetTeacherManualReviewSubmissionByID(id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
	if s.getTeacherManualReviewSubmissionByIDFn != nil {
		return s.getTeacherManualReviewSubmissionByIDFn(id)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) GetTeacherManualReviewSubmissionByIDWithContext(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
	if s.getTeacherManualReviewSubmissionByIDWithContextFn != nil {
		return s.getTeacherManualReviewSubmissionByIDWithContextFn(ctx, id)
	}
	return s.GetTeacherManualReviewSubmissionByID(id)
}

func (s *stubPracticeRepository) IsUniqueViolation(err error) bool {
	if s.isUniqueViolationFn != nil {
		return s.isUniqueViolationFn(err)
	}
	return false
}
