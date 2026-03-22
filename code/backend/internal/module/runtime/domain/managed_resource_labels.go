package domain

import "fmt"

const (
	ManagedByLabelKey           = "managed-by"
	ManagedByLabelValue         = "ctf-platform"
	ChallengeInstanceLabelKey   = "ctf-component"
	ChallengeInstanceLabelValue = "challenge-instance"
)

// ManagedByFilter 返回受管容器/网络的统一标签过滤条件。
func ManagedByFilter() string {
	return fmt.Sprintf("%s=%s", ManagedByLabelKey, ManagedByLabelValue)
}
