package module

var reviewedApplicationConcreteImportExceptions = map[string]struct{}{}

var reviewedCrossModulePrivateImportExceptions = map[string]struct{}{}

var reviewedDomainInternalImportExceptions = map[string]struct{}{}

var reviewedRuntimeHostExecutorUsageFiles = map[string]struct{}{
	"../app/composition/runtime_module.go":                {},
	"../app/composition/runtime_node_execution_router.go": {},
	"runtime/infrastructure/agentclient/bridge.go":        {},
	"runtime/infrastructure/agentserver/service.go":       {},
	"runtime/infrastructure/engine.go":                    {},
	"runtime/ports/runtime_host_executor.go":              {},
}

var moduleDependencyBaseline = map[string]struct{}{
	"assessment -> contest":        {},
	"assessment -> challenge":      {},
	"assessment -> identity":       {},
	"assessment -> practice":       {},
	"assessment -> teaching_query": {},
	"auth -> identity":             {},
	"challenge -> identity":        {},
	"challenge -> runtime":         {},
	"contest -> auth":              {},
	"contest -> challenge":         {},
	"contest -> identity":          {},
	"contest -> instance":          {},
	"contest -> runtime":           {},
	"instance -> identity":         {},
	"instance -> contest":          {},
	"ops -> auth":                  {},
	"ops -> challenge":             {},
	"ops -> contest":               {},
	"ops -> identity":              {},
	"ops -> practice":              {},
	"practice -> challenge":        {},
	"practice -> contest":          {},
	"practice -> identity":         {},
	"practice -> instance":         {},
	"practice -> runtime":          {},
	"runtime -> contest":           {},
	"teaching_query -> challenge":  {},
	"teaching_query -> identity":   {},
	"teaching_query -> contest":    {},
	"teaching_query -> assessment": {},
}

var reviewedTransactionBoundaryFunctions = map[string]struct{}{
	"challenge/infrastructure/repository.go#CreateWithHints":                                                 {},
	"challenge/infrastructure/repository.go#UpdateWithHints":                                                 {},
	"challenge/infrastructure/tag_repository.go#AttachTagsInTx":                                              {},
	"contest/infrastructure/contest_awd_runtime_recovery_repository.go#AddPausedDurationToActiveAWDContests": {},
	"contest/infrastructure/contest_status_update_repository.go#applyStatusTransitionWithUpdates":            {},
	"contest/infrastructure/team_membership_lifecycle_repository.go#CreateWithMember":                        {},
	"contest/infrastructure/team_membership_lifecycle_repository.go#DeleteWithMembers":                       {},
	"contest/infrastructure/team_membership_repository.go#AddMemberWithLock":                                 {},
	"contest/infrastructure/team_membership_repository.go#RemoveMember":                                      {},
	"identity/infrastructure/repository.go#Create":                                                           {},
	"identity/infrastructure/repository.go#Update":                                                           {},
	"ops/infrastructure/notification_repository.go#CreateBatch":                                              {},
	"practice/infrastructure/repository.go#CreateAWDServiceOperation":                                        {},
	"practice/infrastructure/repository.go#ResetInstanceRuntimeForRestart":                                   {},
	"runtime/infrastructure/repository.go#BumpAWDDefenseWorkspaceRevision":                                   {},
	"runtime/infrastructure/repository.go#finalizeInstanceRuntime":                                           {},
	"runtime/infrastructure/repository.go#updateStatusAndReleasePortWithCurrentStatus":                       {},
}

var reviewedOversizedRuntimeModuleFiles = map[string]struct{}{}

var reviewedTimeNowFiles = map[string]struct{}{}
