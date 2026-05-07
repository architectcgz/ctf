package commands

import (
	"context"
	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/internal/platform/events"
	"errors"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

type stubPracticeRuntimeService struct {
	cleanupRuntimeFn          func(ctx context.Context, instance *model.Instance) error
	createTopologyFn          func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error)
	createContainerFn         func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error)
	inspectManagedContainerFn func(ctx context.Context, containerID string) (*practiceports.ManagedContainerState, error)
}

func (s *stubPracticeRuntimeService) CleanupRuntime(ctx context.Context, instance *model.Instance) error {
	if s.cleanupRuntimeFn == nil {
		return nil
	}
	return s.cleanupRuntimeFn(ctx, instance)
}

func (s *stubPracticeRuntimeService) CreateTopology(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
	if s.createTopologyFn == nil {
		return nil, errors.New("unexpected CreateTopology call")
	}
	return s.createTopologyFn(ctx, req)
}

func (s *stubPracticeRuntimeService) CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (string, string, int, int, error) {
	if s.createContainerFn == nil {
		return "", "", 0, 0, errors.New("unexpected CreateContainer call")
	}
	return s.createContainerFn(ctx, imageName, env, reservedHostPort)
}

func (s *stubPracticeRuntimeService) InspectManagedContainer(ctx context.Context, containerID string) (*practiceports.ManagedContainerState, error) {
	if s.inspectManagedContainerFn == nil {
		return nil, errors.New("unexpected InspectManagedContainer call")
	}
	return s.inspectManagedContainerFn(ctx, containerID)
}

type stubPracticeEventBus struct {
	publishFn func(ctx context.Context, evt events.Event) error
}

func (s *stubPracticeEventBus) Subscribe(string, events.Handler) {}

func (s *stubPracticeEventBus) Publish(ctx context.Context, evt events.Event) error {
	if s.publishFn != nil {
		return s.publishFn(ctx, evt)
	}
	return nil
}

func requireEventually(t *testing.T, timeout time.Duration, check func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition was not satisfied before timeout")
}

func newPracticeCommandTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("%s/%s.sqlite", t.TempDir(), strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Image{},
		&model.Challenge{},
		&model.ChallengeTopology{},
		&model.Contest{},
		&model.ContestAWDService{},
		&model.ContestRegistration{},
		&model.User{},
		&model.Team{},
		&model.Instance{},
		&model.AWDServiceOperation{},
		&model.AWDDefenseWorkspace{},
		&model.PortAllocation{},
		&model.Submission{},
	); err != nil {
		t.Fatalf("migrate practice command tables: %v", err)
	}
	return db
}

func reserveClosedLoopbackPort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen loopback port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatalf("close loopback listener: %v", err)
	}
	return port
}

func parseHTTPServerEndpoint(t *testing.T, rawURL string) (string, int) {
	t.Helper()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}

	port, err := strconv.Atoi(parsed.Port())
	if err != nil {
		t.Fatalf("parse server port: %v", err)
	}
	return parsed.Hostname(), port
}

func assertAWDDefenseWorkspaceShellNode(t *testing.T, node practiceports.TopologyCreateNode) {
	t.Helper()

	if node.Image != awdDefenseWorkspaceShellImage {
		t.Fatalf("unexpected workspace shell image: %q", node.Image)
	}
	if !reflect.DeepEqual(node.Env, awdDefenseWorkspaceShellEnv) {
		t.Fatalf("unexpected workspace shell env: %+v", node.Env)
	}
	wantCommand := []string{"/bin/sh", "-lc", awdDefenseWorkspaceBootstrapCommand}
	if !reflect.DeepEqual(node.Command, wantCommand) {
		t.Fatalf("unexpected workspace shell command: %+v", node.Command)
	}
}

type stubAssessmentService struct {
	updateFn func(ctx context.Context, userID int64, dimension string) error
}

func (s *stubAssessmentService) UpdateSkillProfileForDimension(ctx context.Context, userID int64, dimension string) error {
	if s.updateFn == nil {
		return nil
	}
	return s.updateFn(ctx, userID, dimension)
}

type stubScoreUpdater struct {
	updateFn func(ctx context.Context, userID int64) error
	lockWait time.Duration
}

func (s *stubScoreUpdater) UpdateUserScore(ctx context.Context, userID int64) error {
	if s.updateFn == nil {
		return nil
	}
	return s.updateFn(ctx, userID)
}

func (s *stubScoreUpdater) lockTimeout() time.Duration {
	return s.lockWait
}

type stubPracticeChallengeContract struct {
	findByIDWithContextFn                func(ctx context.Context, id int64) (*model.Challenge, error)
	findChallengeTopologyByChallengeIDFn func(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error)
}

func (s *stubPracticeChallengeContract) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *stubPracticeChallengeContract) FindChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
	if s.findChallengeTopologyByChallengeIDFn != nil {
		return s.findChallengeTopologyByChallengeIDFn(ctx, challengeID)
	}
	return nil, nil
}

type stubPracticeImageStore struct {
	findByIDFn func(ctx context.Context, id int64) (*model.Image, error)
}

func (s *stubPracticeImageStore) FindByID(ctx context.Context, id int64) (*model.Image, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

type stubPracticeInstanceStore struct {
	findByIDWithContextFn                   func(ctx context.Context, id int64) (*model.Instance, error)
	updateRuntimeWithContextFn              func(ctx context.Context, instance *model.Instance) error
	finishActiveAWDServiceOperationFn       func(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error
	refreshInstanceExpiryWithContextFn      func(ctx context.Context, instanceID int64, expiresAt time.Time) error
	updateStatusAndReleasePortWithContextFn func(ctx context.Context, id int64, status string) error
	findByUserAndChallengeWithContextFn     func(ctx context.Context, userID, challengeID int64) (*model.Instance, error)
}

func (s *stubPracticeInstanceStore) FindByID(ctx context.Context, id int64) (*model.Instance, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *stubPracticeInstanceStore) UpdateRuntime(ctx context.Context, instance *model.Instance) error {
	if s.updateRuntimeWithContextFn != nil {
		return s.updateRuntimeWithContextFn(ctx, instance)
	}
	return nil
}

func (s *stubPracticeInstanceStore) FinishActiveAWDServiceOperationForInstance(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error {
	if s.finishActiveAWDServiceOperationFn != nil {
		return s.finishActiveAWDServiceOperationFn(ctx, instanceID, status, errorMessage, finishedAt)
	}
	return nil
}

func (s *stubPracticeInstanceStore) RefreshInstanceExpiry(ctx context.Context, instanceID int64, expiresAt time.Time) error {
	if s.refreshInstanceExpiryWithContextFn != nil {
		return s.refreshInstanceExpiryWithContextFn(ctx, instanceID, expiresAt)
	}
	return nil
}

func (s *stubPracticeInstanceStore) UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error {
	if s.updateStatusAndReleasePortWithContextFn != nil {
		return s.updateStatusAndReleasePortWithContextFn(ctx, id, status)
	}
	return nil
}

func (s *stubPracticeInstanceStore) FindByUserAndChallenge(ctx context.Context, userID, challengeID int64) (*model.Instance, error) {
	if s.findByUserAndChallengeWithContextFn != nil {
		return s.findByUserAndChallengeWithContextFn(ctx, userID, challengeID)
	}
	return nil, nil
}

func (s *stubPracticeInstanceStore) ListPendingInstances(ctx context.Context, limit int) ([]*model.Instance, error) {
	return []*model.Instance{}, nil
}

func (s *stubPracticeInstanceStore) TryTransitionStatus(ctx context.Context, id int64, fromStatus, toStatus string) (bool, error) {
	return false, nil
}

func (s *stubPracticeInstanceStore) CountInstancesByStatus(ctx context.Context, statuses []string) (int64, error) {
	return 0, nil
}

type practiceServiceContextKey string
