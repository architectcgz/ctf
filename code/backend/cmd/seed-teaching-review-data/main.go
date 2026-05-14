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
	seedClassName                  = "信安2401"
	seedTeacherUsername            = "zhaoxiaofeng"
	defaultPassword                = "Password123"
	seedUserAgent                  = "seed-teaching-review-data/1.0"
	seedAWDContestTitle            = "教学复盘样本 AWD 迁移演练"
	coverageExpansionMinChallenges = 48
	coverageMaxStudentsPerCategory = 4
)

const (
	seedAWDCaptainStudent   = "student"
	seedAWDCaptainTeacher   = "teacher"
	seedAWDCaptainSynthetic = "synthetic"
)

var coverageStudentSeeds = []userSeed{
	{Username: "guhaoran", Name: "顾浩然"},
	{Username: "xushiyu", Name: "徐诗雨"},
	{Username: "liangzixuan", Name: "梁梓轩"},
	{Username: "tianruohan", Name: "田若涵"},
	{Username: "songyiran", Name: "宋奕然"},
	{Username: "hejingyi", Name: "何静怡"},
	{Username: "wenhaoyu", Name: "温浩宇"},
	{Username: "qinchuxi", Name: "秦初曦"},
	{Username: "zhujiale", Name: "朱嘉乐"},
	{Username: "luoyanxi", Name: "罗妍希"},
	{Username: "pengzihan", Name: "彭子涵"},
	{Username: "shenmingrui", Name: "沈铭睿"},
	{Username: "caoruiwen", Name: "曹芮雯"},
	{Username: "fuyuze", Name: "付宇泽"},
	{Username: "linwenqi", Name: "林文琪"},
	{Username: "duyichen", Name: "杜奕辰"},
	{Username: "jiangzimo", Name: "蒋梓墨"},
	{Username: "maoruiting", Name: "毛瑞婷"},
	{Username: "yangjunhao", Name: "杨骏豪"},
	{Username: "zhaomengyao", Name: "赵梦瑶"},
	{Username: "wuxichen", Name: "吴希辰"},
	{Username: "sunyuhan", Name: "孙语涵"},
	{Username: "tanjiahao", Name: "谭嘉豪"},
	{Username: "zhouyuxin", Name: "周雨昕"},
}

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

type awdChallengeRef struct {
	ID         int64
	Name       string
	Category   string
	Difficulty string
}

type awdChallengeCatalog struct {
	byCategory map[string][]awdChallengeRef
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

type awdTeamSeed struct {
	Name        string
	CaptainRole string
	MemberCount int
}

type awdServiceSeed struct {
	TeamIndex      int
	ServiceStatus  string
	AttackReceived int
	SLAScore       int
	DefenseScore   int
	AttackScore    int
	CheckerType    model.AWDCheckerType
	CheckResult    string
	UpdatedOffset  time.Duration
}

type awdAttackSeed struct {
	Offset             time.Duration
	AttackerTeamIndex  int
	VictimTeamIndex    int
	SubmittedFlag      string
	AttackType         string
	Source             string
	SubmittedByStudent bool
	IsSuccess          bool
	ScoreGained        int
}

type awdTrafficSeed struct {
	Offset            time.Duration
	AttackerTeamIndex int
	VictimTeamIndex   int
	Method            string
	Path              string
	StatusCode        int
	Source            string
}

type awdRoundSeed struct {
	RoundNumber  int
	StartOffset  time.Duration
	Duration     time.Duration
	AttackScore  int
	DefenseScore int
	Services     []awdServiceSeed
	Attacks      []awdAttackSeed
	Traffic      []awdTrafficSeed
}

type awdScenario struct {
	ChallengeCategory string
	ChallengeIndex    int
	StartOffset       time.Duration
	Teams             []awdTeamSeed
	Rounds            []awdRoundSeed
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
	Label    string
	User     userSeed
	Profiles map[string]float64
	Sessions []sessionSeed
	AWD      *awdScenario
}

type seededStudentResult struct {
	User                 *model.User
	ScenarioLabel        string
	PracticeSessionCount int
	AWDAttackCount       int
	Recommendations      dto.TeacherRecommendationResp
	Archive              *assessmentcmd.ReviewArchiveData
}

type categoryCoverage struct {
	Published int
	Used      int
}

type seedCoverageSummary struct {
	PublishedChallenges          int
	UsedPracticeChallenges       int
	StudentsWithRecommendations  int
	UniqueTopRecommendationCount int
	ByCategory                   map[string]categoryCoverage
}

type seedResult struct {
	ClassName            string
	Teacher              *model.User
	PracticeSessionCount int
	AWDAttackCount       int
	Students             []seededStudentResult
	ClassReview          *dto.TeacherClassReviewResp
	Coverage             seedCoverageSummary
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
	awdCatalog, err := loadAWDChallengeCatalog(ctx, db)
	if err != nil {
		return nil, err
	}
	scenarios := buildStudentScenarios(catalog)

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
			if err := seedStudentScenario(tx, teacher, student, scenario, catalog, awdCatalog); err != nil {
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
		assessmentinfra.NewRecommendationCacheStore(cache),
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
	readmodelRepo := readmodelinfra.NewRepository(db)
	classInsightService := readmodelqueries.NewClassInsightService(
		readmodelRepo,
		recommendationService,
		zap.NewNop(),
	)
	studentReviewService := readmodelqueries.NewStudentReviewService(
		readmodelRepo,
		recommendationService,
	)

	classReview, err := classInsightService.GetClassReview(ctx, teacher.ID, model.RoleTeacher, seedClassName)
	if err != nil {
		return nil, fmt.Errorf("load class review: %w", err)
	}

	results := make([]seededStudentResult, 0, len(scenarios))
	practiceSessionCount := 0
	awdAttackCount := 0
	sort.SliceStable(scenarios, func(i, j int) bool {
		return scenarios[i].User.StudentNo < scenarios[j].User.StudentNo
	})
	for _, scenario := range scenarios {
		student := studentsByUsername[scenario.User.Username]
		if student == nil {
			continue
		}
		recommendations, recErr := studentReviewService.GetStudentRecommendations(ctx, teacher.ID, model.RoleTeacher, student.ID, 3)
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
		practiceSessionCount += len(scenario.Sessions)
		awdAttackCount += countAWDAttacks(scenario.AWD)
		results = append(results, seededStudentResult{
			User:                 student,
			ScenarioLabel:        scenario.Label,
			PracticeSessionCount: len(scenario.Sessions),
			AWDAttackCount:       countAWDAttacks(scenario.AWD),
			Recommendations:      studentRecommendations,
			Archive:              archive,
		})
	}

	return &seedResult{
		ClassName:            seedClassName,
		Teacher:              teacher,
		PracticeSessionCount: practiceSessionCount,
		AWDAttackCount:       awdAttackCount,
		Students:             results,
		ClassReview:          classReview,
		Coverage:             buildSeedCoverageSummary(catalog, scenarios, results),
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

func loadAWDChallengeCatalog(ctx context.Context, db *gorm.DB) (*awdChallengeCatalog, error) {
	var rows []model.AWDChallenge
	err := db.WithContext(ctx).
		Where("status = ?", model.AWDChallengeStatusPublished).
		Order("created_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("load published awd challenges: %w", err)
	}

	catalog := &awdChallengeCatalog{byCategory: make(map[string][]awdChallengeRef)}
	for _, row := range rows {
		category := strings.ToLower(strings.TrimSpace(row.Category))
		if category == "" {
			continue
		}
		catalog.byCategory[category] = append(catalog.byCategory[category], awdChallengeRef{
			ID:         row.ID,
			Name:       row.Name,
			Category:   category,
			Difficulty: row.Difficulty,
		})
	}
	return catalog, nil
}

func buildStudentScenarios(catalog *challengeCatalog) []studentScenario {
	scenarios := buildBaseStudentScenarios()
	return append(scenarios, buildCoverageStudentScenarios(catalog)...)
}

func buildCoverageStudentScenarios(catalog *challengeCatalog) []studentScenario {
	if !shouldExpandCoverage(catalog) {
		return nil
	}

	scenarios := make([]studentScenario, 0, len(model.AllDimensions)*2)
	seedIndex := 0
	for dimIndex, dimension := range model.AllDimensions {
		items := catalog.byCategory[dimension]
		studentsPerCategory := coverageStudentsPerCategory(len(items))
		if studentsPerCategory == 0 {
			continue
		}
		for variant := 0; variant < studentsPerCategory; variant++ {
			if seedIndex >= len(coverageStudentSeeds) {
				return scenarios
			}
			spec := coverageStudentSeeds[seedIndex]
			seedIndex++
			scenarios = append(scenarios, buildCoverageStudentScenario(spec, dimIndex, variant, catalog))
		}
	}
	return scenarios
}

func shouldExpandCoverage(catalog *challengeCatalog) bool {
	if catalog == nil {
		return false
	}
	return catalog.totalCount() >= coverageExpansionMinChallenges
}

func coverageStudentsPerCategory(challengeCount int) int {
	if challengeCount < 4 {
		return 0
	}
	students := challengeCount / 4
	if students < 1 {
		students = 1
	}
	if students > coverageMaxStudentsPerCategory {
		students = coverageMaxStudentsPerCategory
	}
	return students
}

func buildCoverageStudentScenario(
	spec userSeed,
	dimensionIndex int,
	variant int,
	catalog *challengeCatalog,
) studentScenario {
	weakDimension := model.AllDimensions[dimensionIndex%len(model.AllDimensions)]
	strongDimension := nextCoverageDimension(catalog, dimensionIndex+1, weakDimension)
	steadyDimension := nextCoverageDimension(catalog, dimensionIndex+2, weakDimension, strongDimension)
	studentSerial := dimensionIndex*coverageMaxStudentsPerCategory + variant + 1

	spec.Role = model.RoleStudent
	spec.ClassName = seedClassName
	spec.StudentNo = fmt.Sprintf("20243102%02d", studentSerial)
	spec.Email = fmt.Sprintf("%s@xinan.example.edu.cn", spec.StudentNo)

	weakItems := catalog.byCategory[weakDimension]
	weakSolvedPrefix := coverageWeakSolvedPrefixCount(len(weakItems), variant)
	weakBase := coverageWeakAttemptBase(len(weakItems), weakSolvedPrefix)

	labelSuffix := []string{
		"高试错补量",
		"稳态训练补量",
		"短时冲刺补量",
		"交替练习补量",
	}[variant%4]
	baseOffset := -6*24*time.Hour + time.Duration(dimensionIndex*11+variant*7)*time.Hour
	sessions := make([]sessionSeed, 0, 4+weakSolvedPrefix)
	if strongDimension != "" {
		strongItems := catalog.byCategory[strongDimension]
		strongBase := variant % len(strongItems)
		sessions = append(sessions,
			buildSuccessfulCoverageSession(strongDimension, strongBase, baseOffset+4*time.Hour, 50*time.Minute, variant, spec.Name),
		)
	}
	for solvedIndex := 0; solvedIndex < weakSolvedPrefix; solvedIndex++ {
		solvedOffsetHours := 12 + solvedIndex*8
		if variant > 0 && solvedIndex == weakSolvedPrefix-1 {
			// Break the wrong-submission streak without pulling all successful activity out of the week tail.
			solvedOffsetHours = 33
		}
		sessions = append(sessions, buildSuccessfulCoverageSession(
			weakDimension,
			solvedIndex,
			baseOffset+time.Duration(solvedOffsetHours)*time.Hour,
			38*time.Minute,
			variant+dimensionIndex*10+solvedIndex,
			spec.Name,
		))
	}
	sessions = append(sessions,
		buildWeakCoverageSession(weakDimension, weakBase, baseOffset+22*time.Hour, 65*time.Minute, false, variant),
		buildWeakCoverageSession(weakDimension, (weakBase+1)%len(weakItems), baseOffset+43*time.Hour, 70*time.Minute, variant%2 == 1, variant+1),
	)
	if steadyDimension != "" {
		steadyItems := catalog.byCategory[steadyDimension]
		steadyBase := (variant + dimensionIndex) % len(steadyItems)
		sessions = append(sessions, buildSuccessfulCoverageSession(steadyDimension, steadyBase, baseOffset+65*time.Hour, 45*time.Minute, variant+dimensionIndex, spec.Name))
	}

	return studentScenario{
		Label:    fmt.Sprintf("扩展覆盖 / %s 弱项 / %s", weakDimension, labelSuffix),
		User:     spec,
		Profiles: buildCoverageProfiles(weakDimension, strongDimension, steadyDimension, variant),
		Sessions: sessions,
	}
}

func nextCoverageDimension(catalog *challengeCatalog, startIndex int, exclude ...string) string {
	if catalog == nil || len(model.AllDimensions) == 0 {
		return ""
	}

	excluded := make(map[string]struct{}, len(exclude))
	for _, dimension := range exclude {
		if dimension == "" {
			continue
		}
		excluded[dimension] = struct{}{}
	}

	for offset := 0; offset < len(model.AllDimensions); offset++ {
		dimension := model.AllDimensions[(startIndex+offset)%len(model.AllDimensions)]
		if _, skip := excluded[dimension]; skip {
			continue
		}
		if len(catalog.byCategory[dimension]) == 0 {
			continue
		}
		return dimension
	}
	return ""
}

func coverageWeakSolvedPrefixCount(challengeCount int, variant int) int {
	if challengeCount <= 2 || variant <= 0 {
		return 0
	}
	maxPrefix := challengeCount - 2
	if variant > maxPrefix {
		return maxPrefix
	}
	return variant
}

func coverageWeakAttemptBase(challengeCount int, solvedPrefix int) int {
	if challengeCount <= 1 {
		return 0
	}
	base := solvedPrefix * 2
	if base+1 < challengeCount {
		return base
	}
	if solvedPrefix < challengeCount-1 {
		return solvedPrefix
	}
	return challengeCount - 2
}

func buildCoverageProfiles(
	weakDimension string,
	strongDimension string,
	steadyDimension string,
	variant int,
) map[string]float64 {
	profiles := make(map[string]float64, len(model.AllDimensions))
	for index, dimension := range model.AllDimensions {
		profiles[dimension] = 0.48 + float64((index+variant)%3)*0.04
	}
	profiles[weakDimension] = 0.18 + float64(variant%3)*0.06
	if strongDimension != "" && strongDimension != weakDimension {
		profiles[strongDimension] = 0.78 - float64(variant%3)*0.03
	}
	if steadyDimension != "" && steadyDimension != weakDimension && steadyDimension != strongDimension {
		profiles[steadyDimension] = 0.64 - float64(variant%2)*0.04
	}
	return profiles
}

func buildSuccessfulCoverageSession(
	category string,
	index int,
	startOffset time.Duration,
	duration time.Duration,
	variant int,
	studentName string,
) sessionSeed {
	return sessionSeed{
		ChallengeCategory: category,
		ChallengeIndex:    index,
		StartOffset:       startOffset,
		Duration:          duration,
		Access:            true,
		ProxyRequests: []proxySeed{
			{Offset: 5 * time.Minute, Method: "GET", Path: "/", Status: 200},
			{Offset: 18 * time.Minute, Method: "POST", Path: "/solve", Status: 200, PayloadPreview: fmt.Sprintf("variant=%d", variant)},
		},
		Submissions: []submissionSeed{
			{Offset: 32 * time.Minute, Flag: fmt.Sprintf("flag{%s-success-%d}", category, variant), Correct: true},
		},
		Writeup: &writeupSeed{
			Offset:      40 * time.Minute,
			Title:       fmt.Sprintf("%s 维度训练记录", strings.TrimSpace(studentName)),
			Content:     fmt.Sprintf("围绕 %s 维度完成一次可复盘的成功训练，保留了关键入口、利用路径和提交结论。", category),
			Published:   true,
			Recommended: variant%3 == 0,
		},
	}
}

func buildWeakCoverageSession(
	category string,
	index int,
	startOffset time.Duration,
	duration time.Duration,
	finishWithSuccess bool,
	variant int,
) sessionSeed {
	submissions := []submissionSeed{
		{Offset: 21 * time.Minute, Flag: fmt.Sprintf("flag{%s-miss-%d-a}", category, variant), Correct: false},
		{Offset: 39 * time.Minute, Flag: fmt.Sprintf("flag{%s-miss-%d-b}", category, variant), Correct: false},
	}
	if finishWithSuccess {
		submissions = append(submissions, submissionSeed{
			Offset:  56 * time.Minute,
			Flag:    fmt.Sprintf("flag{%s-hit-%d}", category, variant),
			Correct: true,
		})
	}
	return sessionSeed{
		ChallengeCategory: category,
		ChallengeIndex:    index,
		StartOffset:       startOffset,
		Duration:          duration,
		Access:            true,
		ProxyRequests: []proxySeed{
			{Offset: 6 * time.Minute, Method: "GET", Path: "/", Status: 200},
			{Offset: 15 * time.Minute, Method: "GET", Path: "/hint", Status: 200},
			{Offset: 30 * time.Minute, Method: "POST", Path: "/attempt", Status: 400, PayloadPreview: fmt.Sprintf("dimension=%s", category)},
		},
		Submissions: submissions,
	}
}

func buildBaseStudentScenarios() []studentScenario {
	return []studentScenario{
		{
			Label: "稳定闭环 + AWD 迁移",
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
			AWD: &awdScenario{
				ChallengeCategory: "web",
				ChallengeIndex:    0,
				StartOffset:       -1*24*time.Hour + 6*time.Hour,
				Teams: []awdTeamSeed{
					{Name: "林宸熙攻防组", CaptainRole: seedAWDCaptainStudent, MemberCount: 2},
					{Name: "教学靶场防守组", CaptainRole: seedAWDCaptainTeacher, MemberCount: 2},
					{Name: "旁路渗透样本队", CaptainRole: seedAWDCaptainSynthetic, MemberCount: 3},
				},
				Rounds: []awdRoundSeed{
					{
						RoundNumber:  1,
						StartOffset:  0,
						Duration:     32 * time.Minute,
						AttackScore:  80,
						DefenseScore: 50,
						Services: []awdServiceSeed{
							{TeamIndex: 0, ServiceStatus: model.AWDServiceStatusUp, AttackReceived: 1, SLAScore: 30, DefenseScore: 40, AttackScore: 15, CheckerType: model.AWDCheckerTypeHTTPStandard, CheckResult: `{"status":"ok","latency_ms":78}`, UpdatedOffset: 25 * time.Minute},
							{TeamIndex: 1, ServiceStatus: model.AWDServiceStatusCompromised, AttackReceived: 2, SLAScore: 25, DefenseScore: 28, AttackScore: 20, CheckerType: model.AWDCheckerTypeHTTPStandard, CheckResult: `{"status":"partial","latency_ms":181}`, UpdatedOffset: 26 * time.Minute},
							{TeamIndex: 2, ServiceStatus: model.AWDServiceStatusUp, AttackReceived: 1, SLAScore: 28, DefenseScore: 35, AttackScore: 18, CheckerType: model.AWDCheckerTypeHTTPStandard, CheckResult: `{"status":"ok","latency_ms":96}`, UpdatedOffset: 27 * time.Minute},
						},
						Attacks: []awdAttackSeed{
							{Offset: 12 * time.Minute, AttackerTeamIndex: 0, VictimTeamIndex: 1, SubmittedFlag: "awd{web-replay-miss}", Source: model.AWDAttackSourceSubmission, SubmittedByStudent: true, IsSuccess: false, ScoreGained: 0},
							{Offset: 19 * time.Minute, AttackerTeamIndex: 2, VictimTeamIndex: 0, AttackType: model.AWDAttackTypeServiceExploit, Source: model.AWDAttackSourceManual, IsSuccess: true, ScoreGained: 90},
						},
						Traffic: []awdTrafficSeed{
							{Offset: 8 * time.Minute, AttackerTeamIndex: 0, VictimTeamIndex: 1, Method: "GET", Path: "/health", StatusCode: 200, Source: model.AWDTrafficSourceRuntimeProxy},
							{Offset: 14 * time.Minute, AttackerTeamIndex: 0, VictimTeamIndex: 1, Method: "POST", Path: "/api/replay", StatusCode: 403, Source: model.AWDTrafficSourceRuntimeProxy},
							{Offset: 20 * time.Minute, AttackerTeamIndex: 2, VictimTeamIndex: 0, Method: "GET", Path: "/debug", StatusCode: 200, Source: model.AWDAttackSourceManual},
						},
					},
					{
						RoundNumber:  2,
						StartOffset:  40 * time.Minute,
						Duration:     34 * time.Minute,
						AttackScore:  90,
						DefenseScore: 45,
						Services: []awdServiceSeed{
							{TeamIndex: 0, ServiceStatus: model.AWDServiceStatusUp, AttackReceived: 1, SLAScore: 35, DefenseScore: 45, AttackScore: 32, CheckerType: model.AWDCheckerTypeHTTPStandard, CheckResult: `{"status":"ok","latency_ms":72}`, UpdatedOffset: 28 * time.Minute},
							{TeamIndex: 1, ServiceStatus: model.AWDServiceStatusDown, AttackReceived: 3, SLAScore: 10, DefenseScore: 12, AttackScore: 5, CheckerType: model.AWDCheckerTypeHTTPStandard, CheckResult: `{"status":"down","error":"timeout"}`, UpdatedOffset: 30 * time.Minute},
							{TeamIndex: 2, ServiceStatus: model.AWDServiceStatusCompromised, AttackReceived: 2, SLAScore: 22, DefenseScore: 26, AttackScore: 18, CheckerType: model.AWDCheckerTypeHTTPStandard, CheckResult: `{"status":"partial","latency_ms":204}`, UpdatedOffset: 31 * time.Minute},
						},
						Attacks: []awdAttackSeed{
							{Offset: 9 * time.Minute, AttackerTeamIndex: 0, VictimTeamIndex: 1, SubmittedFlag: "awd{web-replay-hit}", Source: model.AWDAttackSourceSubmission, SubmittedByStudent: true, IsSuccess: true, ScoreGained: 120},
							{Offset: 17 * time.Minute, AttackerTeamIndex: 1, VictimTeamIndex: 2, Source: model.AWDAttackSourceManual, IsSuccess: false, ScoreGained: 0},
							{Offset: 24 * time.Minute, AttackerTeamIndex: 2, VictimTeamIndex: 0, AttackType: model.AWDAttackTypeServiceExploit, Source: model.AWDAttackSourceLegacy, IsSuccess: true, ScoreGained: 60},
						},
						Traffic: []awdTrafficSeed{
							{Offset: 6 * time.Minute, AttackerTeamIndex: 0, VictimTeamIndex: 1, Method: "GET", Path: "/flag", StatusCode: 200, Source: model.AWDTrafficSourceRuntimeProxy},
							{Offset: 10 * time.Minute, AttackerTeamIndex: 0, VictimTeamIndex: 1, Method: "POST", Path: "/exploit", StatusCode: 200, Source: model.AWDTrafficSourceRuntimeProxy},
							{Offset: 25 * time.Minute, AttackerTeamIndex: 2, VictimTeamIndex: 0, Method: "GET", Path: "/backup", StatusCode: 302, Source: model.AWDTrafficSourceRuntimeProxy},
						},
					},
					{
						RoundNumber:  3,
						StartOffset:  82 * time.Minute,
						Duration:     30 * time.Minute,
						AttackScore:  70,
						DefenseScore: 60,
						Services: []awdServiceSeed{
							{TeamIndex: 0, ServiceStatus: model.AWDServiceStatusUp, AttackReceived: 0, SLAScore: 40, DefenseScore: 48, AttackScore: 28, CheckerType: model.AWDCheckerTypeHTTPStandard, CheckResult: `{"status":"ok","latency_ms":69}`, UpdatedOffset: 23 * time.Minute},
							{TeamIndex: 1, ServiceStatus: model.AWDServiceStatusUp, AttackReceived: 1, SLAScore: 34, DefenseScore: 40, AttackScore: 16, CheckerType: model.AWDCheckerTypeHTTPStandard, CheckResult: `{"status":"ok","latency_ms":87}`, UpdatedOffset: 24 * time.Minute},
							{TeamIndex: 2, ServiceStatus: model.AWDServiceStatusDown, AttackReceived: 2, SLAScore: 8, DefenseScore: 10, AttackScore: 6, CheckerType: model.AWDCheckerTypeHTTPStandard, CheckResult: `{"status":"down","error":"container_exit"}`, UpdatedOffset: 25 * time.Minute},
						},
						Attacks: []awdAttackSeed{
							{Offset: 11 * time.Minute, AttackerTeamIndex: 0, VictimTeamIndex: 2, SubmittedFlag: "awd{web-lateral-sync}", Source: model.AWDAttackSourceSubmission, SubmittedByStudent: true, IsSuccess: true, ScoreGained: 100},
							{Offset: 18 * time.Minute, AttackerTeamIndex: 1, VictimTeamIndex: 0, Source: model.AWDAttackSourceManual, IsSuccess: false, ScoreGained: 0},
						},
						Traffic: []awdTrafficSeed{
							{Offset: 7 * time.Minute, AttackerTeamIndex: 0, VictimTeamIndex: 2, Method: "GET", Path: "/metrics", StatusCode: 200, Source: model.AWDTrafficSourceRuntimeProxy},
							{Offset: 12 * time.Minute, AttackerTeamIndex: 0, VictimTeamIndex: 2, Method: "POST", Path: "/sync", StatusCode: 201, Source: model.AWDTrafficSourceRuntimeProxy},
							{Offset: 19 * time.Minute, AttackerTeamIndex: 1, VictimTeamIndex: 0, Method: "GET", Path: "/health", StatusCode: 200, Source: model.AWDTrafficSourceRuntimeProxy},
						},
					},
				},
			},
		},
		{
			Label: "高试错后成功，但缺少复盘",
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
						{Offset: 79 * time.Minute, Flag: "flag{still-not-right}", Correct: false},
						{Offset: 96 * time.Minute, Flag: "flag{crypto-postcard-shift-3}", Correct: true},
					},
				},
			},
		},
		{
			Label: "单维弱项暴露型",
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
			Label: "低活跃掉队型",
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
			Label: "已形成闭环，但 Web 仍需补基础",
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
			Label: "早期完成后停滞，补样本后可形成明确弱项",
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
				{
					ChallengeCategory: "forensics",
					ChallengeIndex:    0,
					StartOffset:       -2*24*time.Hour + 150*time.Minute,
					Duration:          42 * time.Minute,
					Access:            true,
					ProxyRequests: []proxySeed{
						{Offset: 6 * time.Minute, Method: "GET", Path: "/mailbox.eml", Status: 200},
					},
					Submissions: []submissionSeed{
						{Offset: 29 * time.Minute, Flag: "flag{forensics-halfway}", Correct: false},
					},
				},
			},
		},
		{
			Label: "高频试错未命中型",
			User: userSeed{
				Username:  "songyuehan",
				Name:      "宋月涵",
				Email:     "2024310107@xinan.example.edu.cn",
				Role:      model.RoleStudent,
				ClassName: seedClassName,
				StudentNo: "2024310107",
			},
			Profiles: map[string]float64{
				"web":       0.51,
				"crypto":    0.28,
				"forensics": 0.47,
				"misc":      0.56,
				"pwn":       0.39,
				"reverse":   0.45,
			},
			Sessions: []sessionSeed{
				{
					ChallengeCategory: "crypto",
					ChallengeIndex:    1,
					StartOffset:       -36 * time.Hour,
					Duration:          115 * time.Minute,
					Access:            true,
					ProxyRequests: []proxySeed{
						{Offset: 8 * time.Minute, Method: "GET", Path: "/backup.bin", Status: 200},
						{Offset: 22 * time.Minute, Method: "POST", Path: "/xor", Status: 400, PayloadPreview: "key=0x12"},
						{Offset: 37 * time.Minute, Method: "POST", Path: "/xor", Status: 400, PayloadPreview: "key=0x6f"},
						{Offset: 53 * time.Minute, Method: "POST", Path: "/stream", Status: 500, PayloadPreview: "nonce=15"},
					},
					Submissions: []submissionSeed{
						{Offset: 28 * time.Minute, Flag: "flag{stream-fail-1}", Correct: false},
						{Offset: 46 * time.Minute, Flag: "flag{stream-fail-2}", Correct: false},
						{Offset: 69 * time.Minute, Flag: "flag{stream-fail-3}", Correct: false},
						{Offset: 92 * time.Minute, Flag: "flag{stream-fail-4}", Correct: false},
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
	_ = teacherID
	seedUserIDs, extraUserIDs, err := collectSeedUserIDs(tx, userIDs)
	if err != nil {
		return err
	}
	if len(seedUserIDs) == 0 {
		return nil
	}
	if err := resetSeededAWDData(tx); err != nil {
		return err
	}
	if err := tx.Where("user_id IN ?", seedUserIDs).Delete(&model.AuditLog{}).Error; err != nil {
		return fmt.Errorf("delete audit logs: %w", err)
	}
	if err := tx.Where("user_id IN ?", seedUserIDs).Delete(&model.SubmissionWriteup{}).Error; err != nil {
		return fmt.Errorf("delete submission writeups: %w", err)
	}
	if err := tx.Where("user_id IN ? AND contest_id IS NULL", seedUserIDs).Delete(&model.Submission{}).Error; err != nil {
		return fmt.Errorf("delete submissions: %w", err)
	}
	if err := tx.Where("user_id IN ? AND contest_id IS NULL", seedUserIDs).Delete(&model.Instance{}).Error; err != nil {
		return fmt.Errorf("delete instances: %w", err)
	}
	if err := tx.Where("user_id IN ?", seedUserIDs).Delete(&model.SkillProfile{}).Error; err != nil {
		return fmt.Errorf("delete skill profiles: %w", err)
	}
	if err := tx.Where("class_name = ? OR user_id IN ?", seedClassName, seedUserIDs).Delete(&model.Report{}).Error; err != nil {
		return fmt.Errorf("delete reports: %w", err)
	}
	if len(extraUserIDs) > 0 {
		if err := tx.Where("user_id IN ?", extraUserIDs).Delete(&model.UserRole{}).Error; err != nil {
			return fmt.Errorf("delete stale user roles: %w", err)
		}
		if err := tx.Unscoped().Where("id IN ?", extraUserIDs).Delete(&model.User{}).Error; err != nil {
			return fmt.Errorf("delete stale seed users: %w", err)
		}
	}
	return nil
}

func collectSeedUserIDs(tx *gorm.DB, currentUserIDs []int64) ([]int64, []int64, error) {
	var seedUsers []model.User
	if err := tx.Unscoped().
		Where("class_name = ? OR username = ?", seedClassName, seedTeacherUsername).
		Find(&seedUsers).Error; err != nil {
		return nil, nil, fmt.Errorf("find seed users: %w", err)
	}

	if len(seedUsers) == 0 {
		return nil, nil, nil
	}

	keepSet := make(map[int64]struct{}, len(currentUserIDs))
	for _, userID := range currentUserIDs {
		keepSet[userID] = struct{}{}
	}

	seedUserIDs := make([]int64, 0, len(seedUsers))
	extraUserIDs := make([]int64, 0, len(seedUsers))
	for _, user := range seedUsers {
		seedUserIDs = append(seedUserIDs, user.ID)
		if _, ok := keepSet[user.ID]; ok {
			continue
		}
		extraUserIDs = append(extraUserIDs, user.ID)
	}
	return seedUserIDs, extraUserIDs, nil
}

func resetSeededAWDData(tx *gorm.DB) error {
	var contestIDs []int64
	if err := tx.Unscoped().Model(&model.Contest{}).
		Where("mode = ? AND title LIKE ?", model.ContestModeAWD, seedAWDContestTitle+"%").
		Pluck("id", &contestIDs).Error; err != nil {
		return fmt.Errorf("find seeded awd contests: %w", err)
	}
	if len(contestIDs) == 0 {
		return nil
	}

	var roundIDs []int64
	if err := tx.Model(&model.AWDRound{}).Where("contest_id IN ?", contestIDs).Pluck("id", &roundIDs).Error; err != nil {
		return fmt.Errorf("find seeded awd rounds: %w", err)
	}

	if err := tx.Where("contest_id IN ?", contestIDs).Delete(&model.AWDTrafficEvent{}).Error; err != nil {
		return fmt.Errorf("delete awd traffic events: %w", err)
	}
	if len(roundIDs) > 0 {
		if err := tx.Where("round_id IN ?", roundIDs).Delete(&model.AWDTeamService{}).Error; err != nil {
			return fmt.Errorf("delete awd team services: %w", err)
		}
		if err := tx.Where("round_id IN ?", roundIDs).Delete(&model.AWDAttackLog{}).Error; err != nil {
			return fmt.Errorf("delete awd attack logs: %w", err)
		}
	}
	if err := tx.Where("contest_id IN ?", contestIDs).Delete(&model.TeamMember{}).Error; err != nil {
		return fmt.Errorf("delete awd team members: %w", err)
	}
	if err := tx.Unscoped().Where("contest_id IN ?", contestIDs).Delete(&model.Team{}).Error; err != nil {
		return fmt.Errorf("delete awd teams: %w", err)
	}
	if err := tx.Where("contest_id IN ?", contestIDs).Delete(&model.AWDRound{}).Error; err != nil {
		return fmt.Errorf("delete awd rounds: %w", err)
	}
	if err := tx.Unscoped().Where("id IN ?", contestIDs).Delete(&model.Contest{}).Error; err != nil {
		return fmt.Errorf("delete awd contests: %w", err)
	}
	return nil
}

func seedStudentScenario(
	tx *gorm.DB,
	teacher *model.User,
	student *model.User,
	scenario studentScenario,
	catalog *challengeCatalog,
	awdCatalog *awdChallengeCatalog,
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
	if scenario.AWD != nil {
		if err := seedStudentAWDScenario(tx, teacher, student, scenario.AWD, awdCatalog, now); err != nil {
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

func seedStudentAWDScenario(
	tx *gorm.DB,
	teacher *model.User,
	student *model.User,
	scenario *awdScenario,
	catalog *awdChallengeCatalog,
	base time.Time,
) error {
	if scenario == nil {
		return nil
	}
	if len(scenario.Teams) == 0 {
		return errors.New("awd scenario requires teams")
	}
	if len(scenario.Rounds) == 0 {
		return errors.New("awd scenario requires rounds")
	}
	challenge, err := catalog.pick(scenario.ChallengeCategory, scenario.ChallengeIndex)
	if err != nil {
		return err
	}

	contestStart := base.Add(scenario.StartOffset)
	contestEnd := seedAWDScenarioEndAt(contestStart, scenario.Rounds)
	contest := &model.Contest{
		Title:         fmt.Sprintf("%s - %s", seedAWDContestTitle, student.Username),
		Description:   "用于教学复盘样本的 AWD 迁移演练数据。",
		Mode:          model.ContestModeAWD,
		StartTime:     contestStart.Add(-30 * time.Minute),
		EndTime:       contestEnd,
		Status:        model.ContestStatusEnded,
		StatusVersion: 1,
		CreatedAt:     contestStart.Add(-45 * time.Minute),
		UpdatedAt:     contestEnd,
	}
	if err := tx.Create(contest).Error; err != nil {
		return fmt.Errorf("create seed awd contest for %s: %w", student.Username, err)
	}

	serviceID := 7000 + challenge.ID
	teams := make([]*model.Team, 0, len(scenario.Teams))
	teamScores := make(map[int64]int, len(scenario.Teams))
	teamLastSolveAt := make(map[int64]time.Time, len(scenario.Teams))

	for teamIndex, teamSeed := range scenario.Teams {
		memberCount := teamSeed.MemberCount
		if memberCount < 1 {
			memberCount = 1
		}
		team := &model.Team{
			ContestID:  contest.ID,
			Name:       teamSeed.Name,
			CaptainID:  resolveSeedAWDCaptainID(teamSeed.CaptainRole, teacher, student, contest.ID, teamIndex),
			InviteCode: fmt.Sprintf("A%04d%d", student.ID%10000, teamIndex),
			MaxMembers: max(memberCount, 4),
			CreatedAt:  contestStart.Add(-20 * time.Minute),
			UpdatedAt:  contestEnd,
		}
		if err := tx.Create(team).Error; err != nil {
			return fmt.Errorf("create awd team %d for %s: %w", teamIndex, student.Username, err)
		}
		teams = append(teams, team)

		for memberIndex := 0; memberIndex < memberCount; memberIndex++ {
			memberID := seedAWDTeamMemberID(contest.ID, teamIndex, memberIndex)
			if memberIndex == 0 {
				memberID = team.CaptainID
			}
			if err := tx.Create(&model.TeamMember{
				ContestID: contest.ID,
				TeamID:    team.ID,
				UserID:    memberID,
				JoinedAt:  contestStart.Add(-18 * time.Minute),
				CreatedAt: contestStart.Add(-18 * time.Minute),
			}).Error; err != nil {
				return fmt.Errorf("create awd team member %d for %s: %w", memberIndex, team.Name, err)
			}
		}
	}

	for _, roundSeed := range scenario.Rounds {
		roundStart := contestStart.Add(roundSeed.StartOffset)
		roundEnd := roundStart.Add(roundSeed.Duration)
		round := &model.AWDRound{
			ContestID:    contest.ID,
			RoundNumber:  roundSeed.RoundNumber,
			Status:       model.AWDRoundStatusFinished,
			StartedAt:    timePtr(roundStart),
			EndedAt:      timePtr(roundEnd),
			AttackScore:  roundSeed.AttackScore,
			DefenseScore: roundSeed.DefenseScore,
			CreatedAt:    roundStart.Add(-2 * time.Minute),
			UpdatedAt:    roundEnd,
		}
		if err := tx.Create(round).Error; err != nil {
			return fmt.Errorf("create awd round %d for %s: %w", roundSeed.RoundNumber, student.Username, err)
		}

		for serviceIndex, serviceSeed := range roundSeed.Services {
			team, err := seedAWDTeamAt(teams, serviceSeed.TeamIndex)
			if err != nil {
				return fmt.Errorf("round %d service team: %w", roundSeed.RoundNumber, err)
			}
			updatedAt := roundStart.Add(serviceSeed.UpdatedOffset)
			if serviceSeed.UpdatedOffset <= 0 {
				updatedAt = roundEnd.Add(-time.Minute)
			}
			checkResult := strings.TrimSpace(serviceSeed.CheckResult)
			if checkResult == "" {
				checkResult = `{}`
			}
			if err := tx.Create(&model.AWDTeamService{
				RoundID:        round.ID,
				TeamID:         team.ID,
				ServiceID:      serviceID,
				AWDChallengeID: challenge.ID,
				ServiceStatus:  serviceSeed.ServiceStatus,
				CheckResult:    checkResult,
				CheckerType:    serviceSeed.CheckerType,
				AttackReceived: serviceSeed.AttackReceived,
				SLAScore:       serviceSeed.SLAScore,
				DefenseScore:   serviceSeed.DefenseScore,
				AttackScore:    serviceSeed.AttackScore,
				CreatedAt:      roundStart.Add(time.Duration(serviceIndex+1) * time.Minute),
				UpdatedAt:      updatedAt,
			}).Error; err != nil {
				return fmt.Errorf("create awd team service for %s round %d: %w", team.Name, roundSeed.RoundNumber, err)
			}
			teamScores[team.ID] += serviceSeed.SLAScore + serviceSeed.DefenseScore + serviceSeed.AttackScore
		}

		for attackIndex, attackSeed := range roundSeed.Attacks {
			attackerTeam, err := seedAWDTeamAt(teams, attackSeed.AttackerTeamIndex)
			if err != nil {
				return fmt.Errorf("round %d attack attacker team: %w", roundSeed.RoundNumber, err)
			}
			victimTeam, err := seedAWDTeamAt(teams, attackSeed.VictimTeamIndex)
			if err != nil {
				return fmt.Errorf("round %d attack victim team: %w", roundSeed.RoundNumber, err)
			}
			attackAt := roundStart.Add(attackSeed.Offset)
			attackType := strings.TrimSpace(attackSeed.AttackType)
			if attackType == "" {
				attackType = model.AWDAttackTypeFlagCapture
			}
			source := strings.TrimSpace(attackSeed.Source)
			if source == "" {
				source = model.AWDAttackSourceManual
			}
			var submittedBy *int64
			if attackSeed.SubmittedByStudent {
				submittedBy = int64Ptr(student.ID)
			}
			if err := tx.Create(&model.AWDAttackLog{
				RoundID:           round.ID,
				AttackerTeamID:    attackerTeam.ID,
				VictimTeamID:      victimTeam.ID,
				ServiceID:         serviceID,
				AWDChallengeID:    challenge.ID,
				AttackType:        attackType,
				Source:            source,
				SubmittedFlag:     attackSeed.SubmittedFlag,
				SubmittedByUserID: submittedBy,
				IsSuccess:         attackSeed.IsSuccess,
				ScoreGained:       attackSeed.ScoreGained,
				CreatedAt:         attackAt.Add(time.Duration(attackIndex) * time.Second),
			}).Error; err != nil {
				return fmt.Errorf("create awd attack log for %s/%s: %w", attackerTeam.Name, challenge.Name, err)
			}
			if attackSeed.IsSuccess {
				teamScores[attackerTeam.ID] += attackSeed.ScoreGained
				if lastSolveAt, ok := teamLastSolveAt[attackerTeam.ID]; !ok || attackAt.After(lastSolveAt) {
					teamLastSolveAt[attackerTeam.ID] = attackAt
				}
			}
		}

		for trafficIndex, trafficSeed := range roundSeed.Traffic {
			attackerTeam, err := seedAWDTeamAt(teams, trafficSeed.AttackerTeamIndex)
			if err != nil {
				return fmt.Errorf("round %d traffic attacker team: %w", roundSeed.RoundNumber, err)
			}
			victimTeam, err := seedAWDTeamAt(teams, trafficSeed.VictimTeamIndex)
			if err != nil {
				return fmt.Errorf("round %d traffic victim team: %w", roundSeed.RoundNumber, err)
			}
			source := strings.TrimSpace(trafficSeed.Source)
			if source == "" {
				source = model.AWDTrafficSourceRuntimeProxy
			}
			if err := tx.Create(&model.AWDTrafficEvent{
				ContestID:      contest.ID,
				RoundID:        round.ID,
				AttackerTeamID: attackerTeam.ID,
				VictimTeamID:   victimTeam.ID,
				ServiceID:      serviceID,
				AWDChallengeID: challenge.ID,
				Method:         trafficSeed.Method,
				Path:           trafficSeed.Path,
				StatusCode:     trafficSeed.StatusCode,
				Source:         source,
				CreatedAt:      roundStart.Add(trafficSeed.Offset).Add(time.Duration(trafficIndex) * time.Second),
			}).Error; err != nil {
				return fmt.Errorf("create awd traffic event for %s/%s: %w", attackerTeam.Name, challenge.Name, err)
			}
		}
	}

	for _, team := range teams {
		team.TotalScore = teamScores[team.ID]
		if lastSolveAt, ok := teamLastSolveAt[team.ID]; ok {
			team.LastSolveAt = timePtr(lastSolveAt)
		}
		if err := tx.Save(team).Error; err != nil {
			return fmt.Errorf("update awd team score for %s: %w", team.Name, err)
		}
	}

	return nil
}

func seedAWDScenarioEndAt(contestStart time.Time, rounds []awdRoundSeed) time.Time {
	contestEnd := contestStart.Add(2 * time.Hour)
	for _, round := range rounds {
		roundEnd := contestStart.Add(round.StartOffset).Add(round.Duration)
		if roundEnd.After(contestEnd) {
			contestEnd = roundEnd
		}
	}
	return contestEnd
}

func resolveSeedAWDCaptainID(role string, teacher *model.User, student *model.User, contestID int64, teamIndex int) int64 {
	switch strings.TrimSpace(role) {
	case seedAWDCaptainTeacher:
		return teacher.ID
	case seedAWDCaptainStudent:
		return student.ID
	default:
		return seedAWDTeamMemberID(contestID, teamIndex, 0)
	}
}

func seedAWDTeamMemberID(contestID int64, teamIndex, memberIndex int) int64 {
	return 9_000_000_000 + contestID*100 + int64(teamIndex*10+memberIndex)
}

func seedAWDTeamAt(teams []*model.Team, index int) (*model.Team, error) {
	if index < 0 || index >= len(teams) {
		return nil, fmt.Errorf("team index %d out of range", index)
	}
	return teams[index], nil
}

func (c *challengeCatalog) pick(category string, index int) (challengeRef, error) {
	normalized := strings.ToLower(strings.TrimSpace(category))
	items := c.byCategory[normalized]
	if index < 0 || index >= len(items) {
		return challengeRef{}, fmt.Errorf("challenge category %s does not have index %d", normalized, index)
	}
	return items[index], nil
}

func (c *challengeCatalog) totalCount() int {
	if c == nil {
		return 0
	}
	total := 0
	for _, items := range c.byCategory {
		total += len(items)
	}
	return total
}

func (c *awdChallengeCatalog) pick(category string, index int) (awdChallengeRef, error) {
	normalized := strings.ToLower(strings.TrimSpace(category))
	items := c.byCategory[normalized]
	if index < 0 || index >= len(items) {
		return awdChallengeRef{}, fmt.Errorf("awd challenge category %s does not have index %d", normalized, index)
	}
	return items[index], nil
}

func countAWDAttacks(scenario *awdScenario) int {
	if scenario == nil {
		return 0
	}
	total := 0
	for _, round := range scenario.Rounds {
		total += len(round.Attacks)
	}
	return total
}

func buildSeedCoverageSummary(
	catalog *challengeCatalog,
	scenarios []studentScenario,
	results []seededStudentResult,
) seedCoverageSummary {
	summary := seedCoverageSummary{
		ByCategory: make(map[string]categoryCoverage, len(model.AllDimensions)),
	}
	if catalog == nil {
		return summary
	}

	usedByCategory := make(map[string]map[int64]struct{}, len(model.AllDimensions))
	usedPracticeChallenges := make(map[int64]struct{})
	for _, dimension := range model.AllDimensions {
		items := catalog.byCategory[dimension]
		summary.PublishedChallenges += len(items)
		summary.ByCategory[dimension] = categoryCoverage{Published: len(items)}
		usedByCategory[dimension] = make(map[int64]struct{})
	}

	for _, scenario := range scenarios {
		for _, session := range scenario.Sessions {
			challenge, err := catalog.pick(session.ChallengeCategory, session.ChallengeIndex)
			if err != nil {
				continue
			}
			usedPracticeChallenges[challenge.ID] = struct{}{}
			usedByCategory[challenge.Category][challenge.ID] = struct{}{}
		}
	}
	summary.UsedPracticeChallenges = len(usedPracticeChallenges)

	uniqueTopRecommendations := make(map[int64]struct{})
	for _, result := range results {
		if len(result.Recommendations.Challenges) == 0 {
			continue
		}
		summary.StudentsWithRecommendations++
		top := result.Recommendations.Challenges[0]
		uniqueTopRecommendations[top.ChallengeID] = struct{}{}
	}
	summary.UniqueTopRecommendationCount = len(uniqueTopRecommendations)

	for _, dimension := range model.AllDimensions {
		entry := summary.ByCategory[dimension]
		entry.Used = len(usedByCategory[dimension])
		summary.ByCategory[dimension] = entry
	}

	return summary
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
	fmt.Printf("样本规模: %d 名学生，%d 个练习会话，%d 条 AWD 攻击记录\n", len(result.Students), result.PracticeSessionCount, result.AWDAttackCount)
	fmt.Printf(
		"题库覆盖: published=%d, practice_used=%d, recommendation_ready=%d, unique_top_recommendations=%d\n",
		result.Coverage.PublishedChallenges,
		result.Coverage.UsedPracticeChallenges,
		result.Coverage.StudentsWithRecommendations,
		result.Coverage.UniqueTopRecommendationCount,
	)
	fmt.Println("分类覆盖:")
	for _, dimension := range model.AllDimensions {
		coverage, ok := result.Coverage.ByCategory[dimension]
		if !ok {
			continue
		}
		fmt.Printf("  - %s: published=%d, used=%d\n", dimension, coverage.Published, coverage.Used)
	}
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
		if student.ScenarioLabel != "" {
			fmt.Printf("  样本画像: %s\n", student.ScenarioLabel)
		}
		fmt.Printf("  训练足迹: practice=%d, awd=%d\n", student.PracticeSessionCount, student.AWDAttackCount)
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
