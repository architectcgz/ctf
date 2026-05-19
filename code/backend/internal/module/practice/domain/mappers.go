package domain

import (
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
)

func InstanceRespFromModel(inst *instancecontracts.Instance, publicHost, accessHost string) *instancecontracts.InstanceResp {
	resp := practiceResponseMapperInst.ToInstanceRespBasePtr(inst)
	if resp == nil {
		return nil
	}
	resp.AccessURL = runtimecontracts.ResolveRuntimePublicAccessURL(inst.AccessURL, publicHost, accessHost)
	resp.Access = instancecontracts.BuildInstanceAccessInfo(resp.AccessURL)
	resp.RemainingExtends = RemainingExtends(inst)
	return resp
}

func RemainingExtends(inst *instancecontracts.Instance) int {
	remaining := inst.MaxExtends - inst.ExtendCount
	if remaining < 0 {
		return 0
	}
	return remaining
}
