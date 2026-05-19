package domain

import practiceentity "ctf-platform/internal/module/practice/entity"

func CalculateChallengeScore(challenge *practiceentity.Challenge) int {
	if challenge == nil {
		return 0
	}
	if challenge.Points < 0 {
		return 0
	}
	return challenge.Points
}
