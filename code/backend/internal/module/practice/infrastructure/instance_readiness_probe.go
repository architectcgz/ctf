package infrastructure

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type instanceReadinessProbe struct{}

func NewInstanceReadinessProbe() practiceports.PracticeInstanceReadinessProbe {
	return instanceReadinessProbe{}
}

func (instanceReadinessProbe) ProbeAccessURL(ctx context.Context, accessURL string, timeout time.Duration) error {
	parsed, err := url.Parse(accessURL)
	if err != nil {
		return err
	}
	if strings.EqualFold(parsed.Scheme, model.ChallengeTargetProtocolTCP) {
		return probeTCPAccessURL(ctx, parsed, timeout)
	}

	probeCtx, cancel := withOptionalTimeout(ctx, timeout)
	defer cancel()

	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, accessURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 512))
	return nil
}

func probeTCPAccessURL(ctx context.Context, parsed *url.URL, timeout time.Duration) error {
	host := parsed.Host
	if strings.TrimSpace(host) == "" {
		return fmt.Errorf("tcp access url missing host")
	}
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", host)
	if err != nil {
		return err
	}
	return conn.Close()
}

func withOptionalTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}
