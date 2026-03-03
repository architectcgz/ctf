package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/middleware"
	"ctf-platform/pkg/apperror"
)

type Envelope struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	RequestID string `json:"request_id"`
}

func Success(c *gin.Context, data any) {
	SuccessWithStatus(c, http.StatusOK, data)
}

func SuccessWithStatus(c *gin.Context, status int, data any) {
	c.JSON(status, Envelope{
		Code:      0,
		Message:   "success",
		Data:      data,
		RequestID: requestID(c),
	})
}

func Error(c *gin.Context, err *apperror.Error) {
	c.JSON(err.HTTPStatus, Envelope{
		Code:      err.Code,
		Message:   err.Message,
		Data:      nil,
		RequestID: requestID(c),
	})
}

func FromError(c *gin.Context, err error) {
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		Error(c, appErr)
		return
	}
	Error(c, apperror.ErrInternal)
}

func ValidationError(c *gin.Context, _ error) {
	Error(c, apperror.ErrInvalidParams)
}

func InternalError(c *gin.Context) {
	Error(c, apperror.ErrInternal)
}

func requestID(c *gin.Context) string {
	raw, ok := c.Get(middleware.RequestIDKey)
	if !ok {
		return ""
	}
	if requestID, ok := raw.(string); ok {
		return requestID
	}
	return ""
}
