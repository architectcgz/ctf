package infrastructure

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type AWDHTTPRuntimeAdapter struct {
	client         *http.Client
	defaultTimeout time.Duration
}

func NewAWDHTTPRuntimeAdapter(client *http.Client, defaultTimeout time.Duration) *AWDHTTPRuntimeAdapter {
	if client == nil {
		client = &http.Client{}
	}
	return &AWDHTTPRuntimeAdapter{
		client:         client,
		defaultTimeout: normalizedAWDHTTPRuntimeTimeout(defaultTimeout),
	}
}

func (a *AWDHTTPRuntimeAdapter) Execute(ctx context.Context, request contestports.AWDHTTPRequest) (contestports.AWDHTTPResponse, error) {
	if a == nil {
		a = NewAWDHTTPRuntimeAdapter(nil, 0)
	}

	reqCtx, cancel := context.WithTimeout(ctx, a.timeout(request.Timeout))
	defer cancel()

	httpRequest, err := http.NewRequestWithContext(reqCtx, request.Method, strings.TrimSpace(request.URL), bytes.NewBufferString(request.Body))
	if err != nil {
		return contestports.AWDHTTPResponse{}, &contestports.AWDHTTPRuntimeError{
			Kind: contestports.AWDHTTPRuntimeErrorKindRequestBuild,
			Err:  err,
		}
	}
	for key, value := range request.Headers {
		httpRequest.Header.Set(key, value)
	}

	client := a.clientForRequest(request.AccessURL, request.RuntimeDetails)
	response, err := client.Do(httpRequest)
	if err != nil {
		return contestports.AWDHTTPResponse{}, &contestports.AWDHTTPRuntimeError{
			Kind: contestports.AWDHTTPRuntimeErrorKindRequestExecute,
			Err:  err,
		}
	}
	defer response.Body.Close()

	if !request.ReadBody {
		_, _ = io.Copy(io.Discard, response.Body)
		return contestports.AWDHTTPResponse{StatusCode: response.StatusCode}, nil
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return contestports.AWDHTTPResponse{StatusCode: response.StatusCode}, &contestports.AWDHTTPRuntimeError{
			Kind: contestports.AWDHTTPRuntimeErrorKindResponseRead,
			Err:  err,
		}
	}
	return contestports.AWDHTTPResponse{
		StatusCode: response.StatusCode,
		Body:       string(body),
	}, nil
}

func (a *AWDHTTPRuntimeAdapter) timeout(requestTimeout time.Duration) time.Duration {
	if requestTimeout > 0 {
		return requestTimeout
	}
	return normalizedAWDHTTPRuntimeTimeout(a.defaultTimeout)
}

func normalizedAWDHTTPRuntimeTimeout(value time.Duration) time.Duration {
	if value <= 0 {
		return 3 * time.Second
	}
	return value
}

func (a *AWDHTTPRuntimeAdapter) clientForRequest(accessURL, runtimeDetails string) *http.Client {
	if a == nil || a.client == nil {
		return NewAWDHTTPRuntimeAdapter(nil, 0).client
	}

	targetHost, dialIP := resolveAWDHTTPDialOverride(accessURL, runtimeDetails)
	if targetHost == "" || dialIP == "" {
		return a.client
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if baseTransport, ok := a.client.Transport.(*http.Transport); ok && baseTransport != nil {
		transport = baseTransport.Clone()
	}
	baseDialContext := transport.DialContext
	if baseDialContext == nil {
		baseDialContext = (&net.Dialer{}).DialContext
	}

	transport.Proxy = nil
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err == nil && strings.EqualFold(host, targetHost) {
			address = net.JoinHostPort(dialIP, port)
		}
		return baseDialContext(ctx, network, address)
	}

	return &http.Client{
		Transport:     transport,
		CheckRedirect: a.client.CheckRedirect,
		Jar:           a.client.Jar,
	}
}

func resolveAWDHTTPDialOverride(accessURL, runtimeDetails string) (string, string) {
	parsed, err := url.Parse(strings.TrimSpace(accessURL))
	if err != nil {
		return "", ""
	}
	targetHost := strings.TrimSpace(parsed.Hostname())
	if targetHost == "" {
		return "", ""
	}

	resolved := model.ResolveRuntimeAliasAccessURL(accessURL, runtimeDetails)
	resolvedParsed, err := url.Parse(strings.TrimSpace(resolved))
	if err != nil {
		return "", ""
	}
	dialIP := strings.TrimSpace(resolvedParsed.Hostname())
	if dialIP == "" || strings.EqualFold(dialIP, targetHost) || net.ParseIP(dialIP) == nil {
		return "", ""
	}
	return targetHost, dialIP
}

var _ contestports.AWDHTTPRuntime = (*AWDHTTPRuntimeAdapter)(nil)
