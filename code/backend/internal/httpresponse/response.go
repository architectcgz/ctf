package httpresponse

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"ctf-platform/internal/apperror"
)

type Envelope struct {
	Code      int          `json:"code"`
	Message   string       `json:"message"`
	Data      any          `json:"data"`
	RequestID string       `json:"request_id"`
	Errors    []FieldError `json:"errors,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
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

func Error(c *gin.Context, err *apperror.AppError) {
	c.JSON(apperror.HTTPStatus(err), NewEnvelope(c, err, nil))
}

func FromError(c *gin.Context, err error) {
	recordError(c, err)

	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		Error(c, appErr)
		return
	}
	Error(c, apperror.ErrInternal)
}

func ValidationError(c *gin.Context, err error) {
	recordError(c, err)

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		fields := make([]FieldError, 0, len(validationErrors))
		for _, item := range validationErrors {
			fields = append(fields, FieldError{
				Field:   item.Field(),
				Message: item.Error(),
			})
		}
		c.JSON(apperror.HTTPStatus(apperror.ErrValidationFailed), NewEnvelope(c, apperror.ErrValidationFailed, fields))
		return
	}
	Error(c, apperror.ErrInvalidParams)
}

func InternalError(c *gin.Context) {
	Error(c, apperror.ErrInternal)
}

func InvalidParams(c *gin.Context, message string) {
	err := apperror.ErrInvalidParams.WithMessage(message)
	Error(c, err)
}

func Page(c *gin.Context, list any, total int64, page, pageSize int) {
	Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func NewEnvelope(c *gin.Context, err *apperror.AppError, fieldErrors []FieldError) Envelope {
	return Envelope{
		Code:      err.Code,
		Message:   err.Message,
		Data:      nil,
		RequestID: requestID(c),
		Errors:    fieldErrors,
	}
}

func requestID(c *gin.Context) string {
	return c.GetString("request_id")
}

func recordError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	var appErr *apperror.AppError
	if errors.As(err, &appErr) && appErr.Cause != nil {
		_ = c.Error(appErr.Cause)
		return
	}

	_ = c.Error(err)
}
