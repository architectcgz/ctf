package app

import (
	"ctf-platform/tests/system/http/practiceflow"
	"testing"
)

func TestPracticeFlow_PublishedChallengeGeneratesTeacherEvidenceAndAuditTrail(t *testing.T) {
	practiceflow.VerifyPublishedChallengeGeneratesTeacherEvidenceAndAuditTrail(t)
}

func TestPracticeFlow_UnpublishedChallengeCannotBeSolved(t *testing.T) {
	practiceflow.VerifyUnpublishedChallengeCannotBeSolved(t)
}
