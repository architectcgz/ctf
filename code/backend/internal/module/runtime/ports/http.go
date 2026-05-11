package ports

import (
	"context"

	"ctf-platform/internal/model"
	instanceports "ctf-platform/internal/module/instance/ports"
)

type CountRunningRepository interface {
	CountRunning(ctx context.Context) (int64, error)
}

type InstanceLookupRepository = instanceports.InstanceLookupRepository

type InstanceUserLookupRepository = instanceports.InstanceUserLookupRepository

type InstanceAccessRepository = instanceports.InstanceAccessRepository

type UserVisibleInstanceRepository = instanceports.UserVisibleInstanceRepository

type TeacherInstanceQueryRepository = instanceports.TeacherInstanceQueryRepository

type InstanceExtendRepository = instanceports.InstanceExtendRepository

type InstanceStatusRepository = instanceports.InstanceStatusRepository

type RuntimeCleaner = instanceports.RuntimeCleaner

type TeacherInstanceFilter = instanceports.TeacherInstanceFilter

type AWDDefenseWorkspaceLookupRepository interface {
	FindAWDDefenseWorkspace(ctx context.Context, contestID, teamID, serviceID int64) (*model.AWDDefenseWorkspace, error)
}

type AWDDefenseWorkspaceWriteRepository interface {
	UpsertAWDDefenseWorkspace(ctx context.Context, workspace *model.AWDDefenseWorkspace) error
	BumpAWDDefenseWorkspaceRevision(ctx context.Context, contestID, teamID, serviceID, instanceID int64, seedSignature string) error
}

type UserVisibleInstanceRow = instanceports.UserVisibleInstanceRow

type TeacherInstanceRow = instanceports.TeacherInstanceRow

type ProxyTicketClaims = instanceports.ProxyTicketClaims

const (
	ProxyTicketPurposeInstanceAccess = instanceports.ProxyTicketPurposeInstanceAccess
	ProxyTicketPurposeAWDAttack      = instanceports.ProxyTicketPurposeAWDAttack
	ProxyTicketPurposeAWDDefenseSSH  = instanceports.ProxyTicketPurposeAWDDefenseSSH
)

type AWDTargetProxyScope = instanceports.AWDTargetProxyScope

type AWDDefenseSSHScope = instanceports.AWDDefenseSSHScope

type AWDDefenseSSHSession = instanceports.AWDDefenseSSHSession

type ContainerDirectoryEntry struct {
	Name string
	Type string
	Size int64
}

type ProxyTicketStore = instanceports.ProxyTicketStore

type ProxyTicketInstanceReader = instanceports.ProxyTicketInstanceReader

type ProxyTrafficEventRecorder interface {
	RecordRuntimeProxyTrafficEvent(ctx context.Context, instanceID, userID int64, method, requestPath string, statusCode int) error
	RecordAWDProxyTrafficEvent(ctx context.Context, event model.AWDProxyTrafficEventInput) error
}
