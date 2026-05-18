package model

import contestentity "ctf-platform/internal/module/contest/entity"

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
)

type AWDRound = contestentity.AWDRound

type AWDTeamService = contestentity.AWDTeamService

type AWDAttackLog = contestentity.AWDAttackLog

type AWDTrafficEvent = contestentity.AWDTrafficEvent

type AWDProxyTrafficEventInput = contestentity.AWDProxyTrafficEventInput
