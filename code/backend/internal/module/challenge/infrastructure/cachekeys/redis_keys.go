package cachekeys

import "fmt"

const challengeSolvedCountPrefix = "challenge:solved_count:"

func ChallengeSolvedCountKey(challengeID int64) string {
	return fmt.Sprintf("%s%d", challengeSolvedCountPrefix, challengeID)
}
