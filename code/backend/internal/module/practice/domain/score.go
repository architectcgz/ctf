package domain

import "ctf-platform/internal/model"

var difficultyWeights = map[string]float64{
	model.ChallengeDifficultyBeginner: 1.0,
	model.ChallengeDifficultyEasy:     1.2,
	model.ChallengeDifficultyMedium:   1.5,
	model.ChallengeDifficultyHard:     2.0,
	model.ChallengeDifficultyInsane:   3.0,
}

func CalculateChallengeScore(challenge *model.Challenge) int {
	if challenge == nil {
		return 0
	}

	weight := difficultyWeights[challenge.Difficulty]
	if weight == 0 {
		weight = 1.0
	}

	return int(float64(challenge.Points) * weight)
}
