package infrastructure

import (
	"context"
	"time"

	"ctf-platform/internal/model"

	"gorm.io/gorm"
)

type ImageRepository struct {
	db *gorm.DB
}

func NewImageRepository(db *gorm.DB) *ImageRepository {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *ImageRepository) Create(ctx context.Context, image *model.Image) error {
	return r.dbWithContext(ctx).Create(image).Error
}

func (r *ImageRepository) FindByID(ctx context.Context, id int64) (*model.Image, error) {
	var image model.Image
	err := r.dbWithContext(ctx).Where("id = ?", id).First(&image).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *ImageRepository) FindByNameTag(ctx context.Context, name, tag string) (*model.Image, error) {
	var image model.Image
	err := r.dbWithContext(ctx).Where("name = ? AND tag = ?", name, tag).First(&image).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *ImageRepository) List(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error) {
	var images []*model.Image
	var total int64

	query := r.dbWithContext(ctx).Model(&model.Image{}).Where("deleted_at IS NULL")
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&images).Error
	return images, total, err
}

func (r *ImageRepository) Update(ctx context.Context, image *model.Image) error {
	return r.dbWithContext(ctx).Save(image).Error
}

func (r *ImageRepository) Delete(ctx context.Context, id int64) error {
	return r.dbWithContext(ctx).Delete(&model.Image{}, id).Error
}

func (r *ImageRepository) CreateImageBuildJob(ctx context.Context, job *model.ImageBuildJob) error {
	return r.dbWithContext(ctx).Create(job).Error
}

func (r *ImageRepository) FindImageBuildJobByID(ctx context.Context, id int64) (*model.ImageBuildJob, error) {
	var job model.ImageBuildJob
	err := r.dbWithContext(ctx).Where("id = ?", id).First(&job).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *ImageRepository) ListPendingImageBuildJobs(ctx context.Context, limit int) ([]*model.ImageBuildJob, error) {
	if limit <= 0 {
		limit = 1
	}
	var jobs []*model.ImageBuildJob
	err := r.dbWithContext(ctx).
		Where("status = ?", model.ImageBuildJobStatusPending).
		Order("created_at ASC, id ASC").
		Limit(limit).
		Find(&jobs).Error
	return jobs, err
}

func (r *ImageRepository) TryStartImageBuildJob(ctx context.Context, id int64, startedAt time.Time) (bool, error) {
	result := r.dbWithContext(ctx).
		Model(&model.ImageBuildJob{}).
		Where("id = ? AND status = ?", id, model.ImageBuildJobStatusPending).
		Updates(map[string]any{
			"status":     model.ImageBuildJobStatusBuilding,
			"started_at": startedAt,
			"updated_at": startedAt,
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected == 1, nil
}

func (r *ImageRepository) UpdateImageBuildJob(ctx context.Context, job *model.ImageBuildJob) error {
	return r.dbWithContext(ctx).Save(job).Error
}
