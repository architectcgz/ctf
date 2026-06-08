package contracts

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
	ComposeProjectLabelKey      = "com.docker.compose.project"
	ComposeServiceLabelKey      = "com.docker.compose.service"
	ComposeServiceAWD           = "awd"
	ComposeServiceJeopardy      = "jeopardy"
)

// ProjectFilter returns the common label filter for CTF-managed resources.
func ProjectFilter() string {
	return fmt.Sprintf("%s=%s", ProjectLabelKey, ProjectLabelValue)
}

// ManagedByFilter returns the common label filter for platform-managed resources.
func ManagedByFilter() string {
	return fmt.Sprintf("%s=%s", ManagedByLabelKey, ManagedByLabelValue)
}

// ManagedProjectLabels returns the baseline labels shared by platform-managed resources.
func ManagedProjectLabels() map[string]string {
	return map[string]string{
		ProjectLabelKey:        ProjectLabelValue,
		ManagedByLabelKey:      ManagedByLabelValue,
		ComposeProjectLabelKey: ProjectLabelValue,
	}
}

// ChallengeInstanceLabels returns labels used for challenge instance containers and networks.
func ChallengeInstanceLabels(service string) map[string]string {
	labels := ManagedProjectLabels()
	labels[ChallengeInstanceLabelKey] = ChallengeInstanceLabelValue
	labels[ComposeServiceLabelKey] = normalizeComposeService(service)
	return labels
}

// CheckerSandboxLabels returns labels used for AWD checker sandbox containers.
func CheckerSandboxLabels() map[string]string {
	labels := ManagedProjectLabels()
	labels[CheckerRoleLabelKey] = CheckerRoleLabelValue
	labels[ComposeServiceLabelKey] = ComposeServiceAWD
	return labels
}

func normalizeComposeService(service string) string {
	switch service {
	case ComposeServiceAWD:
		return ComposeServiceAWD
	default:
		return ComposeServiceJeopardy
	}
}
