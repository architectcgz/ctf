package ports

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type InstanceScope struct {
	ContestID     *int64
	ContestMode   string
	TeamID        *int64
	ServiceID     *int64
	FlagSubjectID int64
	ShareScope    model.ShareScope
}

type TopologyCreateNode struct {
	Key             string
	Image           string
	Env             map[string]string
	ServicePort     int
	ServiceProtocol string
	IsEntryPoint    bool
	NetworkKeys     []string
	NetworkAliases  []string
	Resources       *model.ResourceLimits
}

type TopologyCreateNetwork struct {
	Key      string
	Name     string
	Internal bool
	Shared   bool
}

type TopologyCreateRequest struct {
	Networks                   []TopologyCreateNetwork
	Nodes                      []TopologyCreateNode
	Policies                   []model.TopologyTrafficPolicy
	ReservedHostPort           int
	DisableEntryPortPublishing bool
}

type TopologyCreateResult struct {
	PrimaryContainerID string
	NetworkID          string
	AccessURL          string
	RuntimeDetails     model.InstanceRuntimeDetails
}

type PracticeInstanceScopeLockRepository interface {
	LockInstanceScope(ctx context.Context, userID, challengeID int64, scope InstanceScope) error
}

type PracticeScopedExistingInstanceRepository interface {
	FindScopedExistingInstance(ctx context.Context, userID, challengeID int64, scope InstanceScope) (*model.Instance, error)
}

type PracticeScopedRestartableInstanceRepository interface {
	FindScopedRestartableInstance(ctx context.Context, userID, challengeID int64, scope InstanceScope) (*model.Instance, error)
}

type PracticeScopedRunningCountRepository interface {
	CountScopedRunningInstances(ctx context.Context, userID int64, scope InstanceScope) (int, error)
}

type PracticeInstanceExpiryRepository interface {
	RefreshInstanceExpiry(ctx context.Context, instanceID int64, expiresAt time.Time) error
}

type PracticeInstanceRestartRepository interface {
	ResetInstanceRuntimeForRestart(ctx context.Context, instanceID int64, status string, expiresAt time.Time) error
}

type PracticeInstanceCreateRepository interface {
	CreateInstance(ctx context.Context, instance *model.Instance) error
}

type PracticeAWDServiceOperationCreateRepository interface {
	CreateAWDServiceOperation(ctx context.Context, operation *model.AWDServiceOperation) error
}

type PracticeAWDServiceOperationFinishRepository interface {
	FinishAWDServiceOperation(ctx context.Context, operationID int64, status, errorMessage string, finishedAt time.Time) error
}

type PracticePortReservationRepository interface {
	ReserveAvailablePort(ctx context.Context, start, end int) (int, error)
	BindReservedPort(ctx context.Context, port int, instanceID int64) error
}

type PracticeInstanceStartTxRepository interface {
	PracticeInstanceScopeLockRepository
	PracticeScopedExistingInstanceRepository
	PracticeScopedRunningCountRepository
	PracticeInstanceExpiryRepository
	PracticeInstanceCreateRepository
	PracticeAWDServiceOperationCreateRepository
	PracticePortReservationRepository
}

type PracticeInstanceRestartTxRepository interface {
	PracticeInstanceScopeLockRepository
	PracticeScopedRestartableInstanceRepository
	PracticeInstanceRestartRepository
	PracticeAWDServiceOperationCreateRepository
}

type PracticeAWDServiceOperationTxRepository interface {
	PracticeAWDServiceOperationCreateRepository
}

type PracticeInstanceStartTxManager interface {
	WithinInstanceStartTx(ctx context.Context, fn func(txRepo PracticeInstanceStartTxRepository) error) error
}

type PracticeInstanceRestartTxManager interface {
	WithinInstanceRestartTx(ctx context.Context, fn func(txRepo PracticeInstanceRestartTxRepository) error) error
}

type PracticeAWDServiceOperationTxManager interface {
	WithinAWDServiceOperationTx(ctx context.Context, fn func(txRepo PracticeAWDServiceOperationTxRepository) error) error
}

type PracticeContestLookupRepository interface {
	FindContestByID(ctx context.Context, contestID int64) (*model.Contest, error)
}

type PracticeContestChallengeLookupRepository interface {
	FindContestChallenge(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error)
}

type PracticeContestAWDServiceRepository interface {
	FindContestAWDService(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error)
	ListContestAWDServices(ctx context.Context, contestID int64) ([]*model.ContestAWDService, error)
}

type PracticeContestAWDInstanceRepository interface {
	ListContestAWDInstances(ctx context.Context, contestID int64) ([]*model.Instance, error)
}

type PracticeContestTeamRepository interface {
	FindContestTeam(ctx context.Context, contestID, teamID int64) (*model.Team, error)
	ListContestTeams(ctx context.Context, contestID int64) ([]*model.Team, error)
}

type PracticeContestRegistrationRepository interface {
	FindContestRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error)
}

type PracticeSubmissionWriteRepository interface {
	CreateSubmission(ctx context.Context, submission *model.Submission) error
	UpdateSubmission(ctx context.Context, submission *model.Submission) error
}

type PracticeSolvedSubmissionRepository interface {
	FindCorrectSubmission(ctx context.Context, userID, challengeID int64) (*model.Submission, error)
}

type PracticeSubmissionHistoryRepository interface {
	ListChallengeSubmissions(ctx context.Context, userID, challengeID int64, limit int) ([]model.Submission, error)
}

type PracticeSubmissionConstraintRepository interface {
	IsUniqueViolation(err error) bool
}

type PracticeUserLookupRepository interface {
	FindUserByID(ctx context.Context, userID int64) (*model.User, error)
}

type PracticeManualReviewListRepository interface {
	ListTeacherManualReviewSubmissions(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]TeacherManualReviewSubmissionRecord, int64, error)
}

type PracticeManualReviewLookupRepository interface {
	GetTeacherManualReviewSubmissionByID(ctx context.Context, id int64) (*TeacherManualReviewSubmissionRecord, error)
}

type TeacherManualReviewSubmissionRecord struct {
	Submission      model.Submission
	StudentUsername string
	StudentName     string
	ClassName       string
	ChallengeTitle  string
	ReviewerName    string
}

type PracticeScoreRepository interface {
	FindChallengeScore(ctx context.Context, challengeID int64) (*model.Challenge, error)
	FindChallengesScores(ctx context.Context, challengeIDs []int64) ([]model.Challenge, error)
	ListSolvedChallengeIDs(ctx context.Context, userID int64) ([]int64, error)
	UpsertUserScore(ctx context.Context, userScore *model.UserScore) error
}

type PracticeRankingRepository interface {
	FindUserScore(ctx context.Context, userID int64) (*model.UserScore, error)
	ListTopUserScores(ctx context.Context, limit int) ([]model.UserScore, error)
	FindUsersByIDs(ctx context.Context, userIDs []int64) ([]model.User, error)
}

type PracticeInstanceLookupRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Instance, error)
	FindByUserAndChallenge(ctx context.Context, userID, challengeID int64) (*model.Instance, error)
}

type PracticeInstanceRuntimeWriteRepository interface {
	UpdateRuntime(ctx context.Context, instance *model.Instance) error
	RefreshInstanceExpiry(ctx context.Context, instanceID int64, expiresAt time.Time) error
}

type PracticeInstanceAWDOperationRepository interface {
	FinishActiveAWDServiceOperationForInstance(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error
}

type PracticeInstanceStatusRepository interface {
	UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error
	TryTransitionStatus(ctx context.Context, id int64, fromStatus, toStatus string) (bool, error)
}

type PracticePendingInstanceRepository interface {
	ListPendingInstances(ctx context.Context, limit int) ([]*model.Instance, error)
}

type PracticeInstanceStatsRepository interface {
	CountInstancesByStatus(ctx context.Context, statuses []string) (int64, error)
}

type RuntimeInstanceService interface {
	CleanupRuntime(ctx context.Context, instance *model.Instance) error
	CreateTopology(ctx context.Context, req *TopologyCreateRequest) (*TopologyCreateResult, error)
	CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error)
}
