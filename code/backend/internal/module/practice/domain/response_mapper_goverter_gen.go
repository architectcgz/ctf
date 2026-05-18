package domain

import contracts "ctf-platform/internal/module/instance/contracts"

type practiceResponseMapperImpl struct{}

func (c *practiceResponseMapperImpl) ToInstanceRespBase(source contracts.Instance) contracts.InstanceResp {
	var contractsInstanceResp contracts.InstanceResp
	contractsInstanceResp.ID = source.ID
	contractsInstanceResp.ChallengeID = source.ChallengeID
	contractsInstanceResp.Status = source.Status
	contractsInstanceResp.ShareScope = source.ShareScope
	contractsInstanceResp.AccessURL = source.AccessURL
	contractsInstanceResp.ExpiresAt = CopyTime(source.ExpiresAt)
	contractsInstanceResp.ExtendCount = source.ExtendCount
	contractsInstanceResp.MaxExtends = source.MaxExtends
	contractsInstanceResp.CreatedAt = CopyTime(source.CreatedAt)
	return contractsInstanceResp
}
func (c *practiceResponseMapperImpl) ToInstanceRespBasePtr(source *contracts.Instance) *contracts.InstanceResp {
	var pContractsInstanceResp *contracts.InstanceResp
	if source != nil {
		contractsInstanceResp := c.ToInstanceRespBase((*source))
		pContractsInstanceResp = &contractsInstanceResp
	}
	return pContractsInstanceResp
}
