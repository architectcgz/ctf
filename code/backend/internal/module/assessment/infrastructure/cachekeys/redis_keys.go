package cachekeys

import "fmt"

const redisNamespace = "ctf"

const recommendationUserPrefix = "recommendation:user:"
const publishedDimensionTotalsKey = "assessment:published_dimension_totals"

func withNamespace(key string) string {
	return redisNamespace + ":" + key
}

func RecommendationKey(userID int64) string {
	return withNamespace(fmt.Sprintf("%s%d", recommendationUserPrefix, userID))
}

func PublishedDimensionTotalsKey() string {
	return withNamespace(publishedDimensionTotalsKey)
}
