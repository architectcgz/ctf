package commands

import (
	"ctf-platform/internal/dto"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func awdCheckerPreviewResultFromDTO(item *dto.AWDCheckerPreviewResp) *contestdomain.AWDCheckerPreviewResult {
	if item == nil {
		return nil
	}
	return &contestdomain.AWDCheckerPreviewResult{
		CheckerType:   item.CheckerType,
		ServiceStatus: item.ServiceStatus,
		CheckResult:   item.CheckResult,
		PreviewContext: contestdomain.AWDCheckerPreviewContext{
			ServiceID:      item.PreviewContext.ServiceID,
			AccessURL:      item.PreviewContext.AccessURL,
			PreviewFlag:    item.PreviewContext.PreviewFlag,
			RoundNumber:    item.PreviewContext.RoundNumber,
			TeamID:         item.PreviewContext.TeamID,
			AWDChallengeID: item.PreviewContext.AWDChallengeID,
		},
		PreviewToken: item.PreviewToken,
	}
}

func awdCheckerPreviewResultToDTO(item *contestdomain.AWDCheckerPreviewResult) *dto.AWDCheckerPreviewResp {
	if item == nil {
		return nil
	}
	return &dto.AWDCheckerPreviewResp{
		CheckerType:   item.CheckerType,
		ServiceStatus: item.ServiceStatus,
		CheckResult:   item.CheckResult,
		PreviewContext: dto.AWDCheckerPreviewContextResp{
			ServiceID:      item.PreviewContext.ServiceID,
			AccessURL:      item.PreviewContext.AccessURL,
			PreviewFlag:    item.PreviewContext.PreviewFlag,
			RoundNumber:    item.PreviewContext.RoundNumber,
			TeamID:         item.PreviewContext.TeamID,
			AWDChallengeID: item.PreviewContext.AWDChallengeID,
		},
		PreviewToken: item.PreviewToken,
	}
}
