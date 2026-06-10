package runtime

import "testing"

func TestBuildExposesCoreContainerRuntimeServices(t *testing.T) {
	t.Parallel()

	module := Build(Deps{})
	if module == nil {
		t.Fatal("expected container runtime module")
	}
	if module.ProvisioningService == nil {
		t.Fatal("expected provisioning service to be constructed by container runtime module")
	}
	if module.CleanupService == nil {
		t.Fatal("expected cleanup service to be constructed by container runtime module")
	}
	if module.ContainerFiles == nil {
		t.Fatal("expected container file writer to be constructed by container runtime module")
	}
}
