package contracts

import (
	"context"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	instanceports "ctf-platform/internal/module/instance/ports"
)

// InstanceCommandService 定义实例 owner 对外暴露的命令能力。
type InstanceCommandService interface {
	DestroyInstance(ctx context.Context, instanceID, userID int64) error
	ExtendInstance(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error)
	DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error
}

// InstanceQueryService 定义实例 owner 对外暴露的查询能力。
type InstanceQueryService interface {
	GetAccessURL(ctx context.Context, instanceID, userID int64) (string, error)
	GetUserInstances(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error)
	ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error)
}

// ProxyTicketClaims 复用实例 owner 的稳定 ticket claim 结构。
type ProxyTicketClaims = instanceports.ProxyTicketClaims

// ProxyTicketService 定义实例 owner 对外暴露的访问 ticket 能力。
type ProxyTicketService interface {
	IssueTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, time.Time, error)
	IssueAWDTargetTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID, victimTeamID int64) (string, time.Time, error)
	IssueAWDDefenseSSHTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64) (string, time.Time, error)
	ResolveTicket(ctx context.Context, ticket string) (*ProxyTicketClaims, error)
	ResolveAWDTargetAccessURL(ctx context.Context, claims *ProxyTicketClaims, contestID, serviceID, victimTeamID int64) (string, error)
}

// AWDDefenseWorkbenchService 定义实例 owner 对外暴露的 AWD defense workbench 能力。
type AWDDefenseWorkbenchService interface {
	ReadAWDDefenseFile(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, filePath string) (*dto.AWDDefenseFileResp, error)
	ListAWDDefenseDirectory(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, dirPath string) (*dto.AWDDefenseDirectoryResp, error)
	SaveAWDDefenseFile(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, req dto.AWDDefenseFileSaveReq) (*dto.AWDDefenseFileSaveResp, error)
	RunAWDDefenseCommand(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, req dto.AWDDefenseCommandReq) (*dto.AWDDefenseCommandResp, error)
}

// MaintenanceService 定义实例 owner 对外暴露的后台维护能力。
type MaintenanceService interface {
	CleanExpiredInstances(ctx context.Context) error
	ReconcileLostActiveRuntimes(ctx context.Context) error
	CleanupOrphans(ctx context.Context) error
}
