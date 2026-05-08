package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/infrastructure/postgres"
	infraredis "ctf-platform/internal/infrastructure/redis"
	"ctf-platform/internal/model"
	assessmentcmd "ctf-platform/internal/module/assessment/application/commands"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmentinfra "ctf-platform/internal/module/assessment/infrastructure"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	readmodelqueries "ctf-platform/internal/module/teaching_readmodel/application/queries"
	readmodelinfra "ctf-platform/internal/module/teaching_readmodel/infrastructure"
	rediskeys "ctf-platform/internal/pkg/redis"
)

const (
	seedClassName       = "信安2401"
	seedTeacherUsername = "zhaoxiaofeng"
	defaultPassword     = "Password123"
	seedUserAgent       = "seed-teaching-review-data/1.0"
)

type userSeed struct {
	Username  string
	Name      string
	Email     string
	Role      string
	ClassName string
	StudentNo string
	TeacherNo string
}

type challengeRef struct {
	ID         int64
	Title      string
	Category   string
	Difficulty string
	Points     int
	FlagType   string
}

type challengeCatalog struct {
	byCategory map[string][]challengeRef
}

type proxySeed struct {
	Offset         time.Duration
	Method         string
	Path           string
	Query          string
	Status         int
	PayloadPreview string
}

type submissionSeed struct {
	Offset  time.Duration
	Flag    string
	Correct bool
}

type writeupSeed struct {
	Offset      time.Duration
	Title       string
	Content     string
	Published   bool
	Recommended bool
}

type sessionSeed struct {
	ChallengeCategory string
	ChallengeIndex    int
	StartOffset       time.Duration
	Duration          time.Duration
	Access            bool
	ProxyRequests     []proxySeed
	Submissions       []submissionSeed
	Writeup           *writeupSeed
}

type studentScenario struct {
	User     userSeed
	Profiles map[string]float64
	Sessions []sessionSeed
}

type seededStudentResult struct {
	User            *model.User
	Recommendations dto.TeacherRecommendationResp
	Archive         *assessmentcmd.ReviewArchiveData
}

type seedResult struct {
	ClassName   string
	Teacher     *model.User
	Students    []seededStudentResult
	ClassReview *dto.TeacherClassReviewResp
}

func main() {
	env := strings.TrimSpace(os.Getenv("APP_ENV"))
	if env == "" {
		env = "dev"
	}
	if env == "dev" {
		setDefaultEnv("CTF_POSTGRES_PORT", "15432")
		setDefaultEnv("CTF_POSTGRES_PASSWORD", "postgres123456")
		setDefaultEnv("CTF_REDIS_ADDR", "127.0.0.1:16379")
		setDefaultEnv("CTF_REDIS_PASSWORD", "redis123456")
	}
	if strings.TrimSpace(os.Getenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET")) == "" {
		if err := os.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET", "dev-integration-secret-123456789"); err != nil {
			panic(fmt.Errorf("set default CTF_CONTAINER_FLAG_GLOBAL_SECRET: %w", err))
		}
	}

	cfg, err := config.Load(env)
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	ctx := context.Background()
	db, err := postgres.Open(ctx, cfg.Postgres)
	if err != nil {
		panic(fmt.Errorf("open postgres: %w", err))
	}

	cache, cacheErr := infraredis.NewClient(ctx, cfg.Redis)
	if cacheErr != nil {
		fmt.Fprintf(os.Stderr, "warn: redis unavailable, recommendation cache will not be cleared: %v\n", cacheErr)
		cache = nil
	}
	if cache != nil {
		defer func() {
			_ = cache.Close()
		}()
	}

	result, err := seedTeachingReviewData(ctx, db, cache, cfg)
	if err != nil {
		panic(err)
	}

	printSeedSummary(result)
}

func seedTeachingReviewData(ctx context.Context, db *gorm.DB, cache *redislib.Client, cfg *config.Config) (*seedResult, error) {
	if ctx == nil {
		return nil, errors.New("seed teaching review data requires context")
	}
	if cfg == nil {
		return nil, errors.New("seed teaching review data requires config")
	}

	catalog, err := loadChallengeCatalog(ctx, db)
	if err != nil {
		return nil, err
	}
	scenarios := buildStudentScenarios()

	teacherSpec := userSeed{
		Username:  seedTeacherUsername,
		Name:      "赵晓峰",
		Email:     "zhaoxiaofeng@xinan.example.edu.cn",
		Role:      model.RoleTeacher,
		ClassName: seedClassName,
		TeacherNo: "T20264001",
	}

	var teacher *model.User
	studentsByUsername := make(map[string]*model.User, len(scenarios))
	seededUserIDs := make([]int64, 0, len(scenarios)+1)

	err = db.Transaction(func(tx *gorm.DB) error {
		teacher, err = upsertUser(tx, teacherSpec)
		if err != nil {
			return err
		}
		seededUserIDs = append(seededUserIDs, teacher.ID)

		for _, scenario := range scenarios {
			student, upsertErr := upsertUser(tx, scenario.User)
			if upsertErr != nil {
				return upsertErr
			}
			studentsByUsername[scenario.User.Username] = student
			seededUserIDs = append(seededUserIDs, student.ID)
		}

		if err := resetSeededData(tx, seededUserIDs, teacher.ID); err != nil {
			return err
		}

		for _, scenario := range scenarios {
			student := studentsByUsername[scenario.User.Username]
			if student == nil {
				return fmt.Errorf("student not found after upsert: %s", scenario.User.Username)
			}
			if err := seedStudentScenario(tx, teacher, student, scenario, catalog); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if cache != nil {
		cacheKeys := make([]string, 0, len(scenarios))
		for _, scenario := range scenarios {
			student := studentsByUsername[scenario.User.Username]
			if student == nil {
				continue
			}
			cacheKeys = append(cacheKeys, rediskeys.RecommendationKey(student.ID))
		}
		if len(cacheKeys) > 0 {
			if cacheErr := cache.Del(ctx, cacheKeys...).Err(); cacheErr != nil {
				fmt.Fprintf(os.Stderr, "warn: clear recommendation cache failed: %v\n", cacheErr)
			}
		}
	}

	assessmentRepo := assessmentinfra.NewRepository(db)
	reportRepo := assessmentinfra.NewReportRepository(db)
	profileReader := assessmentqry.NewProfileService(assessmentRepo)
	recommendationService := assessmentqry.NewRecommendationService(
		assessmentRepo,
		challengeinfra.NewRepository(db),
		cache,
		cfg.Recommendation,
		zap.NewNop(),
	)
	reportService := assessmentcmd.NewReportService(
		reportRepo,
		reportRepo,
		reportRepo,
		reportRepo,
		reportRepo,
		reportRepo,
		reportRepo,
		profileReader,
		cfg.Report,
		zap.NewNop(),
	)
	readmodelService := readmodelqueries.NewQueryService(
		readmodelinfra.NewRepository(db),
		recommendationService,
		cfg.Pagination,
		zap.NewNop(),
	)

	classReview, err := readmodelService.GetClassReview(ctx, teacher.ID, model.RoleTeacher, seedClassName)
	if err != nil {
		return nil, fmt.Errorf("load class review: %w", err)
	}

	results := make([]seededStudentResult, 0, len(scenarios))
	sort.SliceStable(scenarios, func(i, j int) bool {
		return scenarios[i].User.StudentNo < scenarios[j].User.StudentNo
	})
	for _, scenario := range scenarios {
		student := studentsByUsername[scenario.User.Username]
		if student == nil {
			continue
		}
		recommendations, recErr := readmodelService.GetStudentRecommendations(ctx, teacher.ID, model.RoleTeacher, student.ID, 3)
		if recErr != nil {
			return nil, fmt.Errorf("load recommendations for %s: %w", student.Username, recErr)
		}
		studentRecommendations := dto.TeacherRecommendationResp{}
		if recommendations != nil {
			studentRecommendations = *recommendations
		}
		archive, archiveErr := reportService.GetStudentReviewArchive(ctx, teacher.ID, student.ID)
		if archiveErr != nil {
			return nil, fmt.Errorf("load review archive for %s: %w", student.Username, archiveErr)
		}
		results = append(results, seededStudentResult{
			User:            student,
			Recommendations: studentRecommendations,
			Archive:         archive,
		})
	}

	return &seedResult{
		ClassName:   seedClassName,
		Teacher:     teacher,
		Students:    results,
		ClassReview: classReview,
	}, nil
}

func loadChallengeCatalog(ctx context.Context, db *gorm.DB) (*challengeCatalog, error) {
	var rows []model.Challenge
	err := db.WithContext(ctx).
		Where("status = ?", model.ChallengeStatusPublished).
		Order(`
			CASE difficulty
				WHEN 'beginner' THEN 1
				WHEN 'easy' THEN 2
				WHEN 'medium' THEN 3
				WHEN 'hard' THEN 4
				WHEN 'insane' THEN 5
				ELSE 6
			END ASC
		`).
		Order("points ASC").
		Order("created_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("load published challenges: %w", err)
	}
	if len(rows) == 0 {
		return nil, errors.New("no published challenges found")
	}

	catalog := &challengeCatalog{byCategory: make(map[string][]challengeRef)}
	for _, row := range rows {
		category := strings.ToLower(strings.TrimSpace(row.Category))
		if category == "" {
			continue
		}
		catalog.byCategory[category] = append(catalog.byCategory[category], challengeRef{
			ID:         row.ID,
			Title:      row.Title,
			Category:   category,
			Difficulty: row.Difficulty,
			Points:     row.Points,
			FlagType:   row.FlagType,
		})
	}

	required := []string{"web", "crypto", "forensics", "misc", "pwn", "reverse"}
	for _, category := range required {
		if len(catalog.byCategory[category]) == 0 {
			return nil, fmt.Errorf("required published challenge category missing: %s", category)
		}
	}
	return catalog, nil
}

func buildStudentScenarios() []studentScenario {
	return []studentScenario{
		{
			User: userSeed{
				Username:  "linchenxi",
				Name:      "林宸熙",
				Email:     "2024310101@xinan.example.edu.cn",
				Role:      model.RoleStudent,
				ClassName: seedClassName,
				StudentNo: "2024310101",
			},
			Profiles: map[string]float64{
				"web":       0.82,
				"crypto":    0.47,
				"forensics": 0.44,
				"misc":      0.71,
				"pwn":       0.26,
				"reverse":   0.42,
			},
			Sessions: []sessionSeed{
				{
					ChallengeCategory: "web",
					ChallengeIndex:    0,
					StartOffset:       -6*24*time.Hour + 2*time.Hour,
					Duration:          85 * time.Minute,
					Access:            true,
					ProxyRequests: []proxySeed{
						{Offset: 8 * time.Minute, Method: "GET", Path: "/", Status: 200},
						{Offset: 17 * time.Minute, Method: "GET", Path: "/assets/app.js", Status: 200},
						{Offset: 31 * time.Minute, Method: "POST", Path: "/download", Status: 200, PayloadPreview: "ticket=notes"},
					},
					Submissions: []submissionSeed{
						{Offset: 55 * time.Minute, Flag: "flag{web-note-ticket}", Correct: true},
					},
					Writeup: &writeupSeed{
						Offset:      80 * time.Minute,
						Title:       "从请求头和下载参数定位内部笔记入口",
						Content:     "先确认页面入口，再通过下载参数和返回包定位真正的敏感文件路径，最后整理成可复盘步骤。",
						Published:   true,
						Recommended: true,
					},
				},
				{
					ChallengeCategory: "misc",
					ChallengeIndex:    0,
					StartOffset:       -5*24*time.Hour + 90*time.Minute,
					Duration:          40 * time.Minute,
					Access:            true,
					ProxyRequests: []proxySeed{
						{Offset: 5 * time.Minute, Method: "GET", Path: "/hint", Status: 200},
					},
					Submissions: []submissionSeed{
						{Offset: 24 * time.Minute, Flag: "flag{comment-sticky-note}", Correct: true},
					},
				},
			},
		},
		{
			User: userSeed{
				Username:  "zhangyuchen",
				Name:      "张雨辰",
				Email:     "2024310102@xinan.example.edu.cn",
				Role:      model.RoleStudent,
				ClassName: seedClassName,
				StudentNo: "2024310102",
			},
			Profiles: map[string]float64{
				"web":       0.55,
				"crypto":    0.22,
				"forensics": 0.48,
				"misc":      0.52,
				"pwn":       0.45,
				"reverse":   0.51,
			},
			Sessions: []sessionSeed{
				{
					ChallengeCategory: "crypto",
					ChallengeIndex:    0,
					StartOffset:       -4*24*time.Hour + 75*time.Minute,
					Duration:          110 * time.Minute,
					Access:            true,
					ProxyRequests: []proxySeed{
						{Offset: 6 * time.Minute, Method: "GET", Path: "/cipher.txt", Status: 200},
						{Offset: 19 * time.Minute, Method: "POST", Path: "/decode", Status: 400, PayloadPreview: "rotation=11"},
						{Offset: 48 * time.Minute, Method: "POST", Path: "/decode", Status: 200, PayloadPreview: "rotation=3"},
					},
					Submissions: []submissionSeed{
						{Offset: 33 * time.Minute, Flag: "flag{shift-mail-11}", Correct: false},
						{Offset: 61 * time.Minute, Flag: "flag{wrong-frequency}", Correct: false},
						{Offset: 96 * time.Minute, Flag: "flag{crypto-postcard-shift-3}", Correct: true},
					},
				},
			},
		},
		{
			User: userSeed{
				Username:  "wangzihan",
				Name:      "王梓涵",
				Email:     "2024310103@xinan.example.edu.cn",
				Role:      model.RoleStudent,
				ClassName: seedClassName,
				StudentNo: "2024310103",
			},
			Profiles: map[string]float64{
				"web":       0.44,
				"crypto":    0.41,
				"forensics": 0.58,
				"misc":      0.46,
				"pwn":       0.43,
				"reverse":   0.18,
			},
			Sessions: []sessionSeed{
				{
					ChallengeCategory: "forensics",
					ChallengeIndex:    0,
					StartOffset:       -3*24*time.Hour + 3*time.Hour,
					Duration:          35 * time.Minute,
					Access:            true,
					ProxyRequests: []proxySeed{
						{Offset: 4 * time.Minute, Method: "GET", Path: "/trash-bin.zip", Status: 200},
					},
					Submissions: []submissionSeed{
						{Offset: 22 * time.Minute, Flag: "flag{forensics-recycle-note}", Correct: true},
					},
				},
				{
					ChallengeCategory: "reverse",
					ChallengeIndex:    0,
					StartOffset:       -2*24*time.Hour + 4*time.Hour,
					Duration:          50 * time.Minute,
					Access:            true,
					ProxyRequests: []proxySeed{
						{Offset: 9 * time.Minute, Method: "GET", Path: "/checker.bin", Status: 200},
						{Offset: 27 * time.Minute, Method: "POST", Path: "/verify", Status: 400, PayloadPreview: "candidate=debug"},
					},
					Submissions: []submissionSeed{
						{Offset: 41 * time.Minute, Flag: "flag{reverse-not-yet}", Correct: false},
					},
				},
			},
		},
		{
			User: userSeed{
				Username:  "chensiyuan",
				Name:      "陈思远",
				Email:     "2024310104@xinan.example.edu.cn",
				Role:      model.RoleStudent,
				ClassName: seedClassName,
				StudentNo: "2024310104",
			},
			Profiles: map[string]float64{
				"web":       0.46,
				"crypto":    0.45,
				"forensics": 0.44,
				"misc":      0.49,
				"pwn":       0.15,
				"reverse":   0.41,
			},
			Sessions: []sessionSeed{
				{
					ChallengeCategory: "pwn",
					ChallengeIndex:    0,
					StartOffset:       -9*24*time.Hour + 2*time.Hour,
					Duration:          25 * time.Minute,
					Access:            true,
					ProxyRequests: []proxySeed{
						{Offset: 6 * time.Minute, Method: "GET", Path: "/gate", Status: 200},
					},
					Submissions: []submissionSeed{
						{Offset: 19 * time.Minute, Flag: "flag{too-short}", Correct: false},
					},
				},
			},
		},
		{
			User: userSeed{
				Username:  "limuyang",
				Name:      "李沐阳",
				Email:     "2024310105@xinan.example.edu.cn",
				Role:      model.RoleStudent,
				ClassName: seedClassName,
				StudentNo: "2024310105",
			},
			Profiles: map[string]float64{
				"web":       0.33,
				"crypto":    0.56,
				"forensics": 0.49,
				"misc":      0.43,
				"pwn":       0.45,
				"reverse":   0.74,
			},
			Sessions: []sessionSeed{
				{
					ChallengeCategory: "reverse",
					ChallengeIndex:    0,
					StartOffset:       -5*24*time.Hour + 4*time.Hour,
					Duration:          95 * time.Minute,
					Access:            true,
					ProxyRequests: []proxySeed{
						{Offset: 7 * time.Minute, Method: "GET", Path: "/download/checker", Status: 200},
						{Offset: 39 * time.Minute, Method: "POST", Path: "/api/trace", Status: 200, PayloadPreview: "trace_id=7f1"},
					},
					Submissions: []submissionSeed{
						{Offset: 72 * time.Minute, Flag: "flag{reverse-checker-pass}", Correct: true},
					},
					Writeup: &writeupSeed{
						Offset:    90 * time.Minute,
						Title:     "定位校验分支后回推正确输入",
						Content:   "通过拆分关键校验分支和输入变换流程，逆推出最终校验通过的字符串。",
						Published: true,
					},
				},
				{
					ChallengeCategory: "web",
					ChallengeIndex:    1,
					StartOffset:       -1*24*time.Hour + 3*time.Hour,
					Duration:          45 * time.Minute,
					Access:            true,
					ProxyRequests: []proxySeed{
						{Offset: 8 * time.Minute, Method: "GET", Path: "/login", Status: 200},
					},
					Submissions: []submissionSeed{
						{Offset: 31 * time.Minute, Flag: "flag{header-gate-pass}", Correct: true},
					},
				},
			},
		},
		{
			User: userSeed{
				Username:  "zhoujianning",
				Name:      "周嘉宁",
				Email:     "2024310106@xinan.example.edu.cn",
				Role:      model.RoleStudent,
				ClassName: seedClassName,
				StudentNo: "2024310106",
			},
			Profiles: map[string]float64{
				"web":       0.48,
				"crypto":    0.61,
				"forensics": 0.31,
				"misc":      0.67,
				"pwn":       0.42,
				"reverse":   0.45,
			},
			Sessions: []sessionSeed{
				{
					ChallengeCategory: "misc",
					ChallengeIndex:    0,
					StartOffset:       -8*24*time.Hour + 90*time.Minute,
					Duration:          55 * time.Minute,
					Access:            true,
					ProxyRequests: []proxySeed{
						{Offset: 6 * time.Minute, Method: "GET", Path: "/comment.txt", Status: 200},
					},
					Submissions: []submissionSeed{
						{Offset: 29 * time.Minute, Flag: "flag{comment-sticky-note}", Correct: true},
					},
					Writeup: &writeupSeed{
						Offset:    52 * time.Minute,
						Title:     "从注释和静态资源恢复隐藏便签",
						Content:   "先定位注释，再顺着静态资源和编码痕迹还原最终便签内容。",
						Published: true,
					},
				},
				{
					ChallengeCategory: "crypto",
					ChallengeIndex:    1,
					StartOffset:       -8*24*time.Hour + 4*time.Hour,
					Duration:          70 * time.Minute,
					Access:            true,
					ProxyRequests: []proxySeed{
						{Offset: 9 * time.Minute, Method: "GET", Path: "/stream.bin", Status: 200},
					},
					Submissions: []submissionSeed{
						{Offset: 58 * time.Minute, Flag: "flag{stream-backup-ticket}", Correct: true},
					},
				},
			},
		},
	}
}

func upsertUser(tx *gorm.DB, spec userSeed) (*model.User, error) {
	var user model.User
	err := tx.Unscoped().Where("username = ?", spec.Username).First(&user).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		user = model.User{
			Username:  spec.Username,
			Name:      spec.Name,
			Email:     spec.Email,
			Role:      spec.Role,
			ClassName: spec.ClassName,
			Status:    model.UserStatusActive,
			StudentNo: spec.StudentNo,
			TeacherNo: spec.TeacherNo,
		}
		if passwordErr := user.SetPassword(defaultPassword); passwordErr != nil {
			return nil, fmt.Errorf("set password for %s: %w", spec.Username, passwordErr)
		}
		if createErr := tx.Create(&user).Error; createErr != nil {
			return nil, fmt.Errorf("create user %s: %w", spec.Username, createErr)
		}
	case err != nil:
		return nil, fmt.Errorf("find user %s: %w", spec.Username, err)
	default:
		updatedUser := user
		updatedUser.Name = spec.Name
		updatedUser.Email = spec.Email
		updatedUser.Role = spec.Role
		updatedUser.ClassName = spec.ClassName
		updatedUser.Status = model.UserStatusActive
		updatedUser.StudentNo = spec.StudentNo
		updatedUser.TeacherNo = spec.TeacherNo
		updatedUser.DeletedAt = gorm.DeletedAt{}
		if passwordErr := updatedUser.SetPassword(defaultPassword); passwordErr != nil {
			return nil, fmt.Errorf("reset password for %s: %w", spec.Username, passwordErr)
		}
		if updateErr := tx.Unscoped().Model(&user).Updates(map[string]any{
			"name":                  updatedUser.Name,
			"email":                 updatedUser.Email,
			"role":                  updatedUser.Role,
			"class_name":            updatedUser.ClassName,
			"status":                updatedUser.Status,
			"student_no":            updatedUser.StudentNo,
			"teacher_no":            updatedUser.TeacherNo,
			"password_hash":         updatedUser.PasswordHash,
			"failed_login_attempts": 0,
			"last_failed_login_at":  nil,
			"locked_until":          nil,
			"deleted_at":            nil,
			"updated_at":            time.Now().UTC(),
		}).Error; updateErr != nil {
			return nil, fmt.Errorf("update user %s: %w", spec.Username, updateErr)
		}
		user = updatedUser
	}

	if err := ensureUserRole(tx, user.ID, user.Role); err != nil {
		return nil, err
	}
	return &user, nil
}

func ensureUserRole(tx *gorm.DB, userID int64, roleCode string) error {
	var role model.Role
	if err := tx.Where("code = ?", roleCode).First(&role).Error; err != nil {
		return fmt.Errorf("find role %s: %w", roleCode, err)
	}
	if err := tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
		return fmt.Errorf("clear user roles for %d: %w", userID, err)
	}
	if err := tx.Create(&model.UserRole{
		UserID:    userID,
		RoleID:    role.ID,
		CreatedAt: time.Now().UTC(),
	}).Error; err != nil {
		return fmt.Errorf("assign role %s to user %d: %w", roleCode, userID, err)
	}
	return nil
}

func resetSeededData(tx *gorm.DB, userIDs []int64, teacherID int64) error {
	if len(userIDs) == 0 {
		return nil
	}
	if err := tx.Where("user_id IN ?", userIDs).Delete(&model.AuditLog{}).Error; err != nil {
		return fmt.Errorf("delete audit logs: %w", err)
	}
	if err := tx.Where("user_id IN ?", userIDs).Delete(&model.SubmissionWriteup{}).Error; err != nil {
		return fmt.Errorf("delete submission writeups: %w", err)
	}
	if err := tx.Where("user_id IN ? AND contest_id IS NULL", userIDs).Delete(&model.Submission{}).Error; err != nil {
		return fmt.Errorf("delete submissions: %w", err)
	}
	if err := tx.Where("user_id IN ? AND contest_id IS NULL", userIDs).Delete(&model.Instance{}).Error; err != nil {
		return fmt.Errorf("delete instances: %w", err)
	}
	if err := tx.Where("user_id IN ?", userIDs).Delete(&model.SkillProfile{}).Error; err != nil {
		return fmt.Errorf("delete skill profiles: %w", err)
	}
	if err := tx.Where("class_name = ? OR user_id IN ?", seedClassName, userIDs).Delete(&model.Report{}).Error; err != nil {
		return fmt.Errorf("delete reports: %w", err)
	}
	if err := tx.Where("user_id = ?", teacherID).Delete(&model.UserRole{}).Error; err != nil {
		return fmt.Errorf("reset teacher role: %w", err)
	}
	return ensureUserRole(tx, teacherID, model.RoleTeacher)
}

func seedStudentScenario(
	tx *gorm.DB,
	teacher *model.User,
	student *model.User,
	scenario studentScenario,
	catalog *challengeCatalog,
) error {
	now := time.Now().UTC().Truncate(time.Second)
	for _, session := range scenario.Sessions {
		challenge, err := catalog.pick(session.ChallengeCategory, session.ChallengeIndex)
		if err != nil {
			return err
		}
		if err := createSession(tx, teacher, student, challenge, session, now); err != nil {
			return err
		}
	}
	return upsertSkillProfiles(tx, student.ID, scenario.Profiles, now)
}

func createSession(
	tx *gorm.DB,
	teacher *model.User,
	student *model.User,
	challenge challengeRef,
	session sessionSeed,
	base time.Time,
) error {
	startAt := base.Add(session.StartOffset)
	endAt := startAt.Add(session.Duration)
	instance := &model.Instance{
		UserID:         student.ID,
		ChallengeID:    challenge.ID,
		ContainerID:    fmt.Sprintf("seed-%s-%d", student.Username, challenge.ID),
		Status:         model.InstanceStatusStopped,
		AccessURL:      fmt.Sprintf("http://127.0.0.1:%d", 32000+(student.ID%1000)+(challenge.ID%100)),
		Nonce:          fmt.Sprintf("seed-%d-%d", student.ID, challenge.ID),
		ExpiresAt:      endAt.Add(90 * time.Minute),
		DestroyedAt:    timePtr(endAt),
		ExtendCount:    0,
		MaxExtends:     2,
		CreatedAt:      startAt,
		UpdatedAt:      endAt,
		RuntimeDetails: `{"seed":"teaching-review-data"}`,
		ShareScope:     model.ShareScopePerUser,
	}
	if err := tx.Create(instance).Error; err != nil {
		return fmt.Errorf("create instance for %s/%s: %w", student.Username, challenge.Title, err)
	}

	if session.Access {
		if err := tx.Create(&model.AuditLog{
			UserID:       int64Ptr(student.ID),
			Action:       model.AuditActionRead,
			ResourceType: "instance_access",
			ResourceID:   int64Ptr(instance.ID),
			Detail:       proxyDetailJSON("GET", fmt.Sprintf("/api/v1/instances/%d/access", instance.ID), "", 200, ""),
			IPAddress:    seedIPAddress(student.ID),
			UserAgent:    stringPtr(seedUserAgent),
			CreatedAt:    startAt.Add(2 * time.Minute),
		}).Error; err != nil {
			return fmt.Errorf("create instance access log for %s/%s: %w", student.Username, challenge.Title, err)
		}
	}

	for _, proxy := range session.ProxyRequests {
		if err := tx.Create(&model.AuditLog{
			UserID:       int64Ptr(student.ID),
			Action:       auditActionForMethod(proxy.Method),
			ResourceType: "instance_proxy_request",
			ResourceID:   int64Ptr(instance.ID),
			Detail:       proxyDetailJSON(proxy.Method, proxy.Path, proxy.Query, proxy.Status, proxy.PayloadPreview),
			IPAddress:    seedIPAddress(student.ID),
			UserAgent:    stringPtr(seedUserAgent),
			CreatedAt:    startAt.Add(proxy.Offset),
		}).Error; err != nil {
			return fmt.Errorf("create proxy audit log for %s/%s: %w", student.Username, challenge.Title, err)
		}
	}

	for _, submission := range session.Submissions {
		submittedAt := startAt.Add(submission.Offset)
		score := 0
		if submission.Correct {
			score = challenge.Points
		}
		record := &model.Submission{
			UserID:       student.ID,
			ChallengeID:  challenge.ID,
			Flag:         submission.Flag,
			IsCorrect:    submission.Correct,
			ReviewStatus: model.SubmissionReviewStatusNotRequired,
			Score:        score,
			SubmittedAt:  submittedAt,
			UpdatedAt:    submittedAt,
		}
		if err := tx.Create(record).Error; err != nil {
			return fmt.Errorf("create submission for %s/%s: %w", student.Username, challenge.Title, err)
		}
	}

	if session.Writeup != nil {
		writeupAt := startAt.Add(session.Writeup.Offset)
		submissionStatus := model.SubmissionWriteupStatusDraft
		visibilityStatus := model.SubmissionWriteupVisibilityHidden
		var publishedAt *time.Time
		var recommendedAt *time.Time
		var recommendedBy *int64
		if session.Writeup.Published {
			submissionStatus = model.SubmissionWriteupStatusPublished
			visibilityStatus = model.SubmissionWriteupVisibilityVisible
			publishedAt = timePtr(writeupAt)
		}
		if session.Writeup.Recommended {
			recommendedAt = timePtr(writeupAt.Add(15 * time.Minute))
			recommendedBy = int64Ptr(teacher.ID)
		}
		if err := tx.Create(&model.SubmissionWriteup{
			UserID:           student.ID,
			ChallengeID:      challenge.ID,
			Title:            session.Writeup.Title,
			Content:          session.Writeup.Content,
			SubmissionStatus: submissionStatus,
			VisibilityStatus: visibilityStatus,
			IsRecommended:    session.Writeup.Recommended,
			RecommendedAt:    recommendedAt,
			RecommendedBy:    recommendedBy,
			PublishedAt:      publishedAt,
			CreatedAt:        writeupAt,
			UpdatedAt:        writeupAt,
		}).Error; err != nil {
			return fmt.Errorf("create writeup for %s/%s: %w", student.Username, challenge.Title, err)
		}
	}

	return nil
}

func upsertSkillProfiles(tx *gorm.DB, userID int64, profiles map[string]float64, now time.Time) error {
	dimensions := make([]string, 0, len(profiles))
	for dimension := range profiles {
		dimensions = append(dimensions, dimension)
	}
	sort.Strings(dimensions)
	for idx, dimension := range dimensions {
		score := profiles[dimension]
		record := &model.SkillProfile{
			UserID:    userID,
			Dimension: dimension,
			Score:     score,
			UpdatedAt: now.Add(time.Duration(idx) * time.Second),
		}
		if err := tx.Where("user_id = ? AND dimension = ?", userID, dimension).
			Assign(record).
			FirstOrCreate(record).Error; err != nil {
			return fmt.Errorf("upsert skill profile %d/%s: %w", userID, dimension, err)
		}
	}
	return nil
}

func (c *challengeCatalog) pick(category string, index int) (challengeRef, error) {
	normalized := strings.ToLower(strings.TrimSpace(category))
	items := c.byCategory[normalized]
	if index < 0 || index >= len(items) {
		return challengeRef{}, fmt.Errorf("challenge category %s does not have index %d", normalized, index)
	}
	return items[index], nil
}

func proxyDetailJSON(method, path, query string, status int, payloadPreview string) string {
	detail := map[string]any{
		"method":       strings.ToUpper(strings.TrimSpace(method)),
		"target_path":  path,
		"target_query": query,
		"status":       status,
	}
	if strings.TrimSpace(payloadPreview) != "" {
		detail["payload_preview"] = payloadPreview
	}
	data, err := json.Marshal(detail)
	if err != nil {
		return `{}`
	}
	return string(data)
}

func auditActionForMethod(method string) string {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "POST":
		return model.AuditActionSubmit
	case "PUT", "PATCH":
		return model.AuditActionUpdate
	case "DELETE":
		return model.AuditActionDelete
	default:
		return model.AuditActionRead
	}
}

func printSeedSummary(result *seedResult) {
	fmt.Printf("教学复盘样本数据已写入班级 %s\n", result.ClassName)
	fmt.Printf("教师账号: %s / %s (%s)\n", result.Teacher.Username, defaultPassword, result.Teacher.Name)
	fmt.Println()
	fmt.Println("班级复盘结论:")
	for _, item := range result.ClassReview.Items {
		fmt.Printf("- [%s] %s\n", item.Severity, item.Summary)
		if item.Evidence != "" {
			fmt.Printf("  证据: %s\n", item.Evidence)
		}
		if item.Action != "" {
			fmt.Printf("  建议: %s\n", item.Action)
		}
		if item.Recommendation != nil {
			fmt.Printf(
				"  推荐题: %s (%s/%s) - %s\n",
				item.Recommendation.Title,
				item.Recommendation.Category,
				item.Recommendation.Difficulty,
				item.Recommendation.Summary,
			)
			if item.Recommendation.Evidence != "" {
				fmt.Printf("  推荐依据: %s\n", item.Recommendation.Evidence)
			}
		}
	}
	fmt.Println()
	fmt.Println("学生复盘摘要:")
	for _, student := range result.Students {
		name := strings.TrimSpace(student.User.Name)
		if name == "" {
			name = student.User.Username
		}
		fmt.Printf("- %s (%s / %s)\n", name, student.User.Username, student.User.StudentNo)
		if len(student.Recommendations.Challenges) == 0 {
			fmt.Println("  推荐题: 无")
		} else {
			top := student.Recommendations.Challenges[0]
			fmt.Printf("  推荐题: %s (%s/%s) - %s\n", top.Title, top.Category, top.Difficulty, top.Summary)
			if top.Evidence != "" {
				fmt.Printf("  推荐依据: %s\n", top.Evidence)
			}
		}
		fmt.Printf(
			"  归档摘要: solved=%d, attempts=%d, evidence=%d, writeups=%d\n",
			student.Archive.Summary.TotalSolved,
			student.Archive.Summary.TotalAttempts,
			student.Archive.Summary.EvidenceEventCount,
			student.Archive.Summary.WriteupCount,
		)
		if len(student.Archive.TeacherObservations.Items) == 0 {
			fmt.Println("  观察结论: 无")
			continue
		}
		for _, observation := range student.Archive.TeacherObservations.Items {
			fmt.Printf("  观察结论: [%s] %s - %s\n", observation.Severity, observation.Code, observation.Summary)
			if observation.Evidence != "" {
				fmt.Printf("  观察证据: %s\n", observation.Evidence)
			}
			if observation.Action != "" {
				fmt.Printf("  建议动作: %s\n", observation.Action)
			}
		}
	}
}

func seedIPAddress(userID int64) string {
	return fmt.Sprintf("10.24.1.%d", (userID%200)+10)
}

func int64Ptr(v int64) *int64 {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

func timePtr(v time.Time) *time.Time {
	return &v
}

func setDefaultEnv(key, value string) {
	if strings.TrimSpace(os.Getenv(key)) != "" {
		return
	}
	if err := os.Setenv(key, value); err != nil {
		panic(fmt.Errorf("set default %s: %w", key, err))
	}
}
