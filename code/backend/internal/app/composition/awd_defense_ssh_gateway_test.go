package composition

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/model"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/pkg/errcode"
	"golang.org/x/crypto/ssh"
)

type stubAWDDefenseSSHGatewayProxyTickets struct {
	claims *runtimeports.ProxyTicketClaims
	err    error
}

type stubRuntimeHTTPProxyTicketReader struct {
	scope *runtimeports.AWDDefenseSSHScope
}

func (s stubRuntimeHTTPProxyTicketReader) FindByID(context.Context, int64) (*model.Instance, error) {
	return nil, nil
}

func (s stubRuntimeHTTPProxyTicketReader) FindAWDTargetProxyScope(context.Context, int64, int64, int64, int64) (*runtimeports.AWDTargetProxyScope, error) {
	return nil, nil
}

func (s stubRuntimeHTTPProxyTicketReader) FindAWDDefenseSSHScope(context.Context, int64, int64, int64) (*runtimeports.AWDDefenseSSHScope, error) {
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

func (s stubAWDDefenseSSHGatewayProxyTickets) ResolveTicket(context.Context, string) (*runtimeports.ProxyTicketClaims, error) {
	return s.claims, s.err
}

func (s stubAWDDefenseSSHGatewayProxyTickets) ResolveAWDTargetAccessURL(context.Context, *runtimeports.ProxyTicketClaims, int64, int64, int64) (string, error) {
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
			claims: &runtimeports.ProxyTicketClaims{
				UserID:               1001,
				Username:             "student",
				Role:                 model.RoleStudent,
				InstanceID:           9001,
				ContestID:            &contestID,
				ShareScope:           model.InstanceSharingPerTeam,
				Purpose:              runtimeports.ProxyTicketPurposeAWDDefenseSSH,
				AWDAttackerTeamID:    &teamID,
				AWDServiceID:         &serviceID,
				AWDChallengeID:       &challengeID,
				AWDWorkspaceRevision: &workspaceRevision,
			},
		},
		stubRuntimeHTTPProxyTicketReader{
			scope: &runtimeports.AWDDefenseSSHScope{
				InstanceID:        9001,
				ContestID:         contestID,
				TeamID:            teamID,
				ServiceID:         serviceID,
				AWDChallengeID:    challengeID,
				WorkspaceRevision: workspaceRevision,
				ContainerID:       "workspace-ctr",
				ShareScope:        model.InstanceSharingPerTeam,
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
			claims: &runtimeports.ProxyTicketClaims{
				UserID:               1002,
				Username:             "student",
				Role:                 model.RoleStudent,
				InstanceID:           9002,
				ContestID:            &contestID,
				ShareScope:           model.InstanceSharingPerTeam,
				Purpose:              runtimeports.ProxyTicketPurposeAWDDefenseSSH,
				AWDAttackerTeamID:    &teamID,
				AWDServiceID:         &serviceID,
				AWDChallengeID:       &challengeID,
				AWDWorkspaceRevision: &claimedRevision,
			},
		},
		stubRuntimeHTTPProxyTicketReader{
			scope: &runtimeports.AWDDefenseSSHScope{
				InstanceID:        9002,
				ContestID:         contestID,
				TeamID:            teamID,
				ServiceID:         serviceID,
				AWDChallengeID:    challengeID,
				WorkspaceRevision: currentRevision,
				ContainerID:       "workspace-ctr",
				ShareScope:        model.InstanceSharingPerTeam,
			},
		},
		nil,
		"",
		2222,
		nil,
	)

	_, err := gateway.authenticate(context.Background(), "student+52+72", "ticket-secret")
	if err == nil || err.Error() != errcode.ErrForbidden.Error() {
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
