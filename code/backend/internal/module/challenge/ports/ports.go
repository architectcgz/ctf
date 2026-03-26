package ports

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type ChallengeCommandRepository interface {
	CreateWithHints(challenge *model.Challenge, hints []*model.ChallengeHint) error
	FindByID(id int64) (*model.Challenge, error)
	Update(challenge *model.Challenge) error
	UpdateWithHints(challenge *model.Challenge, hints []*model.ChallengeHint, replaceHints bool) error
	Delete(id int64) error
	HasRunningInstances(challengeID int64) (bool, error)
}

type ChallengeFlagRepository interface {
	FindByID(id int64) (*model.Challenge, error)
	Update(challenge *model.Challenge) error
}

type ChallengeQueryRepository interface {
	FindByID(id int64) (*model.Challenge, error)
	FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error)
	List(query *dto.ChallengeQuery) ([]*model.Challenge, int64, error)
	ListHintsByChallengeID(challengeID int64) ([]*model.ChallengeHint, error)
	ListHintsByChallengeIDWithContext(ctx context.Context, challengeID int64) ([]*model.ChallengeHint, error)
	GetUnlockedHintIDsWithContext(ctx context.Context, userID, challengeID int64) (map[int64]bool, error)
	GetSolvedStatusWithContext(ctx context.Context, userID, challengeID int64) (bool, error)
	GetSolvedCountWithContext(ctx context.Context, challengeID int64) (int64, error)
	GetTotalAttemptsWithContext(ctx context.Context, challengeID int64) (int64, error)
	BatchGetSolvedStatusWithContext(ctx context.Context, userID int64, challengeIDs []int64) (map[int64]bool, error)
	BatchGetSolvedCountWithContext(ctx context.Context, challengeIDs []int64) (map[int64]int64, error)
	BatchGetTotalAttemptsWithContext(ctx context.Context, challengeIDs []int64) (map[int64]int64, error)
	ListPublishedWithContext(ctx context.Context, query *dto.ChallengeQuery) ([]*model.Challenge, int64, error)
}

type ChallengeImageUsageRepository interface {
	CountByImageID(imageID int64) (int64, error)
}

type ChallengeWriteupRepository interface {
	FindByID(id int64) (*model.Challenge, error)
	FindWriteupByChallengeID(challengeID int64) (*model.ChallengeWriteup, error)
	UpsertWriteup(writeup *model.ChallengeWriteup) error
	DeleteWriteupByChallengeID(challengeID int64) error
	FindReleasedWriteupByChallengeID(challengeID int64, now time.Time) (*model.ChallengeWriteup, error)
	GetSolvedStatus(userID, challengeID int64) (bool, error)
}

type ChallengeTopologyRepository interface {
	FindByID(id int64) (*model.Challenge, error)
	FindChallengeTopologyByChallengeID(challengeID int64) (*model.ChallengeTopology, error)
	UpsertChallengeTopology(topology *model.ChallengeTopology) error
	DeleteChallengeTopologyByChallengeID(challengeID int64) error
}

type ImageRepository interface {
	Create(image *model.Image) error
	FindByID(id int64) (*model.Image, error)
	FindByNameTag(name, tag string) (*model.Image, error)
	List(name, status string, offset, limit int) ([]*model.Image, int64, error)
	Update(image *model.Image) error
	Delete(id int64) error
}

type EnvironmentTemplateRepository interface {
	Create(template *model.EnvironmentTemplate) error
	Update(template *model.EnvironmentTemplate) error
	Delete(id int64) error
	FindByID(id int64) (*model.EnvironmentTemplate, error)
	List(keyword string) ([]*model.EnvironmentTemplate, error)
	IncrementUsage(id int64) error
}

type TagRepository interface {
	Create(tag *model.Tag) error
	List(tagType string) ([]*model.Tag, error)
	FindByIDs(ids []int64) ([]*model.Tag, error)
	AttachTagsInTx(challengeID int64, tagIDs []int64) error
	DetachFromChallenge(challengeID, tagID int64) error
	FindByChallengeID(challengeID int64) ([]*model.Tag, error)
	Delete(id int64) error
	CountChallengesByTagID(tagID int64) (int64, error)
}

type ImageRuntime interface {
	InspectImageSize(ctx context.Context, imageRef string) (int64, error)
	RemoveImage(ctx context.Context, imageRef string) error
}
