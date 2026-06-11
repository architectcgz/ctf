package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
)

func TestCreateSingleContainerRebindsHostPortAfterPublishConflict(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&practiceCommandImageRow{
		ID:        410,
		Name:      "ctf/web",
		Tag:       "v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	createContainerCalls := 0
	reboundBound := false
	releasedOldPort := false
	service := &serviceCore{
		repo: &stubPracticeRepository{
			reserveAvailablePortExcludingFn: func(ctx context.Context, start, end, excludedPort int) (int, error) {
				if excludedPort != 30021 {
					t.Fatalf("expected original conflicted host port to be excluded, got %d", excludedPort)
				}
				return 30022, nil
			},
			bindReservedPortFn: func(ctx context.Context, port int, instanceID int64) error {
				reboundBound = true
				if port != 30022 {
					t.Fatalf("expected rebound host port 30022, got %d", port)
				}
				if instanceID != 9101 {
					t.Fatalf("expected instance 9101, got %d", instanceID)
				}
				return nil
			},
			releaseReservedPortFn: func(ctx context.Context, port int) error {
				t.Fatalf("expected reserved release path to stay unused, got port %d", port)
				return nil
			},
			releasePortForInstanceFn: func(ctx context.Context, port int, instanceID int64) error {
				releasedOldPort = true
				if port != 30021 {
					t.Fatalf("expected old conflicted host port 30021 to be released, got %d", port)
				}
				if instanceID != 9101 {
					t.Fatalf("expected instance 9101 when releasing old port, got %d", instanceID)
				}
				return nil
			},
		},
		imageRepo: challengeinfra.NewImageRepository(db),
		runtimeService: &stubPracticeRuntimeService{
			createContainerFn: func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int, _ int64) (string, string, int, int, error) {
				createContainerCalls++
				switch createContainerCalls {
				case 1:
					if reservedHostPort != 30021 {
						t.Fatalf("expected first attempt to use reserved host port 30021, got %d", reservedHostPort)
					}
					return "", "", 0, 0, runtimeports.ErrPublishedHostPortConflict
				case 2:
					if reservedHostPort != 30022 {
						t.Fatalf("expected retry to use rebound host port 30022, got %d", reservedHostPort)
					}
					if imageName != "ctf/web:v1" {
						t.Fatalf("expected runtime image ref ctf/web:v1, got %q", imageName)
					}
					if env["FLAG"] != "flag{demo}" {
						t.Fatalf("expected flag env to be forwarded, got %+v", env)
					}
					return "single-rebound-ctr", "single-rebound-net", 30022, 8080, nil
				default:
					t.Fatalf("unexpected CreateContainer call #%d", createContainerCalls)
					return "", "", 0, 0, nil
				}
			},
		},
		config: &config.Config{
			Container: config.ContainerConfig{
				PublicHost: "127.0.0.1",
			},
		},
	}
	instance := &instanceentity.Instance{
		ID:          9101,
		ChallengeID: 410,
		HostPort:    30021,
	}
	challenge := &challengecontracts.PracticeRuntimeChallenge{
		ID:       410,
		ImageID:  int64Ptr(410),
		FlagType: challengecontracts.FlagTypeStatic,
	}

	if err := service.createSingleContainer(context.Background(), instance, toPracticeChallenge(challenge), "flag{demo}"); err != nil {
		t.Fatalf("createSingleContainer() error = %v cause=%v", err, errors.Unwrap(err))
	}
	if createContainerCalls != 2 {
		t.Fatalf("expected one retry after publish conflict, got %d calls", createContainerCalls)
	}
	if !reboundBound {
		t.Fatal("expected rebound host port to be reserved and bound")
	}
	if !releasedOldPort {
		t.Fatal("expected old conflicted host port to be released after successful retry")
	}
	if instance.ContainerID != "single-rebound-ctr" || instance.NetworkID != "single-rebound-net" {
		t.Fatalf("unexpected runtime identifiers after rebound retry: %+v", instance)
	}
	if instance.HostPort != 30022 {
		t.Fatalf("expected instance host port to update to rebound port, got %d", instance.HostPort)
	}
	if instance.AccessURL != "http://127.0.0.1:30022" {
		t.Fatalf("unexpected access url after rebound retry: %s", instance.AccessURL)
	}
}

func TestCreateSingleContainerUsesSingleContainerSubnetPool(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&practiceCommandImageRow{
		ID:        411,
		Name:      "ctf/web-single",
		Tag:       "v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	service := &serviceCore{
		imageRepo: challengeinfra.NewImageRepository(db),
		runtimeService: &stubPracticeRuntimeService{
			createTopologyFn: func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
				if req.SubnetPool != practiceports.SubnetPoolSingleContainer {
					t.Fatalf("expected single container subnet pool, got %q", req.SubnetPool)
				}
				return &practiceports.TopologyCreateResult{
					PrimaryContainerID: "single-ctr",
					NetworkID:          "single-net",
					AccessURL:          "http://127.0.0.1:30031",
					RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
						Networks: []runtimecontracts.InstanceRuntimeNetwork{
							{
								Key:       challengecontracts.TopologyDefaultNetworkKey,
								Name:      "ctf-net-default-single",
								NetworkID: "single-net",
								Subnet:    "10.11.0.0/29",
							},
						},
						Containers: []runtimecontracts.InstanceRuntimeContainer{
							{
								NodeKey:         "default",
								ContainerID:     "single-ctr",
								HostPort:        30031,
								ServicePort:     8080,
								ServiceProtocol: challengecontracts.ChallengeTargetProtocolHTTP,
								IsEntryPoint:    true,
								NetworkKeys:     []string{challengecontracts.TopologyDefaultNetworkKey},
							},
						},
					},
				}, nil
			},
		},
		config: &config.Config{},
	}
	instance := &instanceentity.Instance{
		ID:          9102,
		ChallengeID: 411,
		HostPort:    30031,
	}
	challenge := &challengecontracts.PracticeRuntimeChallenge{
		ID:       411,
		ImageID:  int64Ptr(411),
		FlagType: challengecontracts.FlagTypeStatic,
	}

	if err := service.createSingleContainer(context.Background(), instance, toPracticeChallenge(challenge), "flag{demo}"); err != nil {
		t.Fatalf("createSingleContainer() error = %v", err)
	}
}

func TestReserveReboundProvisioningHostPortReleasesUnboundReservationWhenBindFails(t *testing.T) {
	t.Parallel()

	releasedReservedPort := 0
	service := &serviceCore{
		repo: &stubPracticeRepository{
			reserveAvailablePortExcludingFn: func(ctx context.Context, start, end, excludedPort int) (int, error) {
				if excludedPort != 30031 {
					t.Fatalf("expected excluded port 30031, got %d", excludedPort)
				}
				return 30032, nil
			},
			bindReservedPortFn: func(ctx context.Context, port int, instanceID int64) error {
				if port != 30032 || instanceID != 9301 {
					t.Fatalf("unexpected bind args: port=%d instanceID=%d", port, instanceID)
				}
				return errors.New("bind failed")
			},
			releaseReservedPortFn: func(ctx context.Context, port int) error {
				releasedReservedPort = port
				return nil
			},
			releasePortForInstanceFn: func(ctx context.Context, port int, instanceID int64) error {
				t.Fatalf("did not expect instance-aware release during bind rollback: port=%d instanceID=%d", port, instanceID)
				return nil
			},
		},
		config: &config.Config{
			Container: config.ContainerConfig{
				PortRangeStart: 30030,
				PortRangeEnd:   30040,
			},
		},
	}

	instance := &instanceentity.Instance{
		ID:       9301,
		HostPort: 30031,
	}
	err := service.reserveReboundProvisioningHostPort(context.Background(), instance, 30031)
	if err == nil || err.Error() != "bind failed" {
		t.Fatalf("expected bind failure, got %v", err)
	}
	if releasedReservedPort != 30032 {
		t.Fatalf("expected reserved port 30032 to be released, got %d", releasedReservedPort)
	}
	if instance.HostPort != 30031 {
		t.Fatalf("expected instance host port to stay unchanged on bind failure, got %d", instance.HostPort)
	}
}
