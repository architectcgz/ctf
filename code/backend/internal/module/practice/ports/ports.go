package ports

import (
	"context"
	"errors"
	"time"

	"ctf-platform/internal/model"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceentity "ctf-platform/internal/module/practice/entity"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

var ErrPracticeContestNotFound = errors.New("practice contest not found")
var ErrPracticeContestChallengeNotFound = errors.New("practice contest challenge not found")
var ErrPracticeContestAWDServiceNotFound = errors.New("practice contest awd service not found")
var ErrPracticeContestTeamNotFound = errors.New("practice contest team not found")
var ErrPracticeContestRegistrationNotFound = errors.New("practice contest registration not found")
var ErrPracticeChallengeNotFound = errors.New("practice challenge not found")
var ErrPracticeChallengeTopologyNotFound = errors.New("practice challenge topology not found")
var ErrPracticeUserScoreNotFound = errors.New("practice user score not found")
var ErrPracticeManualReviewSubmissionNotFound = errors.New("practice manual review submission not found")
var ErrPracticeSolvedSubmissionNotFound = errors.New("practice solved submission not found")
var ErrPracticeUserNotFound = errors.New("practice user not found")

const (
	ContestModeAWD = "awd"

	ContestStatusRegistration = "registration"
	ContestStatusRunning      = "running"
	ContestStatusFrozen       = "frozen"
	ContestStatusEnded        = "ended"

	ContestRegistrationStatusPending  = "pending"
	ContestRegistrationStatusApproved = "approved"

	SubmissionReviewStatusNotRequired = "not_required"
	SubmissionReviewStatusPending     = "pending"
	SubmissionReviewStatusApproved    = "approved"
	SubmissionReviewStatusRejected    = "rejected"
)

type ContestRecord struct {
	ID            int64
	Mode          string
	EndTime       time.Time
	PausedSeconds int64
	Status        string
}

type ContestChallengeRecord struct {
	ContestID   int64
	ChallengeID int64
	IsVisible   bool
}

type ContestAWDServiceRecord struct {
	ID              int64
	ContestID       int64
	AWDChallengeID  int64
	DisplayName     string
	ServiceSnapshot string
	ScoreConfig     string
	IsVisible       bool
}

type ContestTeamRecord struct {
	ID        int64
	ContestID int64
	Name      string
	CaptainID int64
}

type SubmissionRecord struct {
	ID            int64
	UserID        int64
	ChallengeID   int64
	ContestID     *int64
	TeamID        *int64
	Flag          string
	IsCorrect     bool
	ReviewStatus  string
	ReviewedBy    *int64
	ReviewedAt    *time.Time
	ReviewComment string
	Score         int
	SubmittedAt   time.Time
	UpdatedAt     time.Time
}

type InstanceScope struct {
	ContestID     *int64
	ContestMode   string
	TeamID        *int64
	ServiceID     *int64
	FlagSubjectID int64
	ShareScope    instancecontracts.ShareScope
}

type TopologyCreateNode struct {
	Key             string
	Image           string
	Env             map[string]string
	Command         []string
	WorkingDir      string
	ServicePort     int
	ServiceProtocol string
	IsEntryPoint    bool
	NetworkKeys     []string
	NetworkAliases  []string
	Mounts          []model.ContainerMount
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
	ContainerName              string
}

type TopologyCreateResult struct {
	PrimaryContainerID string
	NetworkID          string
	AccessURL          string
	RuntimeDetails     model.InstanceRuntimeDetails
}

type ManagedContainerState = runtimeports.ManagedContainerState

type PracticeInstanceScopeLockRepository interface {
	LockInstanceScope(ctx context.Context, userID, challengeID int64, scope InstanceScope) error
}

type PracticeScopedExistingInstanceRepository interface {
	FindScopedExistingInstance(ctx context.Context, userID, challengeID int64, scope InstanceScope) (*instancecontracts.Instance, error)
}

type PracticeScopedRestartableInstanceRepository interface {
	FindScopedRestartableInstance(ctx context.Context, userID, challengeID int64, scope InstanceScope) (*instancecontracts.Instance, error)
}

type PracticeScopedRunningCountRepository interface {
	CountScopedRunningInstances(ctx context.Context, userID int64, scope InstanceScope) (int, error)
}

type PracticeInstanceExpiryRepository interface {
	RefreshInstanceExpiry(ctx context.Context, instanceID int64, expiresAt time.Time) error
}

type PracticeInstanceRestartRepository interface {
	ResetInstanceRuntimeForRestart(ctx context.Context, instanceID int64, status string, expiresAt time.Time, preserveHostPort bool) error
}

type PracticeInstanceRestartHostPortRepository interface {
	IsHostPortReusableForRestart(ctx context.Context, instanceID int64, hostPort int) (bool, error)
}

type PracticeInstanceCreateRepository interface {
	CreateInstance(ctx context.Context, instance *instancecontracts.Instance) error
}

type PracticeAWDServiceOperationCreateRepository interface {
	CreateAWDServiceOperation(ctx context.Context, operation *runtimecontracts.AWDServiceOperation) error
}

type PracticeAWDServiceOperationFinishRepository interface {
	FinishAWDServiceOperation(ctx context.Context, operationID int64, status, errorMessage string, finishedAt time.Time) error
}

type PracticePortReservationRepository interface {
	ReserveAvailablePort(ctx context.Context, start, end int) (int, error)
	ReserveAvailablePortExcluding(ctx context.Context, start, end, excludedPort int) (int, error)
	BindReservedPort(ctx context.Context, port int, instanceID int64) error
	ReleaseReservedPort(ctx context.Context, port int) error
	ReleasePortForInstance(ctx context.Context, port int, instanceID int64) error
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
	PracticeInstanceRestartHostPortRepository
	PracticeAWDServiceOperationCreateRepository
	PracticePortReservationRepository
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
	FindContestByID(ctx context.Context, contestID int64) (*ContestRecord, error)
}

type PracticeDesiredAWDContestRepository interface {
	ListDesiredRuntimeAWDContests(ctx context.Context) ([]*ContestRecord, error)
}

type PracticeContestChallengeLookupRepository interface {
	FindContestChallenge(ctx context.Context, contestID, challengeID int64) (*ContestChallengeRecord, error)
}

type PracticeContestAWDServiceRepository interface {
	FindContestAWDService(ctx context.Context, contestID, serviceID int64) (*ContestAWDServiceRecord, error)
	ListContestAWDServices(ctx context.Context, contestID int64) ([]*ContestAWDServiceRecord, error)
}

type ContestAWDServiceRuntimeSubject struct {
	ServiceID        int64
	ChallengeID      int64
	Visible          bool
	SeedSignature    string
	RuntimeChallenge *model.Challenge
	RuntimeTopology  *model.ChallengeTopology
	WorkspaceConfig  *ContestAWDDefenseWorkspaceConfig
}

type ContestAWDDefenseWorkspaceConfig struct {
	SeedRoot        string
	WorkspaceRoots  []ContestAWDDefenseWorkspaceRoot
	RuntimeMounts   []ContestAWDDefenseRuntimeMount
	CheckerTokenEnv string
}

type ContestAWDDefenseWorkspaceRoot struct {
	Source   string
	ReadOnly bool
}

type ContestAWDDefenseRuntimeMount struct {
	Source   string
	Target   string
	ReadOnly bool
}

type PracticeContestAWDServiceRuntimeSubjectRepository interface {
	FindContestAWDServiceRuntimeSubject(ctx context.Context, contestID, serviceID int64) (*ContestAWDServiceRuntimeSubject, error)
}

type PracticeContestAWDInstanceRepository interface {
	ListContestAWDInstances(ctx context.Context, contestID int64) ([]*instancecontracts.Instance, error)
}

type PracticeAWDDefenseWorkspaceLookupRepository interface {
	FindAWDDefenseWorkspace(ctx context.Context, contestID, teamID, serviceID int64) (*runtimecontracts.AWDDefenseWorkspace, error)
}

type PracticeAWDDefenseWorkspaceWriteRepository interface {
	UpsertAWDDefenseWorkspace(ctx context.Context, workspace *runtimecontracts.AWDDefenseWorkspace) error
	BumpAWDDefenseWorkspaceRevision(ctx context.Context, contestID, teamID, serviceID, instanceID int64, seedSignature string) error
}

type PracticeContestTeamRepository interface {
	FindContestTeam(ctx context.Context, contestID, teamID int64) (*ContestTeamRecord, error)
	ListContestTeams(ctx context.Context, contestID int64) ([]*ContestTeamRecord, error)
}

type ContestParticipation struct {
	Status string
	TeamID *int64
}

type PracticeContestRegistrationRepository interface {
	FindContestRegistration(ctx context.Context, contestID, userID int64) (*ContestParticipation, error)
}

type PracticeAWDScopeControlRepository interface {
	ListContestAWDScopeControls(ctx context.Context, contestID int64) ([]*runtimecontracts.AWDScopeControl, error)
	ListScopeAWDScopeControls(ctx context.Context, contestID, teamID, serviceID int64) ([]*runtimecontracts.AWDScopeControl, error)
	UpsertAWDScopeControl(ctx context.Context, control *runtimecontracts.AWDScopeControl) error
	DeleteAWDScopeControl(ctx context.Context, contestID, teamID int64, scopeType, controlType string, serviceID int64) error
}

type PracticeContestScopeRepository interface {
	PracticeContestLookupRepository
	PracticeContestChallengeLookupRepository
	PracticeContestAWDServiceRepository
	PracticeContestAWDServiceRuntimeSubjectRepository
	PracticeContestTeamRepository
	PracticeContestRegistrationRepository
}

type PracticeRuntimeSubjectRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Challenge, error)
	FindChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error)
}

type PracticeSubmissionWriteRepository interface {
	CreateSubmission(ctx context.Context, submission *SubmissionRecord) error
	UpdateSubmission(ctx context.Context, submission *SubmissionRecord) error
}

type PracticeSubmissionUpdateRepository interface {
	UpdateSubmission(ctx context.Context, submission *SubmissionRecord) error
}

type PracticeSolvedSubmissionRepository interface {
	FindCorrectSubmission(ctx context.Context, userID, challengeID int64) (*SubmissionRecord, error)
}

type PracticeSubmissionHistoryRepository interface {
	ListChallengeSubmissions(ctx context.Context, userID, challengeID int64, limit int) ([]SubmissionRecord, error)
}

type PracticeSubmissionConstraintRepository interface {
	IsUniqueViolation(err error) bool
}

type PracticeUserLookupRepository interface {
	FindUserByID(ctx context.Context, userID int64) (*model.User, error)
}

type PracticeManualReviewListRepository interface {
	ListTeacherManualReviewSubmissions(ctx context.Context, query *practicecontracts.TeacherManualReviewSubmissionQuery) ([]TeacherManualReviewSubmissionRecord, int64, error)
}

type PracticeManualReviewLookupRepository interface {
	GetTeacherManualReviewSubmissionByID(ctx context.Context, id int64) (*TeacherManualReviewSubmissionRecord, error)
}

type PracticeManualReviewRepository interface {
	PracticeSubmissionUpdateRepository
	PracticeSolvedSubmissionRepository
	PracticeUserLookupRepository
	PracticeManualReviewListRepository
	PracticeManualReviewLookupRepository
}

type TeacherManualReviewSubmissionRecord struct {
	Submission      SubmissionRecord
	StudentUsername string
	StudentName     string
	ClassName       string
	ChallengeTitle  string
	ReviewerName    string
}

type CategoryProgressStat struct {
	Category string `gorm:"column:category"`
	Solved   int    `gorm:"column:solved"`
	Total    int    `gorm:"column:total"`
}

type DifficultyProgressStat struct {
	Difficulty string `gorm:"column:difficulty"`
	Solved     int    `gorm:"column:solved"`
	Total      int    `gorm:"column:total"`
}

type UserProgressCategorySnapshot struct {
	Category string
	Solved   int
	Total    int
}

type UserProgressDifficultySnapshot struct {
	Difficulty string
	Solved     int
	Total      int
}

type UserProgressSnapshot struct {
	TotalScore      int
	TotalSolved     int
	Rank            int
	CategoryStats   []UserProgressCategorySnapshot
	DifficultyStats []UserProgressDifficultySnapshot
}

type TimelineEventRecord struct {
	Type        string
	ChallengeID int64
	Title       string
	Timestamp   time.Time
	IsCorrect   *bool
	Points      *int
	Detail      string
}

type TimelineEventSnapshot struct {
	Type        string
	ChallengeID int64
	Title       string
	Timestamp   time.Time
	IsCorrect   *bool
	Points      *int
	Detail      string
}

type TimelineSnapshot struct {
	Events []TimelineEventSnapshot
}

type PracticeProgressQueryRepository interface {
	GetUserProgress(ctx context.Context, userID int64) (totalScore int, totalSolved int, err error)
	GetUserRank(ctx context.Context, userID int64) (int, error)
	GetCategoryStats(ctx context.Context, userID int64) ([]CategoryProgressStat, error)
	GetDifficultyStats(ctx context.Context, userID int64) ([]DifficultyProgressStat, error)
}

type PracticeTimelineQueryRepository interface {
	GetUserTimeline(ctx context.Context, userID int64, limit, offset int) ([]TimelineEventRecord, error)
}

type PracticeUserProgressCache interface {
	GetUserProgress(ctx context.Context, userID int64) (*UserProgressSnapshot, bool, error)
	StoreUserProgress(ctx context.Context, userID int64, resp *UserProgressSnapshot, ttl time.Duration) error
	DeleteUserProgress(ctx context.Context, userID int64) error
}

type PracticeChallengeScoreRepository interface {
	FindChallengeScore(ctx context.Context, challengeID int64) (*model.Challenge, error)
	FindChallengesScores(ctx context.Context, challengeIDs []int64) ([]model.Challenge, error)
}

type PracticeSolvedChallengeRepository interface {
	ListSolvedChallengeIDs(ctx context.Context, userID int64) ([]int64, error)
}

type PracticeUserScoreWriteRepository interface {
	UpsertUserScore(ctx context.Context, userScore *practiceentity.UserScore) error
}

type PracticeUserScoreReadRepository interface {
	FindUserScore(ctx context.Context, userID int64) (*practiceentity.UserScore, error)
}

type PracticeRankingListRepository interface {
	ListTopUserScores(ctx context.Context, limit int) ([]practiceentity.UserScore, error)
}

type PracticeUserDirectoryRepository interface {
	FindUsersByIDs(ctx context.Context, userIDs []int64) ([]model.User, error)
}

type PracticeScoreLockLease interface {
	Key(ctx context.Context) string
	Release(ctx context.Context) (bool, error)
}

type PracticeScoreStateStore interface {
	AcquireUserScoreUpdateLock(ctx context.Context, userID int64, ttl time.Duration) (PracticeScoreLockLease, bool, error)
	LoadUserScoreCache(ctx context.Context, userID int64) (*practicecontracts.UserScoreInfo, bool, error)
	StoreUserScoreCache(ctx context.Context, info *practicecontracts.UserScoreInfo, ttl time.Duration) error
	SyncUserScoreState(ctx context.Context, info *practicecontracts.UserScoreInfo, ttl time.Duration) error
}

type PracticeFlagSubmitRateLimitStore interface {
	AllowFlagSubmit(ctx context.Context, userID, challengeID int64, limit int, window time.Duration) (bool, error)
}

type DesiredAWDReconcileState struct {
	FailureCount    int
	LastFailureAt   time.Time
	NextAttemptAt   time.Time
	SuppressedUntil time.Time
	LastError       string
}

type PracticeDesiredAWDReconcileStateStore interface {
	LoadDesiredAWDReconcileState(ctx context.Context, contestID, teamID, serviceID int64) (*DesiredAWDReconcileState, bool, error)
	StoreDesiredAWDReconcileState(ctx context.Context, contestID, teamID, serviceID int64, state *DesiredAWDReconcileState) error
	DeleteDesiredAWDReconcileState(ctx context.Context, contestID, teamID, serviceID int64) error
}

type PracticeInstanceReadinessProbe interface {
	ProbeAccessURL(ctx context.Context, accessURL string, timeout time.Duration) error
}

type PracticeInstanceLookupRepository interface {
	FindByID(ctx context.Context, id int64) (*instancecontracts.Instance, error)
	FindByUserAndChallenge(ctx context.Context, userID, challengeID int64) (*instancecontracts.Instance, error)
}

type PracticeInstanceRuntimeWriteRepository interface {
	UpdateRuntime(ctx context.Context, instance *instancecontracts.Instance) error
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
	ListPendingInstances(ctx context.Context, limit int) ([]*instancecontracts.Instance, error)
}

type PracticeInstanceStatsRepository interface {
	CountInstancesByStatus(ctx context.Context, statuses []string) (int64, error)
}

type RuntimeInstanceService interface {
	CleanupRuntime(ctx context.Context, instance *instancecontracts.Instance) error
	CreateTopology(ctx context.Context, req *TopologyCreateRequest) (*TopologyCreateResult, error)
	CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error)
	InspectManagedContainer(ctx context.Context, containerID string) (*ManagedContainerState, error)
}
