package ports

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type ChallengeCommandRepository interface {
	CreateWithHints(challenge *model.Challenge, hints []*model.ChallengeHint) error
	CreateWithHintsWithContext(ctx context.Context, challenge *model.Challenge, hints []*model.ChallengeHint) error
	FindByID(id int64) (*model.Challenge, error)
	FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error)
	Update(challenge *model.Challenge) error
	UpdateWithContext(ctx context.Context, challenge *model.Challenge) error
	UpdateWithHints(challenge *model.Challenge, hints []*model.ChallengeHint, replaceHints bool) error
	UpdateWithHintsWithContext(ctx context.Context, challenge *model.Challenge, hints []*model.ChallengeHint, replaceHints bool) error
	Delete(id int64) error
	DeleteWithContext(ctx context.Context, id int64) error
	HasRunningInstances(challengeID int64) (bool, error)
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
	FindByID(id int64) (*model.Challenge, error)
	FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error)
	Update(challenge *model.Challenge) error
	UpdateWithContext(ctx context.Context, challenge *model.Challenge) error
}

type ChallengeQueryRepository interface {
	FindByID(id int64) (*model.Challenge, error)
	FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error)
	List(query *dto.ChallengeQuery) ([]*model.Challenge, int64, error)
	ListWithContext(ctx context.Context, query *dto.ChallengeQuery) ([]*model.Challenge, int64, error)
	ListHintsByChallengeID(challengeID int64) ([]*model.ChallengeHint, error)
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
	CreateAWDServiceTemplate(template *model.AWDServiceTemplate) error
	CreateAWDServiceTemplateWithContext(ctx context.Context, template *model.AWDServiceTemplate) error
	FindAWDServiceTemplateByID(id int64) (*model.AWDServiceTemplate, error)
	FindAWDServiceTemplateByIDWithContext(ctx context.Context, id int64) (*model.AWDServiceTemplate, error)
	UpdateAWDServiceTemplate(template *model.AWDServiceTemplate) error
	UpdateAWDServiceTemplateWithContext(ctx context.Context, template *model.AWDServiceTemplate) error
	DeleteAWDServiceTemplate(id int64) error
	DeleteAWDServiceTemplateWithContext(ctx context.Context, id int64) error
}

type AWDServiceTemplateQueryRepository interface {
	FindAWDServiceTemplateByID(id int64) (*model.AWDServiceTemplate, error)
	FindAWDServiceTemplateByIDWithContext(ctx context.Context, id int64) (*model.AWDServiceTemplate, error)
	ListAWDServiceTemplates(query *dto.AWDServiceTemplateQuery) ([]*model.AWDServiceTemplate, int64, error)
	ListAWDServiceTemplatesWithContext(ctx context.Context, query *dto.AWDServiceTemplateQuery) ([]*model.AWDServiceTemplate, int64, error)
}

type ChallengeImageUsageRepository interface {
	CountByImageID(imageID int64) (int64, error)
	CountByImageIDWithContext(ctx context.Context, imageID int64) (int64, error)
}

type ChallengeWriteupRepository interface {
	FindByID(id int64) (*model.Challenge, error)
	FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error)
	FindUserByID(userID int64) (*model.User, error)
	FindUserByIDWithContext(ctx context.Context, userID int64) (*model.User, error)
	FindWriteupByChallengeID(challengeID int64) (*model.ChallengeWriteup, error)
	FindWriteupByChallengeIDWithContext(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error)
	UpsertWriteup(writeup *model.ChallengeWriteup) error
	UpsertWriteupWithContext(ctx context.Context, writeup *model.ChallengeWriteup) error
	DeleteWriteupByChallengeID(challengeID int64) error
	DeleteWriteupByChallengeIDWithContext(ctx context.Context, challengeID int64) error
	FindReleasedWriteupByChallengeID(challengeID int64, now time.Time) (*model.ChallengeWriteup, error)
	FindReleasedWriteupByChallengeIDWithContext(ctx context.Context, challengeID int64, now time.Time) (*model.ChallengeWriteup, error)
	GetSolvedStatus(userID, challengeID int64) (bool, error)
	GetSolvedStatusWithContext(ctx context.Context, userID, challengeID int64) (bool, error)
	FindSubmissionWriteupByUserChallenge(userID, challengeID int64) (*model.SubmissionWriteup, error)
	FindSubmissionWriteupByUserChallengeWithContext(ctx context.Context, userID, challengeID int64) (*model.SubmissionWriteup, error)
	FindSubmissionWriteupByID(id int64) (*model.SubmissionWriteup, error)
	FindSubmissionWriteupByIDWithContext(ctx context.Context, id int64) (*model.SubmissionWriteup, error)
	UpsertSubmissionWriteup(writeup *model.SubmissionWriteup) error
	UpsertSubmissionWriteupWithContext(ctx context.Context, writeup *model.SubmissionWriteup) error
	GetTeacherSubmissionWriteupByID(id int64) (*TeacherSubmissionWriteupRecord, error)
	GetTeacherSubmissionWriteupByIDWithContext(ctx context.Context, id int64) (*TeacherSubmissionWriteupRecord, error)
	ListTeacherSubmissionWriteups(query *dto.TeacherSubmissionWriteupQuery) ([]TeacherSubmissionWriteupRecord, int64, error)
	ListTeacherSubmissionWriteupsWithContext(ctx context.Context, query *dto.TeacherSubmissionWriteupQuery) ([]TeacherSubmissionWriteupRecord, int64, error)
	ListRecommendedSolutionsByChallengeID(challengeID int64, now time.Time) ([]RecommendedSolutionRecord, error)
	ListRecommendedSolutionsByChallengeIDWithContext(ctx context.Context, challengeID int64, now time.Time) ([]RecommendedSolutionRecord, error)
	ListCommunitySolutionsByChallengeID(challengeID int64, query *dto.CommunityChallengeSolutionQuery) ([]CommunitySolutionRecord, int64, error)
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
	FindByID(id int64) (*model.Challenge, error)
	FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error)
	FindChallengeTopologyByChallengeID(challengeID int64) (*model.ChallengeTopology, error)
	FindChallengeTopologyByChallengeIDWithContext(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error)
	UpsertChallengeTopology(topology *model.ChallengeTopology) error
	UpsertChallengeTopologyWithContext(ctx context.Context, topology *model.ChallengeTopology) error
	DeleteChallengeTopologyByChallengeID(challengeID int64) error
	DeleteChallengeTopologyByChallengeIDWithContext(ctx context.Context, challengeID int64) error
}

type ImageRepository interface {
	Create(image *model.Image) error
	CreateWithContext(ctx context.Context, image *model.Image) error
	FindByID(id int64) (*model.Image, error)
	FindByIDWithContext(ctx context.Context, id int64) (*model.Image, error)
	FindByNameTag(name, tag string) (*model.Image, error)
	FindByNameTagWithContext(ctx context.Context, name, tag string) (*model.Image, error)
	List(name, status string, offset, limit int) ([]*model.Image, int64, error)
	ListWithContext(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error)
	Update(image *model.Image) error
	UpdateWithContext(ctx context.Context, image *model.Image) error
	Delete(id int64) error
	DeleteWithContext(ctx context.Context, id int64) error
}

type EnvironmentTemplateRepository interface {
	Create(template *model.EnvironmentTemplate) error
	CreateWithContext(ctx context.Context, template *model.EnvironmentTemplate) error
	Update(template *model.EnvironmentTemplate) error
	UpdateWithContext(ctx context.Context, template *model.EnvironmentTemplate) error
	Delete(id int64) error
	DeleteWithContext(ctx context.Context, id int64) error
	FindByID(id int64) (*model.EnvironmentTemplate, error)
	FindByIDWithContext(ctx context.Context, id int64) (*model.EnvironmentTemplate, error)
	List(keyword string) ([]*model.EnvironmentTemplate, error)
	ListWithContext(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error)
	IncrementUsage(id int64) error
	IncrementUsageWithContext(ctx context.Context, id int64) error
}

type TagRepository interface {
	Create(tag *model.Tag) error
	CreateWithContext(ctx context.Context, tag *model.Tag) error
	List(tagType string) ([]*model.Tag, error)
	ListWithContext(ctx context.Context, tagType string) ([]*model.Tag, error)
	FindByIDs(ids []int64) ([]*model.Tag, error)
	FindByIDsWithContext(ctx context.Context, ids []int64) ([]*model.Tag, error)
	AttachTagsInTx(challengeID int64, tagIDs []int64) error
	AttachTagsInTxWithContext(ctx context.Context, challengeID int64, tagIDs []int64) error
	DetachFromChallenge(challengeID, tagID int64) error
	DetachFromChallengeWithContext(ctx context.Context, challengeID, tagID int64) error
	FindByChallengeID(challengeID int64) ([]*model.Tag, error)
	FindByChallengeIDWithContext(ctx context.Context, challengeID int64) ([]*model.Tag, error)
	Delete(id int64) error
	DeleteWithContext(ctx context.Context, id int64) error
	CountChallengesByTagID(tagID int64) (int64, error)
	CountChallengesByTagIDWithContext(ctx context.Context, tagID int64) (int64, error)
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
