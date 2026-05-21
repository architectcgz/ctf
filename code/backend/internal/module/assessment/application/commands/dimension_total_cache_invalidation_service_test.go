package commands_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	assessmentcmd "ctf-platform/internal/module/assessment/application/commands"
	assessmentinfra "ctf-platform/internal/module/assessment/infrastructure"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	platformevents "ctf-platform/internal/platform/events"
)

func TestDimensionTotalCacheInvalidationServiceRemovesCachedTotalsOnChallengeEvent(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer redisClient.Close()

	store := assessmentinfra.NewDimensionTotalCacheStore(redisClient)
	if err := store.StorePublishedDimensionTotals(context.Background(), map[string]int{"web": 300}, time.Minute); err != nil {
		t.Fatalf("StorePublishedDimensionTotals() error = %v", err)
	}

	service := assessmentcmd.NewDimensionTotalCacheInvalidationService(store, zap.NewNop())
	bus := platformevents.NewBus()
	service.RegisterChallengeEventConsumers(bus)

	err := bus.Publish(context.Background(), platformevents.Event{
		Name: challengecontracts.EventPublishedCatalogChanged,
		Payload: challengecontracts.PublishedCatalogChangedEvent{
			ChallengeID:     11,
			ChangeType:      challengecontracts.ChallengeCatalogChangeTypePublished,
			CurrentStatus:   challengecontracts.ChallengeStatusPublished,
			CurrentCategory: "web",
			CurrentPoints:   300,
		},
	})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	_, found, err := store.LoadPublishedDimensionTotals(context.Background())
	if err != nil {
		t.Fatalf("LoadPublishedDimensionTotals() error = %v", err)
	}
	if found {
		t.Fatal("expected published dimension totals cache to be invalidated")
	}
}

func TestDimensionTotalCacheInvalidationServiceRejectsUnexpectedPayload(t *testing.T) {
	service := assessmentcmd.NewDimensionTotalCacheInvalidationService(nil, zap.NewNop())
	bus := platformevents.NewBus()
	service.RegisterChallengeEventConsumers(bus)

	err := bus.Publish(context.Background(), platformevents.Event{
		Name:    challengecontracts.EventPublishedCatalogChanged,
		Payload: "unexpected",
	})
	if err == nil {
		t.Fatal("expected payload type error")
	}
}
