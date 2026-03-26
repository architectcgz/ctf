package queries

import (
	"context"
	"net/url"
	"strings"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/errcode"
)

const (
	casProviderName     = "cas"
	casLoginPath        = "/api/v1/auth/cas/login"
	casCallbackPath     = "/api/v1/auth/cas/callback"
	defaultCASLoginPath = "/login"
)

type CASService interface {
	Status() *dto.CASStatusResp
	BuildLogin(ctx context.Context) (*dto.CASLoginResp, error)
}

type casService struct {
	config config.CASConfig
}

func NewCASService(cfg config.CASConfig) CASService {
	return &casService{config: cfg}
}

func (s *casService) Status() *dto.CASStatusResp {
	return &dto.CASStatusResp{
		Provider:      casProviderName,
		Enabled:       s.config.Enabled,
		Configured:    s.isConfigured(),
		AutoProvision: s.config.AutoProvision,
		LoginPath:     casLoginPath,
		CallbackPath:  casCallbackPath,
	}
}

func (s *casService) BuildLogin(context.Context) (*dto.CASLoginResp, error) {
	if !s.config.Enabled {
		return nil, errcode.ErrCASDisabled
	}
	if !s.isConfigured() {
		return nil, errcode.ErrCASNotConfigured
	}

	loginURL, err := s.buildLoginURL()
	if err != nil {
		return nil, errcode.ErrCASNotConfigured.WithCause(err)
	}
	return &dto.CASLoginResp{
		Provider:    casProviderName,
		RedirectURL: loginURL,
		CallbackURL: s.config.ServiceURL,
	}, nil
}

func (s *casService) isConfigured() bool {
	return strings.TrimSpace(s.config.BaseURL) != "" && strings.TrimSpace(s.config.ServiceURL) != ""
}

func (s *casService) buildLoginURL() (string, error) {
	loginPath := strings.TrimSpace(s.config.LoginPath)
	if loginPath == "" {
		loginPath = defaultCASLoginPath
	}
	base, err := url.Parse(strings.TrimRight(s.config.BaseURL, "/"))
	if err != nil {
		return "", err
	}
	base.Path = strings.TrimRight(base.Path, "/") + loginPath

	query := base.Query()
	query.Set("service", s.config.ServiceURL)
	base.RawQuery = query.Encode()
	return base.String(), nil
}
