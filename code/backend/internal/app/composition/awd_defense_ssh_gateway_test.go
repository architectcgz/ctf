package composition

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/model"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/pkg/errcode"
)

type stubAWDDefenseSSHGatewayProxyTickets struct {
	claims *runtimeports.ProxyTicketClaims
	err    error
}

func (s stubAWDDefenseSSHGatewayProxyTickets) IssueAWDDefenseSSHTicket(context.Context, authctx.CurrentUser, int64, int64) (string, time.Time, error) {
	return "", time.Time{}, nil
}

func (s stubAWDDefenseSSHGatewayProxyTickets) ResolveTicket(context.Context, string) (*runtimeports.ProxyTicketClaims, error) {
	return s.claims, s.err
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
		2222,
		nil,
	)

	_, err := gateway.authenticate(context.Background(), "student+52+72", "ticket-secret")
	if err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected forbidden error for stale workspace revision, got %v", err)
	}
}
