package infrastructure

import (
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
		defaultTimeout: defaultTimeout,
	}
}

func (a *AWDHTTPRuntimeAdapter) Execute(ctx context.Context, request contestports.AWDHTTPRequest) (contestports.AWDHTTPResponse, error) {
	timeout := request.Timeout
	if timeout <= 0 {
		timeout = a.defaultTimeout
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, request.Method, request.URL, strings.NewReader(request.Body))
	if err != nil {
		return contestports.AWDHTTPResponse{}, newAWDHTTPRuntimeError(contestports.AWDHTTPRuntimeErrorKindRequestBuild, err)
	}
	for key, value := range request.Headers {
		req.Header.Set(key, value)
	}

	client := a.client
	if client == nil {
		client = &http.Client{}
	}
	if targetHost, dialIP := resolveAWDHTTPDialOverride(request.AccessURL, request.RuntimeDetails); targetHost != "" && dialIP != "" {
		client = cloneAWDHTTPRuntimeClientWithDialOverride(client, targetHost, dialIP)
	}

	resp, err := client.Do(req)
	if err != nil {
		return contestports.AWDHTTPResponse{}, newAWDHTTPRuntimeError(contestports.AWDHTTPRuntimeErrorKindRequestExecute, err)
	}

	if !request.ReadBody {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		return contestports.AWDHTTPResponse{StatusCode: resp.StatusCode}, nil
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return contestports.AWDHTTPResponse{}, newAWDHTTPRuntimeError(contestports.AWDHTTPRuntimeErrorKindResponseRead, err)
	}

	return contestports.AWDHTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       string(bodyBytes),
	}, nil
}

func newAWDHTTPRuntimeError(kind contestports.AWDHTTPRuntimeErrorKind, err error) error {
	return &contestports.AWDHTTPRuntimeError{
		Kind: kind,
		Err:  err,
	}
}

func cloneAWDHTTPRuntimeClientWithDialOverride(client *http.Client, targetHost, dialIP string) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if baseTransport, ok := client.Transport.(*http.Transport); ok && baseTransport != nil {
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

	clonedClient := *client
	clonedClient.Transport = transport
	return &clonedClient
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
