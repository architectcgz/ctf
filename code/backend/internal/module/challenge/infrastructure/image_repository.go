package infrastructure

import (
	"context"

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

	query := r.dbWithContext(ctx).Model(&model.Image{})
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
