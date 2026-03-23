package application

import "math"

func CalculateDynamicScore(baseScore, minScore, decay float64, solveCount int64) int {
	return calculateDynamicScore(baseScore, minScore, decay, solveCount)
}

func calculateDynamicScore(baseScore, minScore, decay float64, solveCount int64) int {
	score := baseScore * math.Pow(decay, float64(solveCount))
	return int(math.Round(math.Max(minScore, score)))
}
