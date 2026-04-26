package commands

import (
	"strings"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func validateAndNormalizeContestAWDFields(
	contest *model.Contest,
	checkerType string,
	checkerConfig map[string]any,
	slaScore int,
	defenseScore int,
) (model.AWDCheckerType, string, error) {
	hasConfig := strings.TrimSpace(checkerType) != "" || len(checkerConfig) > 0 || slaScore > 0 || defenseScore > 0
	if contest == nil {
		return "", "", errcode.ErrInvalidParams
	}
	if contest.Mode != model.ContestModeAWD {
		if hasConfig {
			return "", "", errcode.ErrInvalidParams
		}
		return "", "{}", nil
	}
	if slaScore < 0 || slaScore > contestdomain.AWDMaxServiceSLAScore {
		return "", "", errcode.ErrInvalidParams
	}
	if defenseScore < 0 || defenseScore > contestdomain.AWDMaxServiceDefenseScore {
		return "", "", errcode.ErrInvalidParams
	}

	normalizedType := contestdomain.NormalizeAWDCheckerType(checkerType)
	if len(checkerConfig) > 0 && normalizedType == "" {
		return "", "", errcode.ErrInvalidParams
	}

	rawConfig, err := contestdomain.MarshalAWDCheckerConfig(checkerConfig)
	if err != nil {
		return "", "", errcode.ErrInvalidParams.WithCause(err)
	}
	return normalizedType, rawConfig, nil
}
