package domain

import (
	"strconv"
	"strings"

	"ctf-platform/pkg/crypto"
)

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
