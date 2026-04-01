package infrastructure

import (
	"strings"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *Repository) FindWriteupByChallengeID(challengeID int64) (*model.ChallengeWriteup, error) {
	var writeup model.ChallengeWriteup
	err := r.db.Where("challenge_id = ?", challengeID).First(&writeup).Error
	if err != nil {
		return nil, err
	}
	return &writeup, nil
}

func (r *Repository) UpsertWriteup(writeup *model.ChallengeWriteup) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "challenge_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "content", "visibility", "release_at", "created_by", "updated_at"}),
	}).Create(writeup).Error
}

func (r *Repository) DeleteWriteupByChallengeID(challengeID int64) error {
	return r.db.Where("challenge_id = ?", challengeID).Delete(&model.ChallengeWriteup{}).Error
}

func (r *Repository) FindReleasedWriteupByChallengeID(challengeID int64, now time.Time) (*model.ChallengeWriteup, error) {
	var writeup model.ChallengeWriteup
	err := r.db.
		Where("challenge_id = ?", challengeID).
		Where("visibility = ? OR (visibility = ? AND release_at IS NOT NULL AND release_at <= ?)",
			model.WriteupVisibilityPublic,
			model.WriteupVisibilityScheduled,
			now,
		).
		First(&writeup).Error
	if err != nil {
		return nil, err
	}
	return &writeup, nil
}

func (r *Repository) FindUserByID(userID int64) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) FindSubmissionWriteupByUserChallenge(userID, challengeID int64) (*model.SubmissionWriteup, error) {
	var writeup model.SubmissionWriteup
	err := r.db.Where("user_id = ? AND challenge_id = ?", userID, challengeID).First(&writeup).Error
	if err != nil {
		return nil, err
	}
	return &writeup, nil
}

func (r *Repository) FindSubmissionWriteupByID(id int64) (*model.SubmissionWriteup, error) {
	var writeup model.SubmissionWriteup
	err := r.db.Where("id = ?", id).First(&writeup).Error
	if err != nil {
		return nil, err
	}
	return &writeup, nil
}

func (r *Repository) UpsertSubmissionWriteup(writeup *model.SubmissionWriteup) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "challenge_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"contest_id",
			"title",
			"content",
			"submission_status",
			"review_status",
			"submitted_at",
			"reviewed_by",
			"reviewed_at",
			"review_comment",
			"updated_at",
		}),
	}).Create(writeup).Error
}

type teacherSubmissionWriteupRow struct {
	ID               int64
	UserID           int64
	ChallengeID      int64
	ContestID        *int64
	Title            string
	Content          string
	SubmissionStatus string
	ReviewStatus     string
	SubmittedAt      *time.Time
	ReviewedBy       *int64
	ReviewedAt       *time.Time
	ReviewComment    string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	StudentUsername  string
	StudentName      string
	ClassName        string
	ChallengeTitle   string
	ReviewerName     string
}

func (r teacherSubmissionWriteupRow) toRecord() challengeports.TeacherSubmissionWriteupRecord {
	return challengeports.TeacherSubmissionWriteupRecord{
		Submission: model.SubmissionWriteup{
			ID:               r.ID,
			UserID:           r.UserID,
			ChallengeID:      r.ChallengeID,
			ContestID:        r.ContestID,
			Title:            r.Title,
			Content:          r.Content,
			SubmissionStatus: r.SubmissionStatus,
			ReviewStatus:     r.ReviewStatus,
			SubmittedAt:      r.SubmittedAt,
			ReviewedBy:       r.ReviewedBy,
			ReviewedAt:       r.ReviewedAt,
			ReviewComment:    r.ReviewComment,
			CreatedAt:        r.CreatedAt,
			UpdatedAt:        r.UpdatedAt,
		},
		StudentUsername: r.StudentUsername,
		StudentName:     r.StudentName,
		ClassName:       r.ClassName,
		ChallengeTitle:  r.ChallengeTitle,
		ReviewerName:    r.ReviewerName,
	}
}

func (r *Repository) GetTeacherSubmissionWriteupByID(id int64) (*challengeports.TeacherSubmissionWriteupRecord, error) {
	rows, _, err := r.listTeacherSubmissionWriteups(&dto.TeacherSubmissionWriteupQuery{
		Page: 1,
		Size: 1,
	}, func(db *gorm.DB) *gorm.DB {
		return db.Where("sw.id = ?", id)
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	record := rows[0]
	return &record, nil
}

func (r *Repository) ListTeacherSubmissionWriteups(query *dto.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error) {
	return r.listTeacherSubmissionWriteups(query, nil)
}

func (r *Repository) listTeacherSubmissionWriteups(
	query *dto.TeacherSubmissionWriteupQuery,
	extra func(db *gorm.DB) *gorm.DB,
) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error) {
	base := r.db.Table("submission_writeups AS sw").
		Select(strings.TrimSpace(`
			sw.id,
			sw.user_id,
			sw.challenge_id,
			sw.contest_id,
			sw.title,
			sw.content,
			sw.submission_status,
			sw.review_status,
			sw.submitted_at,
			sw.reviewed_by,
			sw.reviewed_at,
			sw.review_comment,
			sw.created_at,
			sw.updated_at,
			u.username AS student_username,
			COALESCE(u.name, '') AS student_name,
			COALESCE(u.class_name, '') AS class_name,
			c.title AS challenge_title,
			COALESCE(reviewer.name, reviewer.username, '') AS reviewer_name
		`)).
		Joins("JOIN users u ON u.id = sw.user_id").
		Joins("JOIN challenges c ON c.id = sw.challenge_id").
		Joins("LEFT JOIN users reviewer ON reviewer.id = sw.reviewed_by")

	if query != nil {
		if query.StudentID != nil {
			base = base.Where("sw.user_id = ?", *query.StudentID)
		}
		if query.ChallengeID != nil {
			base = base.Where("sw.challenge_id = ?", *query.ChallengeID)
		}
		if query.ClassName != "" {
			base = base.Where("u.class_name = ?", query.ClassName)
		}
		if query.SubmissionStatus != "" {
			base = base.Where("sw.submission_status = ?", query.SubmissionStatus)
		}
		if query.ReviewStatus != "" {
			base = base.Where("sw.review_status = ?", query.ReviewStatus)
		}
	}
	if extra != nil {
		base = extra(base)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := 1
	size := 20
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.Size > 0 {
			size = query.Size
		}
	}
	offset := (page - 1) * size

	var rows []teacherSubmissionWriteupRow
	if err := base.Order("sw.updated_at DESC, sw.id DESC").Offset(offset).Limit(size).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	items := make([]challengeports.TeacherSubmissionWriteupRecord, 0, len(rows))
	for _, item := range rows {
		items = append(items, item.toRecord())
	}
	return items, total, nil
}

func (r *Repository) FindChallengeTopologyByChallengeID(challengeID int64) (*model.ChallengeTopology, error) {
	var topology model.ChallengeTopology
	err := r.db.Where("challenge_id = ?", challengeID).First(&topology).Error
	if err != nil {
		return nil, err
	}
	return &topology, nil
}

func (r *Repository) UpsertChallengeTopology(topology *model.ChallengeTopology) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "challenge_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"template_id", "entry_node_key", "spec", "updated_at", "deleted_at"}),
	}).Create(topology).Error
}

func (r *Repository) DeleteChallengeTopologyByChallengeID(challengeID int64) error {
	return r.db.Where("challenge_id = ?", challengeID).Delete(&model.ChallengeTopology{}).Error
}

type TemplateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) Create(template *model.EnvironmentTemplate) error {
	return r.db.Create(template).Error
}

func (r *TemplateRepository) Update(template *model.EnvironmentTemplate) error {
	return r.db.Save(template).Error
}

func (r *TemplateRepository) Delete(id int64) error {
	return r.db.Delete(&model.EnvironmentTemplate{}, id).Error
}

func (r *TemplateRepository) FindByID(id int64) (*model.EnvironmentTemplate, error) {
	var template model.EnvironmentTemplate
	err := r.db.Where("id = ?", id).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *TemplateRepository) List(keyword string) ([]*model.EnvironmentTemplate, error) {
	var templates []*model.EnvironmentTemplate
	db := r.db.Model(&model.EnvironmentTemplate{})
	if keyword != "" {
		pattern := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR description LIKE ?", pattern, pattern)
	}
	err := db.Order("updated_at DESC").Find(&templates).Error
	return templates, err
}

func (r *TemplateRepository) IncrementUsage(id int64) error {
	return r.db.Model(&model.EnvironmentTemplate{}).
		Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}
