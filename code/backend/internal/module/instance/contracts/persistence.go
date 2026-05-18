package contracts

import instanceentity "ctf-platform/internal/module/instance/entity"

type ShareScope = instanceentity.ShareScope

const (
	ShareScopePerUser = instanceentity.ShareScopePerUser
	ShareScopePerTeam = instanceentity.ShareScopePerTeam
	ShareScopeShared  = instanceentity.ShareScopeShared
)

type Instance = instanceentity.Instance

const (
	InstanceStatusPending  = instanceentity.InstanceStatusPending
	InstanceStatusCreating = instanceentity.InstanceStatusCreating
	InstanceStatusRunning  = instanceentity.InstanceStatusRunning
	InstanceStatusStopped  = instanceentity.InstanceStatusStopped
	InstanceStatusExpired  = instanceentity.InstanceStatusExpired
	InstanceStatusFailed   = instanceentity.InstanceStatusFailed
)
