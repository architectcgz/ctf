package cachekeys

import "fmt"

const redisNamespace = "ctf"

const (
	contestDetailPrefix       = "contest:detail:"
	contestChallengesPrefix   = "contest:challenges:"
	rankGlobalKey             = "rank:global"
	rankContestUserPrefix     = "rank:contest:%d:user"
	rankContestTeamPrefix     = "rank:contest:%d:team"
	rankContestFrozenPrefix   = "rank:contest:%d:frozen"
	contestFreezeFlagPrefix   = "contest:freeze_flag:"
	contestStatusUpdateLock   = "contest:status_updater:lock"
	awdSchedulerLock          = "awd:scheduler:lock"
	awdRoundLockPrefix        = "awd:round:lock:"
	awdCurrentRoundFormat     = "awd:%d:current_round"
	awdRoundFlagsFormat       = "awd:%d:round:%d:flags"
	awdServiceStatusFormat    = "awd:%d:service_status"
	awdScoreboardFormat       = "awd:%d:scoreboard"
	awdCheckerPreviewTokenFmt = "awd:%d:checker_preview:%s"
)

func withNamespace(key string) string {
	return redisNamespace + ":" + key
}

func ContestDetailKey(contestID int64) string {
	return withNamespace(fmt.Sprintf("%s%d", contestDetailPrefix, contestID))
}

func ContestChallengesKey(contestID int64) string {
	return withNamespace(fmt.Sprintf("%s%d", contestChallengesPrefix, contestID))
}

func RankGlobalKey() string {
	return withNamespace(rankGlobalKey)
}

func RankContestUserKey(contestID int64) string {
	return withNamespace(fmt.Sprintf(rankContestUserPrefix, contestID))
}

func RankContestTeamKey(contestID int64) string {
	return withNamespace(fmt.Sprintf(rankContestTeamPrefix, contestID))
}

func RankContestFrozenKey(contestID int64) string {
	return withNamespace(fmt.Sprintf(rankContestFrozenPrefix, contestID))
}

func ContestFreezeFlagKey(contestID int64) string {
	return withNamespace(fmt.Sprintf("%s%d", contestFreezeFlagPrefix, contestID))
}

func ContestStatusUpdateLockKey() string {
	return withNamespace(contestStatusUpdateLock)
}

func AWDSchedulerLockKey() string {
	return withNamespace(awdSchedulerLock)
}

func AWDRoundLockKey(contestID int64, roundNumber int) string {
	return withNamespace(fmt.Sprintf("%s%d:%d", awdRoundLockPrefix, contestID, roundNumber))
}

func AWDCurrentRoundKey(contestID int64) string {
	return withNamespace(fmt.Sprintf(awdCurrentRoundFormat, contestID))
}

func AWDRoundFlagsKey(contestID int64, roundID int64) string {
	return withNamespace(fmt.Sprintf(awdRoundFlagsFormat, contestID, roundID))
}

func AWDRoundFlagServiceField(teamID, serviceID int64) string {
	return fmt.Sprintf("%d:s:%d", teamID, serviceID)
}

func AWDServiceStatusKey(contestID int64) string {
	return withNamespace(fmt.Sprintf(awdServiceStatusFormat, contestID))
}

func AWDScoreboardKey(contestID int64) string {
	return withNamespace(fmt.Sprintf(awdScoreboardFormat, contestID))
}

func AWDCheckerPreviewTokenKey(contestID int64, token string) string {
	return withNamespace(fmt.Sprintf(awdCheckerPreviewTokenFmt, contestID, token))
}
