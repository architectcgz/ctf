package commands

import (
	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
)

func buildContestAWDServiceVirtualChallenge(subject *practiceports.ContestAWDServiceRuntimeSubject) *model.Challenge {
	if subject == nil || subject.RuntimeChallenge == nil {
		return nil
	}
	chal := *subject.RuntimeChallenge
	return &chal
}

func buildContestAWDServiceVirtualTopology(subject *practiceports.ContestAWDServiceRuntimeSubject) *model.ChallengeTopology {
	if subject == nil || subject.RuntimeTopology == nil {
		return nil
	}
	topology := *subject.RuntimeTopology
	return &topology
}
