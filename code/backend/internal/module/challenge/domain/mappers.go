package domain

import (
	"errors"
	"sort"
	"strings"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

func NormalizeHintModels(reqHints []dto.ChallengeHintReq) ([]*model.ChallengeHint, error) {
	if reqHints == nil {
		return nil, nil
	}

	hints := make([]*model.ChallengeHint, 0, len(reqHints))
	seenLevels := make(map[int]struct{}, len(reqHints))
	for _, reqHint := range reqHints {
		content := strings.TrimSpace(reqHint.Content)
		if content == "" {
			return nil, errcode.ErrInvalidParams.WithCause(errors.New("提示内容不能为空"))
		}
		if _, exists := seenLevels[reqHint.Level]; exists {
			return nil, errcode.ErrInvalidParams.WithCause(errors.New("提示级别不能重复"))
		}
		seenLevels[reqHint.Level] = struct{}{}
		hints = append(hints, &model.ChallengeHint{
			Level:      reqHint.Level,
			Title:      strings.TrimSpace(reqHint.Title),
			CostPoints: reqHint.CostPoints,
			Content:    content,
		})
	}

	sort.Slice(hints, func(i, j int) bool {
		return hints[i].Level < hints[j].Level
	})
	return hints, nil
}

func ChallengeRespFromModel(challenge *model.Challenge, hints []*model.ChallengeHint) *dto.ChallengeResp {
	adminHints := make([]*dto.ChallengeHintAdminResp, 0, len(hints))
	for _, hint := range hints {
		adminHints = append(adminHints, &dto.ChallengeHintAdminResp{
			ID:         hint.ID,
			Level:      hint.Level,
			Title:      hint.Title,
			CostPoints: hint.CostPoints,
			Content:    hint.Content,
		})
	}

	return &dto.ChallengeResp{
		ID:            challenge.ID,
		Title:         challenge.Title,
		Description:   challenge.Description,
		Category:      challenge.Category,
		Difficulty:    challenge.Difficulty,
		Points:        challenge.Points,
		ImageID:       challenge.ImageID,
		AttachmentURL: challenge.AttachmentURL,
		Hints:         adminHints,
		Status:        challenge.Status,
		CreatedAt:     challenge.CreatedAt,
		UpdatedAt:     challenge.UpdatedAt,
	}
}

func ImageRespFromModel(image *model.Image) *dto.ImageResp {
	return &dto.ImageResp{
		ID:          image.ID,
		Name:        image.Name,
		Tag:         image.Tag,
		Description: image.Description,
		Size:        image.Size,
		Status:      image.Status,
		CreatedAt:   image.CreatedAt,
		UpdatedAt:   image.UpdatedAt,
	}
}

func TagRespFromModel(tag *model.Tag) *dto.TagResp {
	return &dto.TagResp{
		ID:          tag.ID,
		Name:        tag.Name,
		Type:        tag.Type,
		Description: tag.Description,
		CreatedAt:   tag.CreatedAt,
	}
}

func AdminWriteupRespFromModel(item *model.ChallengeWriteup) *dto.AdminChallengeWriteupResp {
	return &dto.AdminChallengeWriteupResp{
		ID:          item.ID,
		ChallengeID: item.ChallengeID,
		Title:       item.Title,
		Content:     item.Content,
		Visibility:  item.Visibility,
		ReleaseAt:   item.ReleaseAt,
		CreatedBy:   item.CreatedBy,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}
