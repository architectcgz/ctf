package commands

import contracts "ctf-platform/internal/module/instance/contracts"

type instanceResponseMapperImpl struct{}

func (c *instanceResponseMapperImpl) ToInstanceResp(source contracts.Instance) contracts.InstanceResp {
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
func (c *instanceResponseMapperImpl) ToInstanceRespPtr(source *contracts.Instance) *contracts.InstanceResp {
	var pContractsInstanceResp *contracts.InstanceResp
	if source != nil {
		contractsInstanceResp := c.ToInstanceResp((*source))
		pContractsInstanceResp = &contractsInstanceResp
	}
	return pContractsInstanceResp
}
