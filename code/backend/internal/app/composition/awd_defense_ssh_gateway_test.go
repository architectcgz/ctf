package composition

import (
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/config"
	containerruntime "ctf-platform/internal/module/container_runtime/runtime"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceports "ctf-platform/internal/module/instance/ports"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

type stubAWDDefenseSSHGatewayProxyTickets struct {
	claims *instanceports.ProxyTicketClaims
	err    error
}

type blockingInteractiveExecutor struct {
	started chan struct{}
	stopped chan struct{}
	once    sync.Once
}

func newBlockingInteractiveExecutor() *blockingInteractiveExecutor {
	return &blockingInteractiveExecutor{
		started: make(chan struct{}),
		stopped: make(chan struct{}),
	}
}

func (e *blockingInteractiveExecutor) ExecContainerInteractive(ctx context.Context, _ string, _ []string, _ io.Reader, _ io.Writer) error {
	e.once.Do(func() {
		close(e.started)
	})
	<-ctx.Done()
	close(e.stopped)
	return ctx.Err()
}

type stubGatewayListener struct{}

func (stubGatewayListener) Accept() (net.Conn, error) {
	return nil, io.EOF
}

func (stubGatewayListener) Close() error {
	return nil
}

func (stubGatewayListener) Addr() net.Addr {
	return &net.TCPAddr{}
}

type stubSSHChannel struct{}

func (stubSSHChannel) Read(_ []byte) (int, error) {
	return 0, io.EOF
}

func (stubSSHChannel) Write(p []byte) (int, error) {
	return len(p), nil
}

func (stubSSHChannel) Close() error {
	return nil
}

func (stubSSHChannel) CloseWrite() error {
	return nil
}

func (stubSSHChannel) SendRequest(_ string, _ bool, _ []byte) (bool, error) {
	return false, nil
}

func (stubSSHChannel) Stderr() io.ReadWriter {
	return stubSSHChannel{}
}

type stubRuntimeHTTPProxyTicketReader struct {
	scope *instanceports.AWDDefenseSSHScope
}

func (s stubRuntimeHTTPProxyTicketReader) FindByID(context.Context, int64) (*instancecontracts.Instance, error) {
	return nil, nil
}

func (s stubRuntimeHTTPProxyTicketReader) FindAWDTargetProxyScope(context.Context, int64, int64, int64, int64) (*instanceports.AWDTargetProxyScope, error) {
	return nil, nil
}

func (s stubRuntimeHTTPProxyTicketReader) FindAWDDefenseSSHScope(context.Context, int64, int64, int64) (*instanceports.AWDDefenseSSHScope, error) {
	return s.scope, nil
}

func (s stubAWDDefenseSSHGatewayProxyTickets) IssueAWDDefenseSSHTicket(context.Context, authctx.CurrentUser, int64, int64) (string, time.Time, error) {
	return "", time.Time{}, nil
}

func (s stubAWDDefenseSSHGatewayProxyTickets) IssueTicket(context.Context, authctx.CurrentUser, int64) (string, time.Time, error) {
	return "", time.Time{}, nil
}

func (s stubAWDDefenseSSHGatewayProxyTickets) IssueAWDTargetTicket(context.Context, authctx.CurrentUser, int64, int64, int64) (string, time.Time, error) {
	return "", time.Time{}, nil
}

func (s stubAWDDefenseSSHGatewayProxyTickets) ResolveTicket(context.Context, string) (*instanceports.ProxyTicketClaims, error) {
	return s.claims, s.err
}

func (s stubAWDDefenseSSHGatewayProxyTickets) ResolveAWDTargetAccessURL(context.Context, *instanceports.ProxyTicketClaims, int64, int64, int64) (string, error) {
	return "", nil
}

func (s stubAWDDefenseSSHGatewayProxyTickets) MaxAge() int {
	return 900
}

func TestAWDDefenseSSHGatewayAuthenticateUsesWorkspaceScope(t *testing.T) {
	t.Parallel()

	contestID := int64(51)
	teamID := int64(61)
	serviceID := int64(71)
	challengeID := int64(81)
	workspaceRevision := int64(7)
	gateway := NewAWDDefenseSSHGateway(
		stubAWDDefenseSSHGatewayProxyTickets{
			claims: &instanceports.ProxyTicketClaims{
				UserID:               1001,
				Username:             "student",
				Role:                 identitycontracts.RoleStudent,
				InstanceID:           9001,
				ContestID:            &contestID,
				ShareScope:           instancecontracts.ShareScopePerTeam,
				Purpose:              instanceports.ProxyTicketPurposeAWDDefenseSSH,
				AWDAttackerTeamID:    &teamID,
				AWDServiceID:         &serviceID,
				AWDChallengeID:       &challengeID,
				AWDWorkspaceRevision: &workspaceRevision,
			},
		},
		stubRuntimeHTTPProxyTicketReader{
			scope: &instanceports.AWDDefenseSSHScope{
				InstanceID:        9001,
				ContestID:         contestID,
				TeamID:            teamID,
				ServiceID:         serviceID,
				AWDChallengeID:    challengeID,
				WorkspaceRevision: workspaceRevision,
				ContainerID:       "workspace-ctr",
				ShareScope:        instancecontracts.ShareScopePerTeam,
			},
		},
		nil,
		"",
		2222,
		nil,
	)

	session, err := gateway.authenticate(context.Background(), "student+51+71", "ticket-secret")
	if err != nil {
		t.Fatalf("authenticate() error = %v", err)
	}
	if session == nil {
		t.Fatal("expected session")
	}
	if session.ContainerID != "workspace-ctr" || session.WorkspaceRevision != workspaceRevision {
		t.Fatalf("unexpected workspace session: %+v", session)
	}
}

func TestAWDDefenseSSHGatewayAuthenticateRejectsStaleWorkspaceRevision(t *testing.T) {
	t.Parallel()

	contestID := int64(52)
	teamID := int64(62)
	serviceID := int64(72)
	challengeID := int64(82)
	claimedRevision := int64(3)
	currentRevision := int64(4)
	gateway := NewAWDDefenseSSHGateway(
		stubAWDDefenseSSHGatewayProxyTickets{
			claims: &instanceports.ProxyTicketClaims{
				UserID:               1002,
				Username:             "student",
				Role:                 identitycontracts.RoleStudent,
				InstanceID:           9002,
				ContestID:            &contestID,
				ShareScope:           instancecontracts.ShareScopePerTeam,
				Purpose:              instanceports.ProxyTicketPurposeAWDDefenseSSH,
				AWDAttackerTeamID:    &teamID,
				AWDServiceID:         &serviceID,
				AWDChallengeID:       &challengeID,
				AWDWorkspaceRevision: &claimedRevision,
			},
		},
		stubRuntimeHTTPProxyTicketReader{
			scope: &instanceports.AWDDefenseSSHScope{
				InstanceID:        9002,
				ContestID:         contestID,
				TeamID:            teamID,
				ServiceID:         serviceID,
				AWDChallengeID:    challengeID,
				WorkspaceRevision: currentRevision,
				ContainerID:       "workspace-ctr",
				ShareScope:        instancecontracts.ShareScopePerTeam,
			},
		},
		nil,
		"",
		2222,
		nil,
	)

	_, err := gateway.authenticate(context.Background(), "student+52+72", "ticket-secret")
	if err == nil || err.Error() != apperror.ErrForbidden.Error() {
		t.Fatalf("expected forbidden error for stale workspace revision, got %v", err)
	}
}

func TestLoadOrCreateAWDDefenseSSHHostKeySignerCreatesAndReusesFile(t *testing.T) {
	t.Parallel()

	hostKeyPath := filepath.Join(t.TempDir(), "runtime", "awd-defense-ssh-host-key.pem")

	firstSigner, err := loadOrCreateAWDDefenseSSHHostKeySigner(hostKeyPath)
	if err != nil {
		t.Fatalf("first loadOrCreateAWDDefenseSSHHostKeySigner() error = %v", err)
	}
	info, err := os.Stat(hostKeyPath)
	if err != nil {
		t.Fatalf("stat host key file: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Fatalf("host key file mode = %o, want 600", mode)
	}

	secondSigner, err := loadOrCreateAWDDefenseSSHHostKeySigner(hostKeyPath)
	if err != nil {
		t.Fatalf("second loadOrCreateAWDDefenseSSHHostKeySigner() error = %v", err)
	}

	firstFingerprint := ssh.FingerprintSHA256(firstSigner.PublicKey())
	secondFingerprint := ssh.FingerprintSHA256(secondSigner.PublicKey())
	if firstFingerprint != secondFingerprint {
		t.Fatalf("expected persistent host key fingerprint, got %q then %q", firstFingerprint, secondFingerprint)
	}
}

func TestLoadOrCreateAWDDefenseSSHHostKeySignerRejectsInvalidExistingFile(t *testing.T) {
	t.Parallel()

	hostKeyPath := filepath.Join(t.TempDir(), "awd-defense-ssh-host-key.pem")
	if err := os.WriteFile(hostKeyPath, []byte("not-a-private-key"), 0o600); err != nil {
		t.Fatalf("write invalid host key file: %v", err)
	}

	_, err := loadOrCreateAWDDefenseSSHHostKeySigner(hostKeyPath)
	if err == nil {
		t.Fatal("expected loadOrCreateAWDDefenseSSHHostKeySigner() to reject invalid host key file")
	}
}

func TestBuildAWDDefenseSSHGatewayUsesNodeRouterInteractiveExecutorWhenAvailable(t *testing.T) {
	t.Parallel()

	cfg, db, cache := newRootTestDependencies(t)
	cfg.Container = config.ContainerConfig{
		DefenseSSHEnabled:     true,
		DefenseSSHHost:        "ssh.ctf.local",
		DefenseSSHPort:        2222,
		DefenseSSHHostKeyPath: filepath.Join(t.TempDir(), "runtime", "awd-defense-ssh-host-key.pem"),
	}

	root, err := BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}

	defaultExecutor := &stubRuntimeNodeHostExecutor{}
	runtime := &ContainerRuntimeModule{
		nodeRouter: &runtimeNodeExecutionRouter{},
		runtime: &containerruntime.Module{
			InteractiveExecutor: defaultExecutor,
		},
	}

	gateway := BuildAWDDefenseSSHGateway(root, runtime)
	if gateway == nil {
		t.Fatal("expected gateway")
	}
	if gateway.executor != runtime.nodeRouter {
		t.Fatalf("expected gateway executor to use node router, got %T", gateway.executor)
	}
}

func TestBuildAWDDefenseSSHGatewayFallsBackToRuntimeInteractiveExecutorWithoutNodeRouter(t *testing.T) {
	t.Parallel()

	cfg, db, cache := newRootTestDependencies(t)
	cfg.Container = config.ContainerConfig{
		DefenseSSHEnabled:     true,
		DefenseSSHHost:        "ssh.ctf.local",
		DefenseSSHPort:        2222,
		DefenseSSHHostKeyPath: filepath.Join(t.TempDir(), "runtime", "awd-defense-ssh-host-key.pem"),
	}

	root, err := BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}

	defaultExecutor := &stubRuntimeNodeHostExecutor{}
	runtime := &ContainerRuntimeModule{
		runtime: &containerruntime.Module{
			InteractiveExecutor: defaultExecutor,
		},
	}

	gateway := BuildAWDDefenseSSHGateway(root, runtime)
	if gateway == nil {
		t.Fatal("expected gateway")
	}
	if gateway.executor != defaultExecutor {
		t.Fatalf("expected gateway executor to use default interactive executor, got %T", gateway.executor)
	}
}

func TestAWDDefenseSSHGatewayStopCancelsActiveInteractiveSession(t *testing.T) {
	t.Parallel()

	executor := newBlockingInteractiveExecutor()
	runCtx, runCancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	close(done)

	gateway := &AWDDefenseSSHGateway{
		executor:    executor,
		logger:      zap.NewNop(),
		listener:    stubGatewayListener{},
		done:        done,
		runCancel:   runCancel,
		activeConns: map[net.Conn]struct{}{},
	}

	session := &instanceports.AWDDefenseSSHSession{
		InstanceID:        9003,
		WorkspaceRevision: 5,
		ContainerID:       "workspace-ctr",
	}

	gateway.workerWG.Add(1)
	go func() {
		defer gateway.workerWG.Done()
		gateway.runContainerCommand(runCtx, stubSSHChannel{}, session, []string{"/bin/sh"})
	}()

	select {
	case <-executor.started:
	case <-time.After(time.Second):
		t.Fatal("expected interactive executor to start")
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := gateway.Stop(stopCtx); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	select {
	case <-executor.stopped:
	case <-time.After(time.Second):
		t.Fatal("expected Stop() to cancel active interactive session")
	}
}
