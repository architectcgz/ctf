package http

import (
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestqry "ctf-platform/internal/module/contest/application/queries"
)

type ContestListSummaryResp struct {
	DraftCount       int64 `json:"draft_count"`
	RegisteringCount int64 `json:"registering_count"`
	RunningCount     int64 `json:"running_count"`
	FrozenCount      int64 `json:"frozen_count"`
	EndedCount       int64 `json:"ended_count"`
}

type ContestPageResp struct {
	List     []*contestcmd.ContestResp `json:"list"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"page_size"`
	Summary  ContestListSummaryResp    `json:"summary"`
}

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
		List:     contestRequestMapper.ToContestCommandResps(contests),
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
	return contestRequestMapper.ToContestCommandRespPtr(result)
}
