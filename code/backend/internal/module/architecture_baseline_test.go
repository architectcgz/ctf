package module

var reviewedRuntimeHostExecutorUsageFiles = map[string]struct{}{
	"../app/composition/container_runtime_module.go":          {},
	"../app/composition/runtime_node_execution_router.go":     {},
	"container_runtime/infrastructure/agentclient/client.go":  {},
	"container_runtime/infrastructure/agentserver/service.go": {},
	"container_runtime/infrastructure/engine.go":              {},
	"container_runtime/ports/runtime_host_executor.go":        {},
}

type moduleDependencyCategory string

const (
	moduleDependencyProviderContract  moduleDependencyCategory = "provider_contract"
	moduleDependencyPortBoundary      moduleDependencyCategory = "port_boundary"
	moduleDependencyRuntimeCapability moduleDependencyCategory = "runtime_capability"
	moduleDependencyEventConsumer     moduleDependencyCategory = "event_consumer"
	moduleDependencyQueryAggregation  moduleDependencyCategory = "query_aggregation"
	moduleDependencyAnalysisReadModel moduleDependencyCategory = "analysis_read_model"
)

var reviewedModuleDependencyCategories = map[moduleDependencyCategory]struct{}{
	moduleDependencyProviderContract:  {},
	moduleDependencyPortBoundary:      {},
	moduleDependencyRuntimeCapability: {},
	moduleDependencyEventConsumer:     {},
	moduleDependencyQueryAggregation:  {},
	moduleDependencyAnalysisReadModel: {},
}

type moduleDependencyReview struct {
	categories []moduleDependencyCategory
	rationale  string
}

func reviewedDependency(rationale string, categories ...moduleDependencyCategory) moduleDependencyReview {
	return moduleDependencyReview{
		categories: categories,
		rationale:  rationale,
	}
}

var moduleDependencyBaseline = map[string]moduleDependencyReview{
	"assessment -> contest": reviewedDependency(
		"assessment owns profile, recommendation, and report analysis outputs and reads contest facts through contracts.",
		moduleDependencyAnalysisReadModel,
		moduleDependencyProviderContract,
	),
	"assessment -> challenge": reviewedDependency(
		"assessment builds challenge-facing analysis and recommendation facts from challenge contracts.",
		moduleDependencyAnalysisReadModel,
		moduleDependencyProviderContract,
	),
	"assessment -> identity": reviewedDependency(
		"assessment resolves user and class-facing identity facts through identity contracts for analysis outputs.",
		moduleDependencyAnalysisReadModel,
		moduleDependencyProviderContract,
	),
	"assessment -> practice": reviewedDependency(
		"assessment consumes practice progress and submission facts through practice contracts for reports and recommendations.",
		moduleDependencyAnalysisReadModel,
		moduleDependencyProviderContract,
	),
	"auth -> identity": reviewedDependency(
		"auth delegates user/profile lookup and mutation to identity-owned contracts.",
		moduleDependencyProviderContract,
	),
	"challenge -> identity": reviewedDependency(
		"challenge reads identity-owned user facts through contracts for challenge workflows and query views.",
		moduleDependencyProviderContract,
	),
	"challenge -> container_runtime": reviewedDependency(
		"challenge uses container_runtime-owned capability contracts for image and topology runtime details.",
		moduleDependencyRuntimeCapability,
		moduleDependencyProviderContract,
	),
	"contest -> auth": reviewedDependency(
		"contest HTTP realtime entrypoints parse auth-owned ticket/token contracts.",
		moduleDependencyProviderContract,
	),
	"contest -> container_runtime": reviewedDependency(
		"contest uses container_runtime capability contracts and ports for AWD checker/runtime execution.",
		moduleDependencyRuntimeCapability,
		moduleDependencyProviderContract,
		moduleDependencyPortBoundary,
	),
	"contest -> challenge": reviewedDependency(
		"contest reads challenge-owned catalog facts and implements challenge-facing ports for contest/AWD behavior.",
		moduleDependencyProviderContract,
		moduleDependencyPortBoundary,
	),
	"contest -> identity": reviewedDependency(
		"contest reads identity-owned user/team member facts through identity contracts.",
		moduleDependencyProviderContract,
	),
	"contest -> instance": reviewedDependency(
		"contest owns AWD access decisions and implements instance-facing ports while reading stable instance contracts.",
		moduleDependencyProviderContract,
		moduleDependencyPortBoundary,
	),
	"instance -> identity": reviewedDependency(
		"instance reads identity-owned user facts through contracts for instance ownership and teacher queries.",
		moduleDependencyProviderContract,
	),
	"ops -> auth": reviewedDependency(
		"ops notification/admin entrypoints use auth-owned token contracts.",
		moduleDependencyProviderContract,
	),
	"ops -> challenge": reviewedDependency(
		"ops consumes challenge publish/check events through challenge contracts.",
		moduleDependencyEventConsumer,
		moduleDependencyProviderContract,
	),
	"ops -> contest": reviewedDependency(
		"ops consumes contest realtime events and contest-facing relay ports without owning contest state.",
		moduleDependencyEventConsumer,
		moduleDependencyProviderContract,
		moduleDependencyPortBoundary,
	),
	"ops -> identity": reviewedDependency(
		"ops reads identity-owned user facts through contracts for notification and admin views.",
		moduleDependencyProviderContract,
	),
	"ops -> practice": reviewedDependency(
		"ops consumes practice accepted-flag events through practice contracts for notifications.",
		moduleDependencyEventConsumer,
		moduleDependencyProviderContract,
	),
	"practice -> challenge": reviewedDependency(
		"practice reads challenge-owned facts through contracts for submissions, topology, and progress views.",
		moduleDependencyProviderContract,
	),
	"practice -> container_runtime": reviewedDependency(
		"practice uses container_runtime capability contracts and ports for managed practice instances.",
		moduleDependencyRuntimeCapability,
		moduleDependencyProviderContract,
		moduleDependencyPortBoundary,
	),
	"practice -> contest": reviewedDependency(
		"practice reads contest-owned facts through contracts for contest-scoped practice behavior.",
		moduleDependencyProviderContract,
	),
	"practice -> identity": reviewedDependency(
		"practice reads identity-owned user facts through contracts for submissions and progress views.",
		moduleDependencyProviderContract,
	),
	"practice -> instance": reviewedDependency(
		"practice consumes instance-owned contracts for practice instance lifecycle and runtime context.",
		moduleDependencyProviderContract,
	),
	"teaching_analysis -> challenge": reviewedDependency(
		"teaching_analysis aggregates teacher-facing read models from challenge-owned contracts.",
		moduleDependencyQueryAggregation,
		moduleDependencyProviderContract,
	),
	"teaching_analysis -> identity": reviewedDependency(
		"teaching_analysis aggregates teacher-facing read models from identity-owned contracts.",
		moduleDependencyQueryAggregation,
		moduleDependencyProviderContract,
	),
	"teaching_analysis -> contest": reviewedDependency(
		"teaching_analysis aggregates teacher-facing read models from contest-owned contracts.",
		moduleDependencyQueryAggregation,
		moduleDependencyProviderContract,
	),
	"teaching_analysis -> assessment": reviewedDependency(
		"teaching_analysis reuses assessment-owned recommendation contracts for teacher-facing student review.",
		moduleDependencyQueryAggregation,
		moduleDependencyProviderContract,
	),
}

var reviewedTransactionBoundaryFunctions = map[string]struct{}{
	"challenge/infrastructure/repository.go#CreateWithHints":                                                      {},
	"challenge/infrastructure/repository.go#UpdateWithHints":                                                      {},
	"challenge/infrastructure/tag_repository.go#AttachTagsInTx":                                                   {},
	"contest/infrastructure/contest_awd_runtime_recovery_repository.go#AddPausedDurationToActiveAWDContests":      {},
	"contest/infrastructure/awd_runtime_state_repository.go#BumpAWDDefenseWorkspaceRevision":                      {},
	"contest/infrastructure/contest_status_update_repository.go#UpdateContestWithRealtimeRelay":                   {},
	"contest/infrastructure/contest_status_update_repository.go#applyStatusTransitionWithUpdatesAndRealtimeRelay": {},
	"contest/infrastructure/team_membership_lifecycle_repository.go#CreateWithMember":                             {},
	"contest/infrastructure/team_membership_lifecycle_repository.go#DeleteWithMembers":                            {},
	"contest/infrastructure/team_membership_repository.go#AddMemberWithLock":                                      {},
	"contest/infrastructure/team_membership_repository.go#RemoveMember":                                           {},
	"identity/infrastructure/repository.go#Create":                                                                {},
	"identity/infrastructure/repository.go#Update":                                                                {},
	"ops/infrastructure/notification_repository.go#CreateBatch":                                                   {},
	"practice/infrastructure/repository.go#CreateAWDServiceOperation":                                             {},
	"practice/infrastructure/repository.go#ResetInstanceRuntimeForRestart":                                        {},
}
