package domain

import contestcontracts "ctf-platform/internal/module/contest/contracts"

func BuildAWDCheckerToken(contestID, teamID, serviceID, challengeID int64, secret string) string {
	return contestcontracts.BuildAWDCheckerToken(contestID, teamID, serviceID, challengeID, secret)
}

func BuildAWDCheckerPreviewToken(contestID, serviceID, challengeID int64, secret string) string {
	return contestcontracts.BuildAWDCheckerPreviewToken(contestID, serviceID, challengeID, secret)
}
