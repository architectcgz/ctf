package contest

import (
	"ctf-platform/internal/model"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrTeamFull = errors.New("team is full")
var ErrAlreadyJoinedContest = errors.New("user already joined another team in contest")

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
			ContestID: team.ContestID,
			TeamID:    team.ID,
			UserID:    captainID,
			JoinedAt:  time.Now(),
		}
		if err := tx.Create(member).Error; err != nil {
			return err
		}
		return bindContestRegistrationTeam(tx, team.ContestID, captainID, &team.ID)
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
		var team model.Team
		if err := tx.Where("id = ?", id).First(&team).Error; err != nil {
			return err
		}

		var userIDs []int64
		if err := tx.Model(&model.TeamMember{}).Where("team_id = ?", id).Pluck("user_id", &userIDs).Error; err != nil {
			return err
		}
		if err := tx.Where("team_id = ?", id).Delete(&model.TeamMember{}).Error; err != nil {
			return err
		}
		if len(userIDs) > 0 {
			if err := tx.Model(&model.ContestRegistration{}).
				Where("contest_id = ? AND user_id IN ?", team.ContestID, userIDs).
				Updates(map[string]any{
					"team_id":    nil,
					"updated_at": time.Now(),
				}).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&model.Team{}, id).Error
	})
}

func (r *TeamRepository) AddMember(member *model.TeamMember) error {
	return r.db.Create(member).Error
}

func (r *TeamRepository) AddMemberWithLock(contestID, teamID, userID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var team model.Team
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND contest_id = ?", teamID, contestID).
			First(&team).Error; err != nil {
			return err
		}

		var existingCount int64
		if err := tx.Model(&model.TeamMember{}).
			Where("contest_id = ? AND user_id = ?", contestID, userID).
			Count(&existingCount).Error; err != nil {
			return err
		}
		if existingCount > 0 {
			return ErrAlreadyJoinedContest
		}

		var memberCount int64
		if err := tx.Model(&model.TeamMember{}).Where("team_id = ?", teamID).Count(&memberCount).Error; err != nil {
			return err
		}
		if memberCount >= int64(team.MaxMembers) {
			return ErrTeamFull
		}

		member := &model.TeamMember{
			ContestID: contestID,
			TeamID:    teamID,
			UserID:    userID,
			JoinedAt:  time.Now(),
		}
		if err := tx.Create(member).Error; err != nil {
			return err
		}
		return bindContestRegistrationTeam(tx, contestID, userID, &teamID)
	})
}

func (r *TeamRepository) RemoveMember(teamID, userID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var team model.Team
		if err := tx.Where("id = ?", teamID).First(&team).Error; err != nil {
			return err
		}
		if err := tx.Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&model.TeamMember{}).Error; err != nil {
			return err
		}
		return bindContestRegistrationTeam(tx, team.ContestID, userID, nil)
	})
}

func (r *TeamRepository) FindContestRegistration(contestID, userID int64) (*model.ContestRegistration, error) {
	var registration model.ContestRegistration
	err := r.db.Where("contest_id = ? AND user_id = ?", contestID, userID).First(&registration).Error
	return &registration, err
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

func (r *TeamRepository) FindUsersByIDs(ids []int64) ([]*model.User, error) {
	if len(ids) == 0 {
		return []*model.User{}, nil
	}

	var users []*model.User
	err := r.db.Where("id IN ?", ids).Find(&users).Error
	return users, err
}

func IsUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, constraint)
	}
	return false
}

func bindContestRegistrationTeam(tx *gorm.DB, contestID, userID int64, teamID *int64) error {
	result := tx.Model(&model.ContestRegistration{}).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		Updates(map[string]any{
			"team_id":    teamID,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
