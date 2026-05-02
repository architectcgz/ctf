package domain

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

func InstanceRespFromModel(inst *model.Instance) *dto.InstanceResp {
	resp := practiceResponseMapperInst.ToInstanceResp(inst)
	resp.Access = dto.BuildInstanceAccessInfo(inst.AccessURL)
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
