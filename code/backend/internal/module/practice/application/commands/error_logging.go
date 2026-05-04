package commands

import (
	"errors"

	"go.uber.org/zap"
)

func wrappedErrorCauseField(err error) zap.Field {
	cause := errors.Unwrap(err)
	if cause == nil {
		return zap.Skip()
	}
	return zap.NamedError("cause", cause)
}
