package contracts

const EventInstanceStoppingCleanupWakeup = "instance.stopping_cleanup.wakeup"

type InstanceStoppingCleanupWakeupEvent struct {
	InstanceID int64
}
