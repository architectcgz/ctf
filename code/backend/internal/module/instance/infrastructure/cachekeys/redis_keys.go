package cachekeys

const redisNamespace = "ctf"

const (
	containerCleanupLockKey        = "container:cleanup:lock"
	stoppingCleanupLockKey         = "instance:stopping-cleanup:lock"
	platformRuntimeStateKey        = "platform:runtime:state"
	platformRuntimeRecoveryLockKey = "platform:runtime:recovery:lock"
)

func withNamespace(key string) string {
	return redisNamespace + ":" + key
}

func ContainerCleanupLockKey() string {
	return withNamespace(containerCleanupLockKey)
}

func StoppingCleanupLockKey() string {
	return withNamespace(stoppingCleanupLockKey)
}

func PlatformRuntimeStateKey() string {
	return withNamespace(platformRuntimeStateKey)
}

func PlatformRuntimeRecoveryLockKey() string {
	return withNamespace(platformRuntimeRecoveryLockKey)
}
