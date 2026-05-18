package commands

import contracts "ctf-platform/internal/module/instance/contracts"

type practiceCommandResponseMapperImpl struct{}

func (c *practiceCommandResponseMapperImpl) ToAdminAWDInstanceItemResp(source adminAWDInstanceItemRespSource) AdminAWDInstanceItemResp {
	var commandsAdminAWDInstanceItemResp AdminAWDInstanceItemResp
	commandsAdminAWDInstanceItemResp.TeamID = source.TeamID
	commandsAdminAWDInstanceItemResp.ServiceID = source.ServiceID
	commandsAdminAWDInstanceItemResp.Instance = c.pContractsInstanceRespToPContractsInstanceResp(source.Instance)
	return commandsAdminAWDInstanceItemResp
}
func (c *practiceCommandResponseMapperImpl) ToAdminAWDInstanceItemRespPtr(source adminAWDInstanceItemRespSource) *AdminAWDInstanceItemResp {
	commandsAdminAWDInstanceItemResp := c.ToAdminAWDInstanceItemResp(source)
	return &commandsAdminAWDInstanceItemResp
}
func (c *practiceCommandResponseMapperImpl) ToChallengeSubmissionRecordRespBase(source challengeSubmissionRecordRespSource) ChallengeSubmissionRecordResp {
	var commandsChallengeSubmissionRecordResp ChallengeSubmissionRecordResp
	commandsChallengeSubmissionRecordResp.ID = source.ID
	commandsChallengeSubmissionRecordResp.SubmittedAt = CopyTime(source.SubmittedAt)
	return commandsChallengeSubmissionRecordResp
}
func (c *practiceCommandResponseMapperImpl) ToChallengeSubmissionRecordRespBasePtr(source *challengeSubmissionRecordRespSource) *ChallengeSubmissionRecordResp {
	var pCommandsChallengeSubmissionRecordResp *ChallengeSubmissionRecordResp
	if source != nil {
		commandsChallengeSubmissionRecordResp := c.ToChallengeSubmissionRecordRespBase((*source))
		pCommandsChallengeSubmissionRecordResp = &commandsChallengeSubmissionRecordResp
	}
	return pCommandsChallengeSubmissionRecordResp
}
func (c *practiceCommandResponseMapperImpl) pContractsInstanceAccessInfoToPContractsInstanceAccessInfo(source *contracts.InstanceAccessInfo) *contracts.InstanceAccessInfo {
	var pContractsInstanceAccessInfo *contracts.InstanceAccessInfo
	if source != nil {
		var contractsInstanceAccessInfo contracts.InstanceAccessInfo
		contractsInstanceAccessInfo.Protocol = (*source).Protocol
		contractsInstanceAccessInfo.Host = (*source).Host
		contractsInstanceAccessInfo.Port = (*source).Port
		contractsInstanceAccessInfo.Command = (*source).Command
		pContractsInstanceAccessInfo = &contractsInstanceAccessInfo
	}
	return pContractsInstanceAccessInfo
}
func (c *practiceCommandResponseMapperImpl) pContractsInstanceRespToPContractsInstanceResp(source *contracts.InstanceResp) *contracts.InstanceResp {
	var pContractsInstanceResp *contracts.InstanceResp
	if source != nil {
		var contractsInstanceResp contracts.InstanceResp
		contractsInstanceResp.ID = (*source).ID
		contractsInstanceResp.ChallengeID = (*source).ChallengeID
		contractsInstanceResp.Status = (*source).Status
		contractsInstanceResp.ShareScope = (*source).ShareScope
		contractsInstanceResp.AccessURL = (*source).AccessURL
		contractsInstanceResp.Access = c.pContractsInstanceAccessInfoToPContractsInstanceAccessInfo((*source).Access)
		contractsInstanceResp.ExpiresAt = CopyTime((*source).ExpiresAt)
		contractsInstanceResp.ExtendCount = (*source).ExtendCount
		contractsInstanceResp.MaxExtends = (*source).MaxExtends
		contractsInstanceResp.RemainingExtends = (*source).RemainingExtends
		contractsInstanceResp.CreatedAt = CopyTime((*source).CreatedAt)
		pContractsInstanceResp = &contractsInstanceResp
	}
	return pContractsInstanceResp
}
