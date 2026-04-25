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
	Key          string
	Image        string
	Env          map[string]string
	ServicePort  int
	IsEntryPoint bool
	NetworkKeys  []string
	Resources    *model.ResourceLimits
}

type TopologyCreateNetwork struct {
	Key      string
	Internal bool
}

type TopologyCreateRequest struct {
	Networks         []TopologyCreateNetwork
	Nodes            []TopologyCreateNode
	Policies         []model.TopologyTrafficPolicy
	ReservedHostPort int
}

type TopologyCreateResult struct {
	PrimaryContainerID string
	NetworkID          string
	AccessURL          string
	RuntimeDetails     model.InstanceRuntimeDetails
}

type PracticeCommandTxRepository interface {
	LockInstanceScope(userID, challengeID int64, scope InstanceScope) error
	LockInstanceScopeWithContext(ctx context.Context, userID, challengeID int64, scope InstanceScope) error
	FindScopedExistingInstance(userID, challengeID int64, scope InstanceScope) (*model.Instance, error)
	FindScopedExistingInstanceWithContext(ctx context.Context, userID, challengeID int64, scope InstanceScope) (*model.Instance, error)
	CountScopedRunningInstances(userID int64, scope InstanceScope) (int, error)
	CountScopedRunningInstancesWithContext(ctx context.Context, userID int64, scope InstanceScope) (int, error)
	RefreshInstanceExpiry(ctx context.Context, instanceID int64, expiresAt time.Time) error
	CreateInstance(instance *model.Instance) error
	CreateInstanceWithContext(ctx context.Context, instance *model.Instance) error
	ReserveAvailablePort(start, end int) (int, error)
	ReserveAvailablePortWithContext(ctx context.Context, start, end int) (int, error)
	BindReservedPort(port int, instanceID int64) error
	BindReservedPortWithContext(ctx context.Context, port int, instanceID int64) error
	CreateSubmission(submission *model.Submission) error
}

type PracticeCommandRepository interface {
	WithinTransaction(ctx context.Context, fn func(txRepo PracticeCommandTxRepository) error) error
	FindContestByIDWithContext(ctx context.Context, contestID int64) (*model.Contest, error)
	FindContestChallengeWithContext(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error)
	FindContestAWDServiceWithContext(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error)
	FindContestRegistrationWithContext(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error)
	CreateSubmission(submission *model.Submission) error
	CreateSubmissionWithContext(ctx context.Context, submission *model.Submission) error
	FindCorrectSubmission(userID, challengeID int64) (*model.Submission, error)
	FindCorrectSubmissionWithContext(ctx context.Context, userID, challengeID int64) (*model.Submission, error)
	ListChallengeSubmissions(userID, challengeID int64, limit int) ([]model.Submission, error)
	ListChallengeSubmissionsWithContext(ctx context.Context, userID, challengeID int64, limit int) ([]model.Submission, error)
	UpdateSubmission(submission *model.Submission) error
	UpdateSubmissionWithContext(ctx context.Context, submission *model.Submission) error
	FindUserByID(userID int64) (*model.User, error)
	FindUserByIDWithContext(ctx context.Context, userID int64) (*model.User, error)
	ListTeacherManualReviewSubmissions(query *dto.TeacherManualReviewSubmissionQuery) ([]TeacherManualReviewSubmissionRecord, int64, error)
	ListTeacherManualReviewSubmissionsWithContext(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]TeacherManualReviewSubmissionRecord, int64, error)
	GetTeacherManualReviewSubmissionByID(id int64) (*TeacherManualReviewSubmissionRecord, error)
	GetTeacherManualReviewSubmissionByIDWithContext(ctx context.Context, id int64) (*TeacherManualReviewSubmissionRecord, error)
	IsUniqueViolation(err error) bool
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
	FindChallengeScoreWithContext(ctx context.Context, challengeID int64) (*model.Challenge, error)
	FindChallengesScoresWithContext(ctx context.Context, challengeIDs []int64) ([]model.Challenge, error)
	ListSolvedChallengeIDsWithContext(ctx context.Context, userID int64) ([]int64, error)
	UpsertUserScoreWithContext(ctx context.Context, userScore *model.UserScore) error
}

type PracticeRankingRepository interface {
	FindUserScoreWithContext(ctx context.Context, userID int64) (*model.UserScore, error)
	ListTopUserScoresWithContext(ctx context.Context, limit int) ([]model.UserScore, error)
	FindUsersByIDsWithContext(ctx context.Context, userIDs []int64) ([]model.User, error)
}

type InstanceRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Instance, error)
	UpdateRuntime(ctx context.Context, instance *model.Instance) error
	RefreshInstanceExpiry(ctx context.Context, instanceID int64, expiresAt time.Time) error
	UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error
	FindByUserAndChallenge(ctx context.Context, userID, challengeID int64) (*model.Instance, error)
	ListPendingInstances(ctx context.Context, limit int) ([]*model.Instance, error)
	TryTransitionStatus(ctx context.Context, id int64, fromStatus, toStatus string) (bool, error)
	CountInstancesByStatus(ctx context.Context, statuses []string) (int64, error)
}

type RuntimeInstanceService interface {
	CleanupRuntime(ctx context.Context, instance *model.Instance) error
	CreateTopology(ctx context.Context, req *TopologyCreateRequest) (*TopologyCreateResult, error)
	CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error)
}
