package commands

import (
	"context"
	"errors"
	"testing"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type awdPreviewRuntimeChallengeLookupStub struct {
	findByIDFn func(context.Context, int64) (*challengecontracts.AWDChallenge, error)
}

func (s awdPreviewRuntimeChallengeLookupStub) FindAWDChallengeByID(ctx context.Context, id int64) (*challengecontracts.AWDChallenge, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return &challengecontracts.AWDChallenge{ID: id}, nil
}

func (s awdPreviewRuntimeChallengeLookupStub) ListAWDChallenges(context.Context, *challengecontracts.AWDChallengeQuery) ([]*challengecontracts.AWDChallenge, int64, error) {
	return nil, 0, errors.New("unexpected ListAWDChallenges call")
}

type awdPreviewRuntimeImageStoreStub struct {
	findByIDFn func(context.Context, int64) (*challengeentity.Image, error)
}

func (s awdPreviewRuntimeImageStoreStub) FindByID(ctx context.Context, id int64) (*challengeentity.Image, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return &challengeentity.Image{ID: id}, nil
}

var _ challengeports.AWDChallengeQueryRepository = (*awdPreviewRuntimeChallengeLookupStub)(nil)
var _ challengecontracts.ImageStore = (*awdPreviewRuntimeImageStoreStub)(nil)

func TestAWDServiceLoadPreviewRuntimeDefinitionTreatsPreviewChallengeSentinelAsNotFound(t *testing.T) {
	t.Parallel()

	service := &AWDService{
		awdChallengeRepo: awdPreviewRuntimeChallengeLookupStub{
			findByIDFn: func(context.Context, int64) (*challengecontracts.AWDChallenge, error) {
				return nil, contestports.ErrContestAWDPreviewChallengeNotFound
			},
		},
	}

	_, _, err := service.loadPreviewRuntimeDefinition(context.Background(), nil, 3201)
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrNotFound.Code {
		t.Fatalf("expected errcode.ErrNotFound, got %v", err)
	}
}

func TestAWDServicePrepareCheckerPreviewAccessURLAllowsExplicitURLWhenPreviewChallengeNotFound(t *testing.T) {
	t.Parallel()

	service := &AWDService{
		awdChallengeRepo: awdPreviewRuntimeChallengeLookupStub{
			findByIDFn: func(context.Context, int64) (*challengecontracts.AWDChallenge, error) {
				return nil, contestports.ErrContestAWDPreviewChallengeNotFound
			},
		},
		flagSecret: "preview-secret",
	}

	accessURL, checkerTokenEnv, checkerToken, cleanup, err := service.prepareCheckerPreviewAccessURL(
		context.Background(),
		32,
		nil,
		3201,
		" http://preview.internal ",
		"flag{preview}",
	)
	if err != nil {
		t.Fatalf("prepareCheckerPreviewAccessURL() error = %v", err)
	}
	if accessURL != "http://preview.internal" {
		t.Fatalf("unexpected access url: %s", accessURL)
	}
	if checkerTokenEnv != "" || checkerToken != "" {
		t.Fatalf("expected empty checker token context, got env=%q token=%q", checkerTokenEnv, checkerToken)
	}
	if cleanup != nil {
		t.Fatal("expected no cleanup callback for explicit url fallback")
	}
}

func TestAWDServicePrepareCheckerPreviewAccessURLRejectsExplicitURLWhenPreviewImageNotFound(t *testing.T) {
	t.Parallel()

	service := &AWDService{
		imageRepo: awdPreviewRuntimeImageStoreStub{
			findByIDFn: func(context.Context, int64) (*challengeentity.Image, error) {
				return nil, contestports.ErrContestAWDPreviewImageNotFound
			},
		},
		awdChallengeRepo: awdPreviewRuntimeChallengeLookupStub{
			findByIDFn: func(context.Context, int64) (*challengecontracts.AWDChallenge, error) {
				return &challengecontracts.AWDChallenge{
					ID:             3202,
					DeploymentMode: challengecontracts.AWDDeploymentModeSingleContainer,
					RuntimeConfig:  `{"image_id":9901}`,
				}, nil
			},
		},
	}

	_, _, _, _, err := service.prepareCheckerPreviewAccessURL(
		context.Background(),
		32,
		nil,
		3202,
		"http://preview.internal",
		"flag{preview}",
	)
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrNotFound.Code {
		t.Fatalf("expected errcode.ErrNotFound, got %v", err)
	}
}
