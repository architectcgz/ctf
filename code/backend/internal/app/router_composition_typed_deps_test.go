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

	repositoryContent, err := os.ReadFile(filepath.Join("..", "module", "runtime", "infrastructure", "repository.go"))
	if err != nil {
		t.Fatalf("read runtime repository.go: %v", err)
	}
	allocationContent, err := os.ReadFile(filepath.Join("..", "module", "runtime", "infrastructure", "allocation_repository.go"))
	if err != nil {
		t.Fatalf("read runtime allocation_repository.go: %v", err)
	}

	repositorySource := string(repositoryContent)
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
		"func (r *Repository) ReleaseRuntimeAllocationsForInstance",
		"func (r *Repository) ReserveAvailablePort",
		"func (r *Repository) ReserveAvailableSubnet",
		"func (r *Repository) SyncInstanceHostPortForRestart",
	}
	for _, marker := range blocked {
		if strings.Contains(repositorySource, marker) {
			t.Fatalf("runtime Repository should not own allocation marker %s", marker)
		}
	}
}

func TestRuntimeRepositoryDoesNotOwnAWDPersistence(t *testing.T) {
	t.Parallel()

	repositoryContent, err := os.ReadFile(filepath.Join("..", "module", "runtime", "infrastructure", "repository.go"))
	if err != nil {
		t.Fatalf("read runtime repository.go: %v", err)
	}
	awdContent, err := os.ReadFile(filepath.Join("..", "module", "runtime", "infrastructure", "awd_repository.go"))
	if err != nil {
		t.Fatalf("read runtime awd_repository.go: %v", err)
	}

	repositorySource := string(repositoryContent)
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
		"func (r *Repository) FindAWDDefenseWorkspace",
		"func (r *Repository) UpsertAWDDefenseWorkspace",
		"func (r *Repository) BumpAWDDefenseWorkspaceRevision",
		"func (r *Repository) FindRunningAWDDefenseWorkspaceByInstanceID",
		"func (r *Repository) CreateAWDServiceOperation",
		"func (r *Repository) FinishActiveAWDServiceOperationForInstance",
		"func (r *Repository) FinishAWDServiceOperation",
	}
	for _, marker := range blocked {
		if strings.Contains(repositorySource, marker) {
			t.Fatalf("runtime Repository should not own AWD marker %s", marker)
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
		"newInstanceMaintenanceRepository(root.DB(), instanceRepo, allocationRepo, awdRepo, repo)",
		"newPracticeInstanceRepository(root.DB(), instanceRepo, allocationRepo, awdRepo, repo)",
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
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("runtime composition should inject runtime persistence marker %s", marker)
		}
	}
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
