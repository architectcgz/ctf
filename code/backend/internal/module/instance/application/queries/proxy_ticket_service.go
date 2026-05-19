package queries

import (
	"context"
	"crypto/rand"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/authctx"
	instanceentity "ctf-platform/internal/module/instance/entity"
	instanceports "ctf-platform/internal/module/instance/ports"
)

type ProxyTicketService struct {
	instanceReader instanceports.ProxyTicketInstanceReader
	store          instanceports.ProxyTicketStore
	ticketTTL      time.Duration
}

func NewProxyTicketService(store instanceports.ProxyTicketStore, instanceReader instanceports.ProxyTicketInstanceReader, ticketTTL time.Duration) *ProxyTicketService {
	return &ProxyTicketService{
		instanceReader: instanceReader,
		store:          store,
		ticketTTL:      ticketTTL,
	}
}

func (s *ProxyTicketService) IssueTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, time.Time, error) {
	if s == nil || s.store == nil || s.instanceReader == nil || s.ticketTTL <= 0 {
		return "", time.Time{}, errProxyTicketServiceUnavailable()
	}

	ticket, err := generateProxyToken(32)
	if err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}

	instance, err := s.instanceReader.FindByID(ctx, instanceID)
	if err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}
	if instance == nil {
		return "", time.Time{}, apperror.ErrNotFound
	}

	claims := instanceports.ProxyTicketClaims{
		UserID:     user.UserID,
		Username:   user.Username,
		Role:       user.Role,
		InstanceID: instanceID,
		ContestID:  instance.ContestID,
		ShareScope: instance.ShareScope,
		Purpose:    instanceports.ProxyTicketPurposeInstanceAccess,
		IssuedAt:   time.Now().UTC(),
	}
	expiresAt := time.Now().Add(s.ticketTTL).UTC()

	if err := s.store.SaveProxyTicket(ctx, ticket, claims, s.ticketTTL); err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}

	return ticket, expiresAt, nil
}

func (s *ProxyTicketService) IssueAWDTargetTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID, victimTeamID int64) (string, time.Time, error) {
	if s == nil || s.store == nil || s.instanceReader == nil || s.ticketTTL <= 0 {
		return "", time.Time{}, errProxyTicketServiceUnavailable()
	}

	scope, err := s.instanceReader.FindAWDTargetProxyScope(ctx, user.UserID, contestID, serviceID, victimTeamID)
	if err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}
	if scope == nil {
		return "", time.Time{}, apperror.ErrForbidden
	}
	if scope.AttackerTeamID == scope.VictimTeamID {
		return "", time.Time{}, apperror.ErrForbidden
	}

	ticket, err := generateProxyToken(32)
	if err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}

	claims := instanceports.ProxyTicketClaims{
		UserID:            user.UserID,
		Username:          user.Username,
		Role:              user.Role,
		InstanceID:        scope.InstanceID,
		ContestID:         &scope.ContestID,
		ShareScope:        scope.ShareScope,
		Purpose:           instanceports.ProxyTicketPurposeAWDAttack,
		AWDAttackerTeamID: &scope.AttackerTeamID,
		AWDVictimTeamID:   &scope.VictimTeamID,
		AWDServiceID:      &scope.ServiceID,
		AWDChallengeID:    &scope.AWDChallengeID,
		IssuedAt:          time.Now().UTC(),
	}
	expiresAt := time.Now().Add(s.ticketTTL).UTC()

	if err := s.store.SaveProxyTicket(ctx, ticket, claims, s.ticketTTL); err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}

	return ticket, expiresAt, nil
}

func (s *ProxyTicketService) IssueAWDDefenseSSHTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64) (string, time.Time, error) {
	if s == nil || s.store == nil || s.instanceReader == nil || s.ticketTTL <= 0 {
		return "", time.Time{}, errProxyTicketServiceUnavailable()
	}

	scope, err := s.instanceReader.FindAWDDefenseSSHScope(ctx, user.UserID, contestID, serviceID)
	if err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}
	if scope == nil || scope.ContainerID == "" || scope.WorkspaceRevision <= 0 {
		return "", time.Time{}, apperror.ErrForbidden
	}

	ticket, err := generateProxyToken(32)
	if err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}

	claims := instanceports.ProxyTicketClaims{
		UserID:               user.UserID,
		Username:             user.Username,
		Role:                 user.Role,
		InstanceID:           scope.InstanceID,
		ContestID:            &scope.ContestID,
		ShareScope:           scope.ShareScope,
		Purpose:              instanceports.ProxyTicketPurposeAWDDefenseSSH,
		AWDAttackerTeamID:    &scope.TeamID,
		AWDServiceID:         &scope.ServiceID,
		AWDChallengeID:       &scope.AWDChallengeID,
		AWDWorkspaceRevision: &scope.WorkspaceRevision,
		IssuedAt:             time.Now().UTC(),
	}
	expiresAt := time.Now().Add(s.ticketTTL).UTC()

	if err := s.store.SaveProxyTicket(ctx, ticket, claims, s.ticketTTL); err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}

	return ticket, expiresAt, nil
}

func (s *ProxyTicketService) ResolveAWDTargetAccessURL(ctx context.Context, claims *instanceports.ProxyTicketClaims, contestID, serviceID, victimTeamID int64) (string, error) {
	if s == nil || s.instanceReader == nil {
		return "", errProxyTicketServiceUnavailable()
	}
	if claims == nil || claims.Purpose != instanceports.ProxyTicketPurposeAWDAttack {
		return "", instancecontracts.ErrProxyTicketInvalid
	}
	if claims.ContestID == nil || *claims.ContestID != contestID ||
		claims.AWDServiceID == nil || *claims.AWDServiceID != serviceID ||
		claims.AWDVictimTeamID == nil || *claims.AWDVictimTeamID != victimTeamID {
		return "", apperror.ErrForbidden
	}

	scope, err := s.instanceReader.FindAWDTargetProxyScope(ctx, claims.UserID, contestID, serviceID, victimTeamID)
	if err != nil {
		return "", apperror.ErrInternal.WithCause(err)
	}
	if scope == nil || scope.InstanceID != claims.InstanceID || scope.AttackerTeamID == scope.VictimTeamID {
		return "", apperror.ErrForbidden
	}
	if scope.Status != instanceentity.InstanceStatusRunning || strings.TrimSpace(scope.AccessURL) == "" {
		return "", apperror.ErrServiceUnavailable.WithCause(fmt.Errorf("awd target instance %d status=%s", scope.InstanceID, scope.Status))
	}
	return scope.AccessURL, nil
}

func (s *ProxyTicketService) ResolveTicket(ctx context.Context, ticket string) (*instanceports.ProxyTicketClaims, error) {
	if s == nil || s.store == nil {
		return nil, errProxyTicketServiceUnavailable()
	}
	if ticket == "" {
		return nil, instancecontracts.ErrProxyTicketInvalid
	}

	claims, err := s.store.FindProxyTicket(ctx, ticket)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if claims == nil {
		return nil, instancecontracts.ErrProxyTicketInvalid
	}
	if claims.UserID <= 0 || claims.InstanceID <= 0 || claims.Username == "" || claims.Role == "" || claims.ShareScope == "" {
		return nil, instancecontracts.ErrProxyTicketInvalid
	}
	if claims.Purpose == instanceports.ProxyTicketPurposeAWDAttack && (claims.ContestID == nil || claims.AWDAttackerTeamID == nil || claims.AWDVictimTeamID == nil || claims.AWDServiceID == nil || claims.AWDChallengeID == nil) {
		return nil, instancecontracts.ErrProxyTicketInvalid
	}
	if claims.Purpose == instanceports.ProxyTicketPurposeAWDDefenseSSH && (claims.ContestID == nil || claims.AWDAttackerTeamID == nil || claims.AWDServiceID == nil || claims.AWDChallengeID == nil || claims.AWDWorkspaceRevision == nil || *claims.AWDWorkspaceRevision <= 0) {
		return nil, instancecontracts.ErrProxyTicketInvalid
	}

	return claims, nil
}

func (s *ProxyTicketService) MaxAge() int {
	if s == nil || s.ticketTTL <= 0 {
		return 0
	}
	return int(s.ticketTTL.Seconds())
}

func errProxyTicketServiceUnavailable() error {
	return apperror.ErrInternal.WithCause(fmt.Errorf("proxy ticket service is not configured"))
}

func generateProxyToken(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
