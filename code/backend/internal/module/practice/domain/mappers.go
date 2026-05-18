package domain

import (
	"ctf-platform/internal/model"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

func InstanceRespFromModel(inst *model.Instance, publicHost, accessHost string) *instancecontracts.InstanceResp {
	resp := practiceResponseMapperInst.ToInstanceRespBasePtr(inst)
	if resp == nil {
		return nil
	}
	resp.AccessURL = model.ResolveRuntimePublicAccessURL(inst.AccessURL, publicHost, accessHost)
	resp.Access = instancecontracts.BuildInstanceAccessInfo(resp.AccessURL)
	resp.RemainingExtends = RemainingExtends(inst)
	return resp
}

func RemainingExtends(inst *model.Instance) int {
	remaining := inst.MaxExtends - inst.ExtendCount
	if remaining < 0 {
		return 0
	}
	return remaining
}
