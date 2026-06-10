package contracts

import contestentity "ctf-platform/internal/module/contest/entity"

type Contest = contestentity.Contest

const (
	ContestModeJeopardy = contestentity.ContestModeJeopardy
	ContestModeAWD      = contestentity.ContestModeAWD

	ContestStatusDraft        = contestentity.ContestStatusDraft
	ContestStatusRegistration = contestentity.ContestStatusRegistration
	ContestStatusRunning      = contestentity.ContestStatusRunning
	ContestStatusFrozen       = contestentity.ContestStatusFrozen
	ContestStatusEnded        = contestentity.ContestStatusEnded
)

type Team = contestentity.Team

type TeamMember = contestentity.TeamMember

type Submission = contestentity.Submission

const (
	SubmissionReviewStatusNotRequired = contestentity.SubmissionReviewStatusNotRequired
	SubmissionReviewStatusPending     = contestentity.SubmissionReviewStatusPending
	SubmissionReviewStatusApproved    = contestentity.SubmissionReviewStatusApproved
	SubmissionReviewStatusRejected    = contestentity.SubmissionReviewStatusRejected
)

type ContestRegistration = contestentity.ContestRegistration

const (
	ContestRegistrationStatusPending  = contestentity.ContestRegistrationStatusPending
	ContestRegistrationStatusApproved = contestentity.ContestRegistrationStatusApproved
	ContestRegistrationStatusRejected = contestentity.ContestRegistrationStatusRejected
)

type AWDCheckerType = contestentity.AWDCheckerType

const (
	AWDCheckerTypeLegacyProbe  = contestentity.AWDCheckerTypeLegacyProbe
	AWDCheckerTypeHTTPStandard = contestentity.AWDCheckerTypeHTTPStandard
	AWDCheckerTypeTCPStandard  = contestentity.AWDCheckerTypeTCPStandard
	AWDCheckerTypeScript       = contestentity.AWDCheckerTypeScript
)

type AWDCheckerValidationState = contestentity.AWDCheckerValidationState

const (
	AWDCheckerValidationStatePending = contestentity.AWDCheckerValidationStatePending
	AWDCheckerValidationStatePassed  = contestentity.AWDCheckerValidationStatePassed
	AWDCheckerValidationStateFailed  = contestentity.AWDCheckerValidationStateFailed
	AWDCheckerValidationStateStale   = contestentity.AWDCheckerValidationStateStale
)

type ContestChallenge = contestentity.ContestChallenge

type ContestAWDService = contestentity.ContestAWDService

type ContestAWDServiceSnapshot = contestentity.ContestAWDServiceSnapshot

var EncodeContestAWDServiceSnapshot = contestentity.EncodeContestAWDServiceSnapshot

var DecodeContestAWDServiceSnapshot = contestentity.DecodeContestAWDServiceSnapshot

const (
	AWDRoundStatusPending  = contestentity.AWDRoundStatusPending
	AWDRoundStatusRunning  = contestentity.AWDRoundStatusRunning
	AWDRoundStatusFinished = contestentity.AWDRoundStatusFinished

	AWDServiceStatusUp          = contestentity.AWDServiceStatusUp
	AWDServiceStatusDown        = contestentity.AWDServiceStatusDown
	AWDServiceStatusCompromised = contestentity.AWDServiceStatusCompromised

	AWDAttackTypeFlagCapture    = contestentity.AWDAttackTypeFlagCapture
	AWDAttackTypeServiceExploit = contestentity.AWDAttackTypeServiceExploit

	AWDAttackSourceLegacy     = contestentity.AWDAttackSourceLegacy
	AWDAttackSourceManual     = contestentity.AWDAttackSourceManual
	AWDAttackSourceSubmission = contestentity.AWDAttackSourceSubmission

	AWDTrafficSourceRuntimeProxy = contestentity.AWDTrafficSourceRuntimeProxy

	AWDDefenseWorkspaceStatusPending      = contestentity.AWDDefenseWorkspaceStatusPending
	AWDDefenseWorkspaceStatusProvisioning = contestentity.AWDDefenseWorkspaceStatusProvisioning
	AWDDefenseWorkspaceStatusRunning      = contestentity.AWDDefenseWorkspaceStatusRunning
	AWDDefenseWorkspaceStatusFailed       = contestentity.AWDDefenseWorkspaceStatusFailed

	AWDServiceOperationTypeStart    = contestentity.AWDServiceOperationTypeStart
	AWDServiceOperationTypeRestart  = contestentity.AWDServiceOperationTypeRestart
	AWDServiceOperationTypeRecover  = contestentity.AWDServiceOperationTypeRecover
	AWDServiceOperationTypeRecreate = contestentity.AWDServiceOperationTypeRecreate

	AWDServiceOperationRequestedByUser   = contestentity.AWDServiceOperationRequestedByUser
	AWDServiceOperationRequestedByAdmin  = contestentity.AWDServiceOperationRequestedByAdmin
	AWDServiceOperationRequestedBySystem = contestentity.AWDServiceOperationRequestedBySystem

	AWDServiceOperationStatusRequested    = contestentity.AWDServiceOperationStatusRequested
	AWDServiceOperationStatusProvisioning = contestentity.AWDServiceOperationStatusProvisioning
	AWDServiceOperationStatusRecovering   = contestentity.AWDServiceOperationStatusRecovering
	AWDServiceOperationStatusRecovered    = contestentity.AWDServiceOperationStatusRecovered
	AWDServiceOperationStatusSucceeded    = contestentity.AWDServiceOperationStatusSucceeded
	AWDServiceOperationStatusFailed       = contestentity.AWDServiceOperationStatusFailed

	AWDScopeControlScopeTeam        = contestentity.AWDScopeControlScopeTeam
	AWDScopeControlScopeTeamService = contestentity.AWDScopeControlScopeTeamService

	AWDScopeControlTypeRetired                    = contestentity.AWDScopeControlTypeRetired
	AWDScopeControlTypeServiceDisabled            = contestentity.AWDScopeControlTypeServiceDisabled
	AWDScopeControlTypeDesiredReconcileSuppressed = contestentity.AWDScopeControlTypeDesiredReconcileSuppressed
)

type AWDRound = contestentity.AWDRound

type AWDTeamService = contestentity.AWDTeamService

type AWDAttackLog = contestentity.AWDAttackLog

type AWDTrafficEvent = contestentity.AWDTrafficEvent

type AWDProxyTrafficEventInput = contestentity.AWDProxyTrafficEventInput

type AWDDefenseWorkspace = contestentity.AWDDefenseWorkspace

type AWDServiceOperation = contestentity.AWDServiceOperation

type AWDScopeControl = contestentity.AWDScopeControl
