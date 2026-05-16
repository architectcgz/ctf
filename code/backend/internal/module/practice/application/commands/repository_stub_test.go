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
	withinInstanceStartTxFn                func(ctx context.Context, fn func(txRepo practiceports.PracticeInstanceStartTxRepository) error) error
	withinInstanceRestartTxFn              func(ctx context.Context, fn func(txRepo practiceports.PracticeInstanceRestartTxRepository) error) error
	withinAWDServiceOperationTxFn          func(ctx context.Context, fn func(txRepo practiceports.PracticeAWDServiceOperationTxRepository) error) error
	findContestByIDFn                      func(ctx context.Context, contestID int64) (*model.Contest, error)
	listDesiredRuntimeAWDContestsFn        func(ctx context.Context) ([]*model.Contest, error)
	findContestChallengeFn                 func(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error)
	findContestAWDServiceFn                func(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error)
	listContestAWDServicesFn               func(ctx context.Context, contestID int64) ([]*model.ContestAWDService, error)
	listContestAWDInstancesFn              func(ctx context.Context, contestID int64) ([]*model.Instance, error)
	findContestTeamFn                      func(ctx context.Context, contestID, teamID int64) (*model.Team, error)
	listContestTeamsFn                     func(ctx context.Context, contestID int64) ([]*model.Team, error)
	findContestRegistrationFn              func(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error)
	listContestAWDScopeControlsFn          func(ctx context.Context, contestID int64) ([]*model.AWDScopeControl, error)
	listScopeAWDScopeControlsFn            func(ctx context.Context, contestID, teamID, serviceID int64) ([]*model.AWDScopeControl, error)
	upsertAWDScopeControlFn                func(ctx context.Context, control *model.AWDScopeControl) error
	deleteAWDScopeControlFn                func(ctx context.Context, contestID, teamID int64, scopeType, controlType string, serviceID int64) error
	lockInstanceScopeFn                    func(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) error
	findScopedExistingInstanceFn           func(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*model.Instance, error)
	findScopedRestartableInstanceFn        func(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*model.Instance, error)
	countScopedRunningInstancesFn          func(ctx context.Context, userID int64, scope practiceports.InstanceScope) (int, error)
	refreshInstanceExpiryFn                func(instanceID int64, expiresAt time.Time) error
	refreshInstanceExpiryWithContextFn     func(ctx context.Context, instanceID int64, expiresAt time.Time) error
	resetInstanceRuntimeForRestartFn       func(ctx context.Context, instanceID int64, status string, expiresAt time.Time, preserveHostPort bool) error
	isHostPortReusableForRestartFn         func(ctx context.Context, instanceID int64, hostPort int) (bool, error)
	createInstanceFn                       func(ctx context.Context, instance *model.Instance) error
	createAWDServiceOperationFn            func(ctx context.Context, operation *model.AWDServiceOperation) error
	finishAWDServiceOperationFn            func(ctx context.Context, operationID int64, status, errorMessage string, finishedAt time.Time) error
	reserveAvailablePortFn                 func(ctx context.Context, start, end int) (int, error)
	reserveAvailablePortExcludingFn        func(ctx context.Context, start, end, excludedPort int) (int, error)
	bindReservedPortFn                     func(ctx context.Context, port int, instanceID int64) error
	releaseReservedPortFn                  func(ctx context.Context, port int) error
	releasePortForInstanceFn               func(ctx context.Context, port int, instanceID int64) error
	createSubmissionFn                     func(ctx context.Context, submission *model.Submission) error
	findCorrectSubmissionFn                func(ctx context.Context, userID, challengeID int64) (*model.Submission, error)
	listChallengeSubmissionsFn             func(ctx context.Context, userID, challengeID int64, limit int) ([]model.Submission, error)
	updateSubmissionFn                     func(ctx context.Context, submission *model.Submission) error
	findUserByIDFn                         func(ctx context.Context, userID int64) (*model.User, error)
	listTeacherManualReviewSubmissionsFn   func(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error)
	getTeacherManualReviewSubmissionByIDFn func(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error)
	isUniqueViolationFn                    func(err error) bool
}

func (s *stubPracticeRepository) WithinInstanceStartTx(ctx context.Context, fn func(txRepo practiceports.PracticeInstanceStartTxRepository) error) error {
	if s.withinInstanceStartTxFn != nil {
		return s.withinInstanceStartTxFn(ctx, fn)
	}
	return fn(s)
}

func (s *stubPracticeRepository) WithinInstanceRestartTx(ctx context.Context, fn func(txRepo practiceports.PracticeInstanceRestartTxRepository) error) error {
	if s.withinInstanceRestartTxFn != nil {
		return s.withinInstanceRestartTxFn(ctx, fn)
	}
	return fn(s)
}

func (s *stubPracticeRepository) WithinAWDServiceOperationTx(ctx context.Context, fn func(txRepo practiceports.PracticeAWDServiceOperationTxRepository) error) error {
	if s.withinAWDServiceOperationTxFn != nil {
		return s.withinAWDServiceOperationTxFn(ctx, fn)
	}
	return fn(s)
}

func (s *stubPracticeRepository) FindContestByID(ctx context.Context, contestID int64) (*model.Contest, error) {
	if s.findContestByIDFn != nil {
		return s.findContestByIDFn(ctx, contestID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) ListDesiredRuntimeAWDContests(ctx context.Context) ([]*model.Contest, error) {
	if s.listDesiredRuntimeAWDContestsFn != nil {
		return s.listDesiredRuntimeAWDContestsFn(ctx)
	}
	return []*model.Contest{}, nil
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

func (s *stubPracticeRepository) ListContestAWDScopeControls(ctx context.Context, contestID int64) ([]*model.AWDScopeControl, error) {
	if s.listContestAWDScopeControlsFn != nil {
		return s.listContestAWDScopeControlsFn(ctx, contestID)
	}
	return nil, nil
}

func (s *stubPracticeRepository) ListScopeAWDScopeControls(ctx context.Context, contestID, teamID, serviceID int64) ([]*model.AWDScopeControl, error) {
	if s.listScopeAWDScopeControlsFn != nil {
		return s.listScopeAWDScopeControlsFn(ctx, contestID, teamID, serviceID)
	}
	return nil, nil
}

func (s *stubPracticeRepository) UpsertAWDScopeControl(ctx context.Context, control *model.AWDScopeControl) error {
	if s.upsertAWDScopeControlFn != nil {
		return s.upsertAWDScopeControlFn(ctx, control)
	}
	return nil
}

func (s *stubPracticeRepository) DeleteAWDScopeControl(ctx context.Context, contestID, teamID int64, scopeType, controlType string, serviceID int64) error {
	if s.deleteAWDScopeControlFn != nil {
		return s.deleteAWDScopeControlFn(ctx, contestID, teamID, scopeType, controlType, serviceID)
	}
	return nil
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

func (s *stubPracticeRepository) FindScopedRestartableInstance(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*model.Instance, error) {
	if s.findScopedRestartableInstanceFn != nil {
		return s.findScopedRestartableInstanceFn(ctx, userID, challengeID, scope)
	}
	return s.FindScopedExistingInstance(ctx, userID, challengeID, scope)
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

func (s *stubPracticeRepository) ResetInstanceRuntimeForRestart(ctx context.Context, instanceID int64, status string, expiresAt time.Time, preserveHostPort bool) error {
	if s.resetInstanceRuntimeForRestartFn != nil {
		return s.resetInstanceRuntimeForRestartFn(ctx, instanceID, status, expiresAt, preserveHostPort)
	}
	return nil
}

func (s *stubPracticeRepository) IsHostPortReusableForRestart(ctx context.Context, instanceID int64, hostPort int) (bool, error) {
	if s.isHostPortReusableForRestartFn != nil {
		return s.isHostPortReusableForRestartFn(ctx, instanceID, hostPort)
	}
	return true, nil
}

func (s *stubPracticeRepository) CreateInstance(ctx context.Context, instance *model.Instance) error {
	if s.createInstanceFn != nil {
		return s.createInstanceFn(ctx, instance)
	}
	return nil
}

func (s *stubPracticeRepository) CreateAWDServiceOperation(ctx context.Context, operation *model.AWDServiceOperation) error {
	if s.createAWDServiceOperationFn != nil {
		return s.createAWDServiceOperationFn(ctx, operation)
	}
	return nil
}

func (s *stubPracticeRepository) FinishAWDServiceOperation(ctx context.Context, operationID int64, status, errorMessage string, finishedAt time.Time) error {
	if s.finishAWDServiceOperationFn != nil {
		return s.finishAWDServiceOperationFn(ctx, operationID, status, errorMessage, finishedAt)
	}
	return nil
}

func (s *stubPracticeRepository) ReserveAvailablePort(ctx context.Context, start, end int) (int, error) {
	if s.reserveAvailablePortFn != nil {
		return s.reserveAvailablePortFn(ctx, start, end)
	}
	return start, nil
}

func (s *stubPracticeRepository) ReserveAvailablePortExcluding(ctx context.Context, start, end, excludedPort int) (int, error) {
	if s.reserveAvailablePortExcludingFn != nil {
		return s.reserveAvailablePortExcludingFn(ctx, start, end, excludedPort)
	}
	return s.ReserveAvailablePort(ctx, start, end)
}

func (s *stubPracticeRepository) BindReservedPort(ctx context.Context, port int, instanceID int64) error {
	if s.bindReservedPortFn != nil {
		return s.bindReservedPortFn(ctx, port, instanceID)
	}
	return nil
}

func (s *stubPracticeRepository) ReleaseReservedPort(ctx context.Context, port int) error {
	if s.releaseReservedPortFn != nil {
		return s.releaseReservedPortFn(ctx, port)
	}
	return nil
}

func (s *stubPracticeRepository) ReleasePortForInstance(ctx context.Context, port int, instanceID int64) error {
	if s.releasePortForInstanceFn != nil {
		return s.releasePortForInstanceFn(ctx, port, instanceID)
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
