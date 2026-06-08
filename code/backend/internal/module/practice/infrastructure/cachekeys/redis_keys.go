package cachekeys

import "fmt"

const redisNamespace = "ctf"

const (
	userProgressPrefix             = "user:progress:"
	desiredAWDReconcileStatePrefix = "awd:desired_reconcile:state:"
	userScorePrefix                = "score:user:"
	rankingKey                     = "ranking"
	scoreLockPrefix                = "lock:score:"
	provisioningSchedulerLockKey   = "practice:instance:scheduler:lock"
)

func withNamespace(key string) string {
	return redisNamespace + ":" + key
}

func UserProgressKey(userID int64) string {
	return withNamespace(fmt.Sprintf("%s%d", userProgressPrefix, userID))
}

func DesiredAWDReconcileStateKey(contestID, teamID, serviceID int64) string {
	return withNamespace(fmt.Sprintf("%s%d:%d:%d", desiredAWDReconcileStatePrefix, contestID, teamID, serviceID))
}

func UserScoreKey(userID int64) string {
	return withNamespace(fmt.Sprintf("%s%d", userScorePrefix, userID))
}

func RankingKey() string {
	return withNamespace(rankingKey)
}

func ScoreLockKey(userID int64) string {
	return withNamespace(fmt.Sprintf("%s%d", scoreLockPrefix, userID))
}

func ProvisioningSchedulerLockKey() string {
	return withNamespace(provisioningSchedulerLockKey)
}
