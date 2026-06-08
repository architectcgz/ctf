package module

var reviewedApplicationConcreteImportExceptions = map[string]struct{}{}

var reviewedCrossModulePrivateImportExceptions = map[string]struct{}{}

var reviewedDomainInternalImportExceptions = map[string]struct{}{}

var moduleDependencyBaseline = map[string]struct{}{
	"assessment -> contest":        {},
	"assessment -> challenge":      {},
	"assessment -> identity":       {},
	"assessment -> practice":       {},
	"assessment -> teaching_query": {},
	"auth -> identity":             {},
	"challenge -> contest":         {},
	"challenge -> identity":        {},
	"challenge -> instance":        {},
	"challenge -> runtime":         {},
	"contest -> auth":              {},
	"contest -> challenge":         {},
	"contest -> identity":          {},
	"contest -> instance":          {},
	"contest -> runtime":           {},
	"instance -> identity":         {},
	"instance -> contest":          {},
	"instance -> runtime":          {},
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
	"runtime -> challenge":         {},
	"runtime -> contest":           {},
	"runtime -> identity":          {},
	"runtime -> instance":          {},
	"runtime -> ops":               {},
	"teaching_query -> challenge":  {},
	"teaching_query -> identity":   {},
	"teaching_query -> contest":    {},
	"teaching_query -> assessment": {},
}

var reviewedTransactionBoundaryFiles = map[string]struct{}{
	"challenge/infrastructure/repository.go":                            {},
	"challenge/infrastructure/tag_repository.go":                        {},
	"contest/infrastructure/awd_repository.go":                          {},
	"contest/infrastructure/contest_awd_runtime_recovery_repository.go": {},
	"contest/infrastructure/contest_status_update_repository.go":        {},
	"contest/infrastructure/submission_repository.go":                   {},
	"contest/infrastructure/team_membership_lifecycle_repository.go":    {},
	"contest/infrastructure/team_membership_repository.go":              {},
	"identity/infrastructure/repository.go":                             {},
	"ops/infrastructure/notification_repository.go":                     {},
	"practice/infrastructure/repository.go":                             {},
	"runtime/infrastructure/repository.go":                              {},
}

var reviewedOversizedRuntimeModuleFiles = map[string]struct{}{}

var reviewedTimeNowFiles = map[string]struct{}{}
