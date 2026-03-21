package ops

import "ctf-platform/internal/module/system"

type Module struct {
	*system.AuditService
}

func NewModule(auditService *system.AuditService) *Module {
	return &Module{AuditService: auditService}
}
