package ports

import (
	"context"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
)

type AWDDefenseWorkspaceLookupRepository interface {
	FindAWDDefenseWorkspace(ctx context.Context, contestID, teamID, serviceID int64) (*runtimeentity.AWDDefenseWorkspace, error)
}

type AWDDefenseWorkspaceWriteRepository interface {
	UpsertAWDDefenseWorkspace(ctx context.Context, workspace *runtimeentity.AWDDefenseWorkspace) error
	BumpAWDDefenseWorkspaceRevision(ctx context.Context, contestID, teamID, serviceID, instanceID int64, seedSignature string) error
}

type ContainerDirectoryEntry struct {
	Name string
	Type string
	Size int64
}

type ProxyTrafficEventRecorder interface {
	RecordRuntimeProxyTrafficEvent(ctx context.Context, instanceID, userID int64, method, requestPath string, statusCode int) error
	RecordAWDProxyTrafficEvent(ctx context.Context, event contestcontracts.AWDProxyTrafficEventInput) error
}
