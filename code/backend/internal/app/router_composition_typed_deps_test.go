package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeModuleUsesTypedDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("..", "module", "runtime", "runtime", "module.go"))
	if err != nil {
		t.Fatalf("read runtime runtime module: %v", err)
	}

	source := string(content)
	expected := []string{
		"type runtimeModuleDeps struct",
		"repo",
		"runtimeInstanceRepository",
		"countRunningQuery",
		"opsports.RuntimeQuery",
		"cleanupService",
		"*runtimecmd.RuntimeCleanupService",
		"provisioningService",
		"*runtimecmd.ProvisioningService",
		"containerStatsService",
		"*runtimeapp.ContainerStatsService",
		"imageRuntime",
		"challengeports.ImageRuntime",
		"containerFiles",
		"contestports.AWDContainerFileWriter",
		"containerPublicHost",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("runtime runtime module should declare typed deps marker %s", marker)
		}
	}

	blocked := []string{
		"practiceInstanceRepo",
		"practiceInstanceRepository",
		"PracticeInstanceRepository",
		"PracticeRuntimeService",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("runtime runtime module should not keep practice-facing glue marker %s", marker)
		}
	}
}

func TestRuntimeModuleUsesCommandsQueriesServices(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("..", "module", "runtime", "runtime", "module.go"))
	if err != nil {
		t.Fatalf("read runtime runtime module: %v", err)
	}

	source := string(content)
	expected := []string{
		"runtimecmd.NewRuntimeCleanupService(",
		"runtimecmd.NewProvisioningService(",
		"runtimeqry.NewCountRunningService(",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("runtime runtime module should use layered runtime service marker %s", marker)
		}
	}

	blocked := []string{
		"runtimecmd.NewInstanceService(",
		"runtimecmd.NewRuntimeMaintenanceService(",
		"runtimeqry.NewInstanceService(",
		"runtimeqry.NewProxyTicketService(",
		"runtimeapp.NewInstanceService(",
		"runtimeapp.NewQueryService(",
		"runtimeapp.NewProxyTicketService(",
		"runtimeapp.NewRuntimeCleanupService(",
		"runtimeapp.NewRuntimeMaintenanceService(",
		"runtimeapp.NewProvisioningService(",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("runtime runtime module should not keep legacy root service marker %s", marker)
		}
	}
}

func TestAuthModuleUsesTypedDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "auth_module.go"))
	if err != nil {
		t.Fatalf("read auth_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type authModuleDeps struct",
		"users",
		"*identityinfra.Repository",
		"tokenService",
		"authcontracts.TokenService",
		"profileCommands",
		"identitycontracts.ProfileCommandService",
		"profileQueries",
		"identitycontracts.ProfileQueryService",
		"auditRecorder",
		"auditlog.Recorder",
		"buildAuthModuleDeps(",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("auth composition should declare typed deps marker %s", marker)
		}
	}

	blocked := []string{
		"authcmd.NewService(identity.users, identity.TokenService",
		"authcmd.NewCASService(cfg.Auth.CAS, identity.users, identity.TokenService",
		"users:           identity.users,",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("auth composition should not keep direct module dependency marker %s", marker)
		}
	}
}

func TestOpsModuleUsesTypedDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "ops_module.go"))
	if err != nil {
		t.Fatalf("read ops_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"runtime *opsruntime.Module",
		"opsruntime.Build(",
		"opsruntime.Deps{",
		"RuntimeQuery: runtime.OpsRuntimeQuery",
		"RuntimeStats: runtime.OpsRuntimeStatsProvider",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("ops composition should declare typed deps marker %s", marker)
		}
	}

	blocked := []string{
		"type opsModuleDeps struct",
		"type opsNotificationDeps struct",
		"opsinfra.NewAuditRepository(",
		"opsinfra.NewRiskRepository(",
		"opsinfra.NewNotificationRepository(",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("ops composition should not keep concrete marker %s", marker)
		}
	}
}

func TestContestModuleDepsAvoidConcreteContestRepositories(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("..", "module", "contest", "runtime", "module.go"))
	if err != nil {
		t.Fatalf("read contest runtime module: %v", err)
	}

	source := string(content)
	blocked := []string{
		"challenge         *ChallengeModule",
		"runtime           *RuntimeModule",
		"bindRealtime",
		"SetRealtimeBroadcaster",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("contest runtime deps should not keep direct module dependency field %s", marker)
		}
	}
}

func TestContestRuntimeUsesTypedCrossModuleDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("..", "module", "contest", "runtime", "module.go"))
	if err != nil {
		t.Fatalf("read contest runtime module: %v", err)
	}

	source := string(content)
	expected := []string{
		"type Deps struct",
		"ChallengeCatalog",
		"challengecontracts.ContestChallengeContract",
		"FlagValidator",
		"challengecontracts.FlagValidator",
		"ContainerFiles",
		"contestports.AWDContainerFileWriter",
		"CheckerRunner",
		"contestports.CheckerRunner",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("contest runtime should declare typed cross-module deps marker %s", marker)
		}
	}

	blocked := []string{
		"challenge         *ChallengeModule",
		"runtime           *RuntimeModule",
		"deps.challenge.Catalog",
		"deps.challenge.FlagValidator",
		"NewDockerCheckerRunner(",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("contest runtime should not keep direct module dependency marker %s", marker)
		}
	}
}

func TestChallengeModuleUsesTypedPortsDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "challenge_module.go"))
	if err != nil {
		t.Fatalf("read challenge_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type ChallengeModule = challengeruntime.Module",
		"challengeruntime.Build(",
		"challengeruntime.Deps{",
		"Events:       root.Events",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("challenge composition should declare typed deps marker %s", marker)
		}
	}

	blocked := []string{
		"type challengeModuleDeps struct",
		"challengeinfra.NewRepository(",
		"challengeinfra.NewImageRepository(",
		"challengeinfra.NewTemplateRepository(",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("challenge composition deps should not keep concrete repository field %s", marker)
		}
	}
}

func TestPracticeModuleUsesTypedPortsDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "practice_module.go"))
	if err != nil {
		t.Fatalf("read practice_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type PracticeModule = practiceruntime.Module",
		"practiceruntime.Build(",
		"practiceruntime.Deps{",
		"instance.PracticeInstanceRepository",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("practice composition should declare typed deps marker %s", marker)
		}
	}

	blocked := []string{
		"type practiceModuleDeps struct",
		"practiceinfra.NewRepository(",
		"buildPracticeHandler(",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("practice composition should not keep wiring marker %s", marker)
		}
	}
}

func TestPracticeModuleUsesTypedCrossModuleDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "practice_module.go"))
	if err != nil {
		t.Fatalf("read practice_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"practiceruntime.Deps{",
		"instance.PracticeInstanceRepository",
		"instance.PracticeRuntimeService",
		"challenge.Catalog",
		"challenge.ImageStore",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("practice composition should declare typed cross-module deps marker %s", marker)
		}
	}

	blocked := []string{
		"type practiceModuleExternalDeps struct",
		"buildPracticeModuleExternalDeps(",
		"assessment.ProfileService",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("practice composition should not keep runtime private marker %s", marker)
		}
	}
}

func TestAssessmentModuleUsesTypedPortsDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "assessment_module.go"))
	if err != nil {
		t.Fatalf("read assessment_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type AssessmentModule = assessmentruntime.Module",
		"assessmentruntime.Build(",
		"assessmentruntime.Deps{",
		"ChallengeRepo: challenge.Catalog",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("assessment composition should declare typed deps marker %s", marker)
		}
	}

	blocked := []string{
		"type assessmentModuleDeps struct",
		"assessmentinfra.NewRepository(",
		"assessmentinfra.NewReportRepository(",
		"assessmentinfra.NewTeacherAWDReviewRepository(",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("assessment composition should not keep wiring marker %s", marker)
		}
	}
}

func TestAssessmentModuleUsesTypedCrossModuleDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "assessment_module.go"))
	if err != nil {
		t.Fatalf("read assessment_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"assessmentruntime.Deps{",
		"ChallengeRepo: challenge.Catalog",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("assessment composition should declare typed cross-module deps marker %s", marker)
		}
	}
}

func TestPracticeModuleAvoidsRuntimeBridgeGlue(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "practice_module.go"))
	if err != nil {
		t.Fatalf("read practice_module.go: %v", err)
	}

	source := string(content)
	blocked := []string{
		"type practiceRuntimeCleanerBridge interface",
		"type practiceRuntimeRepositoryBridge interface",
		"type practiceRuntimeInstanceService interface",
		"type practiceRuntimeProvisioningBridge interface",
		"type practiceRuntimeInstanceServiceAdapter struct",
		"newPracticeRuntimeInstanceServiceAdapter(",
		"toRuntimeTopologyCreateRequest(",
		"fromRuntimeTopologyCreateResult(",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("practice composition should not keep runtime bridge marker %s", marker)
		}
	}
}

func TestRuntimeModuleUsesExternalPortsForCrossModuleDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("..", "module", "runtime", "runtime", "module.go"))
	if err != nil {
		t.Fatalf("read runtime runtime module: %v", err)
	}

	source := string(content)
	expected := []string{
		"challengeports.ImageRuntime",
		"opsports.RuntimeQuery",
		"contestports.AWDContainerFileWriter",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("runtime runtime module should use external ports marker %s", marker)
		}
	}

	blocked := []string{
		"contestinfra.AWDContainerFileWriter",
		"practiceRuntimeRepositoryBridge",
		"practiceRuntimeInstanceService",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("runtime runtime module should not keep bridge marker %s", marker)
		}
	}
}

func TestIdentityModuleUsesTypedDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "identity_module.go"))
	if err != nil {
		t.Fatalf("read identity_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type identityModuleDeps struct",
		"users *identityinfra.Repository",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("identity composition should declare typed deps marker %s", marker)
		}
	}

	if !strings.Contains(source, "users: identityinfra.NewRepository(root.DB())") {
		t.Fatalf("identity composition should build repository in buildIdentityModuleDeps")
	}
	if strings.Contains(source, "tokenService") {
		t.Fatalf("identity composition should not keep token service wiring")
	}
}

func TestTeachingQueryModuleUsesTypedDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "teaching_query_module.go"))
	if err != nil {
		t.Fatalf("read teaching_query_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type TeachingQueryModule = teachingqueryruntime.Module",
		"teachingqueryruntime.Build(",
		"teachingqueryruntime.Deps{",
		"Users:           teachingQueryUserLookupAdapter{users: identity.Users},",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("teaching query composition should declare typed deps marker %s", marker)
		}
	}

	blocked := []string{
		"type teachingReadmodelModuleDeps struct",
		"queryinfra.NewRepository(",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("teaching query composition should not keep wiring marker %s", marker)
		}
	}
}
