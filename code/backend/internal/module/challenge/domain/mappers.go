package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
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

func ChallengeRespFromModel(challenge *model.Challenge, hints []*model.ChallengeHint) *dto.ChallengeResp {
	resp := challengeResponseMapperInst.ToChallengeResp(challenge)
	if len(hints) == 0 {
		return resp
	}
	resp.Hints = make([]*dto.ChallengeHintAdminResp, 0, len(hints))
	for _, hint := range hints {
		resp.Hints = append(resp.Hints, challengeResponseMapperInst.ToChallengeHintAdminResp(hint))
	}
	return resp
}

func AWDChallengeRespFromModel(challenge *model.AWDChallenge) *dto.AWDChallengeResp {
	if challenge == nil {
		return nil
	}
	resp := challengeResponseMapperInst.ToAWDChallengeResp(challenge)
	resp.CheckerType = strings.TrimSpace(string(challenge.CheckerType))
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

func ImageRespFromModel(image *model.Image) *dto.ImageResp {
	resp := challengeResponseMapperInst.ToImageResp(image)
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

func TagRespFromModel(tag *model.Tag) *dto.TagResp {
	return challengeResponseMapperInst.ToTagResp(tag)
}

func AdminWriteupRespFromModel(item *model.ChallengeWriteup) *dto.AdminChallengeWriteupResp {
	return challengeResponseMapperInst.ToAdminChallengeWriteupResp(item)
}

func SubmissionWriteupRespFromModel(item *model.SubmissionWriteup) *dto.SubmissionWriteupResp {
	return challengeResponseMapperInst.ToSubmissionWriteupResp(item)
}

func TeacherSubmissionWriteupItemRespFromRecord(item challengeports.TeacherSubmissionWriteupRecord) *dto.TeacherSubmissionWriteupItemResp {
	resp := challengeResponseMapperInst.ToTeacherSubmissionWriteupItemRespBase(&item.Submission)
	resp.StudentUsername = item.StudentUsername
	resp.StudentName = item.StudentName
	resp.StudentNo = item.StudentNo
	resp.ClassName = item.ClassName
	resp.ChallengeTitle = item.ChallengeTitle
	resp.ContentPreview = buildContentPreview(item.Submission.Content)
	return resp
}

func TeacherSubmissionWriteupDetailRespFromRecord(item challengeports.TeacherSubmissionWriteupRecord) *dto.TeacherSubmissionWriteupDetailResp {
	resp := &dto.TeacherSubmissionWriteupDetailResp{
		SubmissionWriteupResp: *SubmissionWriteupRespFromModel(&item.Submission),
		StudentUsername:       item.StudentUsername,
		StudentName:           item.StudentName,
		StudentNo:             item.StudentNo,
		ClassName:             item.ClassName,
		ChallengeTitle:        item.ChallengeTitle,
	}
	return resp
}

func RecommendedSolutionRespFromRecord(item challengeports.RecommendedSolutionRecord) *dto.RecommendedChallengeSolutionResp {
	resp := challengeResponseMapperInst.ToRecommendedChallengeSolutionRespBase(item)
	resp.ID = item.SourceType + "-" + strconv.FormatInt(item.SourceID, 10)
	return resp
}

func CommunitySolutionRespFromRecord(item challengeports.CommunitySolutionRecord) *dto.CommunityChallengeSolutionResp {
	resp := challengeResponseMapperInst.ToCommunityChallengeSolutionRespBase(&item.Submission)
	resp.ContentPreview = buildContentPreview(item.Submission.Content)
	resp.AuthorName = item.AuthorName
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
