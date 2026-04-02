package ports

import (
	"fmt"

	ctfws "ctf-platform/pkg/websocket"
)

type RealtimeBroadcaster interface {
	SendToChannel(channel string, message ctfws.Envelope) int
}

func AnnouncementChannel(contestID int64) string {
	return fmt.Sprintf("contest:%d:announcements", contestID)
}

func ScoreboardChannel(contestID int64) string {
	return fmt.Sprintf("contest:%d:scoreboard", contestID)
}
