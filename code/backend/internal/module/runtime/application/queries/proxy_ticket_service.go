package queries

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"ctf-platform/internal/authctx"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/pkg/errcode"
)

type ProxyTicketService struct {
	instanceReader runtimeports.ProxyTicketInstanceReader
	store          runtimeports.ProxyTicketStore
	ticketTTL      time.Duration
}

func NewProxyTicketService(store runtimeports.ProxyTicketStore, instanceReader runtimeports.ProxyTicketInstanceReader, ticketTTL time.Duration) *ProxyTicketService {
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
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}

	instance, err := s.instanceReader.FindByID(ctx, instanceID)
	if err != nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}
	if instance == nil {
		return "", time.Time{}, errcode.ErrNotFound
	}

	claims := runtimeports.ProxyTicketClaims{
		UserID:     user.UserID,
		Username:   user.Username,
		Role:       user.Role,
		InstanceID: instanceID,
		ContestID:  instance.ContestID,
		ShareScope: instance.ShareScope,
		Purpose:    runtimeports.ProxyTicketPurposeInstanceAccess,
		IssuedAt:   time.Now().UTC(),
	}
	expiresAt := time.Now().Add(s.ticketTTL).UTC()

	if err := s.store.SaveProxyTicket(ctx, ticket, claims, s.ticketTTL); err != nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}

	return ticket, expiresAt, nil
}

func (s *ProxyTicketService) IssueAWDTargetTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID, victimTeamID int64) (string, time.Time, error) {
	if s == nil || s.store == nil || s.instanceReader == nil || s.ticketTTL <= 0 {
		return "", time.Time{}, errProxyTicketServiceUnavailable()
	}

	scope, err := s.instanceReader.FindAWDTargetProxyScope(ctx, user.UserID, contestID, serviceID, victimTeamID)
	if err != nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}
	if scope == nil {
		return "", time.Time{}, errcode.ErrForbidden
	}
	if scope.AttackerTeamID == scope.VictimTeamID {
		return "", time.Time{}, errcode.ErrForbidden
	}

	ticket, err := generateProxyToken(32)
	if err != nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}

	claims := runtimeports.ProxyTicketClaims{
		UserID:            user.UserID,
		Username:          user.Username,
		Role:              user.Role,
		InstanceID:        scope.InstanceID,
		ContestID:         &scope.ContestID,
		ShareScope:        scope.ShareScope,
		Purpose:           runtimeports.ProxyTicketPurposeAWDAttack,
		AWDAttackerTeamID: &scope.AttackerTeamID,
		AWDVictimTeamID:   &scope.VictimTeamID,
		AWDServiceID:      &scope.ServiceID,
		AWDChallengeID:    &scope.ChallengeID,
		IssuedAt:          time.Now().UTC(),
	}
	expiresAt := time.Now().Add(s.ticketTTL).UTC()

	if err := s.store.SaveProxyTicket(ctx, ticket, claims, s.ticketTTL); err != nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}

	return ticket, expiresAt, nil
}

func (s *ProxyTicketService) ResolveAWDTargetAccessURL(ctx context.Context, claims *runtimeports.ProxyTicketClaims, contestID, serviceID, victimTeamID int64) (string, error) {
	if s == nil || s.instanceReader == nil {
		return "", errProxyTicketServiceUnavailable()
	}
	if claims == nil || claims.Purpose != runtimeports.ProxyTicketPurposeAWDAttack {
		return "", errcode.ErrProxyTicketInvalid
	}
	if claims.ContestID == nil || *claims.ContestID != contestID ||
		claims.AWDServiceID == nil || *claims.AWDServiceID != serviceID ||
		claims.AWDVictimTeamID == nil || *claims.AWDVictimTeamID != victimTeamID {
		return "", errcode.ErrForbidden
	}

	scope, err := s.instanceReader.FindAWDTargetProxyScope(ctx, claims.UserID, contestID, serviceID, victimTeamID)
	if err != nil {
		return "", errcode.ErrInternal.WithCause(err)
	}
	if scope == nil || scope.InstanceID != claims.InstanceID || scope.AttackerTeamID == scope.VictimTeamID {
		return "", errcode.ErrForbidden
	}
	return scope.AccessURL, nil
}

func (s *ProxyTicketService) ResolveTicket(ctx context.Context, ticket string) (*runtimeports.ProxyTicketClaims, error) {
	if s == nil || s.store == nil {
		return nil, errProxyTicketServiceUnavailable()
	}
	if ticket == "" {
		return nil, errcode.ErrProxyTicketInvalid
	}

	claims, err := s.store.FindProxyTicket(ctx, ticket)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if claims == nil {
		return nil, errcode.ErrProxyTicketInvalid
	}
	if claims.UserID <= 0 || claims.InstanceID <= 0 || claims.Username == "" || claims.Role == "" || claims.ShareScope == "" {
		return nil, errcode.ErrProxyTicketInvalid
	}
	if claims.Purpose == runtimeports.ProxyTicketPurposeAWDAttack && (claims.ContestID == nil || claims.AWDAttackerTeamID == nil || claims.AWDVictimTeamID == nil || claims.AWDServiceID == nil || claims.AWDChallengeID == nil) {
		return nil, errcode.ErrProxyTicketInvalid
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
	return errcode.ErrInternal.WithCause(fmt.Errorf("proxy ticket service is not configured"))
}

func generateProxyToken(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
