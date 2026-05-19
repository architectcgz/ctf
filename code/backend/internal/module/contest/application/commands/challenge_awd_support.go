package commands

import (
	"strings"

	"ctf-platform/internal/apperror"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
)

func validateAndNormalizeContestAWDFields(
	contest *contestentity.Contest,
	checkerType string,
	checkerConfig map[string]any,
	slaScore int,
	defenseScore int,
) (contestentity.AWDCheckerType, string, error) {
	hasConfig := strings.TrimSpace(checkerType) != "" || len(checkerConfig) > 0 || slaScore > 0 || defenseScore > 0
	if contest == nil {
		return "", "", apperror.ErrInvalidParams
	}
	if contest.Mode != contestentity.ContestModeAWD {
		if hasConfig {
			return "", "", apperror.ErrInvalidParams
		}
		return "", "{}", nil
	}
	if slaScore < 0 || slaScore > contestdomain.AWDMaxServiceSLAScore {
		return "", "", apperror.ErrInvalidParams
	}
	if defenseScore < 0 || defenseScore > contestdomain.AWDMaxServiceDefenseScore {
		return "", "", apperror.ErrInvalidParams
	}

	normalizedType := contestdomain.NormalizeAWDCheckerType(checkerType)
	if len(checkerConfig) > 0 && normalizedType == "" {
		return "", "", apperror.ErrInvalidParams
	}

	rawConfig, err := contestdomain.MarshalAWDCheckerConfig(checkerConfig)
	if err != nil {
		return "", "", apperror.ErrInvalidParams.WithCause(err)
	}
	return normalizedType, rawConfig, nil
}
