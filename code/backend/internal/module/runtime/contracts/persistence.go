package contracts

import runtimeentity "ctf-platform/internal/module/runtime/entity"

type AWDDefenseWorkspace = runtimeentity.AWDDefenseWorkspace

const (
	AWDDefenseWorkspaceStatusPending      = runtimeentity.AWDDefenseWorkspaceStatusPending
	AWDDefenseWorkspaceStatusProvisioning = runtimeentity.AWDDefenseWorkspaceStatusProvisioning
	AWDDefenseWorkspaceStatusRunning      = runtimeentity.AWDDefenseWorkspaceStatusRunning
	AWDDefenseWorkspaceStatusFailed       = runtimeentity.AWDDefenseWorkspaceStatusFailed
)

type AWDServiceOperation = runtimeentity.AWDServiceOperation

const (
	AWDServiceOperationTypeStart    = runtimeentity.AWDServiceOperationTypeStart
	AWDServiceOperationTypeRestart  = runtimeentity.AWDServiceOperationTypeRestart
	AWDServiceOperationTypeRecover  = runtimeentity.AWDServiceOperationTypeRecover
	AWDServiceOperationTypeRecreate = runtimeentity.AWDServiceOperationTypeRecreate

	AWDServiceOperationRequestedByUser   = runtimeentity.AWDServiceOperationRequestedByUser
	AWDServiceOperationRequestedByAdmin  = runtimeentity.AWDServiceOperationRequestedByAdmin
	AWDServiceOperationRequestedBySystem = runtimeentity.AWDServiceOperationRequestedBySystem

	AWDServiceOperationStatusRequested    = runtimeentity.AWDServiceOperationStatusRequested
	AWDServiceOperationStatusProvisioning = runtimeentity.AWDServiceOperationStatusProvisioning
	AWDServiceOperationStatusRecovering   = runtimeentity.AWDServiceOperationStatusRecovering
	AWDServiceOperationStatusRecovered    = runtimeentity.AWDServiceOperationStatusRecovered
	AWDServiceOperationStatusSucceeded    = runtimeentity.AWDServiceOperationStatusSucceeded
	AWDServiceOperationStatusFailed       = runtimeentity.AWDServiceOperationStatusFailed
)
