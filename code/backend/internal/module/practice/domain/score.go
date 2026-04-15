package domain

import "ctf-platform/internal/model"

func CalculateChallengeScore(challenge *model.Challenge) int {
	if challenge == nil {
		return 0
	}
	if challenge.Points < 0 {
		return 0
	}
	return challenge.Points
}
