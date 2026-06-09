package contracts

import challengeentity "ctf-platform/internal/module/challenge/entity"

type AWDServiceType = challengeentity.AWDServiceType

const (
	AWDServiceTypeWebHTTP        = challengeentity.AWDServiceTypeWebHTTP
	AWDServiceTypeBinaryTCP      = challengeentity.AWDServiceTypeBinaryTCP
	AWDServiceTypeMultiContainer = challengeentity.AWDServiceTypeMultiContainer
)

type AWDDeploymentMode = challengeentity.AWDDeploymentMode

const (
	AWDDeploymentModeSingleContainer = challengeentity.AWDDeploymentModeSingleContainer
	AWDDeploymentModeTopology        = challengeentity.AWDDeploymentModeTopology
)

type AWDChallengeStatus = challengeentity.AWDChallengeStatus

const (
	AWDChallengeStatusDraft     = challengeentity.AWDChallengeStatusDraft
	AWDChallengeStatusPublished = challengeentity.AWDChallengeStatusPublished
	AWDChallengeStatusArchived  = challengeentity.AWDChallengeStatusArchived
)

type AWDReadinessStatus = challengeentity.AWDReadinessStatus

const (
	AWDReadinessStatusPending = challengeentity.AWDReadinessStatusPending
	AWDReadinessStatusPassed  = challengeentity.AWDReadinessStatusPassed
	AWDReadinessStatusFailed  = challengeentity.AWDReadinessStatusFailed
)

type AWDCheckerType = challengeentity.AWDCheckerType

const (
	AWDCheckerTypeLegacyProbe  = challengeentity.AWDCheckerTypeLegacyProbe
	AWDCheckerTypeHTTPStandard = challengeentity.AWDCheckerTypeHTTPStandard
	AWDCheckerTypeTCPStandard  = challengeentity.AWDCheckerTypeTCPStandard
	AWDCheckerTypeScript       = challengeentity.AWDCheckerTypeScript
)

type AWDChallenge = challengeentity.AWDChallenge

const (
	ChallengeStatusDraft     = "draft"
	ChallengeStatusPublished = "published"
	ChallengeStatusArchived  = "archived"
)

const (
	FlagTypeStatic       = "static"
	FlagTypeDynamic      = "dynamic"
	FlagTypeRegex        = "regex"
	FlagTypeManualReview = "manual_review"
)

const (
	InstanceSharingPerUser = "per_user"
	InstanceSharingPerTeam = "per_team"
	InstanceSharingShared  = "shared"
)

const (
	ChallengeTargetProtocolHTTP = challengeentity.ChallengeTargetProtocolHTTP
	ChallengeTargetProtocolTCP  = challengeentity.ChallengeTargetProtocolTCP
)
