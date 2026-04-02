package commands

import (
	"time"

	contestports "ctf-platform/internal/module/contest/ports"
	ctfws "ctf-platform/pkg/websocket"
)

func broadcastContestRealtimeEvent(broadcaster contestports.RealtimeBroadcaster, channel string, message ctfws.Envelope) {
	if broadcaster == nil || channel == "" {
		return
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now().UTC()
	}
	broadcaster.SendToChannel(channel, message)
}
