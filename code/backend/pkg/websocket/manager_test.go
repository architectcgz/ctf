package websocket

import (
	"testing"

	"go.uber.org/zap"

	"ctf-platform/internal/authctx"
)

func TestManagerSendToChannelOnlyDeliversSubscribedClients(t *testing.T) {
	t.Parallel()

	manager := &Manager{
		clients:  make(map[int64]map[string]*client),
		channels: make(map[string]map[string]*client),
		logger:   zap.NewNop(),
	}

	subscribed := &client{
		id:       "subscribed",
		user:     authctx.CurrentUser{UserID: 1001},
		send:     make(chan Envelope, 1),
		stop:     make(chan struct{}),
		channels: map[string]struct{}{"contest:42:announcements": {}},
		logger:   zap.NewNop(),
	}
	otherChannel := &client{
		id:       "other",
		user:     authctx.CurrentUser{UserID: 1002},
		send:     make(chan Envelope, 1),
		stop:     make(chan struct{}),
		channels: map[string]struct{}{"contest:99:announcements": {}},
		logger:   zap.NewNop(),
	}
	plainClient := &client{
		id:     "plain",
		user:   authctx.CurrentUser{UserID: 1003},
		send:   make(chan Envelope, 1),
		stop:   make(chan struct{}),
		logger: zap.NewNop(),
	}

	manager.register(subscribed)
	manager.register(otherChannel)
	manager.register(plainClient)

	sent := manager.SendToChannel("contest:42:announcements", Envelope{Type: "contest.announcement.created"})
	if sent != 1 {
		t.Fatalf("expected 1 channel delivery, got %d", sent)
	}

	select {
	case message := <-subscribed.send:
		if message.Type != "contest.announcement.created" {
			t.Fatalf("unexpected message type: %s", message.Type)
		}
	default:
		t.Fatal("expected subscribed client to receive message")
	}

	select {
	case <-otherChannel.send:
		t.Fatal("did not expect other channel client to receive message")
	default:
	}

	select {
	case <-plainClient.send:
		t.Fatal("did not expect plain client to receive channel message")
	default:
	}
}

func TestManagerSendToUserStillDeliversAcrossClientTypes(t *testing.T) {
	t.Parallel()

	manager := &Manager{
		clients:  make(map[int64]map[string]*client),
		channels: make(map[string]map[string]*client),
		logger:   zap.NewNop(),
	}

	channelClient := &client{
		id:       "channel-client",
		user:     authctx.CurrentUser{UserID: 2001},
		send:     make(chan Envelope, 1),
		stop:     make(chan struct{}),
		channels: map[string]struct{}{"contest:7:scoreboard": {}},
		logger:   zap.NewNop(),
	}
	plainClient := &client{
		id:     "plain-client",
		user:   authctx.CurrentUser{UserID: 2001},
		send:   make(chan Envelope, 1),
		stop:   make(chan struct{}),
		logger: zap.NewNop(),
	}

	manager.register(channelClient)
	manager.register(plainClient)

	sent := manager.SendToUser(2001, Envelope{Type: "notification.created"})
	if sent != 2 {
		t.Fatalf("expected 2 user deliveries, got %d", sent)
	}
}
