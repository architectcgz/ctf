package app

import (
	"ctf-platform/tests/system/http/practiceflow"
	"testing"
)

func TestPracticeFlow_PublishedChallengeLifecycleAndAccess(t *testing.T) {
	practiceflow.VerifyPublishedChallengeLifecycleAndAccess(t)
}

func TestPracticeFlow_PublishedChallengeSubmissionsAndProgress(t *testing.T) {
	practiceflow.VerifyPublishedChallengeSubmissionsAndProgress(t)
}
