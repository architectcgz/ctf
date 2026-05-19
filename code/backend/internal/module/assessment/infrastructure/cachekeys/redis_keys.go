package cachekeys

import "fmt"

const redisNamespace = "ctf"

const recommendationUserPrefix = "recommendation:user:"

func withNamespace(key string) string {
	return redisNamespace + ":" + key
}

func RecommendationKey(userID int64) string {
	return withNamespace(fmt.Sprintf("%s%d", recommendationUserPrefix, userID))
}
