package cachekeys

const redisNamespace = "ctf"

const (
	containerCleanupLockKey = "container:cleanup:lock"
	platformRuntimeStateKey = "platform:runtime:state"
)

func withNamespace(key string) string {
	return redisNamespace + ":" + key
}

func ContainerCleanupLockKey() string {
	return withNamespace(containerCleanupLockKey)
}

func PlatformRuntimeStateKey() string {
	return withNamespace(platformRuntimeStateKey)
}
