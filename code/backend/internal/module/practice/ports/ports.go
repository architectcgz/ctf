package ports

import (
	"context"

	"ctf-platform/internal/model"
)

type InstanceScope struct {
	ContestID     *int64
	TeamID        *int64
	FlagSubjectID int64
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
	LockInstanceScope(userID int64, scope InstanceScope) error
	FindScopedExistingInstance(userID, challengeID int64, scope InstanceScope) (*model.Instance, error)
	CountScopedRunningInstances(userID int64, scope InstanceScope) (int, error)
	CreateInstance(instance *model.Instance) error
	ReserveAvailablePort(start, end int) (int, error)
	BindReservedPort(port int, instanceID int64) error
}

type PracticeCommandRepository interface {
	WithinTransaction(ctx context.Context, fn func(txRepo PracticeCommandTxRepository) error) error
	FindContestByIDWithContext(ctx context.Context, contestID int64) (*model.Contest, error)
	FindContestChallengeWithContext(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error)
	FindContestRegistrationWithContext(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error)
	CreateSubmission(submission *model.Submission) error
	FindCorrectSubmission(userID, challengeID int64) (*model.Submission, error)
	IsUniqueViolation(err error) bool
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
	UpdateRuntime(instance *model.Instance) error
	UpdateStatusAndReleasePort(id int64, status string) error
	FindByUserAndChallenge(userID, challengeID int64) (*model.Instance, error)
}

type RuntimeInstanceService interface {
	CleanupRuntime(instance *model.Instance) error
	CreateTopology(ctx context.Context, req *TopologyCreateRequest) (*TopologyCreateResult, error)
	CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error)
}
