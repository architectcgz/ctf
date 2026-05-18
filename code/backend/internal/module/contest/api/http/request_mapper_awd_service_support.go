package http

import (
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func contestAWDServiceResultsToResp(results []contestqry.ContestAWDServiceResult) []*contestcmd.ContestAWDServiceResp {
	resp := make([]*contestcmd.ContestAWDServiceResp, 0, len(results))
	for i := range results {
		item := results[i]
		respItem := contestResponseMapper.ToContestAWDServiceCommandRespPtr(&item)
		if respItem == nil {
			resp = append(resp, nil)
			continue
		}
		respItem.LastPreviewResult = contestResponseMapper.ToAWDCheckerPreviewCommandRespPtr(
			contestdomain.ParseAWDCheckerPreviewResult(item.LastPreviewResultRaw),
		)
		resp = append(resp, respItem)
	}
	return resp
}
