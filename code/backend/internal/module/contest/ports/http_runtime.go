package ports

import (
	"context"
	"time"
)

type AWDHTTPRuntime interface {
	Execute(ctx context.Context, request AWDHTTPRequest) (AWDHTTPResponse, error)
}

type AWDHTTPRequest struct {
	AccessURL      string
	RuntimeDetails string
	URL            string
	Method         string
	Headers        map[string]string
	Body           string
	ReadBody       bool
	Timeout        time.Duration
}

type AWDHTTPResponse struct {
	StatusCode int
	Body       string
}

type AWDHTTPRuntimeErrorKind string

const (
	AWDHTTPRuntimeErrorKindRequestBuild   AWDHTTPRuntimeErrorKind = "request_build"
	AWDHTTPRuntimeErrorKindRequestExecute AWDHTTPRuntimeErrorKind = "request_execute"
	AWDHTTPRuntimeErrorKindResponseRead   AWDHTTPRuntimeErrorKind = "response_read"
)

type AWDHTTPRuntimeError struct {
	Kind AWDHTTPRuntimeErrorKind
	Err  error
}

func (e *AWDHTTPRuntimeError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return string(e.Kind)
	}
	return string(e.Kind) + ": " + e.Err.Error()
}

func (e *AWDHTTPRuntimeError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
