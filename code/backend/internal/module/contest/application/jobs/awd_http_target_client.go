package jobs

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"

	"ctf-platform/internal/model"
)

func (u *AWDRoundUpdater) httpClientForAWDTarget(accessURL, runtimeDetails string) *http.Client {
	client := u.httpClient
	if client == nil {
		client = &http.Client{Timeout: normalizedAWDCheckerTimeout(u.cfg.CheckerTimeout)}
	}

	targetHost, dialIP := resolveAWDHTTPDialOverride(accessURL, runtimeDetails)
	if targetHost == "" || dialIP == "" {
		return client
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if baseTransport, ok := client.Transport.(*http.Transport); ok && baseTransport != nil {
		transport = baseTransport.Clone()
	}
	baseDialContext := transport.DialContext
	if baseDialContext == nil {
		dialer := &net.Dialer{}
		baseDialContext = dialer.DialContext
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
		CheckRedirect: client.CheckRedirect,
		Jar:           client.Jar,
		Timeout:       client.Timeout,
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
