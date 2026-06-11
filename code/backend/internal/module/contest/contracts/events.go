package contracts

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	EventAnnouncementCreated = "contest.announcement_created"
	EventAnnouncementDeleted = "contest.announcement_deleted"
	EventFlagAccepted        = "contest.flag_accepted"
	EventAWDAttackAccepted   = "contest.awd.attack_accepted"
	EventAWDPreviewProgress  = "contest.awd_preview_progress"
	EventScoreboardUpdated   = "contest.scoreboard_updated"

	RealtimeDeliveryChannel = "channel"
	RealtimeDeliveryUser    = "user"
)

func AnnouncementChannel(contestID int64) string {
	return fmt.Sprintf("contest:%d:announcements", contestID)
}

func ScoreboardChannel(contestID int64) string {
	return fmt.Sprintf("contest:%d:scoreboard", contestID)
}

type AnnouncementCreatedEvent struct {
	ContestID      int64
	AnnouncementID int64
	Title          string
	Content        string
	CreatedAt      time.Time
	OccurredAt     time.Time
}

type AnnouncementDeletedEvent struct {
	ContestID      int64
	AnnouncementID int64
	OccurredAt     time.Time
}

type FlagAcceptedEvent struct {
	UserID      int64
	ContestID   int64
	ChallengeID int64
	Dimension   string
	OccurredAt  time.Time
}

type AWDAttackAcceptedEvent struct {
	UserID         int64
	ContestID      int64
	AWDChallengeID int64
	Dimension      string
	OccurredAt     time.Time
}

type AWDPreviewProgressEvent struct {
	UserID           int64
	ContestID        int64
	PreviewRequestID string
	PhaseKey         string
	PhaseLabel       string
	Detail           string
	Attempt          int
	TotalAttempts    int
	Status           string
	Error            string
	OccurredAt       time.Time
}

type ScoreboardUpdatedEvent struct {
	ContestID  int64
	OccurredAt time.Time
}

type RealtimeRelayEvent struct {
	EventName       string    `json:"event_name"`
	Delivery        string    `json:"delivery"`
	Channel         string    `json:"channel,omitempty"`
	RecipientUserID *int64    `json:"recipient_user_id,omitempty"`
	MessageType     string    `json:"message_type"`
	Payload         any       `json:"payload,omitempty"`
	Timestamp       time.Time `json:"timestamp"`
}

type AnnouncementRealtimePayload struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type AnnouncementCreatedRelayPayload struct {
	ContestID    int64                       `json:"contest_id"`
	Announcement AnnouncementRealtimePayload `json:"announcement"`
}

type AnnouncementDeletedRelayPayload struct {
	ContestID      int64 `json:"contest_id"`
	AnnouncementID int64 `json:"announcement_id"`
}

type ScoreboardUpdatedRelayPayload struct {
	ContestID int64 `json:"contest_id"`
}

type AWDPreviewProgressRelayPayload struct {
	ContestID        int64  `json:"contest_id"`
	PreviewRequestID string `json:"preview_request_id"`
	PhaseKey         string `json:"phase_key"`
	PhaseLabel       string `json:"phase_label"`
	Detail           string `json:"detail"`
	Attempt          int    `json:"attempt,omitempty"`
	TotalAttempts    int    `json:"total_attempts,omitempty"`
	Status           string `json:"status"`
	Error            string `json:"error,omitempty"`
}

type realtimeRelayEventWire struct {
	EventName       string          `json:"event_name"`
	Delivery        string          `json:"delivery"`
	Channel         string          `json:"channel,omitempty"`
	RecipientUserID *int64          `json:"recipient_user_id,omitempty"`
	MessageType     string          `json:"message_type"`
	Payload         json.RawMessage `json:"payload,omitempty"`
	Timestamp       time.Time       `json:"timestamp"`
}

func RelayAnnouncementCreated(evt AnnouncementCreatedEvent) RealtimeRelayEvent {
	return RealtimeRelayEvent{
		EventName:   EventAnnouncementCreated,
		Delivery:    RealtimeDeliveryChannel,
		Channel:     AnnouncementChannel(evt.ContestID),
		MessageType: "contest.announcement.created",
		Payload: AnnouncementCreatedRelayPayload{
			ContestID: evt.ContestID,
			Announcement: AnnouncementRealtimePayload{
				ID:        evt.AnnouncementID,
				Title:     evt.Title,
				Content:   evt.Content,
				CreatedAt: evt.CreatedAt,
			},
		},
		Timestamp: contestEventTimestamp(evt.OccurredAt),
	}
}

func RelayAnnouncementDeleted(evt AnnouncementDeletedEvent) RealtimeRelayEvent {
	return RealtimeRelayEvent{
		EventName:   EventAnnouncementDeleted,
		Delivery:    RealtimeDeliveryChannel,
		Channel:     AnnouncementChannel(evt.ContestID),
		MessageType: "contest.announcement.deleted",
		Payload: AnnouncementDeletedRelayPayload{
			ContestID:      evt.ContestID,
			AnnouncementID: evt.AnnouncementID,
		},
		Timestamp: contestEventTimestamp(evt.OccurredAt),
	}
}

func RelayScoreboardUpdated(evt ScoreboardUpdatedEvent) RealtimeRelayEvent {
	return RealtimeRelayEvent{
		EventName:   EventScoreboardUpdated,
		Delivery:    RealtimeDeliveryChannel,
		Channel:     ScoreboardChannel(evt.ContestID),
		MessageType: "scoreboard.updated",
		Payload:     ScoreboardUpdatedRelayPayload{ContestID: evt.ContestID},
		Timestamp:   contestEventTimestamp(evt.OccurredAt),
	}
}

func RelayAWDPreviewProgress(evt AWDPreviewProgressEvent) RealtimeRelayEvent {
	messagePayload := AWDPreviewProgressRelayPayload{
		ContestID:        evt.ContestID,
		PreviewRequestID: strings.TrimSpace(evt.PreviewRequestID),
		PhaseKey:         strings.TrimSpace(evt.PhaseKey),
		PhaseLabel:       strings.TrimSpace(evt.PhaseLabel),
		Detail:           strings.TrimSpace(evt.Detail),
		Status:           strings.TrimSpace(evt.Status),
	}
	if evt.Attempt > 0 {
		messagePayload.Attempt = evt.Attempt
	}
	if evt.TotalAttempts > 0 {
		messagePayload.TotalAttempts = evt.TotalAttempts
	}
	if evt.Error != "" {
		messagePayload.Error = strings.TrimSpace(evt.Error)
	}
	return RealtimeRelayEvent{
		EventName:       EventAWDPreviewProgress,
		Delivery:        RealtimeDeliveryUser,
		RecipientUserID: &evt.UserID,
		MessageType:     "awd.preview.progress",
		Payload:         messagePayload,
		Timestamp:       contestEventTimestamp(evt.OccurredAt),
	}
}

func EncodeRealtimeRelay(relay RealtimeRelayEvent) ([]byte, error) {
	payload, err := json.Marshal(relay.Payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(realtimeRelayEventWire{
		EventName:       relay.EventName,
		Delivery:        relay.Delivery,
		Channel:         relay.Channel,
		RecipientUserID: relay.RecipientUserID,
		MessageType:     relay.MessageType,
		Payload:         payload,
		Timestamp:       relay.Timestamp,
	})
}

func DecodeRealtimeRelay(data []byte) (RealtimeRelayEvent, error) {
	var wire realtimeRelayEventWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return RealtimeRelayEvent{}, err
	}
	payload, err := DecodeRealtimeRelayPayload(wire.EventName, wire.Payload)
	if err != nil {
		return RealtimeRelayEvent{}, err
	}
	return RealtimeRelayEvent{
		EventName:       wire.EventName,
		Delivery:        wire.Delivery,
		Channel:         wire.Channel,
		RecipientUserID: wire.RecipientUserID,
		MessageType:     wire.MessageType,
		Payload:         payload,
		Timestamp:       wire.Timestamp,
	}, nil
}

func DecodeRealtimeRelayPayload(eventName string, payload []byte) (any, error) {
	if len(payload) == 0 {
		return nil, nil
	}
	switch eventName {
	case EventAnnouncementCreated:
		var decoded AnnouncementCreatedRelayPayload
		if err := json.Unmarshal(payload, &decoded); err != nil {
			return nil, err
		}
		return decoded, nil
	case EventAnnouncementDeleted:
		var decoded AnnouncementDeletedRelayPayload
		if err := json.Unmarshal(payload, &decoded); err != nil {
			return nil, err
		}
		return decoded, nil
	case EventScoreboardUpdated:
		var decoded ScoreboardUpdatedRelayPayload
		if err := json.Unmarshal(payload, &decoded); err != nil {
			return nil, err
		}
		return decoded, nil
	case EventAWDPreviewProgress:
		var decoded AWDPreviewProgressRelayPayload
		if err := json.Unmarshal(payload, &decoded); err != nil {
			return nil, err
		}
		return decoded, nil
	default:
		var decoded map[string]any
		if err := json.Unmarshal(payload, &decoded); err != nil {
			return nil, err
		}
		return decoded, nil
	}
}

func contestEventTimestamp(ts time.Time) time.Time {
	if ts.IsZero() {
		return time.Now().UTC()
	}
	return ts.UTC()
}
