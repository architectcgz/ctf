package commands

import (
	"ctf-platform/internal/dto"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func awdCheckerPreviewResultFromDTO(item *dto.AWDCheckerPreviewResp) *contestdomain.AWDCheckerPreviewResult {
	if item == nil {
		return nil
	}
	mapped := awdPreviewResultMapper.ToDomain(*item)
	return &mapped
}

func awdCheckerPreviewResultToDTO(item *contestdomain.AWDCheckerPreviewResult) *dto.AWDCheckerPreviewResp {
	if item == nil {
		return nil
	}
	mapped := awdPreviewResultMapper.ToDTO(*item)
	return &mapped
}
