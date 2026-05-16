package runtime

import "testing"

func TestBuildExposesCoreRuntimeServices(t *testing.T) {
	t.Parallel()

	module := Build(Deps{})
	if module == nil {
		t.Fatal("expected runtime module")
	}
	if module.ProvisioningService == nil {
		t.Fatal("expected provisioning service to be constructed by runtime module")
	}
	if module.CleanupService == nil {
		t.Fatal("expected cleanup service to be constructed by runtime module")
	}
}
