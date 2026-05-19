package contracts

import (
	"net/http"

	"ctf-platform/internal/apperror"
)

var ErrNotificationNotFound = apperror.Define(15001, "通知不存在", http.StatusNotFound)
