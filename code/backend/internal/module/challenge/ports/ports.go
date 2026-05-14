package ports

import (
	"context"
	"errors"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

var (
	ErrChallengeQueryChallengeNotFound          = errors.New("challenge query challenge not found")
	ErrChallengeCommandChallengeNotFound        = errors.New("challenge command challenge not found")
	ErrChallengePublishCheckJobNotFound         = errors.New("challenge publish check job not found")
	ErrChallengeImageNotFound                   = errors.New("challenge image not found")
	ErrChallengeFlagChallengeNotFound           = errors.New("challenge flag challenge not found")
	ErrAWDChallengeNotFound                     = errors.New("awd challenge not found")
	ErrChallengeTopologyChallengeNotFound       = errors.New("challenge topology challenge not found")
	ErrChallengeTopologyNotFound                = errors.New("challenge topology not found")
	ErrChallengeTopologyTemplateNotFound        = errors.New("challenge topology template not found")
	ErrChallengeTopologyPackageRevisionNotFound = errors.New("challenge topology package revision not found")

	ErrChallengeWriteupChallengeNotFound         = errors.New("challenge writeup challenge not found")
	ErrChallengeWriteupRequesterNotFound         = errors.New("challenge writeup requester not found")
	ErrChallengeOfficialWriteupNotFound          = errors.New("challenge official writeup not found")
	ErrChallengeReleasedWriteupNotFound          = errors.New("challenge released writeup not found")
	ErrChallengeSubmissionWriteupNotFound        = errors.New("challenge submission writeup not found")
	ErrChallengeSubmissionWriteupDetailNotFound  = errors.New("challenge submission writeup detail not found")
	ErrChallengeTeacherSubmissionWriteupNotFound = errors.New("challenge teacher submission writeup not found")
)

type ChallengeWriteRepository interface {
	CreateWithHints(ctx context.Context, challenge *model.Challenge, hints []*model.ChallengeHint) error
	FindByID(ctx context.Context, id int64) (*model.Challenge, error)
	Update(ctx context.Context, challenge *model.Challenge) error
	UpdateWithHints(ctx context.Context, challenge *model.Challenge, hints []*model.ChallengeHint, replaceHints bool) error
	Delete(ctx context.Context, id int64) error
}

type ChallengeInstanceUsageRepository interface {
	HasRunningInstances(ctx context.Context, challengeID int64) (bool, error)
}

type ChallengePublishCheckRepository interface {
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

type ChallengeReadRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Challenge, error)
	List(ctx context.Context, query *dto.ChallengeQuery) ([]*model.Challenge, int64, error)
	ListHintsByChallengeID(ctx context.Context, challengeID int64) ([]*model.ChallengeHint, error)
}

type ChallengePublishedRepository interface {
	ListPublished(ctx context.Context, query *dto.ChallengeQuery) ([]*model.Challenge, int64, error)
}

type ChallengeStatsRepository interface {
	GetSolvedStatus(ctx context.Context, userID, challengeID int64) (bool, error)
	GetSolvedCount(ctx context.Context, challengeID int64) (int64, error)
	GetTotalAttempts(ctx context.Context, challengeID int64) (int64, error)
}

type ChallengeBatchStatsRepository interface {
	BatchGetSolvedStatus(ctx context.Context, userID int64, challengeIDs []int64) (map[int64]bool, error)
	BatchGetSolvedCount(ctx context.Context, challengeIDs []int64) (map[int64]int64, error)
	BatchGetTotalAttempts(ctx context.Context, challengeIDs []int64) (map[int64]int64, error)
}

type ChallengeSolvedCountCache interface {
	GetSolvedCount(ctx context.Context, challengeID int64) (count int64, hit bool, err error)
	StoreSolvedCount(ctx context.Context, challengeID int64, count int64, ttl time.Duration) error
}

type AWDChallengeCommandRepository interface {
	CreateAWDChallenge(ctx context.Context, challenge *model.AWDChallenge) error
	FindAWDChallengeByID(ctx context.Context, id int64) (*model.AWDChallenge, error)
	UpdateAWDChallenge(ctx context.Context, challenge *model.AWDChallenge) error
	DeleteAWDChallenge(ctx context.Context, id int64) error
}

type AWDChallengeQueryRepository interface {
	FindAWDChallengeByID(ctx context.Context, id int64) (*model.AWDChallenge, error)
	ListAWDChallenges(ctx context.Context, query *dto.AWDChallengeQuery) ([]*model.AWDChallenge, int64, error)
}

type ChallengeImageUsageRepository interface {
	CountByImageID(ctx context.Context, imageID int64) (int64, error)
}

type ChallengeWriteupChallengeLookupRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Challenge, error)
}

type ChallengeWriteupUserLookupRepository interface {
	FindUserByID(ctx context.Context, userID int64) (*model.User, error)
}

type ChallengeAdminWriteupRepository interface {
	FindWriteupByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error)
	UpsertWriteup(ctx context.Context, writeup *model.ChallengeWriteup) error
	DeleteWriteupByChallengeID(ctx context.Context, challengeID int64) error
}

type ChallengeReleasedWriteupRepository interface {
	FindReleasedWriteupByChallengeID(ctx context.Context, challengeID int64, now time.Time) (*model.ChallengeWriteup, error)
}

type ChallengeWriteupSolveStatusRepository interface {
	GetSolvedStatus(ctx context.Context, userID, challengeID int64) (bool, error)
}

type ChallengeSubmissionWriteupRepository interface {
	FindSubmissionWriteupByUserChallenge(ctx context.Context, userID, challengeID int64) (*model.SubmissionWriteup, error)
	FindSubmissionWriteupByID(ctx context.Context, id int64) (*model.SubmissionWriteup, error)
	UpsertSubmissionWriteup(ctx context.Context, writeup *model.SubmissionWriteup) error
}

type ChallengeTeacherSubmissionWriteupRepository interface {
	GetTeacherSubmissionWriteupByID(ctx context.Context, id int64) (*TeacherSubmissionWriteupRecord, error)
	ListTeacherSubmissionWriteups(ctx context.Context, query *dto.TeacherSubmissionWriteupQuery) ([]TeacherSubmissionWriteupRecord, int64, error)
}

type ChallengeSolutionQueryRepository interface {
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

type ChallengeTopologyChallengeLookupRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Challenge, error)
}

type ChallengeTopologyReadRepository interface {
	FindChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error)
}

type ChallengeTopologyWriteRepository interface {
	UpsertChallengeTopology(ctx context.Context, topology *model.ChallengeTopology) error
	DeleteChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) error
}

type ChallengePackageRevisionRepository interface {
	CreateChallengePackageRevision(ctx context.Context, revision *model.ChallengePackageRevision) error
	FindChallengePackageRevisionByID(ctx context.Context, id int64) (*model.ChallengePackageRevision, error)
	FindLatestChallengePackageRevisionByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengePackageRevision, error)
	ListChallengePackageRevisionsByChallengeID(ctx context.Context, challengeID int64) ([]*model.ChallengePackageRevision, error)
}

type ImportedPlatformBuildImageRequest struct {
	ChallengeMode  string
	PackageSlug    string
	SuggestedTag   string
	SourceDir      string
	DockerfilePath string
	ContextPath    string
	CreatedBy      int64
}

type ImportedImageResolution struct {
	ImageID  int64
	ImageRef string
}

type ChallengeImportedImageTxStore interface {
	ResolvePlatformBuildImage(ctx context.Context, req ImportedPlatformBuildImageRequest) (*ImportedImageResolution, error)
	ResolveExternalImage(ctx context.Context, packageSlug string, imageRef string) (*ImportedImageResolution, error)
	ResolveExistingImageRef(ctx context.Context, packageSlug string, imageRef string) (*ImportedImageResolution, error)
}

type ChallengeImportTxStore interface {
	ChallengeImportedImageTxStore
	RejectImportedChallengeSlugConflict(ctx context.Context, packageSlug string) error
	FindLegacyChallengeForImportedPackageCreate(ctx context.Context, title string, category string) (*model.Challenge, bool, error)
	CreateImportedChallenge(ctx context.Context, challenge *model.Challenge) error
	UpdateImportedChallenge(ctx context.Context, challenge *model.Challenge, updates map[string]any) error
	ClearPublishCheckJobs(ctx context.Context, challengeID int64) error
	ReplaceImportedHints(ctx context.Context, challengeID int64, hints []model.ChallengeHint) error
	ApplyImportedFlagUpdates(ctx context.Context, challengeID int64, updates map[string]any) error
	NextChallengePackageRevisionNo(ctx context.Context, challengeID int64) (int, error)
	CreateImportedPackageRevision(ctx context.Context, revision *model.ChallengePackageRevision) error
	UpsertImportedTopology(ctx context.Context, topology *model.ChallengeTopology) error
}

type ChallengeImportTxRunner interface {
	WithinChallengeImportTransaction(ctx context.Context, fn func(store ChallengeImportTxStore) error) error
}

type AWDChallengeImportTxStore interface {
	ChallengeImportedImageTxStore
	RejectImportedAWDChallengeSlugConflict(ctx context.Context, slug string) error
	CreateImportedAWDChallenge(ctx context.Context, challenge *model.AWDChallenge) error
}

type AWDChallengeImportTxRunner interface {
	WithinAWDChallengeImportTransaction(ctx context.Context, fn func(store AWDChallengeImportTxStore) error) error
}

type ChallengePackageExportTxStore interface {
	FindChallenge(ctx context.Context, challengeID int64) (*model.Challenge, error)
	FindTopology(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error)
	FindPackageRevisionByID(ctx context.Context, revisionID int64) (*model.ChallengePackageRevision, error)
	NextPackageRevisionNo(ctx context.Context, challengeID int64) (int, error)
	ListChallengeHints(ctx context.Context, challengeID int64) ([]model.ChallengeHint, error)
	FindImageRefByID(ctx context.Context, imageID int64) (string, error)
	CreateExportRevision(ctx context.Context, revision *model.ChallengePackageRevision) error
	MarkTopologyExported(ctx context.Context, topologyID int64, revisionID int64, baselineSpec string, updatedAt time.Time) error
}

type ChallengePackageExportTxRunner interface {
	WithinChallengePackageExportTransaction(ctx context.Context, fn func(store ChallengePackageExportTxStore) error) error
}

type ImageCommandRepository interface {
	Create(ctx context.Context, image *model.Image) error
	FindByID(ctx context.Context, id int64) (*model.Image, error)
	FindByNameTag(ctx context.Context, name, tag string) (*model.Image, error)
	Update(ctx context.Context, image *model.Image) error
	Delete(ctx context.Context, id int64) error
}

type ImageQueryRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Image, error)
	List(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error)
}

type ImageBuildJobRepository interface {
	CreateImageBuildJob(ctx context.Context, job *model.ImageBuildJob) error
	FindImageBuildJobByID(ctx context.Context, id int64) (*model.ImageBuildJob, error)
	ListPendingImageBuildJobs(ctx context.Context, limit int) ([]*model.ImageBuildJob, error)
	TryStartImageBuildJob(ctx context.Context, id int64, startedAt time.Time) (bool, error)
	UpdateImageBuildJob(ctx context.Context, job *model.ImageBuildJob) error
}

type ImageInspectResult struct {
	Size int64
}

type DockerImageBuilder interface {
	Build(ctx context.Context, contextPath, dockerfilePath, localRef string) error
	Tag(ctx context.Context, sourceRef, targetRef string) error
	Push(ctx context.Context, targetRef string) error
	Pull(ctx context.Context, targetRef string) error
	Inspect(ctx context.Context, targetRef string) (ImageInspectResult, error)
}

type RegistryVerifier interface {
	CheckManifest(ctx context.Context, imageRef string) (string, error)
}

type EnvironmentTemplateCommandRepository interface {
	Create(ctx context.Context, template *model.EnvironmentTemplate) error
	Update(ctx context.Context, template *model.EnvironmentTemplate) error
	Delete(ctx context.Context, id int64) error
}

type EnvironmentTemplateQueryRepository interface {
	FindByID(ctx context.Context, id int64) (*model.EnvironmentTemplate, error)
	List(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error)
}

type EnvironmentTemplateUsageRepository interface {
	IncrementUsage(ctx context.Context, id int64) error
}

type TagCommandRepository interface {
	Create(ctx context.Context, tag *model.Tag) error
	FindByIDs(ctx context.Context, ids []int64) ([]*model.Tag, error)
	AttachTagsInTx(ctx context.Context, challengeID int64, tagIDs []int64) error
	DetachFromChallenge(ctx context.Context, challengeID, tagID int64) error
	Delete(ctx context.Context, id int64) error
	CountChallengesByTagID(ctx context.Context, tagID int64) (int64, error)
}

type TagQueryRepository interface {
	List(ctx context.Context, tagType string) ([]*model.Tag, error)
	FindByChallengeID(ctx context.Context, challengeID int64) ([]*model.Tag, error)
}

type ImageRuntime interface {
	InspectImageSize(ctx context.Context, imageRef string) (int64, error)
	RemoveImage(ctx context.Context, imageRef string) error
}

type RuntimeTopologyCreateNode struct {
	Key             string
	Image           string
	Env             map[string]string
	ServicePort     int
	ServiceProtocol string
	IsEntryPoint    bool
	NetworkKeys     []string
	Resources       *model.ResourceLimits
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
