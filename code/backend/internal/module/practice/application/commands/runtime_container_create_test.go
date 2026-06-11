package commands

import (
	"context"
	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
	"testing"
	"time"
)

func TestBuildTopologyCreateRequestKeepsFineGrainedPolicies(t *testing.T) {
	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&practiceCommandImageRow{ID: 1, Name: "ctf/web", Tag: "v1", Status: challengecontracts.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	service := &serviceCore{
		imageRepo: challengeinfra.NewImageRepository(db),
		config:    &config.Config{},
	}

	request, err := service.buildTopologyCreateRequest(context.Background(), 30001, false, toPracticeChallenge(&challengecontracts.PracticeRuntimeChallenge{ImageID: int64Ptr(1)}), "web", challengecontracts.TopologySpec{
		Nodes: []challengecontracts.TopologyNode{
			{Key: "web", ServicePort: 8080, InjectFlag: true},
		},
		Policies: []challengecontracts.TopologyTrafficPolicy{
			{SourceNodeKey: "web", TargetNodeKey: "web", Action: challengecontracts.TopologyPolicyActionAllow, Protocol: challengecontracts.TopologyPolicyProtocolTCP, Ports: []int{8080}},
		},
	}, "flag{demo}")
	if err != nil {
		t.Fatalf("buildTopologyCreateRequest() error = %v", err)
	}
	if len(request.Policies) != 1 {
		t.Fatalf("expected fine-grained policy to be kept, got %+v", request.Policies)
	}
	if request.Policies[0].Protocol != challengecontracts.TopologyPolicyProtocolTCP {
		t.Fatalf("unexpected policy protocol: %+v", request.Policies[0])
	}
}

func TestBuildTopologyCreateRequestRejectsSharedChallengeFlagInjection(t *testing.T) {
	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&practiceCommandImageRow{ID: 2, Name: "ctf/web", Tag: "v2", Status: challengecontracts.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	service := &serviceCore{
		imageRepo: challengeinfra.NewImageRepository(db),
		config:    &config.Config{},
	}

	_, err := service.buildTopologyCreateRequest(context.Background(), 30002, false, toPracticeChallenge(&challengecontracts.PracticeRuntimeChallenge{
		ImageID:         int64Ptr(2),
		InstanceSharing: challengecontracts.InstanceSharingShared,
	}), "web", challengecontracts.TopologySpec{
		Nodes: []challengecontracts.TopologyNode{
			{Key: "web", ServicePort: 8080, InjectFlag: true},
		},
	}, "flag{demo}")
	if err == nil || err.Error() != apperror.ErrInvalidParams.Error() {
		t.Fatalf("expected invalid params for shared topology flag injection, got %v", err)
	}
}

func TestBuildRuntimeContainerNameUsesChallengeSlugAndContestIdentity(t *testing.T) {
	t.Parallel()

	contestID := int64(8)
	teamID := int64(15)
	serviceID := int64(21)
	packageSlug := "Bank Portal"

	got := buildRuntimeContainerName(toPracticeChallenge(&challengecontracts.PracticeRuntimeChallenge{PackageSlug: &packageSlug}), &instanceentity.Instance{
		ContestID: &contestID,
		TeamID:    &teamID,
		ServiceID: &serviceID,
	})
	want := "ctf-instance-bank-portal-c8-t15-s21"
	if got != want {
		t.Fatalf("expected runtime container name %q, got %q", want, got)
	}
}

func TestBuildRuntimeContainerNameIncludesServiceIDWhenChallengeSlugMissing(t *testing.T) {
	t.Parallel()

	contestID := int64(9)
	teamID := int64(16)
	serviceID := int64(22)

	got := buildRuntimeContainerName(toPracticeChallenge(&challengecontracts.PracticeRuntimeChallenge{}), &instanceentity.Instance{
		ContestID: &contestID,
		TeamID:    &teamID,
		ServiceID: &serviceID,
	})
	want := "ctf-instance-challenge-c9-t16-s22"
	if got != want {
		t.Fatalf("expected runtime container name %q, got %q", want, got)
	}
}

func TestApplyAWDStableNetworkToTopologyRequestSkipsContainerNameForMultiNodeTopology(t *testing.T) {
	t.Parallel()

	contestID := int64(10)
	teamID := int64(17)
	serviceID := int64(23)
	packageSlug := "Campus Drive"
	request := &practiceports.TopologyCreateRequest{
		Networks: []practiceports.TopologyCreateNetwork{
			{Key: challengecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []practiceports.TopologyCreateNode{
			{
				Key:          "web",
				IsEntryPoint: true,
				NetworkKeys:  []string{challengecontracts.TopologyDefaultNetworkKey},
			},
			{
				Key:         "worker",
				NetworkKeys: []string{challengecontracts.TopologyDefaultNetworkKey},
			},
		},
	}

	applyAWDStableNetworkToTopologyRequest(&instanceentity.Instance{
		ContestID: &contestID,
		TeamID:    &teamID,
		ServiceID: &serviceID,
	}, toPracticeChallenge(&challengecontracts.PracticeRuntimeChallenge{
		PackageSlug: &packageSlug,
	}), request)

	if request.ContainerName != "" {
		t.Fatalf("expected multi-node AWD topology to skip preferred container name, got %q", request.ContainerName)
	}
	if len(request.Networks) != 1 || request.Networks[0].Name != "ctf-awd-contest-10" || !request.Networks[0].Shared {
		t.Fatalf("expected shared AWD contest network, got %+v", request.Networks)
	}
	if len(request.Nodes[0].NetworkAliases) != 1 || request.Nodes[0].NetworkAliases[0] != "awd-c10-t17-s23" {
		t.Fatalf("expected stable AWD alias on entry node, got %+v", request.Nodes[0].NetworkAliases)
	}
	if len(request.Nodes[1].NetworkAliases) != 0 {
		t.Fatalf("expected non-entry node aliases unchanged, got %+v", request.Nodes[1].NetworkAliases)
	}
}

func stringPtr(value string) *string {
	return &value
}
