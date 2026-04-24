package ports

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type ChallengeCommandRepository interface {
	CreateWithHints(ctx context.Context, challenge *model.Challenge, hints []*model.ChallengeHint) error
	FindByID(ctx context.Context, id int64) (*model.Challenge, error)
	Update(ctx context.Context, challenge *model.Challenge) error
	UpdateWithHints(ctx context.Context, challenge *model.Challenge, hints []*model.ChallengeHint, replaceHints bool) error
	Delete(ctx context.Context, id int64) error
	HasRunningInstances(ctx context.Context, challengeID int64) (bool, error)
	CreatePublishCheckJob(ctx context.Context, job *model.ChallengePublishCheckJob) error
	FindPublishCheckJobByID(ctx context.Context, id int64) (*model.ChallengePublishCheckJob, error)
	FindActivePublishCheckJobByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengePublishCheckJob, error)
	FindLatestPublishCheckJobByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengePublishCheckJob, error)
	ListPendingPublishCheckJobs(ctx context.Context, limit int) ([]*model.ChallengePublishCheckJob, error)
	TryStartPublishCheckJob(ctx context.Context, id int64, startedAt time.Time) (bool, error)
	UpdatePublishCheckJob(ctx context.Context, job *model.ChallengePublishCheckJob) error
}

type ChallengeFlagRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Challenge, error)
	Update(ctx context.Context, challenge *model.Challenge) error
}

type ChallengeQueryRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Challenge, error)
	List(ctx context.Context, query *dto.ChallengeQuery) ([]*model.Challenge, int64, error)
	ListHintsByChallengeID(ctx context.Context, challengeID int64) ([]*model.ChallengeHint, error)
	GetSolvedStatus(ctx context.Context, userID, challengeID int64) (bool, error)
	GetSolvedCount(ctx context.Context, challengeID int64) (int64, error)
	GetTotalAttempts(ctx context.Context, challengeID int64) (int64, error)
	BatchGetSolvedStatus(ctx context.Context, userID int64, challengeIDs []int64) (map[int64]bool, error)
	BatchGetSolvedCount(ctx context.Context, challengeIDs []int64) (map[int64]int64, error)
	BatchGetTotalAttempts(ctx context.Context, challengeIDs []int64) (map[int64]int64, error)
	ListPublished(ctx context.Context, query *dto.ChallengeQuery) ([]*model.Challenge, int64, error)
}

type AWDServiceTemplateCommandRepository interface {
	CreateAWDServiceTemplate(ctx context.Context, template *model.AWDServiceTemplate) error
	FindAWDServiceTemplateByID(ctx context.Context, id int64) (*model.AWDServiceTemplate, error)
	UpdateAWDServiceTemplate(ctx context.Context, template *model.AWDServiceTemplate) error
	DeleteAWDServiceTemplate(ctx context.Context, id int64) error
}

type AWDServiceTemplateQueryRepository interface {
	FindAWDServiceTemplateByID(ctx context.Context, id int64) (*model.AWDServiceTemplate, error)
	ListAWDServiceTemplates(ctx context.Context, query *dto.AWDServiceTemplateQuery) ([]*model.AWDServiceTemplate, int64, error)
}

type ChallengeImageUsageRepository interface {
	CountByImageID(ctx context.Context, imageID int64) (int64, error)
}

type ChallengeWriteupRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Challenge, error)
	FindUserByID(ctx context.Context, userID int64) (*model.User, error)
	FindWriteupByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error)
	UpsertWriteup(ctx context.Context, writeup *model.ChallengeWriteup) error
	DeleteWriteupByChallengeID(ctx context.Context, challengeID int64) error
	FindReleasedWriteupByChallengeID(ctx context.Context, challengeID int64, now time.Time) (*model.ChallengeWriteup, error)
	GetSolvedStatus(ctx context.Context, userID, challengeID int64) (bool, error)
	FindSubmissionWriteupByUserChallenge(ctx context.Context, userID, challengeID int64) (*model.SubmissionWriteup, error)
	FindSubmissionWriteupByID(ctx context.Context, id int64) (*model.SubmissionWriteup, error)
	UpsertSubmissionWriteup(ctx context.Context, writeup *model.SubmissionWriteup) error
	GetTeacherSubmissionWriteupByID(ctx context.Context, id int64) (*TeacherSubmissionWriteupRecord, error)
	ListTeacherSubmissionWriteups(ctx context.Context, query *dto.TeacherSubmissionWriteupQuery) ([]TeacherSubmissionWriteupRecord, int64, error)
	ListRecommendedSolutionsByChallengeID(ctx context.Context, challengeID int64, now time.Time) ([]RecommendedSolutionRecord, error)
	ListCommunitySolutionsByChallengeID(ctx context.Context, challengeID int64, query *dto.CommunityChallengeSolutionQuery) ([]CommunitySolutionRecord, int64, error)
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
	FindByID(ctx context.Context, id int64) (*model.Challenge, error)
	FindChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error)
	UpsertChallengeTopology(ctx context.Context, topology *model.ChallengeTopology) error
	DeleteChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) error
}

type ImageRepository interface {
	Create(ctx context.Context, image *model.Image) error
	FindByID(ctx context.Context, id int64) (*model.Image, error)
	FindByNameTag(ctx context.Context, name, tag string) (*model.Image, error)
	List(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error)
	Update(ctx context.Context, image *model.Image) error
	Delete(ctx context.Context, id int64) error
}

type EnvironmentTemplateRepository interface {
	Create(ctx context.Context, template *model.EnvironmentTemplate) error
	Update(ctx context.Context, template *model.EnvironmentTemplate) error
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*model.EnvironmentTemplate, error)
	List(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error)
	IncrementUsage(ctx context.Context, id int64) error
}

type TagRepository interface {
	Create(ctx context.Context, tag *model.Tag) error
	List(ctx context.Context, tagType string) ([]*model.Tag, error)
	FindByIDs(ctx context.Context, ids []int64) ([]*model.Tag, error)
	AttachTagsInTx(ctx context.Context, challengeID int64, tagIDs []int64) error
	DetachFromChallenge(ctx context.Context, challengeID, tagID int64) error
	FindByChallengeID(ctx context.Context, challengeID int64) ([]*model.Tag, error)
	Delete(ctx context.Context, id int64) error
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
