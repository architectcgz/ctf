package contest

import (
	"ctf-platform/internal/model"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *TeamRepository) CreateWithMember(team *model.Team, captainID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(team).Error; err != nil {
			return err
		}
		member := &model.TeamMember{
			TeamID:   team.ID,
			UserID:   captainID,
			JoinedAt: time.Now(),
		}
		return tx.Create(member).Error
	})
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

func (r *TeamRepository) DeleteWithMembers(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("team_id = ?", id).Delete(&model.TeamMember{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Team{}, id).Error
	})
}

func (r *TeamRepository) AddMember(member *model.TeamMember) error {
	return r.db.Create(member).Error
}

func (r *TeamRepository) AddMemberWithLock(teamID, userID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var team model.Team
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&team, teamID).Error; err != nil {
			return err
		}

		var count int64
		if err := tx.Model(&model.TeamMember{}).Where("team_id = ?", teamID).Count(&count).Error; err != nil {
			return err
		}

		if count >= int64(team.MaxMembers) {
			return gorm.ErrInvalidData
		}

		member := &model.TeamMember{
			TeamID:   teamID,
			UserID:   userID,
			JoinedAt: time.Now(),
		}
		return tx.Create(member).Error
	})
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
		Where("team_members.user_id = ? AND teams.contest_id = ? AND teams.deleted_at IS NULL", userID, contestID).
		First(&team).Error
	return &team, err
}

func (r *TeamRepository) ListByContest(contestID int64) ([]*model.Team, error) {
	var teams []*model.Team
	err := r.db.Where("contest_id = ?", contestID).Order("created_at DESC").Find(&teams).Error
	return teams, err
}

func (r *TeamRepository) GetMemberCountBatch(teamIDs []int64) (map[int64]int, error) {
	type Result struct {
		TeamID int64
		Count  int
	}
	var results []Result
	err := r.db.Model(&model.TeamMember{}).
		Select("team_id, COUNT(*) as count").
		Where("team_id IN ?", teamIDs).
		Group("team_id").
		Scan(&results).Error

	countMap := make(map[int64]int)
	for _, r := range results {
		countMap[r.TeamID] = r.Count
	}
	return countMap, err
}
