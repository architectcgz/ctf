package domain

import (
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

func InstanceRespFromModel(inst *instancecontracts.Instance, publicHost, accessHost string) *instancecontracts.InstanceResp {
	resp := practiceResponseMapperInst.ToInstanceRespBasePtr(inst)
	if resp == nil {
		return nil
	}
	if inst != nil && inst.Status == instancecontracts.InstanceStatusStopping {
		resp.Status = "destroying"
	}
	if resp.Status == instancecontracts.InstanceStatusRunning {
		resp.AccessURL = runtimecontracts.ResolveRuntimePublicAccessURL(inst.AccessURL, publicHost, accessHost)
		resp.Access = instancecontracts.BuildInstanceAccessInfo(resp.AccessURL)
	} else {
		resp.AccessURL = ""
		resp.Access = nil
	}
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
