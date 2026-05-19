package contracts

import (
	"net/http"

	"ctf-platform/internal/apperror"
)

var ErrWSTicketInvalid = apperror.Define(15002, "WebSocket Ticket 无效或已过期", http.StatusUnauthorized)
