package challenge

import (
	"ctf-platform/internal/model"

	"gorm.io/gorm"
)

type ImageRepository struct {
	db *gorm.DB
}

func NewImageRepository(db *gorm.DB) *ImageRepository {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) Create(image *model.Image) error {
	return r.db.Create(image).Error
}

func (r *ImageRepository) FindByID(id int64) (*model.Image, error) {
	var image model.Image
	err := r.db.Where("id = ?", id).First(&image).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *ImageRepository) FindByNameTag(name, tag string) (*model.Image, error) {
	var image model.Image
	err := r.db.Where("name = ? AND tag = ?", name, tag).First(&image).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *ImageRepository) List(name, status string, offset, limit int) ([]*model.Image, int64, error) {
	var images []*model.Image
	var total int64

	query := r.db.Model(&model.Image{})
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

func (r *ImageRepository) Update(image *model.Image) error {
	return r.db.Save(image).Error
}

func (r *ImageRepository) Delete(id int64) error {
	return r.db.Delete(&model.Image{}, id).Error
}
