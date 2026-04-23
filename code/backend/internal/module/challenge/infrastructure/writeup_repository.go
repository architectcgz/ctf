package infrastructure

import (
	"context"
	"sort"
	"strings"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *Repository) FindWriteupByChallengeID(challengeID int64) (*model.ChallengeWriteup, error) {
	return r.FindWriteupByChallengeIDWithContext(context.Background(), challengeID)
}

func (r *Repository) FindWriteupByChallengeIDWithContext(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var writeup model.ChallengeWriteup
	err := r.db.WithContext(ctx).Where("challenge_id = ?", challengeID).First(&writeup).Error
	if err != nil {
		return nil, err
	}
	return &writeup, nil
}

func (r *Repository) UpsertWriteup(writeup *model.ChallengeWriteup) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "challenge_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "content", "visibility", "created_by", "is_recommended", "recommended_at", "recommended_by", "updated_at"}),
	}).Create(writeup).Error
}

func (r *Repository) DeleteWriteupByChallengeID(challengeID int64) error {
	return r.db.Where("challenge_id = ?", challengeID).Delete(&model.ChallengeWriteup{}).Error
}

func (r *Repository) FindReleasedWriteupByChallengeID(challengeID int64, now time.Time) (*model.ChallengeWriteup, error) {
	return r.FindReleasedWriteupByChallengeIDWithContext(context.Background(), challengeID, now)
}

func (r *Repository) FindReleasedWriteupByChallengeIDWithContext(ctx context.Context, challengeID int64, now time.Time) (*model.ChallengeWriteup, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var writeup model.ChallengeWriteup
	err := r.db.WithContext(ctx).
		Where("challenge_id = ?", challengeID).
		Where("visibility = ?", model.WriteupVisibilityPublic).
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
	return r.FindSubmissionWriteupByUserChallengeWithContext(context.Background(), userID, challengeID)
}

func (r *Repository) FindSubmissionWriteupByUserChallengeWithContext(ctx context.Context, userID, challengeID int64) (*model.SubmissionWriteup, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var writeup model.SubmissionWriteup
	err := r.db.WithContext(ctx).Where("user_id = ? AND challenge_id = ?", userID, challengeID).First(&writeup).Error
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
			"visibility_status",
			"is_recommended",
			"recommended_at",
			"recommended_by",
			"published_at",
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
	VisibilityStatus string
	IsRecommended    bool
	RecommendedAt    *time.Time
	RecommendedBy    *int64
	PublishedAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	StudentUsername  string
	StudentName      string
	StudentNo        string
	ClassName        string
	ChallengeTitle   string
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
			VisibilityStatus: r.VisibilityStatus,
			IsRecommended:    r.IsRecommended,
			RecommendedAt:    r.RecommendedAt,
			RecommendedBy:    r.RecommendedBy,
			PublishedAt:      r.PublishedAt,
			CreatedAt:        r.CreatedAt,
			UpdatedAt:        r.UpdatedAt,
		},
		StudentUsername: r.StudentUsername,
		StudentName:     r.StudentName,
		StudentNo:       r.StudentNo,
		ClassName:       r.ClassName,
		ChallengeTitle:  r.ChallengeTitle,
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

type recommendedSolutionRow struct {
	SourceType    string
	SourceID      int64
	ChallengeID   int64
	Title         string
	Content       string
	AuthorName    string
	IsRecommended bool
	RecommendedAt *time.Time
	UpdatedAt     time.Time
}

func (r recommendedSolutionRow) toRecord() challengeports.RecommendedSolutionRecord {
	return challengeports.RecommendedSolutionRecord{
		SourceType:    r.SourceType,
		SourceID:      r.SourceID,
		ChallengeID:   r.ChallengeID,
		Title:         r.Title,
		Content:       r.Content,
		AuthorName:    r.AuthorName,
		IsRecommended: r.IsRecommended,
		RecommendedAt: r.RecommendedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

func (r *Repository) ListRecommendedSolutionsByChallengeID(challengeID int64, now time.Time) ([]challengeports.RecommendedSolutionRecord, error) {
	rows := make([]recommendedSolutionRow, 0)

	var officialRows []recommendedSolutionRow
	if err := r.db.Table("challenge_writeups AS cw").
		Select(strings.TrimSpace(`
			'official' AS source_type,
			cw.id AS source_id,
			cw.challenge_id,
			cw.title,
			cw.content,
			COALESCE(author.name, author.username, '官方题解') AS author_name,
			cw.is_recommended,
			cw.recommended_at,
			cw.updated_at
		`)).
		Joins("LEFT JOIN users author ON author.id = cw.created_by").
		Where("cw.challenge_id = ? AND cw.is_recommended = ?", challengeID, true).
		Where("cw.visibility = ?", model.WriteupVisibilityPublic).
		Order("cw.recommended_at DESC, cw.updated_at DESC").
		Scan(&officialRows).Error; err != nil {
		return nil, err
	}
	rows = append(rows, officialRows...)

	var communityRows []recommendedSolutionRow
	if err := r.db.Table("submission_writeups AS sw").
		Select(strings.TrimSpace(`
			'community' AS source_type,
			sw.id AS source_id,
			sw.challenge_id,
			sw.title,
			sw.content,
			COALESCE(u.name, u.username) AS author_name,
			sw.is_recommended,
			sw.recommended_at,
			sw.updated_at
		`)).
		Joins("JOIN users u ON u.id = sw.user_id").
		Where("sw.challenge_id = ? AND sw.submission_status = ? AND sw.visibility_status = ? AND sw.is_recommended = ?",
			challengeID,
			model.SubmissionWriteupStatusPublished,
			model.SubmissionWriteupVisibilityVisible,
			true,
		).
		Order("sw.recommended_at DESC, sw.updated_at DESC").
		Scan(&communityRows).Error; err != nil {
		return nil, err
	}
	rows = append(rows, communityRows...)

	sort.Slice(rows, func(i, j int) bool {
		left := rows[i].UpdatedAt
		if rows[i].RecommendedAt != nil {
			left = *rows[i].RecommendedAt
		}
		right := rows[j].UpdatedAt
		if rows[j].RecommendedAt != nil {
			right = *rows[j].RecommendedAt
		}
		return left.After(right)
	})

	items := make([]challengeports.RecommendedSolutionRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toRecord())
	}
	return items, nil
}

type communitySolutionRow struct {
	ID               int64
	UserID           int64
	ChallengeID      int64
	ContestID        *int64
	Title            string
	Content          string
	SubmissionStatus string
	VisibilityStatus string
	IsRecommended    bool
	RecommendedAt    *time.Time
	RecommendedBy    *int64
	PublishedAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	AuthorName       string
	ChallengeTitle   string
}

func (r communitySolutionRow) toRecord() challengeports.CommunitySolutionRecord {
	return challengeports.CommunitySolutionRecord{
		Submission: model.SubmissionWriteup{
			ID:               r.ID,
			UserID:           r.UserID,
			ChallengeID:      r.ChallengeID,
			ContestID:        r.ContestID,
			Title:            r.Title,
			Content:          r.Content,
			SubmissionStatus: r.SubmissionStatus,
			VisibilityStatus: r.VisibilityStatus,
			IsRecommended:    r.IsRecommended,
			RecommendedAt:    r.RecommendedAt,
			RecommendedBy:    r.RecommendedBy,
			PublishedAt:      r.PublishedAt,
			CreatedAt:        r.CreatedAt,
			UpdatedAt:        r.UpdatedAt,
		},
		AuthorName:     r.AuthorName,
		ChallengeID:    r.ChallengeID,
		ChallengeTitle: r.ChallengeTitle,
	}
}

func (r *Repository) ListCommunitySolutionsByChallengeID(challengeID int64, query *dto.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error) {
	base := r.db.Table("submission_writeups AS sw").
		Select(strings.TrimSpace(`
			sw.id,
			sw.user_id,
			sw.challenge_id,
			sw.contest_id,
			sw.title,
			sw.content,
			sw.submission_status,
			sw.visibility_status,
			sw.is_recommended,
			sw.recommended_at,
			sw.recommended_by,
			sw.published_at,
			sw.created_at,
			sw.updated_at,
			COALESCE(u.name, u.username) AS author_name,
			c.title AS challenge_title
		`)).
		Joins("JOIN users u ON u.id = sw.user_id").
		Joins("JOIN challenges c ON c.id = sw.challenge_id").
		Where("sw.challenge_id = ? AND sw.submission_status = ? AND sw.visibility_status = ?",
			challengeID,
			model.SubmissionWriteupStatusPublished,
			model.SubmissionWriteupVisibilityVisible,
		)

	if query != nil && strings.TrimSpace(query.Q) != "" {
		pattern := "%" + strings.TrimSpace(query.Q) + "%"
		base = base.Where("sw.title LIKE ? OR u.username LIKE ? OR u.name LIKE ?", pattern, pattern, pattern)
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
		if query.Sort == "oldest" {
			base = base.Order("sw.published_at ASC, sw.id ASC")
		} else {
			base = base.Order("sw.published_at DESC, sw.id DESC")
		}
	} else {
		base = base.Order("sw.published_at DESC, sw.id DESC")
	}

	offset := (page - 1) * size
	var rows []communitySolutionRow
	if err := base.Offset(offset).Limit(size).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	items := make([]challengeports.CommunitySolutionRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toRecord())
	}
	return items, total, nil
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
			sw.visibility_status,
			sw.is_recommended,
			sw.recommended_at,
			sw.recommended_by,
			sw.published_at,
			sw.created_at,
			sw.updated_at,
			u.username AS student_username,
			COALESCE(u.name, '') AS student_name,
			COALESCE(NULLIF(u.student_no, ''), '') AS student_no,
			COALESCE(u.class_name, '') AS class_name,
			c.title AS challenge_title
		`)).
		Joins("JOIN users u ON u.id = sw.user_id").
		Joins("JOIN challenges c ON c.id = sw.challenge_id")

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
		if query.VisibilityStatus != "" {
			base = base.Where("sw.visibility_status = ?", query.VisibilityStatus)
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
	return r.FindChallengeTopologyByChallengeIDWithContext(context.Background(), challengeID)
}

func (r *Repository) FindChallengeTopologyByChallengeIDWithContext(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var topology model.ChallengeTopology
	err := r.db.WithContext(ctx).Where("challenge_id = ?", challengeID).First(&topology).Error
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
	return r.FindByIDWithContext(context.Background(), id)
}

func (r *TemplateRepository) FindByIDWithContext(ctx context.Context, id int64) (*model.EnvironmentTemplate, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var template model.EnvironmentTemplate
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *TemplateRepository) List(keyword string) ([]*model.EnvironmentTemplate, error) {
	return r.ListWithContext(context.Background(), keyword)
}

func (r *TemplateRepository) ListWithContext(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var templates []*model.EnvironmentTemplate
	db := r.db.WithContext(ctx).Model(&model.EnvironmentTemplate{})
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
