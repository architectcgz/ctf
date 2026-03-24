package domain

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/crypto"
)

const (
	AWDCheckSourceScheduler      = "scheduler"
	AWDCheckSourceManualCurrent  = "manual_current_round"
	AWDCheckSourceManualSelected = "manual_selected_round"
	AWDCheckSourceManualService  = "manual_service_check"
)

func AWDRoundRespFromModel(round *model.AWDRound) *dto.AWDRoundResp {
	if round == nil {
		return nil
	}
	return &dto.AWDRoundResp{
		ID:           round.ID,
		ContestID:    round.ContestID,
		RoundNumber:  round.RoundNumber,
		Status:       round.Status,
		StartedAt:    round.StartedAt,
		EndedAt:      round.EndedAt,
		AttackScore:  round.AttackScore,
		DefenseScore: round.DefenseScore,
		CreatedAt:    round.CreatedAt,
		UpdatedAt:    round.UpdatedAt,
	}
}

func AWDTeamServiceRespFromModel(record *model.AWDTeamService, teamName string) *dto.AWDTeamServiceResp {
	if record == nil {
		return nil
	}
	return &dto.AWDTeamServiceResp{
		ID:             record.ID,
		RoundID:        record.RoundID,
		TeamID:         record.TeamID,
		TeamName:       teamName,
		ChallengeID:    record.ChallengeID,
		ServiceStatus:  record.ServiceStatus,
		CheckResult:    ParseAWDCheckResult(record.CheckResult),
		AttackReceived: record.AttackReceived,
		DefenseScore:   record.DefenseScore,
		AttackScore:    record.AttackScore,
		UpdatedAt:      record.UpdatedAt,
	}
}

func AWDAttackLogRespFromModel(record *model.AWDAttackLog, attackerTeam, victimTeam string) *dto.AWDAttackLogResp {
	if record == nil {
		return nil
	}
	return &dto.AWDAttackLogResp{
		ID:             record.ID,
		RoundID:        record.RoundID,
		AttackerTeamID: record.AttackerTeamID,
		AttackerTeam:   attackerTeam,
		VictimTeamID:   record.VictimTeamID,
		VictimTeam:     victimTeam,
		ChallengeID:    record.ChallengeID,
		AttackType:     record.AttackType,
		Source:         NormalizeAWDAttackSource(record.Source),
		SubmittedFlag:  record.SubmittedFlag,
		IsSuccess:      record.IsSuccess,
		ScoreGained:    record.ScoreGained,
		CreatedAt:      record.CreatedAt,
	}
}

func NormalizeAWDAttackSource(value string) string {
	switch strings.TrimSpace(value) {
	case model.AWDAttackSourceManual:
		return model.AWDAttackSourceManual
	case model.AWDAttackSourceSubmission:
		return model.AWDAttackSourceSubmission
	default:
		return model.AWDAttackSourceLegacy
	}
}

func NormalizeAWDCheckSource(value any) string {
	raw, ok := value.(string)
	if !ok {
		return ""
	}
	switch strings.TrimSpace(raw) {
	case AWDCheckSourceScheduler:
		return AWDCheckSourceScheduler
	case AWDCheckSourceManualCurrent:
		return AWDCheckSourceManualCurrent
	case AWDCheckSourceManualSelected:
		return AWDCheckSourceManualSelected
	case AWDCheckSourceManualService:
		return AWDCheckSourceManualService
	default:
		return ""
	}
}

func NormalizedAWDCheckSource(value string) string {
	switch strings.TrimSpace(value) {
	case AWDCheckSourceManualCurrent:
		return AWDCheckSourceManualCurrent
	case AWDCheckSourceManualSelected:
		return AWDCheckSourceManualSelected
	case AWDCheckSourceManualService:
		return AWDCheckSourceManualService
	default:
		return AWDCheckSourceScheduler
	}
}

func MarshalAWDCheckResult(value map[string]any) (string, error) {
	if len(value) == 0 {
		return "{}", nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func NormalizeManualAWDCheckResult(value map[string]any) map[string]any {
	result := make(map[string]any, len(value)+2)
	for key, item := range value {
		result[key] = item
	}
	result["check_source"] = AWDCheckSourceManualService
	if checkedAt, ok := result["checked_at"].(string); !ok || strings.TrimSpace(checkedAt) == "" {
		result["checked_at"] = time.Now().UTC().Format(time.RFC3339)
	}
	return result
}

func ParseAWDCheckResult(value string) map[string]any {
	if strings.TrimSpace(value) == "" {
		return map[string]any{}
	}
	result := make(map[string]any)
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return map[string]any{}
	}
	return result
}

func SortAWDSummaryItems(items []*dto.AWDRoundSummaryItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].TotalScore != items[j].TotalScore {
			return items[i].TotalScore > items[j].TotalScore
		}
		return items[i].TeamID < items[j].TeamID
	})
}

func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique")
}

func BuildAWDRoundFlag(contestID int64, roundNumber int, teamID, challengeID int64, secret, prefix string) string {
	nonce := strings.Join([]string{
		"awd",
		strconv.FormatInt(contestID, 10),
		strconv.Itoa(roundNumber),
		strconv.FormatInt(teamID, 10),
		strconv.FormatInt(challengeID, 10),
	}, ":")
	return crypto.GenerateDynamicFlag(teamID, challengeID, secret, nonce, prefix)
}
