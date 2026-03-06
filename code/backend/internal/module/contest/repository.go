package contest

import (
	"ctf-platform/internal/model"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByID(id int64) (*model.Contest, error) {
	var contest model.Contest
	err := r.db.Where("id = ?", id).First(&contest).Error
	return &contest, err
}

func (r *Repository) Update(contest *model.Contest) error {
	return r.db.Save(contest).Error
}

func (r *Repository) FindTeamByID(teamID int64) (*model.Team, error) {
	var team model.Team
	err := r.db.Where("id = ?", teamID).First(&team).Error
	return &team, err
}

func (r *Repository) FindTeamsByIDs(ids []int64) ([]*model.Team, error) {
	var teams []*model.Team
	err := r.db.Where("id IN ?", ids).Find(&teams).Error
	return teams, err
}
