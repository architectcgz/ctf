package commands

import (
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	practiceentity "ctf-platform/internal/module/practice/entity"
)

func toPracticeChallenge(challenge *challengecontracts.PracticeRuntimeChallenge) *practiceentity.Challenge {
	if challenge == nil {
		return nil
	}
	return &practiceentity.Challenge{
		ID:              challenge.ID,
		PackageSlug:     challenge.PackageSlug,
		Title:           challenge.Title,
		Category:        challenge.Category,
		Difficulty:      challenge.Difficulty,
		Points:          challenge.Points,
		ImageID:         challenge.ImageID,
		Status:          string(challenge.Status),
		FlagType:        challenge.FlagType,
		FlagHash:        challenge.FlagHash,
		FlagSalt:        challenge.FlagSalt,
		FlagRegex:       challenge.FlagRegex,
		FlagPrefix:      challenge.FlagPrefix,
		InstanceSharing: string(challenge.InstanceSharing),
		TargetProtocol:  challenge.TargetProtocol,
		TargetPort:      challenge.TargetPort,
	}
}
