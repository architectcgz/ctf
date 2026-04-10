package queries

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
)

type sharedProofTicketResolver interface {
	ResolveTicket(ctx context.Context, ticket string) (*runtimeports.ProxyTicketClaims, error)
}

type SharedProofService struct {
	repo           runtimeports.SharedProofRepository
	ticketResolver sharedProofTicketResolver
	proofTTL       time.Duration
}

func NewSharedProofService(
	repo runtimeports.SharedProofRepository,
	ticketResolver sharedProofTicketResolver,
	proofTTL time.Duration,
) *SharedProofService {
	return &SharedProofService{
		repo:           repo,
		ticketResolver: ticketResolver,
		proofTTL:       proofTTL,
	}
}

func (s *SharedProofService) IssueSharedProof(ctx context.Context, proxyTicket string) (*dto.SharedProofIssueResp, error) {
	if s == nil || s.repo == nil || s.ticketResolver == nil || s.proofTTL <= 0 {
		return nil, errcode.ErrInternal.WithCause(errors.New("shared proof service is not configured"))
	}
	if proxyTicket == "" {
		return nil, errcode.ErrProxyTicketInvalid
	}

	claims, err := s.ticketResolver.ResolveTicket(ctx, proxyTicket)
	if err != nil {
		return nil, err
	}
	if claims == nil || claims.UserID <= 0 || claims.InstanceID <= 0 || claims.ChallengeID <= 0 {
		return nil, errcode.ErrProxyTicketInvalid
	}
	if claims.ShareScope != model.InstanceSharingShared {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("shared_proof 仅支持共享实例"))
	}

	instance, err := s.repo.FindByID(claims.InstanceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrProxyTicketInvalid
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if instance == nil || instance.ChallengeID != claims.ChallengeID || instance.ShareScope != model.InstanceSharingShared {
		return nil, errcode.ErrProxyTicketInvalid
	}
	if !isRuntimeAccessibleStatus(instance.Status) {
		return nil, errcode.ErrProxyTicketInvalid
	}
	if !sameInt64Ptr(instance.ContestID, claims.ContestID) {
		return nil, errcode.ErrProxyTicketInvalid
	}

	challenge, err := s.repo.FindChallengeByID(claims.ChallengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if challenge.FlagType != model.FlagTypeSharedProof || challenge.InstanceSharing != model.InstanceSharingShared {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("当前题目未启用 shared_proof"))
	}

	proof, err := generateProxyToken(24)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	now := time.Now().UTC()
	expiresAt := now.Add(s.proofTTL)
	if err := s.repo.CreateSharedProof(&model.SharedProof{
		UserID:      claims.UserID,
		ChallengeID: claims.ChallengeID,
		ContestID:   claims.ContestID,
		InstanceID:  claims.InstanceID,
		ProofHash:   crypto.HashSharedProof(proof),
		Status:      model.SharedProofStatusActive,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &dto.SharedProofIssueResp{
		Proof:     proof,
		ExpiresAt: expiresAt,
	}, nil
}

func isRuntimeAccessibleStatus(status string) bool {
	switch status {
	case model.InstanceStatusPending, model.InstanceStatusCreating, model.InstanceStatusRunning:
		return true
	default:
		return false
	}
}

func sameInt64Ptr(left, right *int64) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
