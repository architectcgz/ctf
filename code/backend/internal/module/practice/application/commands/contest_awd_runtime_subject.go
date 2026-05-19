package commands

import (
	practiceentity "ctf-platform/internal/module/practice/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
)

func buildContestAWDServiceVirtualChallenge(subject *practiceports.ContestAWDServiceRuntimeSubject) *practiceentity.Challenge {
	if subject == nil || subject.RuntimeChallenge == nil {
		return nil
	}
	chal := *subject.RuntimeChallenge
	return &chal
}

func buildContestAWDServiceVirtualTopology(subject *practiceports.ContestAWDServiceRuntimeSubject) *practiceports.RuntimeChallengeTopology {
	if subject == nil || subject.RuntimeTopology == nil {
		return nil
	}
	topology := *subject.RuntimeTopology
	return &topology
}
