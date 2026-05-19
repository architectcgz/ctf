package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"ctf-platform/internal/apperror"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

func NormalizeHintModels(reqHints []challengecontracts.ChallengeHintReq) ([]*challengeentity.ChallengeHint, error) {
	if reqHints == nil {
		return nil, nil
	}

	hints := make([]*challengeentity.ChallengeHint, 0, len(reqHints))
	seenLevels := make(map[int]struct{}, len(reqHints))
	for _, reqHint := range reqHints {
		content := strings.TrimSpace(reqHint.Content)
		if content == "" {
			return nil, apperror.ErrInvalidParams.WithCause(errors.New("提示内容不能为空"))
		}
		if _, exists := seenLevels[reqHint.Level]; exists {
			return nil, apperror.ErrInvalidParams.WithCause(errors.New("提示级别不能重复"))
		}
		seenLevels[reqHint.Level] = struct{}{}
		hints = append(hints, &challengeentity.ChallengeHint{
			Level:   reqHint.Level,
			Title:   strings.TrimSpace(reqHint.Title),
			Content: content,
		})
	}

	sort.Slice(hints, func(i, j int) bool {
		return hints[i].Level < hints[j].Level
	})
	return hints, nil
}

func ChallengeRespFromWriteModel(challenge *challengeports.ChallengeWriteModel, hints []*challengeentity.ChallengeHint) *challengecontracts.ChallengeResp {
	resp := challengeResponseMapperInst.ToChallengeRespBasePtr(challenge)
	if resp == nil {
		return nil
	}

	adminHints := make([]*challengecontracts.ChallengeHintAdminResp, 0, len(hints))
	for _, hint := range hints {
		adminHints = append(adminHints, challengeResponseMapperInst.ToChallengeHintAdminRespPtr(hint))
	}
	resp.Hints = adminHints
	return resp
}

func AWDChallengeRespFromModel(challenge *challengeentity.AWDChallenge) *challengecontracts.AWDChallengeResp {
	resp := challengeResponseMapperInst.ToAWDChallengeRespBasePtr(challenge)
	if resp == nil {
		return nil
	}

	resp.CheckerConfig = parseTemplateConfigMap(challenge.CheckerConfig)
	resp.FlagMode = strings.TrimSpace(challenge.FlagMode)
	resp.FlagConfig = parseTemplateConfigMap(challenge.FlagConfig)
	resp.DefenseEntryMode = strings.TrimSpace(challenge.DefenseEntryMode)
	resp.AccessConfig = parseTemplateConfigMap(challenge.AccessConfig)
	resp.RuntimeConfig = parseTemplateConfigMap(challenge.RuntimeConfig)
	return resp
}

func parseTemplateConfigMap(raw string) map[string]any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	result := make(map[string]any)
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func ImageRespFromModel(image *challengeentity.Image) *challengecontracts.ImageResp {
	resp := challengeResponseMapperInst.ToImageRespBasePtr(image)
	if resp == nil {
		return nil
	}
	resp.SizeFormatted = FormatImageSize(image.Size)
	return resp
}

func FormatImageSize(size int64) string {
	if size <= 0 {
		return "0 B"
	}

	value := float64(size)
	units := []string{"B", "KB", "MB", "GB", "TB"}
	unitIndex := 0
	for value >= 1024 && unitIndex < len(units)-1 {
		value = value / 1024
		unitIndex++
	}
	if value == float64(int64(value)) {
		return fmt.Sprintf("%.0f %s", value, units[unitIndex])
	}
	return fmt.Sprintf("%.1f %s", value, units[unitIndex])
}

func TeacherSubmissionWriteupItemRespFromRecord(item challengeports.TeacherSubmissionWriteupRecord) *challengecontracts.TeacherSubmissionWriteupItemResp {
	resp := challengeResponseMapperInst.ToTeacherSubmissionWriteupItemRespBasePtr(&item)
	resp.ContentPreview = buildContentPreview(item.Submission.Content)
	return resp
}

func TeacherSubmissionWriteupDetailRespFromRecord(item challengeports.TeacherSubmissionWriteupRecord) *challengecontracts.TeacherSubmissionWriteupDetailResp {
	return challengeResponseMapperInst.ToTeacherSubmissionWriteupDetailRespPtr(item)
}

func RecommendedSolutionRespFromRecord(item challengeports.RecommendedSolutionRecord) *challengecontracts.RecommendedChallengeSolutionResp {
	resp := challengeResponseMapperInst.ToRecommendedChallengeSolutionRespBasePtr(&item)
	resp.ID = item.SourceType + "-" + strconv.FormatInt(item.SourceID, 10)
	return resp
}

func CommunitySolutionRespFromRecord(item challengeports.CommunitySolutionRecord) *challengecontracts.CommunityChallengeSolutionResp {
	resp := challengeResponseMapperInst.ToCommunityChallengeSolutionRespBasePtr(&item)
	resp.ContentPreview = buildContentPreview(item.Submission.Content)
	return resp
}

func buildContentPreview(content string) string {
	normalized := strings.Join(strings.Fields(strings.TrimSpace(content)), " ")
	runes := []rune(normalized)
	if len(runes) <= 96 {
		return normalized
	}
	return string(runes[:96]) + "..."
}
