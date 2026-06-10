package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestentity "ctf-platform/internal/module/contest/entity"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	opsentity "ctf-platform/internal/module/ops/entity"
	practiceentity "ctf-platform/internal/module/practice/entity"
	"ctf-platform/internal/platform/randomstring"
	flagcrypto "ctf-platform/internal/shared/flagcrypto"
	"ctf-platform/internal/shared/taxonomy"
)

type fullRouterTestEnv struct {
	router *gin.Engine
	db     *gorm.DB
	cache  *redislib.Client

	admin        *identitycontracts.User
	teacher      *identitycontracts.User
	student      *identitycontracts.User
	peerStudent  *identitycontracts.User
	otherTeacher *identitycontracts.User
	otherStudent *identitycontracts.User
	studentPwd   string
	teacherPwd   string
	adminPwd     string
	className    string
	reportDir    string
	image        *appImageRow
	challenge    *appChallengeRow
	template     *challengeentity.EnvironmentTemplate
	contest      *contestcontracts.Contest
	awdContest   *contestcontracts.Contest
	registration *contestcontracts.ContestRegistration
	announcement *contestentity.ContestAnnouncement
	team         *contestcontracts.Team
	awdRound     *contestcontracts.AWDRound
	instance     *instancecontracts.Instance
	notification *opsentity.Notification
	report       *assessmententity.Report
}

func newFullRouterTestEnv(t *testing.T) *fullRouterTestEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	cache := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = cache.Close() })

	db := openFullRouterTestDB(t)
	t.Cleanup(func() {
		if sqlDB, sqlErr := db.DB(); sqlErr == nil {
			_ = sqlDB.Close()
		}
	})

	cfg := newFullRouterTestConfig(t)
	router, err := NewRouter(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	env := &fullRouterTestEnv{
		router:     router,
		db:         db,
		cache:      cache,
		adminPwd:   "Password123",
		teacherPwd: "Password123",
		studentPwd: "Password123",
		className:  "ClassA",
		reportDir:  cfg.Report.StorageDir,
	}

	seedFullRouterData(t, env)
	return env
}

func openFullRouterTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return openInternalAppTestSQLite(t, "full-router.sqlite")
}

func newFullRouterTestConfig(t *testing.T) *config.Config {
	t.Helper()

	cfg := newPracticeFlowTestConfig(t)
	cfg.RateLimit.Global.Enabled = false
	cfg.RateLimit.Login.Enabled = false
	cfg.Score = config.ScoreConfig{
		CacheTTL:        time.Minute,
		LockTimeout:     2 * time.Second,
		MaxRankingLimit: 100,
	}
	cfg.Recommendation = config.RecommendationConfig{
		WeakThreshold: 0.4,
		CacheTTL:      time.Hour,
		DefaultLimit:  6,
		MaxLimit:      20,
	}
	cfg.Report = config.ReportConfig{
		StorageDir:      filepath.Join(t.TempDir(), "reports"),
		DefaultFormat:   assessmententity.ReportFormatPDF,
		PersonalTimeout: 10 * time.Second,
		ClassTimeout:    10 * time.Second,
		FileTTL:         24 * time.Hour,
		MaxWorkers:      1,
	}
	cfg.Dashboard = config.DashboardConfig{
		CacheTTL:       time.Minute,
		AlertThreshold: 80,
		RedisKeyPrefix: "test:dashboard",
	}
	cfg.Contest = config.ContestConfig{
		StatusUpdateInterval:  time.Minute,
		StatusUpdateBatchSize: 100,
		BaseScore:             1000,
		MinScore:              100,
		Decay:                 0.9,
		FirstBloodBonus:       0.1,
		AWD: config.ContestAWDConfig{
			SchedulerInterval:  30 * time.Second,
			SchedulerBatchSize: 100,
			RoundInterval:      5 * time.Minute,
			RoundLockTTL:       30 * time.Second,
			PreviousRoundGrace: 15 * time.Second,
			CheckerTimeout:     2 * time.Second,
			CheckerHealthPath:  "/health",
		},
	}
	return cfg
}

func seedFullRouterData(t *testing.T, env *fullRouterTestEnv) {
	t.Helper()

	seedRoles(t, env.db)

	env.admin = createFullRouterUser(t, env.db, "admin_matrix", env.adminPwd, identitycontracts.RoleAdmin, "")
	env.teacher = createFullRouterUser(t, env.db, "teacher_matrix", env.teacherPwd, identitycontracts.RoleTeacher, env.className)
	env.student = createFullRouterUser(t, env.db, "student_matrix", env.studentPwd, identitycontracts.RoleStudent, env.className)
	env.peerStudent = createFullRouterUser(t, env.db, "student_peer", "Password123", identitycontracts.RoleStudent, env.className)
	env.otherTeacher = createFullRouterUser(t, env.db, "teacher_other", "Password123", identitycontracts.RoleTeacher, "ClassB")
	env.otherStudent = createFullRouterUser(t, env.db, "student_other", "Password123", identitycontracts.RoleStudent, "ClassB")

	env.image = createFlowImage(t, env.db)

	salt, err := randomstring.Generate()
	if err != nil {
		t.Fatalf("generate flag salt: %v", err)
	}
	env.challenge = &appChallengeRow{
		Title:         "Matrix Web Challenge",
		Description:   "challenge for full router integration tests",
		Category:      taxonomy.DimensionWeb,
		Difficulty:    taxonomy.DifficultyEasy,
		Points:        100,
		ImageID:       env.image.ID,
		Status:        challengecontracts.ChallengeStatusPublished,
		FlagType:      challengecontracts.FlagTypeStatic,
		FlagSalt:      salt,
		FlagHash:      flagcrypto.HashStaticFlag("flag{matrix}", salt),
		FlagPrefix:    "flag",
		AttachmentURL: "https://example.com/files/matrix.zip",
		CreatedBy:     &env.teacher.ID,
	}
	if err := env.db.Create(env.challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	hint := &challengeentity.ChallengeHint{
		ChallengeID: env.challenge.ID,
		Level:       1,
		Title:       "入口提示",
		Content:     "先查看登录表单。",
	}
	if err := env.db.Create(hint).Error; err != nil {
		t.Fatalf("create hint: %v", err)
	}

	writeup := &challengeentity.ChallengeWriteup{
		ChallengeID: env.challenge.ID,
		Title:       "题解",
		Content:     "writeup content",
		Visibility:  challengeentity.WriteupVisibilityPublic,
		CreatedBy:   &env.admin.ID,
	}
	if err := env.db.Create(writeup).Error; err != nil {
		t.Fatalf("create writeup: %v", err)
	}

	spec, err := challengecontracts.EncodeTopologySpec(challengecontracts.TopologySpec{
		Networks: []challengecontracts.TopologyNetwork{{Key: challengecontracts.TopologyDefaultNetworkKey, Name: "default"}},
		Nodes: []challengecontracts.TopologyNode{{
			Key:         "web",
			Name:        "Web Node",
			ImageID:     env.image.ID,
			ServicePort: 80,
			InjectFlag:  true,
			Tier:        challengecontracts.TopologyTierPublic,
			NetworkKeys: []string{challengecontracts.TopologyDefaultNetworkKey},
		}},
	})
	if err != nil {
		t.Fatalf("encode topology: %v", err)
	}

	env.template = &challengeentity.EnvironmentTemplate{
		Name:         "Matrix Template",
		Description:  "template for integration tests",
		EntryNodeKey: "web",
		Spec:         spec,
	}
	if err := env.db.Create(env.template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}

	now := time.Now()
	env.contest = &contestcontracts.Contest{
		Title:       "Matrix Jeopardy",
		Description: "contest",
		Mode:        contestcontracts.ContestModeJeopardy,
		StartTime:   now.Add(-time.Hour),
		EndTime:     now.Add(time.Hour),
		Status:      contestcontracts.ContestStatusRunning,
	}
	if err := env.db.Create(env.contest).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}

	env.awdContest = &contestcontracts.Contest{
		Title:       "Matrix AWD",
		Description: "awd contest",
		Mode:        contestcontracts.ContestModeAWD,
		StartTime:   now.Add(-time.Hour),
		EndTime:     now.Add(time.Hour),
		Status:      contestcontracts.ContestStatusRunning,
	}
	if err := env.db.Create(env.awdContest).Error; err != nil {
		t.Fatalf("create awd contest: %v", err)
	}

	contestChallenge := &contestcontracts.ContestChallenge{
		ContestID:   env.contest.ID,
		ChallengeID: env.challenge.ID,
		Points:      100,
		Order:       1,
		IsVisible:   true,
	}
	if err := env.db.Create(contestChallenge).Error; err != nil {
		t.Fatalf("create contest challenge: %v", err)
	}
	awdContestChallenge := &contestcontracts.ContestChallenge{
		ContestID:   env.awdContest.ID,
		ChallengeID: env.challenge.ID,
		Points:      100,
		Order:       1,
		IsVisible:   true,
	}
	if err := env.db.Create(awdContestChallenge).Error; err != nil {
		t.Fatalf("create awd contest challenge: %v", err)
	}

	env.registration = &contestcontracts.ContestRegistration{
		ContestID: env.contest.ID,
		UserID:    env.student.ID,
		Status:    contestcontracts.ContestRegistrationStatusApproved,
	}
	if err := env.db.Create(env.registration).Error; err != nil {
		t.Fatalf("create registration: %v", err)
	}
	awdRegistration := &contestcontracts.ContestRegistration{
		ContestID: env.awdContest.ID,
		UserID:    env.student.ID,
		Status:    contestcontracts.ContestRegistrationStatusApproved,
	}
	if err := env.db.Create(awdRegistration).Error; err != nil {
		t.Fatalf("create awd registration: %v", err)
	}

	env.announcement = &contestentity.ContestAnnouncement{
		ContestID: env.contest.ID,
		Title:     "公告",
		Content:   "contest starts",
		CreatedBy: &env.admin.ID,
	}
	if err := env.db.Create(env.announcement).Error; err != nil {
		t.Fatalf("create announcement: %v", err)
	}

	env.team = &contestcontracts.Team{
		ContestID:  env.contest.ID,
		Name:       "Matrix Team",
		CaptainID:  env.student.ID,
		InviteCode: "MATRIX123",
		MaxMembers: 4,
	}
	if err := env.db.Create(env.team).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := env.db.Create(&contestcontracts.TeamMember{
		ContestID: env.contest.ID,
		TeamID:    env.team.ID,
		UserID:    env.student.ID,
		JoinedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create team member: %v", err)
	}
	env.registration.TeamID = &env.team.ID
	if err := env.db.Save(env.registration).Error; err != nil {
		t.Fatalf("update registration team: %v", err)
	}

	env.awdRound = &contestcontracts.AWDRound{
		ContestID:    env.awdContest.ID,
		RoundNumber:  1,
		Status:       contestcontracts.AWDRoundStatusRunning,
		StartedAt:    &now,
		AttackScore:  50,
		DefenseScore: 50,
	}
	if err := env.db.Create(env.awdRound).Error; err != nil {
		t.Fatalf("create awd round: %v", err)
	}
	if err := env.db.Create(&contestcontracts.AWDTeamService{
		RoundID:        env.awdRound.ID,
		TeamID:         env.team.ID,
		AWDChallengeID: env.challenge.ID,
		ServiceStatus:  contestcontracts.AWDServiceStatusUp,
		CheckResult:    `{"status":"ok"}`,
	}).Error; err != nil {
		t.Fatalf("create awd team service: %v", err)
	}
	if err := env.db.Create(&contestcontracts.AWDAttackLog{
		RoundID:        env.awdRound.ID,
		AttackerTeamID: env.team.ID,
		VictimTeamID:   env.team.ID,
		AWDChallengeID: env.challenge.ID,
		AttackType:     contestcontracts.AWDAttackTypeFlagCapture,
		Source:         contestcontracts.AWDAttackSourceManual,
		IsSuccess:      false,
	}).Error; err != nil {
		t.Fatalf("create awd attack log: %v", err)
	}

	runtimeDetails, err := runtimecontracts.EncodeInstanceRuntimeDetails(runtimecontracts.InstanceRuntimeDetails{
		Containers: []runtimecontracts.InstanceRuntimeContainer{{
			NodeKey:      "web",
			ContainerID:  "ctf-instance",
			ServicePort:  80,
			IsEntryPoint: true,
			HostPort:     30001,
		}},
	})
	if err != nil {
		t.Fatalf("encode runtime details: %v", err)
	}
	env.instance = &instancecontracts.Instance{
		UserID:         env.student.ID,
		ChallengeID:    env.challenge.ID,
		ContainerID:    "ctf-instance",
		NetworkID:      "ctf-network",
		RuntimeDetails: runtimeDetails,
		Status:         instancecontracts.InstanceStatusRunning,
		AccessURL:      "http://127.0.0.1:30001",
		Nonce:          "matrix-nonce",
		ExpiresAt:      now.Add(2 * time.Hour),
		MaxExtends:     2,
	}
	if err := env.db.Create(env.instance).Error; err != nil {
		t.Fatalf("create instance: %v", err)
	}

	if err := env.db.Create(&contestcontracts.Submission{
		UserID:      env.student.ID,
		ChallengeID: env.challenge.ID,
		IsCorrect:   true,
		SubmittedAt: now.Add(-10 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("create submission: %v", err)
	}
	if err := env.db.Create(&practiceentity.UserScore{
		UserID:     env.student.ID,
		TotalScore: 100,
	}).Error; err != nil {
		t.Fatalf("create user score: %v", err)
	}
	if err := env.db.Create(&assessmententity.SkillProfile{
		UserID:    env.student.ID,
		Dimension: taxonomy.DimensionWeb,
		Score:     0.3,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create skill profile: %v", err)
	}

	env.notification = &opsentity.Notification{
		UserID:    env.student.ID,
		Type:      opsentity.NotificationTypeSystem,
		Title:     "通知",
		Content:   "hello",
		IsRead:    false,
		CreatedAt: now,
	}
	if err := env.db.Create(env.notification).Error; err != nil {
		t.Fatalf("create notification: %v", err)
	}

	if err := os.MkdirAll(env.reportDir, 0o755); err != nil {
		t.Fatalf("mkdir report dir: %v", err)
	}
	reportPath := filepath.Join(env.reportDir, "personal-report.pdf")
	if err := os.WriteFile(reportPath, []byte("matrix report"), 0o644); err != nil {
		t.Fatalf("write report file: %v", err)
	}
	expiresAt := now.Add(24 * time.Hour)
	completedAt := now
	env.report = &assessmententity.Report{
		Type:        assessmententity.ReportTypePersonal,
		Format:      assessmententity.ReportFormatPDF,
		UserID:      &env.student.ID,
		Status:      assessmententity.ReportStatusReady,
		FilePath:    reportPath,
		ExpiresAt:   &expiresAt,
		CompletedAt: &completedAt,
	}
	if err := env.db.Create(env.report).Error; err != nil {
		t.Fatalf("create report: %v", err)
	}
}

func seedRoles(t *testing.T, db *gorm.DB) {
	t.Helper()

	roles := []*identitycontracts.Role{
		{Code: identitycontracts.RoleAdmin, Name: "管理员"},
		{Code: identitycontracts.RoleTeacher, Name: "教师"},
		{Code: identitycontracts.RoleStudent, Name: "学生"},
	}
	for _, role := range roles {
		if err := db.Create(role).Error; err != nil {
			t.Fatalf("create role %s: %v", role.Code, err)
		}
	}
}

func createFullRouterUser(t *testing.T, db *gorm.DB, username, password, role, className string) *identitycontracts.User {
	t.Helper()

	user := &identitycontracts.User{
		Username:  username,
		Email:     fmt.Sprintf("%s@example.com", username),
		Role:      role,
		Status:    identitycontracts.UserStatusActive,
		ClassName: className,
		Name:      username,
	}
	setTestPassword(t, user, password)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user %s: %v", username, err)
	}
	return user
}

func performFullRouterRequest(
	t *testing.T,
	router http.Handler,
	method string,
	target string,
	payload any,
	headers map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encode request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, target, &body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}
