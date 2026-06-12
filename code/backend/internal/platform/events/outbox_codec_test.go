package events

import (
	"testing"
	"time"
)

type codecTestPayload struct {
	UserID      int64     `json:"user_id"`
	ChallengeID int64     `json:"challenge_id"`
	OccurredAt  time.Time `json:"occurred_at"`
}

func TestOutboxCodecEncodesAndDecodesTypedPayload(t *testing.T) {
	t.Parallel()

	codec := NewOutboxCodec()
	codec.Register("practice.flag_accepted", 1, func() any { return &codecTestPayload{} })

	occurredAt := time.Date(2026, 6, 12, 8, 30, 0, 0, time.FixedZone("CST", 8*60*60))
	encoded, err := codec.Encode("practice.flag_accepted", 1, codecTestPayload{
		UserID:      7,
		ChallengeID: 11,
		OccurredAt:  occurredAt,
	}, occurredAt)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	if encoded.OccurredAt.Location() != time.UTC {
		t.Fatalf("encoded occurred_at location = %v, want UTC", encoded.OccurredAt.Location())
	}
	if encoded.OccurredAt.Hour() != 0 || encoded.OccurredAt.Day() != 12 {
		t.Fatalf("encoded occurred_at = %s, want UTC-normalized timestamp", encoded.OccurredAt)
	}

	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	payload, ok := decoded.Payload.(*codecTestPayload)
	if !ok {
		t.Fatalf("decoded payload type = %T, want *codecTestPayload", decoded.Payload)
	}
	if payload.UserID != 7 || payload.ChallengeID != 11 {
		t.Fatalf("decoded payload = %+v", payload)
	}
	if payload.OccurredAt.Location() != time.UTC {
		t.Fatalf("decoded payload occurred_at location = %v, want UTC", payload.OccurredAt.Location())
	}
}

func TestOutboxCodecRejectsUnknownEventVersion(t *testing.T) {
	t.Parallel()

	codec := NewOutboxCodec()
	encoded, err := codec.Encode("practice.flag_accepted", 1, codecTestPayload{UserID: 7}, time.Now())
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if _, err := codec.Decode(encoded); err == nil {
		t.Fatal("Decode() error = nil, want unknown event/version error")
	}
}
