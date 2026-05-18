package http

import (
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestqry "ctf-platform/internal/module/contest/application/queries"
)

func buildContestPageResp(
	req ListContestsReq,
	contests []*contestqry.ContestResult,
	total int64,
	summary *contestqry.ContestListSummaryResult,
) ContestPageResp {
	page := req.Page
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 {
		size = 20
	}

	resp := ContestPageResp{
		List:     contestResponseMapper.ToContestCommandResps(contests),
		Total:    total,
		Page:     page,
		PageSize: size,
	}
	if summary != nil {
		resp.Summary = ContestListSummaryResp{
			DraftCount:       summary.DraftCount,
			RegisteringCount: summary.RegistrationCount,
			RunningCount:     summary.RunningCount,
			FrozenCount:      summary.FrozenCount,
			EndedCount:       summary.EndedCount,
		}
	}
	return resp
}

func contestResultToResp(result *contestqry.ContestResult) *contestcmd.ContestResp {
	return contestResponseMapper.ToContestCommandRespPtr(result)
}
