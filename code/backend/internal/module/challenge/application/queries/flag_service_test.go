package queries

import (
	"context"
	"testing"

	"ctf-platform/internal/model"
)

type stubChallengeFlagRepository struct {
	findByIDWithContextFn func(ctx context.Context, id int64) (*model.Challenge, error)
	updateWithContextFn   func(ctx context.Context, challenge *model.Challenge) error
}

func (s *stubChallengeFlagRepository) FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *stubChallengeFlagRepository) UpdateWithContext(ctx context.Context, challenge *model.Challenge) error {
	if s.updateWithContextFn != nil {
		return s.updateWithContextFn(ctx, challenge)
	}
	return nil
}

type challengeFlagContextKey string

func TestFlagServiceGetFlagConfigPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := challengeFlagContextKey("flag-config")
	expectedCtxValue := "ctx-flag-config"
	findCalled := false
	repo := &stubChallengeFlagRepository{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-by-id ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{
				ID:         id,
				FlagType:   model.FlagTypeDynamic,
				FlagPrefix: "flag",
			}, nil
		},
	}
	service, err := NewFlagService(repo, "12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("NewFlagService() error = %v", err)
	}

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.GetFlagConfig(ctx, 42)
	if err != nil {
		t.Fatalf("GetFlagConfig() error = %v", err)
	}
	if !findCalled {
		t.Fatal("expected repository find to be called")
	}
	if resp == nil || resp.FlagType != model.FlagTypeDynamic || resp.FlagPrefix != "flag" || !resp.Configured {
		t.Fatalf("unexpected flag config response: %+v", resp)
	}
}
