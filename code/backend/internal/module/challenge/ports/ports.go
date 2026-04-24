package ports

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type ChallengeCommandRepository interface {
	CreateWithHintsWithContext(ctx context.Context, challenge *model.Challenge, hints []*model.ChallengeHint) error
	FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error)
	UpdateWithContext(ctx context.Context, challenge *model.Challenge) error
	UpdateWithHintsWithContext(ctx context.Context, challenge *model.Challenge, hints []*model.ChallengeHint, replaceHints bool) error
	DeleteWithContext(ctx context.Context, id int64) error
	HasRunningInstancesWithContext(ctx context.Context, challengeID int64) (bool, error)
	CreatePublishCheckJob(ctx context.Context, job *model.ChallengePublishCheckJob) error
	FindPublishCheckJobByID(ctx context.Context, id int64) (*model.ChallengePublishCheckJob, error)
	FindActivePublishCheckJobByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengePublishCheckJob, error)
	FindLatestPublishCheckJobByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengePublishCheckJob, error)
	ListPendingPublishCheckJobs(ctx context.Context, limit int) ([]*model.ChallengePublishCheckJob, error)
	TryStartPublishCheckJob(ctx context.Context, id int64, startedAt time.Time) (bool, error)
	UpdatePublishCheckJob(ctx context.Context, job *model.ChallengePublishCheckJob) error
}

type ChallengeFlagRepository interface {
	FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error)
	UpdateWithContext(ctx context.Context, challenge *model.Challenge) error
}

type ChallengeQueryRepository interface {
	FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error)
	ListWithContext(ctx context.Context, query *dto.ChallengeQuery) ([]*model.Challenge, int64, error)
	ListHintsByChallengeIDWithContext(ctx context.Context, challengeID int64) ([]*model.ChallengeHint, error)
	GetSolvedStatusWithContext(ctx context.Context, userID, challengeID int64) (bool, error)
	GetSolvedCountWithContext(ctx context.Context, challengeID int64) (int64, error)
	GetTotalAttemptsWithContext(ctx context.Context, challengeID int64) (int64, error)
	BatchGetSolvedStatusWithContext(ctx context.Context, userID int64, challengeIDs []int64) (map[int64]bool, error)
	BatchGetSolvedCountWithContext(ctx context.Context, challengeIDs []int64) (map[int64]int64, error)
	BatchGetTotalAttemptsWithContext(ctx context.Context, challengeIDs []int64) (map[int64]int64, error)
	ListPublishedWithContext(ctx context.Context, query *dto.ChallengeQuery) ([]*model.Challenge, int64, error)
}

type AWDServiceTemplateCommandRepository interface {
	CreateAWDServiceTemplateWithContext(ctx context.Context, template *model.AWDServiceTemplate) error
	FindAWDServiceTemplateByIDWithContext(ctx context.Context, id int64) (*model.AWDServiceTemplate, error)
	UpdateAWDServiceTemplateWithContext(ctx context.Context, template *model.AWDServiceTemplate) error
	DeleteAWDServiceTemplateWithContext(ctx context.Context, id int64) error
}

type AWDServiceTemplateQueryRepository interface {
	FindAWDServiceTemplateByIDWithContext(ctx context.Context, id int64) (*model.AWDServiceTemplate, error)
	ListAWDServiceTemplatesWithContext(ctx context.Context, query *dto.AWDServiceTemplateQuery) ([]*model.AWDServiceTemplate, int64, error)
}

type ChallengeImageUsageRepository interface {
	CountByImageIDWithContext(ctx context.Context, imageID int64) (int64, error)
}

type ChallengeWriteupRepository interface {
	FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error)
	FindUserByIDWithContext(ctx context.Context, userID int64) (*model.User, error)
	FindWriteupByChallengeIDWithContext(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error)
	UpsertWriteupWithContext(ctx context.Context, writeup *model.ChallengeWriteup) error
	DeleteWriteupByChallengeIDWithContext(ctx context.Context, challengeID int64) error
	FindReleasedWriteupByChallengeIDWithContext(ctx context.Context, challengeID int64, now time.Time) (*model.ChallengeWriteup, error)
	GetSolvedStatusWithContext(ctx context.Context, userID, challengeID int64) (bool, error)
	FindSubmissionWriteupByUserChallengeWithContext(ctx context.Context, userID, challengeID int64) (*model.SubmissionWriteup, error)
	FindSubmissionWriteupByIDWithContext(ctx context.Context, id int64) (*model.SubmissionWriteup, error)
	UpsertSubmissionWriteupWithContext(ctx context.Context, writeup *model.SubmissionWriteup) error
	GetTeacherSubmissionWriteupByIDWithContext(ctx context.Context, id int64) (*TeacherSubmissionWriteupRecord, error)
	ListTeacherSubmissionWriteupsWithContext(ctx context.Context, query *dto.TeacherSubmissionWriteupQuery) ([]TeacherSubmissionWriteupRecord, int64, error)
	ListRecommendedSolutionsByChallengeIDWithContext(ctx context.Context, challengeID int64, now time.Time) ([]RecommendedSolutionRecord, error)
	ListCommunitySolutionsByChallengeIDWithContext(ctx context.Context, challengeID int64, query *dto.CommunityChallengeSolutionQuery) ([]CommunitySolutionRecord, int64, error)
}

type TeacherSubmissionWriteupRecord struct {
	Submission      model.SubmissionWriteup
	StudentUsername string
	StudentName     string
	StudentNo       string
	ClassName       string
	ChallengeTitle  string
}

type RecommendedSolutionRecord struct {
	SourceType    string
	SourceID      int64
	ChallengeID   int64
	Title         string
	Content       string
	AuthorName    string
	IsRecommended bool
	RecommendedAt *time.Time
	UpdatedAt     time.Time
}

type CommunitySolutionRecord struct {
	Submission     model.SubmissionWriteup
	AuthorName     string
	ChallengeID    int64
	ChallengeTitle string
}

type ChallengeTopologyRepository interface {
	FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error)
	FindChallengeTopologyByChallengeIDWithContext(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error)
	UpsertChallengeTopologyWithContext(ctx context.Context, topology *model.ChallengeTopology) error
	DeleteChallengeTopologyByChallengeIDWithContext(ctx context.Context, challengeID int64) error
}

type ImageRepository interface {
	CreateWithContext(ctx context.Context, image *model.Image) error
	FindByIDWithContext(ctx context.Context, id int64) (*model.Image, error)
	FindByNameTagWithContext(ctx context.Context, name, tag string) (*model.Image, error)
	ListWithContext(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error)
	UpdateWithContext(ctx context.Context, image *model.Image) error
	DeleteWithContext(ctx context.Context, id int64) error
}

type EnvironmentTemplateRepository interface {
	CreateWithContext(ctx context.Context, template *model.EnvironmentTemplate) error
	UpdateWithContext(ctx context.Context, template *model.EnvironmentTemplate) error
	DeleteWithContext(ctx context.Context, id int64) error
	FindByIDWithContext(ctx context.Context, id int64) (*model.EnvironmentTemplate, error)
	ListWithContext(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error)
	IncrementUsageWithContext(ctx context.Context, id int64) error
}

type TagRepository interface {
	Create(ctx context.Context, tag *model.Tag) error
	List(ctx context.Context, tagType string) ([]*model.Tag, error)
	FindByIDs(ctx context.Context, ids []int64) ([]*model.Tag, error)
	AttachTagsInTx(ctx context.Context, challengeID int64, tagIDs []int64) error
	DetachFromChallenge(ctx context.Context, challengeID, tagID int64) error
	FindByChallengeID(ctx context.Context, challengeID int64) ([]*model.Tag, error)
	DeleteWithContext(ctx context.Context, id int64) error
	CountChallengesByTagID(ctx context.Context, tagID int64) (int64, error)
}

type ImageRuntime interface {
	InspectImageSize(ctx context.Context, imageRef string) (int64, error)
	RemoveImage(ctx context.Context, imageRef string) error
}

type RuntimeTopologyCreateNode struct {
	Key          string
	Image        string
	Env          map[string]string
	ServicePort  int
	IsEntryPoint bool
	NetworkKeys  []string
	Resources    *model.ResourceLimits
}

type RuntimeTopologyCreateNetwork struct {
	Key      string
	Internal bool
}

type RuntimeTopologyCreateRequest struct {
	Networks []RuntimeTopologyCreateNetwork
	Nodes    []RuntimeTopologyCreateNode
	Policies []model.TopologyTrafficPolicy
}

type RuntimeTopologyCreateResult struct {
	AccessURL      string
	RuntimeDetails model.InstanceRuntimeDetails
}

type ChallengeRuntimeProbe interface {
	CreateTopology(ctx context.Context, req *RuntimeTopologyCreateRequest) (*RuntimeTopologyCreateResult, error)
	CreateContainer(ctx context.Context, imageName string, env map[string]string) (accessURL string, runtimeDetails model.InstanceRuntimeDetails, err error)
	CleanupRuntimeDetails(ctx context.Context, details model.InstanceRuntimeDetails) error
}
