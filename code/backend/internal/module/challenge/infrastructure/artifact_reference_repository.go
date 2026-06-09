package infrastructure

import (
	"context"

	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"gorm.io/gorm"
)

type ArtifactReferenceRepository struct {
	db *gorm.DB
}

func NewArtifactReferenceRepository(db *gorm.DB) *ArtifactReferenceRepository {
	return &ArtifactReferenceRepository{db: db}
}

func (r *ArtifactReferenceRepository) ListArtifactReferences(ctx context.Context) (challengeports.ArtifactReferences, error) {
	if r == nil || r.db == nil {
		return challengeports.ArtifactReferences{}, nil
	}

	var refs challengeports.ArtifactReferences
	if err := r.db.WithContext(ctx).
		Model(&challengeentity.Challenge{}).
		Where("deleted_at IS NULL").
		Where("attachment_url IS NOT NULL AND attachment_url <> ''").
		Pluck("attachment_url", &refs.AttachmentURLs).Error; err != nil {
		return challengeports.ArtifactReferences{}, err
	}
	if err := r.db.WithContext(ctx).
		Model(&challengeentity.ImageBuildJob{}).
		Where("source_dir <> ''").
		Where("status IN ?", []string{
			challengeentity.ImageBuildJobStatusPending,
			challengeentity.ImageBuildJobStatusBuilding,
			challengeentity.ImageBuildJobStatusPushed,
			challengeentity.ImageBuildJobStatusVerifying,
		}).
		Pluck("source_dir", &refs.ImageBuildSourceDirs).Error; err != nil {
		return challengeports.ArtifactReferences{}, err
	}
	if err := r.db.WithContext(ctx).
		Model(&challengeentity.AWDChallenge{}).
		Where("deleted_at IS NULL").
		Where("checker_config <> ''").
		Pluck("checker_config", &refs.AWDCheckerConfigs).Error; err != nil {
		return challengeports.ArtifactReferences{}, err
	}
	return refs, nil
}
