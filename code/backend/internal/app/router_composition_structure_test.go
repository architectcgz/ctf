package app

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/auditlog"
	assessmenthttp "ctf-platform/internal/module/assessment/api/http"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeports "ctf-platform/internal/module/challenge/ports"
	contesthttp "ctf-platform/internal/module/contest/api/http"
	contestports "ctf-platform/internal/module/contest/ports"
	identityhttp "ctf-platform/internal/module/identity/api/http"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instancehttp "ctf-platform/internal/module/instance/api/http"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	opshttp "ctf-platform/internal/module/ops/api/http"
	opscmd "ctf-platform/internal/module/ops/application/commands"
	opsports "ctf-platform/internal/module/ops/ports"
	practicehttp "ctf-platform/internal/module/practice/api/http"
	practiceports "ctf-platform/internal/module/practice/ports"
	teachinganalysishttp "ctf-platform/internal/module/teaching_analysis/api/http"
	teachinganalysisqueries "ctf-platform/internal/module/teaching_analysis/application/queries"
)

func TestBuildRoot(t *testing.T) {
	t.Parallel()

	cfg, db, cache := newAppTestDependencies(t)

	root, err := composition.BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}
	if root == nil {
		t.Fatal("expected root")
	}
	if root.Events == nil {
		t.Fatal("expected events bus")
	}
}

func TestBuildRootRejectsMismatchedContainerFlagSecret(t *testing.T) {
	t.Parallel()

	cfg, db, cache := newAppTestDependencies(t)
	cfg.Container.FlagGlobalSecret = "first-cluster-secret-123456789012345"
	cfg.Container.FlagGlobalSecretKeyID = "active"
	cfg.Container.ResolvedFlagSecretKeyID = "active"

	root, err := composition.BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("first BuildRoot() error = %v", err)
	}
	root.Cancel()

	nextCfg := *cfg
	nextCfg.Container.FlagGlobalSecret = "second-cluster-secret-12345678901234"
	_, err = composition.BuildRoot(&nextCfg, zap.NewNop(), db, cache)
	if err == nil {
		t.Fatal("expected BuildRoot() to reject mismatched container flag secret")
	}
	if !strings.Contains(err.Error(), "container flag secret fingerprint mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildRootAllowsContainerFlagSecretRotationWhenPreviousActiveKeyConfigured(t *testing.T) {
	t.Parallel()

	cfg, db, cache := newAppTestDependencies(t)
	cfg.Container.FlagGlobalSecret = "old-cluster-secret-123456789012345678"
	cfg.Container.FlagGlobalSecretKeyID = "old"
	cfg.Container.ResolvedFlagSecretKeyID = "old"
	cfg.Container.ResolvedFlagSecrets = map[string]string{
		"default": "old-cluster-secret-123456789012345678",
		"old":     "old-cluster-secret-123456789012345678",
	}

	root, err := composition.BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("first BuildRoot() error = %v", err)
	}
	root.Cancel()

	nextCfg := *cfg
	nextCfg.Container.FlagGlobalSecret = "new-cluster-secret-123456789012345678"
	nextCfg.Container.FlagGlobalSecretKeyID = "new"
	nextCfg.Container.ResolvedFlagSecretKeyID = "new"
	nextCfg.Container.FlagGlobalSecretAllowRotation = true
	nextCfg.Container.ResolvedFlagSecrets = map[string]string{
		"old": "old-cluster-secret-123456789012345678",
		"new": "new-cluster-secret-123456789012345678",
	}
	rotatedRoot, err := composition.BuildRoot(&nextCfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("rotating BuildRoot() error = %v", err)
	}
	rotatedRoot.Cancel()
}

func TestBuildRootRejectsMissingRequiredLegacyContainerFlagSecret(t *testing.T) {
	t.Parallel()

	cfg, db, cache := newAppTestDependencies(t)
	cfg.Container.FlagGlobalSecret = "old-cluster-secret-123456789012345678"
	cfg.Container.FlagGlobalSecretKeyID = "old"
	cfg.Container.ResolvedFlagSecretKeyID = "old"
	cfg.Container.ResolvedFlagSecrets = map[string]string{
		"default": "old-cluster-secret-123456789012345678",
		"old":     "old-cluster-secret-123456789012345678",
	}

	if err := db.Create(&instancecontracts.Instance{
		UserID:      7,
		ChallengeID: 11,
		Status:      "running",
		Nonce:       "legacy-nonce",
		ExpiresAt:   time.Now().UTC().Add(time.Hour),
	}).Error; err != nil {
		t.Fatalf("create legacy instance: %v", err)
	}

	root, err := composition.BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("first BuildRoot() error = %v", err)
	}
	root.Cancel()

	nextCfg := *cfg
	nextCfg.Container.FlagGlobalSecret = "new-cluster-secret-123456789012345678"
	nextCfg.Container.FlagGlobalSecretKeyID = "new"
	nextCfg.Container.ResolvedFlagSecretKeyID = "new"
	nextCfg.Container.FlagGlobalSecretAllowRotation = true
	nextCfg.Container.ResolvedFlagSecrets = map[string]string{
		"old": "old-cluster-secret-123456789012345678",
		"new": "new-cluster-secret-123456789012345678",
	}
	_, err = composition.BuildRoot(&nextCfg, zap.NewNop(), db, cache)
	if err == nil {
		t.Fatal("expected BuildRoot() to reject missing legacy default key")
	}
	if !strings.Contains(err.Error(), "required container flag secret key default is not configured") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpsModuleContractsCompile(t *testing.T) {
	var _ auditlog.Recorder = (*opscmd.AuditService)(nil)
}

func TestTeachingAnalysisModuleContractsCompile(t *testing.T) {
	var _ teachinganalysisqueries.Service = (*teachinganalysisqueries.QueryService)(nil)
}

func TestCompositionModulesExposeContracts(t *testing.T) {
	t.Parallel()

	assertFieldType(t, reflect.TypeOf(composition.IdentityModule{}), "AdminHandler", reflect.TypeOf(&identityhttp.Handler{}))
	assertFieldType(t, reflect.TypeOf(composition.IdentityModule{}), "ProfileCommands", reflect.TypeOf((*identitycontracts.ProfileCommandService)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.IdentityModule{}), "ProfileQueries", reflect.TypeOf((*identitycontracts.ProfileQueryService)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.IdentityModule{}), "Users", reflect.TypeOf((*identitycontracts.UserLookupRepository)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.InstanceModule{}), "Handler", reflect.TypeOf(&instancehttp.Handler{}))
	assertFieldType(t, reflect.TypeOf(composition.InstanceModule{}), "PracticeRuntimeService", reflect.TypeOf((*practiceports.RuntimeInstanceService)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.PracticeModule{}), "Handler", reflect.TypeOf(&practicehttp.Handler{}))
	assertFieldType(t, reflect.TypeOf(composition.ContainerRuntimeModule{}), "ChallengeImageRuntime", reflect.TypeOf((*challengeports.ImageRuntime)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.ContainerRuntimeModule{}), "ChallengeRuntimeProbe", reflect.TypeOf((*challengeports.ChallengeRuntimeProbe)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.ContainerRuntimeModule{}), "OpsRuntimeQuery", reflect.TypeOf((*opsports.RuntimeQuery)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.ContainerRuntimeModule{}), "OpsRuntimeStatsProvider", reflect.TypeOf((*opsports.RuntimeStatsProvider)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.ContainerRuntimeModule{}), "ContestContainerFiles", reflect.TypeOf((*contestports.AWDContainerFileWriter)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.OpsModule{}), "AuditService", reflect.TypeOf((*auditlog.Recorder)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.OpsModule{}), "AuditHandler", reflect.TypeOf(&opshttp.AuditHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.OpsModule{}), "DashboardHandler", reflect.TypeOf(&opshttp.DashboardHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.OpsModule{}), "NotificationHandler", reflect.TypeOf(&opshttp.NotificationHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.OpsModule{}), "RiskHandler", reflect.TypeOf(&opshttp.RiskHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.TeachingAnalysisModule{}), "Handler", reflect.TypeOf(&teachinganalysishttp.Handler{}))
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "Catalog", reflect.TypeOf((*challengecontracts.ChallengeContract)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "FlagValidator", reflect.TypeOf((*challengecontracts.FlagValidator)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "FlagHandler", reflect.TypeOf(&challengehttp.FlagHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "Handler", reflect.TypeOf(&challengehttp.Handler{}))
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "ImageHandler", reflect.TypeOf(&challengehttp.ImageHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "ImageStore", reflect.TypeOf((*challengecontracts.ImageStore)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "TopologyHandler", reflect.TypeOf(&challengehttp.TopologyHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "WriteupHandler", reflect.TypeOf(&challengehttp.WriteupHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.AssessmentModule{}), "Handler", reflect.TypeOf(&assessmenthttp.Handler{}))
	assertFieldType(t, reflect.TypeOf(composition.AssessmentModule{}), "ProfileService", reflect.TypeOf((*assessmentcontracts.ProfileService)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.AssessmentModule{}), "Recommendations", reflect.TypeOf((*assessmentcontracts.RecommendationProvider)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.AssessmentModule{}), "ReportHandler", reflect.TypeOf(&assessmenthttp.ReportHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.AssessmentModule{}), "TeacherAWDReviewHandler", reflect.TypeOf(&assessmenthttp.TeacherAWDReviewHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.ContestModule{}), "AWDHandler", reflect.TypeOf(&contesthttp.AWDHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.ContestModule{}), "ChallengeHandler", reflect.TypeOf(&contesthttp.ChallengeHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.ContestModule{}), "Handler", reflect.TypeOf(&contesthttp.Handler{}))
	assertFieldType(t, reflect.TypeOf(composition.ContestModule{}), "ParticipationHandler", reflect.TypeOf(&contesthttp.ParticipationHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.ContestModule{}), "SubmissionHandler", reflect.TypeOf(&contesthttp.SubmissionHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.ContestModule{}), "TeamHandler", reflect.TypeOf(&contesthttp.TeamHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.PracticeModule{}), "Handler", reflect.TypeOf(&practicehttp.Handler{}))
	assertNoField(t, reflect.TypeOf(composition.AuthModule{}), "TokenService")
	assertNoField(t, reflect.TypeOf(composition.ChallengeModule{}), "FlagService")
	assertNoField(t, reflect.TypeOf(composition.ChallengeModule{}), "ImageRepository")
	assertNoField(t, reflect.TypeOf(composition.ChallengeModule{}), "ImageService")
	assertNoField(t, reflect.TypeOf(composition.ChallengeModule{}), "Repository")
	assertNoField(t, reflect.TypeOf(composition.ContestModule{}), "Repository")
	assertNoField(t, reflect.TypeOf(composition.AssessmentModule{}), "RecommendationService")
	assertNoField(t, reflect.TypeOf(composition.AssessmentModule{}), "ReportService")
	assertNoField(t, reflect.TypeOf(composition.AssessmentModule{}), "Service")
	assertNoField(t, reflect.TypeOf(composition.InstanceModule{}), "Service")
	assertNoField(t, reflect.TypeOf(composition.PracticeModule{}), "Service")
	assertNoField(t, reflect.TypeOf(composition.ContainerRuntimeModule{}), "Handler")
	assertNoField(t, reflect.TypeOf(composition.ContainerRuntimeModule{}), "PracticeRuntimeService")
	assertNoField(t, reflect.TypeOf(composition.ContainerRuntimeModule{}), "Query")
	assertNoField(t, reflect.TypeOf(composition.ContainerRuntimeModule{}), "Repository")
	assertNoField(t, reflect.TypeOf(composition.ContainerRuntimeModule{}), "Service")
	assertNoField(t, reflect.TypeOf(composition.IdentityModule{}), "TokenService")
	assertNoField(t, reflect.TypeOf(composition.IdentityModule{}), "users")
}

func TestCompositionBuildersUseRuntimeAndInstanceModulesForDependencies(t *testing.T) {
	t.Parallel()

	assertFunctionParamType(t, reflect.TypeOf(composition.BuildInstanceModule), 1, reflect.TypeOf(&composition.ContainerRuntimeModule{}))
	assertFunctionParamType(t, reflect.TypeOf(composition.BuildChallengeModule), 1, reflect.TypeOf(&composition.ContainerRuntimeModule{}))
	assertFunctionParamType(t, reflect.TypeOf(composition.BuildContestModule), 2, reflect.TypeOf(&composition.ContainerRuntimeModule{}))
	assertFunctionParamType(t, reflect.TypeOf(composition.BuildOpsModule), 1, reflect.TypeOf(&composition.ContainerRuntimeModule{}))
	assertFunctionParamType(t, reflect.TypeOf(composition.BuildPracticeModule), 2, reflect.TypeOf(&composition.InstanceModule{}))
}

func TestBuildOpsModuleDelegatesToContainerRuntime(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "ops_module.go"))
	if err != nil {
		t.Fatalf("read ops_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"module := opsruntime.Build(",
		"opsruntime.Deps{",
		"return &OpsModule{",
		"m.runtime.BindNotificationHandler(tokenService)",
		"m.NotificationHandler = m.runtime.NotificationHandler",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("ops module should delegate through %s", marker)
		}
	}
}

func TestBuildContainerRuntimeModuleDelegatesToSubBuilders(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "container_runtime_module.go"))
	if err != nil {
		t.Fatalf("read container_runtime_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type ContainerRuntimeModule struct",
		"func BuildContainerRuntimeModule(root *Root) (*ContainerRuntimeModule, error) {",
		"defaultNodeClient, err := buildDefaultNodeRuntimeClient(root, nodeAllocationRepo, defaultNode)",
		"nodeRouter.rememberClient(defaultNode.ID, defaultNodeClient)",
		"module := containerruntime.Build(",
		"containerruntime.Deps{",
		"module.BackgroundJobs",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("runtime module should delegate through %s", marker)
		}
	}

	blocked := []string{
		"type RuntimeModule = ContainerRuntimeModule",
		"func BuildRuntimeModule(root *Root) (*RuntimeModule, error) {",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("container runtime composition should not keep legacy compatibility marker %s", marker)
		}
	}
}

func TestBuildInstanceModuleDelegatesToSubBuilders(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "instance_module.go"))
	if err != nil {
		t.Fatalf("read instance_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"module := runtime.runtime",
		"newCompositeActiveContainerInventory(",
		"newInstanceRuntimeInventoryProvider(instanceinfra.NewContainerInventoryRepository(root.DB()))",
		"contestinfra.NewAWDContainerInventoryRepository(root.DB())",
		"instanceinfra.NewRepository(root.DB())",
		"instancecmd.NewInstanceService(",
		"buildRuntimeProxyTicketService(root, instanceRepo)",
		"instancecmd.NewInstanceMaintenanceService(",
		"instancehttp.NewHandler(m.service",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("instance module should delegate through %s", marker)
		}
	}

	blocked := []string{
		`"awd_defense_ssh_gateway"`,
		`NewAWDDefenseSSHGateway(`,
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("instance module should not keep inline ssh gateway wiring %s", marker)
		}
	}
}

func TestRouterRateLimitStrategyUsesUserAndLoginPrincipalKeys(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile("router.go")
	if err != nil {
		t.Fatalf("read router.go: %v", err)
	}

	source := string(content)
	expected := []string{
		`protected.Use(middleware.RateLimitByUser(`,
		`middleware.RateLimitByLoginPrincipalAndIP(`,
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("router should include rate limit marker %s", marker)
		}
	}

	blocked := []string{
		`engine.Use(middleware.RateLimitByIP(rateChecker, "global"`,
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("router should not keep global IP rate limit marker %s", marker)
		}
	}
}

func TestBuildContestModuleDelegatesToRuntime(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "contest_module.go"))
	if err != nil {
		t.Fatalf("read contest_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"contestruntime.Build(",
		"contestruntime.Deps{",
		"Events:                root.Events",
		"root.RegisterBackgroundJob(",
		"NewLoopBackgroundJob(job.Name, job.Run)",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("contest module should delegate through %s", marker)
		}
	}
}

func TestBuildChallengeModuleDelegatesToSubBuilders(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "challenge_module.go"))
	if err != nil {
		t.Fatalf("read challenge_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"challengeruntime.Build(",
		"root.RegisterBackgroundJob(",
		"NewLoopBackgroundJob(job.Name, job.Run)",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("challenge module should delegate through %s", marker)
		}
	}
}

func TestBuildPracticeModuleDelegatesToSubBuilders(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "practice_module.go"))
	if err != nil {
		t.Fatalf("read practice_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"practiceruntime.Build(",
		"root.RegisterBackgroundJob(",
		"NewLoopBackgroundJob(job.Name, job.Run)",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("practice module should delegate through %s", marker)
		}
	}
}

func TestBuildAssessmentModuleDelegatesToSubBuilders(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "assessment_module.go"))
	if err != nil {
		t.Fatalf("read assessment_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"assessmentruntime.Build(",
		"root.RegisterBackgroundJob(",
		"NewBackgroundJob(job.Name, job.Start, job.Stop)",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("assessment module should delegate through %s", marker)
		}
	}
}

func TestCompositionBuildersAvoidPrivateCrossModuleFields(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("composition", "*_module.go"))
	if err != nil {
		t.Fatalf("glob composition modules: %v", err)
	}

	blocked := []string{
		"identity.users",
		"runtime.practice.",
		"runtime.ops.",
		"runtime.challenge.",
		"runtime.contest.",
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}

		source := string(content)
		for _, marker := range blocked {
			if strings.Contains(source, marker) {
				t.Fatalf("%s should not reference private cross-module field %s", file, marker)
			}
		}
	}
}
