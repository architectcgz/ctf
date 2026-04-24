package commands

import (
	"context"
	"strings"
	"testing"

	"ctf-platform/internal/model"
)

type flagCommandContextRepoStub struct {
	findByIDWithContextFn func(ctx context.Context, id int64) (*model.Challenge, error)
	updateWithContextFn   func(ctx context.Context, challenge *model.Challenge) error
}

func (s *flagCommandContextRepoStub) FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *flagCommandContextRepoStub) UpdateWithContext(ctx context.Context, challenge *model.Challenge) error {
	if s.updateWithContextFn != nil {
		return s.updateWithContextFn(ctx, challenge)
	}
	return nil
}

type flagCommandContextKey string

func TestFlagServiceConfigureWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		call            func(service *FlagService, ctx context.Context) error
		assertChallenge func(t *testing.T, challenge *model.Challenge)
	}{
		{
			name: "static",
			call: func(service *FlagService, ctx context.Context) error {
				return service.ConfigureStaticFlag(ctx, 1, "flag{demo_static}", "flag")
			},
			assertChallenge: func(t *testing.T, challenge *model.Challenge) {
				t.Helper()
				if challenge.FlagType != model.FlagTypeStatic || challenge.FlagPrefix != "flag" || challenge.FlagHash == "" || challenge.FlagSalt == "" {
					t.Fatalf("unexpected static challenge payload: %+v", challenge)
				}
			},
		},
		{
			name: "dynamic",
			call: func(service *FlagService, ctx context.Context) error {
				return service.ConfigureDynamicFlag(ctx, 1, "ctf")
			},
			assertChallenge: func(t *testing.T, challenge *model.Challenge) {
				t.Helper()
				if challenge.FlagType != model.FlagTypeDynamic || challenge.FlagPrefix != "ctf" {
					t.Fatalf("unexpected dynamic challenge payload: %+v", challenge)
				}
			},
		},
		{
			name: "regex",
			call: func(service *FlagService, ctx context.Context) error {
				return service.ConfigureRegexFlag(ctx, 1, `^flag\{user-[0-9]{3}\}$`, "flag")
			},
			assertChallenge: func(t *testing.T, challenge *model.Challenge) {
				t.Helper()
				if challenge.FlagType != model.FlagTypeRegex || challenge.FlagRegex != `^flag\{user-[0-9]{3}\}$` || challenge.FlagPrefix != "flag" {
					t.Fatalf("unexpected regex challenge payload: %+v", challenge)
				}
			},
		},
		{
			name: "manual-review",
			call: func(service *FlagService, ctx context.Context) error {
				return service.ConfigureManualReviewFlag(ctx, 1)
			},
			assertChallenge: func(t *testing.T, challenge *model.Challenge) {
				t.Helper()
				if challenge.FlagType != model.FlagTypeManualReview {
					t.Fatalf("unexpected manual review challenge payload: %+v", challenge)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctxKey := flagCommandContextKey(tt.name)
			expectedCtxValue := "ctx-" + tt.name
			findCalled := false
			updateCalled := false
			repo := &flagCommandContextRepoStub{
				findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
					findCalled = true
					if got := ctx.Value(ctxKey); got != expectedCtxValue {
						t.Fatalf("expected find-challenge ctx value %v, got %v", expectedCtxValue, got)
					}
					return &model.Challenge{ID: id, Title: tt.name, FlagPrefix: "legacy", FlagHash: "legacy-hash", FlagSalt: "legacy-salt", FlagRegex: "legacy-regex"}, nil
				},
				updateWithContextFn: func(ctx context.Context, challenge *model.Challenge) error {
					updateCalled = true
					if got := ctx.Value(ctxKey); got != expectedCtxValue {
						t.Fatalf("expected update-challenge ctx value %v, got %v", expectedCtxValue, got)
					}
					tt.assertChallenge(t, challenge)
					return nil
				},
			}
			service, err := NewFlagService(repo, strings.Repeat("s", 32))
			if err != nil {
				t.Fatalf("NewFlagService() error = %v", err)
			}

			ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
			if err := tt.call(service, ctx); err != nil {
				t.Fatalf("configure flag with context error = %v", err)
			}
			if !findCalled || !updateCalled {
				t.Fatalf("expected repository calls, got find=%v update=%v", findCalled, updateCalled)
			}
		})
	}
}
