package cachekeys

const redisNamespace = "ctf"

const (
	containerCleanupLockKey        = "container:cleanup:lock"
	platformRuntimeStateKey        = "platform:runtime:state"
	platformRuntimeRecoveryLockKey = "platform:runtime:recovery:lock"
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

func PlatformRuntimeRecoveryLockKey() string {
	return withNamespace(platformRuntimeRecoveryLockKey)
}
