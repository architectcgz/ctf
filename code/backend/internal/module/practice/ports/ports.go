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

type PracticeCommandTxRepository interface {
	LockInstanceScope(ctx context.Context, userID, challengeID int64, scope InstanceScope) error
	FindScopedExistingInstance(ctx context.Context, userID, challengeID int64, scope InstanceScope) (*model.Instance, error)
	FindScopedRestartableInstance(ctx context.Context, userID, challengeID int64, scope InstanceScope) (*model.Instance, error)
	CountScopedRunningInstances(ctx context.Context, userID int64, scope InstanceScope) (int, error)
	RefreshInstanceExpiry(ctx context.Context, instanceID int64, expiresAt time.Time) error
	ResetInstanceRuntimeForRestart(ctx context.Context, instanceID int64, status string, expiresAt time.Time) error
	CreateInstance(ctx context.Context, instance *model.Instance) error
	CreateAWDServiceOperation(ctx context.Context, operation *model.AWDServiceOperation) error
	FinishAWDServiceOperation(ctx context.Context, operationID int64, status, errorMessage string, finishedAt time.Time) error
	ReserveAvailablePort(ctx context.Context, start, end int) (int, error)
	BindReservedPort(ctx context.Context, port int, instanceID int64) error
}

type PracticeTransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(txRepo PracticeCommandTxRepository) error) error
}

type PracticeContestCommandRepository interface {
	FindContestByID(ctx context.Context, contestID int64) (*model.Contest, error)
	FindContestChallenge(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error)
	FindContestAWDService(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error)
	ListContestAWDServices(ctx context.Context, contestID int64) ([]*model.ContestAWDService, error)
	ListContestAWDInstances(ctx context.Context, contestID int64) ([]*model.Instance, error)
	FindContestTeam(ctx context.Context, contestID, teamID int64) (*model.Team, error)
	ListContestTeams(ctx context.Context, contestID int64) ([]*model.Team, error)
	FindContestRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error)
}

type PracticeSubmissionCommandRepository interface {
	CreateSubmission(ctx context.Context, submission *model.Submission) error
	FindCorrectSubmission(ctx context.Context, userID, challengeID int64) (*model.Submission, error)
	ListChallengeSubmissions(ctx context.Context, userID, challengeID int64, limit int) ([]model.Submission, error)
	UpdateSubmission(ctx context.Context, submission *model.Submission) error
	IsUniqueViolation(err error) bool
}

type PracticeUserLookupRepository interface {
	FindUserByID(ctx context.Context, userID int64) (*model.User, error)
}

type PracticeManualReviewCommandRepository interface {
	ListTeacherManualReviewSubmissions(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]TeacherManualReviewSubmissionRecord, int64, error)
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

type InstanceRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Instance, error)
	UpdateRuntime(ctx context.Context, instance *model.Instance) error
	FinishActiveAWDServiceOperationForInstance(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error
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
