package domain

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

func InstanceRespFromModel(inst *model.Instance, publicHost, accessHost string) *dto.InstanceResp {
	resp := practiceResponseMapperInst.ToInstanceRespBasePtr(inst)
	if resp == nil {
		return nil
	}
	resp.AccessURL = model.ResolveRuntimePublicAccessURL(inst.AccessURL, publicHost, accessHost)
	resp.Access = dto.BuildInstanceAccessInfo(resp.AccessURL)
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
