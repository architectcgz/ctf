package contest

import (
	"ctf-platform/internal/model"

	"gorm.io/gorm"
)

type TeamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(team *model.Team) error {
	return r.db.Create(team).Error
}

func (r *TeamRepository) FindByID(id int64) (*model.Team, error) {
	var team model.Team
	err := r.db.Where("id = ?", id).First(&team).Error
	return &team, err
}

func (r *TeamRepository) FindByInviteCode(code string) (*model.Team, error) {
	var team model.Team
	err := r.db.Where("invite_code = ?", code).First(&team).Error
	return &team, err
}

func (r *TeamRepository) Delete(id int64) error {
	return r.db.Delete(&model.Team{}, id).Error
}

func (r *TeamRepository) AddMember(member *model.TeamMember) error {
	return r.db.Create(member).Error
}

func (r *TeamRepository) RemoveMember(teamID, userID int64) error {
	return r.db.Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&model.TeamMember{}).Error
}

func (r *TeamRepository) GetMembers(teamID int64) ([]*model.TeamMember, error) {
	var members []*model.TeamMember
	err := r.db.Where("team_id = ?", teamID).Order("joined_at ASC").Find(&members).Error
	return members, err
}

func (r *TeamRepository) GetMemberCount(teamID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.TeamMember{}).Where("team_id = ?", teamID).Count(&count).Error
	return count, err
}

func (r *TeamRepository) FindUserTeamInContest(userID, contestID int64) (*model.Team, error) {
	var team model.Team
	err := r.db.Joins("JOIN team_members ON teams.id = team_members.team_id").
		Where("team_members.user_id = ? AND teams.contest_id = ?", userID, contestID).
		First(&team).Error
	return &team, err
}

func (r *TeamRepository) ListByContest(contestID int64) ([]*model.Team, error) {
	var teams []*model.Team
	err := r.db.Where("contest_id = ?", contestID).Order("created_at DESC").Find(&teams).Error
	return teams, err
}
