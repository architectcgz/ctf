package commands

import (
	"context"
	"fmt"

	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
	platformevents "ctf-platform/internal/platform/events"
)

func enqueuePracticeFlagAcceptedOutboxEvent(
	ctx context.Context,
	repo practiceports.PracticeSubmissionOutboxTxRepository,
	payload practicecontracts.FlagAcceptedEvent,
) error {
	codec := platformevents.NewOutboxCodec()
	event, err := codec.Encode(
		practicecontracts.EventFlagAccepted,
		practicecontracts.EventFlagAcceptedPayloadVersion,
		payload,
		payload.OccurredAt,
	)
	if err != nil {
		return err
	}
	event.Route = platformevents.OutboxRouteHandler
	event.DedupeKey = practiceFlagAcceptedDedupeKey(payload.UserID, payload.ChallengeID)
	return repo.EnqueueOutboxEvent(ctx, event)
}

func practiceFlagAcceptedDedupeKey(userID, challengeID int64) string {
	return fmt.Sprintf("practice:flag_accepted:%d:%d", userID, challengeID)
}
