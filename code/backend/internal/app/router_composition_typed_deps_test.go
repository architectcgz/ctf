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
		"cleanupService",
		"*runtimecmd.RuntimeCleanupService",
		"provisioningService",
		"*runtimecmd.ProvisioningService",
		"containerStatsService",
		"*runtimeapp.ContainerStatsService",
		"imageRuntime",
		"*runtimeapp.ImageRuntimeService",
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
		"opsports.RuntimeQuery",
		"challengeports.ImageRuntime",
		"challengeports.ChallengeRuntimeProbe",
		"runtimeInstanceRepository",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("runtime runtime module should not keep cross-module glue marker %s", marker)
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

func TestPracticeModuleWiresRuntimePortOwnerFromCompositionRoot(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "practice_module.go"))
	if err != nil {
		t.Fatalf("read practice_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"runtimeinfra \"ctf-platform/internal/module/runtime/infrastructure\"",
		"runtimeports \"ctf-platform/internal/module/runtime/ports\"",
		"RuntimePortOwnerFor: runtimePortOwnerFor",
		"func runtimePortOwnerFor(db *gorm.DB) runtimeports.PortReservationOwner",
		"return runtimeinfra.NewAllocationRepository(db)",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("practice composition should wire runtime port owner through composition root marker %s", marker)
		}
	}
}

func TestRuntimeRepositoryDoesNotOwnAllocationPersistence(t *testing.T) {
	t.Parallel()

	allocationContent, err := os.ReadFile(filepath.Join("..", "module", "runtime", "infrastructure", "allocation_repository.go"))
	if err != nil {
		t.Fatalf("read runtime allocation_repository.go: %v", err)
	}

	otherInfrastructure := runtimeInfrastructureSourceExcept(t, "allocation_repository.go")
	allocationSource := string(allocationContent)
	expected := []string{
		"type AllocationRepository struct",
		"func NewAllocationRepository(db *gorm.DB) *AllocationRepository",
		"func (r *AllocationRepository) ReleaseRuntimeAllocationsForInstance",
		"func (r *AllocationRepository) ReserveAvailablePort",
		"func (r *AllocationRepository) ReserveAvailableSubnet",
		"func (r *AllocationRepository) SyncInstanceHostPortForRestart",
	}
	for _, marker := range expected {
		if !strings.Contains(allocationSource, marker) {
			t.Fatalf("runtime allocation repository should own marker %s", marker)
		}
	}

	blocked := []string{
		"ReleaseRuntimeAllocationsForInstance",
		"ReserveAvailablePort",
		"ReserveAvailableSubnet",
		"SyncInstanceHostPortForRestart",
	}
	for _, marker := range blocked {
		if strings.Contains(otherInfrastructure, marker) {
			t.Fatalf("runtime allocation persistence should not leak outside allocation_repository.go marker %s", marker)
		}
	}
}

func TestRuntimeRepositoryDoesNotOwnAWDPersistence(t *testing.T) {
	t.Parallel()

	awdContent, err := os.ReadFile(filepath.Join("..", "module", "runtime", "infrastructure", "awd_repository.go"))
	if err != nil {
		t.Fatalf("read runtime awd_repository.go: %v", err)
	}

	otherInfrastructure := runtimeInfrastructureSourceExcept(t, "awd_repository.go")
	awdSource := string(awdContent)
	expected := []string{
		"type AWDRepository struct",
		"func NewAWDRepository(db *gorm.DB) *AWDRepository",
		"func (r *AWDRepository) FindAWDDefenseWorkspace",
		"func (r *AWDRepository) UpsertAWDDefenseWorkspace",
		"func (r *AWDRepository) BumpAWDDefenseWorkspaceRevision",
		"func (r *AWDRepository) FindRunningAWDDefenseWorkspaceByInstanceID",
		"func (r *AWDRepository) CreateAWDServiceOperation",
		"func (r *AWDRepository) FinishActiveAWDServiceOperationForInstance",
		"func (r *AWDRepository) FinishAWDServiceOperation",
	}
	for _, marker := range expected {
		if !strings.Contains(awdSource, marker) {
			t.Fatalf("runtime AWD repository should own marker %s", marker)
		}
	}

	blocked := []string{
		"FindAWDDefenseWorkspace",
		"UpsertAWDDefenseWorkspace",
		"BumpAWDDefenseWorkspaceRevision",
		"FindRunningAWDDefenseWorkspaceByInstanceID",
		"CreateAWDServiceOperation",
		"FinishActiveAWDServiceOperationForInstance",
		"FinishAWDServiceOperation",
	}
	for _, marker := range blocked {
		if strings.Contains(otherInfrastructure, marker) {
			t.Fatalf("runtime AWD persistence should not leak outside awd_repository.go marker %s", marker)
		}
	}
}

func TestRuntimeRepositoryDoesNotOwnStatePersistence(t *testing.T) {
	t.Parallel()

	stateContent, err := os.ReadFile(filepath.Join("..", "module", "runtime", "infrastructure", "state_repository.go"))
	if err != nil {
		t.Fatalf("read runtime state_repository.go: %v", err)
	}

	otherInfrastructure := runtimeInfrastructureSourceExcept(t, "state_repository.go")
	stateSource := string(stateContent)
	expected := []string{
		"type RuntimeStateRepository struct",
		"func NewRuntimeStateRepository(db *gorm.DB) *RuntimeStateRepository",
		"func (r *RuntimeStateRepository) FindByID",
		"func (r *RuntimeStateRepository) ListActiveContainerIDs",
		"func (r *RuntimeStateRepository) FindRuntimeNodeIDByContainerID",
		"func (r *RuntimeStateRepository) ListInstancesNeedingACLHandleMigration",
		"func (r *RuntimeStateRepository) UpdateInstanceRuntimeDetails",
	}
	for _, marker := range expected {
		if !strings.Contains(stateSource, marker) {
			t.Fatalf("runtime state repository should own marker %s", marker)
		}
	}

	blocked := []string{
		"func (r *RuntimeStateRepository) FindByID",
		"func (r *RuntimeStateRepository) ListActiveContainerIDs",
		"func (r *RuntimeStateRepository) FindRuntimeNodeIDByContainerID",
		"func (r *RuntimeStateRepository) ListInstancesNeedingACLHandleMigration",
		"func (r *RuntimeStateRepository) UpdateInstanceRuntimeDetails",
	}
	for _, marker := range blocked {
		if strings.Contains(otherInfrastructure, marker) {
			t.Fatalf("runtime state persistence should not leak outside state_repository.go marker %s", marker)
		}
	}
}

func TestRuntimeAWDPersistenceWiredFromCompositionRoot(t *testing.T) {
	t.Parallel()

	instanceContent, err := os.ReadFile(filepath.Join("composition", "instance_module.go"))
	if err != nil {
		t.Fatalf("read instance_module.go: %v", err)
	}
	contestContent, err := os.ReadFile(filepath.Join("composition", "contest_module.go"))
	if err != nil {
		t.Fatalf("read contest_module.go: %v", err)
	}

	instanceSource := string(instanceContent)
	instanceExpected := []string{
		"awdRepo := runtimeinfra.NewAWDRepository(root.DB())",
		"newInstanceMaintenanceRepository(root.DB(), instanceRepo, allocationRepo, awdRepo, stateRepo)",
		"newPracticeInstanceRepository(root.DB(), instanceRepo, allocationRepo, awdRepo)",
	}
	for _, marker := range instanceExpected {
		if !strings.Contains(instanceSource, marker) {
			t.Fatalf("instance composition should wire runtime AWD repository marker %s", marker)
		}
	}

	contestSource := string(contestContent)
	contestExpected := []string{
		"runtimeAWDRepo := runtimeinfra.NewAWDRepository(root.DB())",
		"newContestEndedRuntimeStateStore(root.DB(), instanceRepo, allocationRepo, runtimeAWDRepo)",
	}
	for _, marker := range contestExpected {
		if !strings.Contains(contestSource, marker) {
			t.Fatalf("contest composition should wire runtime AWD repository marker %s", marker)
		}
	}
}

func TestRuntimeStatePersistenceWiredFromCompositionRoot(t *testing.T) {
	t.Parallel()

	runtimeContent, err := os.ReadFile(filepath.Join("composition", "runtime_module.go"))
	if err != nil {
		t.Fatalf("read runtime_module.go: %v", err)
	}
	instanceContent, err := os.ReadFile(filepath.Join("composition", "instance_module.go"))
	if err != nil {
		t.Fatalf("read instance_module.go: %v", err)
	}

	runtimeSource := string(runtimeContent)
	runtimeExpected := []string{
		"stateRepo := runtimeinfra.NewRuntimeStateRepository(root.DB())",
		"newRuntimeNodeExecutionRouter(cfg, log.Named(\"runtime_node_router\"), allocationRepo, stateRepo, nodeRepo, defaultNodeName)",
		"migrateLegacyInstanceACLHandles(root.Context(), stateRepo, nodeRouter, defaultNodeClient, log.Named(\"runtime_acl_migration\"))",
	}
	for _, marker := range runtimeExpected {
		if !strings.Contains(runtimeSource, marker) {
			t.Fatalf("runtime composition should wire runtime state repository marker %s", marker)
		}
	}

	instanceSource := string(instanceContent)
	instanceExpected := []string{
		"stateRepo := runtimeinfra.NewRuntimeStateRepository(root.DB())",
		"newInstanceMaintenanceRepository(root.DB(), instanceRepo, allocationRepo, awdRepo, stateRepo)",
	}
	for _, marker := range instanceExpected {
		if !strings.Contains(instanceSource, marker) {
			t.Fatalf("instance composition should wire runtime state repository marker %s", marker)
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
		"*runtimeapp.ImageRuntimeService",
		"runtimeports.ManagedContainerStatsReader",
		"contestports.AWDContainerFileWriter",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("runtime runtime module should use external ports marker %s", marker)
		}
	}

	blocked := []string{
		"contestinfra.AWDContainerFileWriter",
		"opsports.RuntimeQuery",
		"opsports.RuntimeStatsProvider",
		"practiceRuntimeRepositoryBridge",
		"practiceRuntimeInstanceService",
		"challengeports.ImageRuntime",
		"challengeports.ChallengeRuntimeProbe",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("runtime runtime module should not keep bridge marker %s", marker)
		}
	}
}

func TestRuntimeModuleDoesNotConstructRuntimeInfrastructure(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("..", "module", "runtime", "runtime", "module.go"))
	if err != nil {
		t.Fatalf("read runtime runtime module: %v", err)
	}

	source := string(content)
	expected := []string{
		"ProvisioningRepository",
		"runtimecmd.ProvisioningRepository",
		"CleanupRepository",
		"runtimecmd.RuntimeCleanupRepository",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("runtime runtime module should declare injected runtime persistence marker %s", marker)
		}
	}

	blocked := []string{
		"runtimeinfra.NewRepository(",
		"ctf-platform/internal/module/runtime/infrastructure",
		"*gorm.DB",
		"*redislib.Client",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("runtime runtime module should not construct infrastructure marker %s", marker)
		}
	}
}

func TestRuntimeCompositionInjectsRuntimePersistenceIntoRuntimeModule(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "runtime_module.go"))
	if err != nil {
		t.Fatalf("read runtime_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"ProvisioningRepository:",
		"CleanupRepository:",
		"stateRepo := runtimeinfra.NewRuntimeStateRepository(root.DB())",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("runtime composition should inject runtime persistence marker %s", marker)
		}
	}
}

func runtimeInfrastructureSourceExcept(t *testing.T, excluded ...string) string {
	t.Helper()

	skip := make(map[string]struct{}, len(excluded))
	for _, name := range excluded {
		skip[name] = struct{}{}
	}

	files, err := filepath.Glob(filepath.Join("..", "module", "runtime", "infrastructure", "*.go"))
	if err != nil {
		t.Fatalf("glob runtime infrastructure files: %v", err)
	}

	var builder strings.Builder
	for _, file := range files {
		base := filepath.Base(file)
		if strings.HasSuffix(base, "_test.go") {
			continue
		}
		if _, excludedFile := skip[base]; excludedFile {
			continue
		}
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read runtime infrastructure file %s: %v", file, err)
		}
		builder.Write(content)
		builder.WriteByte('\n')
	}
	return builder.String()
}

func TestRuntimeNodeExecutionRouterUsesNarrowRuntimePersistenceDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "runtime_node_execution_router.go"))
	if err != nil {
		t.Fatalf("read runtime_node_execution_router.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type runtimeNodeAllocationRepository interface",
		"type runtimeNodeStateRepository interface",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("runtime node execution router should declare narrow runtime persistence marker %s", marker)
		}
	}

	blocked := []string{
		"*runtimeinfra.Repository",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("runtime node execution router should not depend on concrete runtime repository marker %s", marker)
		}
	}
}

func TestRuntimeACLMigrationUsesNarrowRuntimeStateDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "runtime_acl_migration.go"))
	if err != nil {
		t.Fatalf("read runtime_acl_migration.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type runtimeACLMigrationRepository interface",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("runtime acl migration should declare narrow runtime state marker %s", marker)
		}
	}

	blocked := []string{
		"*runtimeinfra.Repository",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("runtime acl migration should not depend on concrete runtime repository marker %s", marker)
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
