package domain

import (
	"fmt"
	"math"
	"strconv"
)

func CalculateDynamicScore(baseScore, minScore, decay float64, solveCount int64) int {
	score := baseScore * math.Pow(decay, float64(solveCount))
	return int(math.Round(math.Max(minScore, score)))
}

func TeamIDToMember(teamID int64) string {
	return strconv.FormatInt(teamID, 10)
}

func MemberToTeamID(member any) int64 {
	switch value := member.(type) {
	case string:
		id, _ := strconv.ParseInt(value, 10, 64)
		return id
	case []byte:
		id, _ := strconv.ParseInt(string(value), 10, 64)
		return id
	default:
		id, _ := strconv.ParseInt(fmt.Sprint(value), 10, 64)
		return id
	}
}
