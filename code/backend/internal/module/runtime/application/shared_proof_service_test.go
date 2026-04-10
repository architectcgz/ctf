package application_test

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/model"
	runtimeqry "ctf-platform/internal/module/runtime/application/queries"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	flagcrypto "ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
)

type stubSharedProofRepository struct {
	findInstanceByIDFn  func(id int64) (*model.Instance, error)
	findChallengeByIDFn func(challengeID int64) (*model.Challenge, error)
	createSharedProofFn func(proof *model.SharedProof) error
}

func (s *stubSharedProofRepository) FindByID(id int64) (*model.Instance, error) {
	if s.findInstanceByIDFn == nil {
		return nil, nil
	}
	return s.findInstanceByIDFn(id)
}

func (s *stubSharedProofRepository) FindChallengeByID(challengeID int64) (*model.Challenge, error) {
	if s.findChallengeByIDFn == nil {
		return nil, nil
	}
	return s.findChallengeByIDFn(challengeID)
}

func (s *stubSharedProofRepository) CreateSharedProof(proof *model.SharedProof) error {
	if s.createSharedProofFn == nil {
		return nil
	}
	return s.createSharedProofFn(proof)
}

type stubSharedProofTicketResolver struct {
	resolveTicketFn func(ctx context.Context, ticket string) (*runtimeports.ProxyTicketClaims, error)
}

func (s *stubSharedProofTicketResolver) ResolveTicket(ctx context.Context, ticket string) (*runtimeports.ProxyTicketClaims, error) {
	if s.resolveTicketFn == nil {
		return nil, nil
	}
	return s.resolveTicketFn(ctx, ticket)
}

func TestSharedProofServiceIssueSharedProofPersistsBoundProof(t *testing.T) {
	t.Parallel()

	var createdProof *model.SharedProof
	service := runtimeqry.NewSharedProofService(
		&stubSharedProofRepository{
			findInstanceByIDFn: func(id int64) (*model.Instance, error) {
				contestID := int64(3001)
				return &model.Instance{
					ID:          id,
					ContestID:   &contestID,
					ChallengeID: 2001,
					ShareScope:  model.InstanceSharingShared,
					Status:      model.InstanceStatusRunning,
				}, nil
			},
			findChallengeByIDFn: func(challengeID int64) (*model.Challenge, error) {
				return &model.Challenge{
					ID:              challengeID,
					FlagType:        model.FlagTypeSharedProof,
					InstanceSharing: model.InstanceSharingShared,
				}, nil
			},
			createSharedProofFn: func(proof *model.SharedProof) error {
				createdProof = proof
				return nil
			},
		},
		&stubSharedProofTicketResolver{
			resolveTicketFn: func(ctx context.Context, ticket string) (*runtimeports.ProxyTicketClaims, error) {
				contestID := int64(3001)
				return &runtimeports.ProxyTicketClaims{
					UserID:      1001,
					Username:    "alice",
					Role:        "student",
					InstanceID:  4001,
					ChallengeID: 2001,
					ContestID:   &contestID,
					ShareScope:  model.InstanceSharingShared,
					IssuedAt:    time.Now().UTC(),
				}, nil
			},
		},
		2*time.Minute,
	)

	resp, err := service.IssueSharedProof(context.Background(), "ticket-1")
	if err != nil {
		t.Fatalf("IssueSharedProof() error = %v", err)
	}
	if resp == nil || resp.Proof == "" || resp.ExpiresAt.IsZero() {
		t.Fatalf("unexpected shared proof response: %+v", resp)
	}
	if createdProof == nil {
		t.Fatal("expected shared proof to be persisted")
	}
	if createdProof.UserID != 1001 || createdProof.ChallengeID != 2001 || createdProof.InstanceID != 4001 {
		t.Fatalf("unexpected persisted proof scope: %+v", createdProof)
	}
	if createdProof.ContestID == nil || *createdProof.ContestID != 3001 {
		t.Fatalf("unexpected persisted contest scope: %+v", createdProof)
	}
	if createdProof.ProofHash != flagcrypto.HashSharedProof(resp.Proof) {
		t.Fatalf("expected persisted proof hash to match issued proof, got %+v", createdProof)
	}
}

func TestSharedProofServiceIssueSharedProofRejectsNonSharedInstanceScope(t *testing.T) {
	t.Parallel()

	service := runtimeqry.NewSharedProofService(
		&stubSharedProofRepository{
			findInstanceByIDFn: func(id int64) (*model.Instance, error) {
				return &model.Instance{
					ID:          id,
					ChallengeID: 2001,
					ShareScope:  model.InstanceSharingPerUser,
					Status:      model.InstanceStatusRunning,
				}, nil
			},
			findChallengeByIDFn: func(challengeID int64) (*model.Challenge, error) {
				return &model.Challenge{
					ID:              challengeID,
					FlagType:        model.FlagTypeSharedProof,
					InstanceSharing: model.InstanceSharingShared,
				}, nil
			},
		},
		&stubSharedProofTicketResolver{
			resolveTicketFn: func(ctx context.Context, ticket string) (*runtimeports.ProxyTicketClaims, error) {
				return &runtimeports.ProxyTicketClaims{
					UserID:      1001,
					Username:    "alice",
					Role:        "student",
					InstanceID:  4001,
					ChallengeID: 2001,
					ShareScope:  model.InstanceSharingPerUser,
					IssuedAt:    time.Now().UTC(),
				}, nil
			},
		},
		2*time.Minute,
	)

	_, err := service.IssueSharedProof(context.Background(), "ticket-1")
	if err == nil || err.Error() != errcode.ErrInvalidParams.Error() {
		t.Fatalf("expected invalid params for non-shared instance scope, got %v", err)
	}
}
