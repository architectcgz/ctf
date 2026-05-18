package queries

import (
	contracts "ctf-platform/internal/module/instance/contracts"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type instanceResponseMapperImpl struct{}

func (c *instanceResponseMapperImpl) ToInstanceInfo(source runtimeports.UserVisibleInstanceRow) contracts.InstanceInfo {
	var contractsInstanceInfo contracts.InstanceInfo
	contractsInstanceInfo.ID = source.ID
	contractsInstanceInfo.ContestMode = source.ContestMode
	contractsInstanceInfo.ChallengeID = source.ChallengeID
	contractsInstanceInfo.ChallengeTitle = source.ChallengeTitle
	contractsInstanceInfo.Category = source.Category
	contractsInstanceInfo.Difficulty = source.Difficulty
	contractsInstanceInfo.FlagType = source.FlagType
	contractsInstanceInfo.ShareScope = source.ShareScope
	contractsInstanceInfo.ExpiresAt = CopyTime(source.ExpiresAt)
	contractsInstanceInfo.ExtendCount = source.ExtendCount
	contractsInstanceInfo.MaxExtends = source.MaxExtends
	contractsInstanceInfo.CreatedAt = CopyTime(source.CreatedAt)
	return contractsInstanceInfo
}
func (c *instanceResponseMapperImpl) ToInstanceInfoPtr(source *runtimeports.UserVisibleInstanceRow) *contracts.InstanceInfo {
	var pContractsInstanceInfo *contracts.InstanceInfo
	if source != nil {
		contractsInstanceInfo := c.ToInstanceInfo((*source))
		pContractsInstanceInfo = &contractsInstanceInfo
	}
	return pContractsInstanceInfo
}
