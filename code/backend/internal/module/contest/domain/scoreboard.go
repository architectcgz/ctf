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
	id, _ := ParseMemberToTeamID(member)
	return id
}

func ParseMemberToTeamID(member any) (int64, bool) {
	switch value := member.(type) {
	case string:
		id, err := strconv.ParseInt(value, 10, 64)
		return id, err == nil && id > 0
	case []byte:
		id, err := strconv.ParseInt(string(value), 10, 64)
		return id, err == nil && id > 0
	default:
		id, err := strconv.ParseInt(fmt.Sprint(value), 10, 64)
		return id, err == nil && id > 0
	}
}
