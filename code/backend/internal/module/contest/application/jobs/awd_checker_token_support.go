package jobs

import (
	"errors"
	"strings"

	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) resolveAWDCheckerToken(definition contestports.AWDServiceDefinition, contestID, teamID int64) (string, error) {
	if token := strings.TrimSpace(definition.CheckerToken); token != "" {
		return token, nil
	}
	if strings.TrimSpace(definition.CheckerTokenEnv) == "" {
		return "", nil
	}
	if strings.TrimSpace(u.flagSecret) == "" {
		return "", errors.New("checker token secret is not configured")
	}
	if contestID <= 0 || teamID <= 0 || definition.ServiceID <= 0 || definition.AWDChallengeID <= 0 {
		return "", errors.New("checker token scope is incomplete")
	}
	token := contestdomain.BuildAWDCheckerToken(contestID, teamID, definition.ServiceID, definition.AWDChallengeID, u.flagSecret)
	if strings.TrimSpace(token) == "" {
		return "", errors.New("checker token secret is not configured")
	}
	return token, nil
}

func (u *AWDRoundUpdater) resolveAWDCheckerTokenForInstance(
	definition contestports.AWDServiceDefinition,
	instance contestports.AWDServiceInstance,
	contestID, teamID int64,
) (string, error) {
	if token := strings.TrimSpace(definition.CheckerToken); token != "" {
		return token, nil
	}
	if token := loadAWDCheckerTokenFromRuntimeDetails(instance.RuntimeDetails, definition.CheckerTokenEnv); token != "" {
		return token, nil
	}
	return u.resolveAWDCheckerToken(definition, contestID, teamID)
}

func loadAWDCheckerTokenFromRuntimeDetails(rawRuntimeDetails string, checkerTokenEnv string) string {
	if strings.TrimSpace(rawRuntimeDetails) == "" {
		return ""
	}
	details, err := runtimecontracts.DecodeInstanceRuntimeDetails(rawRuntimeDetails)
	if err != nil {
		return ""
	}
	return details.FindAWDCheckerToken(checkerTokenEnv)
}
