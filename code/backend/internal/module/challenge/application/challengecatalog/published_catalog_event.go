package challengecatalog

import (
	"context"
	"strings"

	"go.uber.org/zap"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeports "ctf-platform/internal/module/challenge/ports"
	platformevents "ctf-platform/internal/platform/events"
)

type PublishedState struct {
	ID       int64
	Status   string
	Category string
	Points   int
}

func PublishedStateFromWriteModel(challenge *challengeports.ChallengeWriteModel) PublishedState {
	if challenge == nil {
		return PublishedState{}
	}
	return PublishedState{
		ID:       challenge.ID,
		Status:   strings.TrimSpace(challenge.Status),
		Category: strings.TrimSpace(challenge.Category),
		Points:   challenge.Points,
	}
}

func PublishedStateFromImportedChallenge(challenge *challengeports.ImportedChallenge) PublishedState {
	if challenge == nil {
		return PublishedState{}
	}
	return PublishedState{
		ID:       challenge.ID,
		Status:   strings.TrimSpace(challenge.Status),
		Category: strings.TrimSpace(challenge.Category),
		Points:   challenge.Points,
	}
}

func (state PublishedState) IsPublished() bool {
	return state.Status == challengecontracts.ChallengeStatusPublished
}

func PublishedStateChanged(before, after PublishedState) bool {
	switch {
	case before.IsPublished() != after.IsPublished():
		return true
	case before.IsPublished() && after.IsPublished():
		return before.Category != after.Category || before.Points != after.Points
	default:
		return false
	}
}

func PublishWeakEvent(ctx context.Context, logger *zap.Logger, bus platformevents.Bus, evt platformevents.Event) {
	if bus == nil {
		return
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	if err := bus.Publish(ctx, evt); err != nil {
		logger.Warn("publish_challenge_event_failed", zap.String("event", evt.Name), zap.Error(err))
	}
}

func PublishPublishedCatalogChangedEvent(
	ctx context.Context,
	logger *zap.Logger,
	bus platformevents.Bus,
	changeType string,
	before PublishedState,
	after PublishedState,
) {
	if !PublishedStateChanged(before, after) {
		return
	}
	challengeID := after.ID
	if challengeID <= 0 {
		challengeID = before.ID
	}
	PublishWeakEvent(ctx, logger, bus, platformevents.Event{
		Name: challengecontracts.EventPublishedCatalogChanged,
		Payload: challengecontracts.PublishedCatalogChangedEvent{
			ChallengeID:      challengeID,
			ChangeType:       changeType,
			PreviousStatus:   before.Status,
			CurrentStatus:    after.Status,
			PreviousCategory: before.Category,
			CurrentCategory:  after.Category,
			PreviousPoints:   before.Points,
			CurrentPoints:    after.Points,
		},
	})
}
