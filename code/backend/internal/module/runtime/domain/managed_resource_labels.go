package domain

import "fmt"

const (
	ProjectLabelKey             = "ctf.project"
	ProjectLabelValue           = "ctf"
	ManagedByLabelKey           = "managed-by"
	ManagedByLabelValue         = "ctf-platform"
	ChallengeInstanceLabelKey   = "ctf-component"
	ChallengeInstanceLabelValue = "challenge-instance"
	CheckerRoleLabelKey         = "ctf.role"
	CheckerRoleLabelValue       = "checker-sandbox"
)

// ProjectFilter 返回 ctf 项目资源的统一标签过滤条件。
func ProjectFilter() string {
	return fmt.Sprintf("%s=%s", ProjectLabelKey, ProjectLabelValue)
}

// ManagedByFilter 返回受管容器/网络的统一标签过滤条件。
func ManagedByFilter() string {
	return fmt.Sprintf("%s=%s", ManagedByLabelKey, ManagedByLabelValue)
}

// ManagedProjectLabels 返回平台受管资源共享的基础标签。
func ManagedProjectLabels() map[string]string {
	return map[string]string{
		ProjectLabelKey:   ProjectLabelValue,
		ManagedByLabelKey: ManagedByLabelValue,
	}
}

// ChallengeInstanceLabels 返回题目实例容器/网络的统一标签。
func ChallengeInstanceLabels() map[string]string {
	labels := ManagedProjectLabels()
	labels[ChallengeInstanceLabelKey] = ChallengeInstanceLabelValue
	return labels
}
