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
	withinTransactionFn                    func(ctx context.Context, fn func(txRepo practiceports.PracticeCommandTxRepository) error) error
	findContestByIDFn                      func(ctx context.Context, contestID int64) (*model.Contest, error)
	findContestChallengeFn                 func(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error)
	findContestAWDServiceFn                func(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error)
	listContestAWDServicesFn               func(ctx context.Context, contestID int64) ([]*model.ContestAWDService, error)
	listContestAWDInstancesFn              func(ctx context.Context, contestID int64) ([]*model.Instance, error)
	findContestTeamFn                      func(ctx context.Context, contestID, teamID int64) (*model.Team, error)
	listContestTeamsFn                     func(ctx context.Context, contestID int64) ([]*model.Team, error)
	findContestRegistrationFn              func(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error)
	lockInstanceScopeFn                    func(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) error
	findScopedExistingInstanceFn           func(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*model.Instance, error)
	countScopedRunningInstancesFn          func(ctx context.Context, userID int64, scope practiceports.InstanceScope) (int, error)
	refreshInstanceExpiryFn                func(instanceID int64, expiresAt time.Time) error
	refreshInstanceExpiryWithContextFn     func(ctx context.Context, instanceID int64, expiresAt time.Time) error
	createInstanceFn                       func(ctx context.Context, instance *model.Instance) error
	reserveAvailablePortFn                 func(ctx context.Context, start, end int) (int, error)
	bindReservedPortFn                     func(ctx context.Context, port int, instanceID int64) error
	createSubmissionFn                     func(ctx context.Context, submission *model.Submission) error
	findCorrectSubmissionFn                func(ctx context.Context, userID, challengeID int64) (*model.Submission, error)
	listChallengeSubmissionsFn             func(ctx context.Context, userID, challengeID int64, limit int) ([]model.Submission, error)
	updateSubmissionFn                     func(ctx context.Context, submission *model.Submission) error
	findUserByIDFn                         func(ctx context.Context, userID int64) (*model.User, error)
	listTeacherManualReviewSubmissionsFn   func(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error)
	getTeacherManualReviewSubmissionByIDFn func(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error)
	isUniqueViolationFn                    func(err error) bool
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

func (s *stubPracticeRepository) ListContestAWDServices(ctx context.Context, contestID int64) ([]*model.ContestAWDService, error) {
	if s.listContestAWDServicesFn != nil {
		return s.listContestAWDServicesFn(ctx, contestID)
	}
	return []*model.ContestAWDService{}, nil
}

func (s *stubPracticeRepository) ListContestAWDInstances(ctx context.Context, contestID int64) ([]*model.Instance, error) {
	if s.listContestAWDInstancesFn != nil {
		return s.listContestAWDInstancesFn(ctx, contestID)
	}
	return []*model.Instance{}, nil
}

func (s *stubPracticeRepository) FindContestTeam(ctx context.Context, contestID, teamID int64) (*model.Team, error) {
	if s.findContestTeamFn != nil {
		return s.findContestTeamFn(ctx, contestID, teamID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) ListContestTeams(ctx context.Context, contestID int64) ([]*model.Team, error) {
	if s.listContestTeamsFn != nil {
		return s.listContestTeamsFn(ctx, contestID)
	}
	return []*model.Team{}, nil
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

func (s *stubPracticeRepository) CreateSubmission(ctx context.Context, submission *model.Submission) error {
	if s.createSubmissionFn != nil {
		return s.createSubmissionFn(ctx, submission)
	}
	return nil
}

func (s *stubPracticeRepository) FindCorrectSubmission(ctx context.Context, userID, challengeID int64) (*model.Submission, error) {
	if s.findCorrectSubmissionFn != nil {
		return s.findCorrectSubmissionFn(ctx, userID, challengeID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) ListChallengeSubmissions(ctx context.Context, userID, challengeID int64, limit int) ([]model.Submission, error) {
	if s.listChallengeSubmissionsFn != nil {
		return s.listChallengeSubmissionsFn(ctx, userID, challengeID, limit)
	}
	return nil, nil
}

func (s *stubPracticeRepository) UpdateSubmission(ctx context.Context, submission *model.Submission) error {
	if s.updateSubmissionFn != nil {
		return s.updateSubmissionFn(ctx, submission)
	}
	return nil
}

func (s *stubPracticeRepository) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	if s.findUserByIDFn != nil {
		return s.findUserByIDFn(ctx, userID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) ListTeacherManualReviewSubmissions(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
	if s.listTeacherManualReviewSubmissionsFn != nil {
		return s.listTeacherManualReviewSubmissionsFn(ctx, query)
	}
	return nil, 0, nil
}

func (s *stubPracticeRepository) GetTeacherManualReviewSubmissionByID(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
	if s.getTeacherManualReviewSubmissionByIDFn != nil {
		return s.getTeacherManualReviewSubmissionByIDFn(ctx, id)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) IsUniqueViolation(err error) bool {
	if s.isUniqueViolationFn != nil {
		return s.isUniqueViolationFn(err)
	}
	return false
}
