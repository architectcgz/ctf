package contest

import "math"

func calculateDynamicScore(baseScore, minScore, decay float64, solveCount int64) int {
	score := baseScore * math.Pow(decay, float64(solveCount))
	return int(math.Round(math.Max(minScore, score)))
}
