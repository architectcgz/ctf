package jobs

import (
	"strconv"
	"strings"
)

type awdHTTPCheckerTemplateData struct {
	Flag        string
	Round       int
	TeamID      int64
	ChallengeID int64
}

func renderAWDHTTPCheckerTemplate(templateValue string, data awdHTTPCheckerTemplateData) string {
	replacer := strings.NewReplacer(
		"{{FLAG}}", data.Flag,
		"{{ROUND}}", strconv.Itoa(data.Round),
		"{{TEAM_ID}}", strconv.FormatInt(data.TeamID, 10),
		"{{CHALLENGE_ID}}", strconv.FormatInt(data.ChallengeID, 10),
	)
	return replacer.Replace(templateValue)
}
