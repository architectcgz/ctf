package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/validation"
	"ctf-platform/pkg/errcode"
)

const (
	casProviderName           = "cas"
	casLoginPath              = "/api/v1/auth/cas/login"
	casCallbackPath           = "/api/v1/auth/cas/callback"
	defaultCASLoginPath       = "/login"
	defaultCASValidatePath    = "/serviceValidate"
	defaultCASValidateTimeout = 5 * time.Second
)

type CASProvider interface {
	Status() *dto.CASStatusResp
	BuildLogin(ctx context.Context) (*dto.CASLoginResp, error)
	Authenticate(ctx context.Context, ticket string) (*dto.LoginResp, *TokenPair, error)
}

type casProvider struct {
	config       config.CASConfig
	repo         Repository
	tokenService TokenService
	log          *zap.Logger
	httpClient   *http.Client
}

type casValidateResponse struct {
	XMLName               xml.Name                  `xml:"serviceResponse"`
	AuthenticationSuccess *casAuthenticationSuccess `xml:"authenticationSuccess"`
	AuthenticationFailure *casAuthenticationFailure `xml:"authenticationFailure"`
}

type casAuthenticationSuccess struct {
	User       string        `xml:"user"`
	Attributes casAttributes `xml:"attributes"`
}

type casAuthenticationFailure struct {
	Code    string `xml:"code,attr"`
	Message string `xml:",chardata"`
}

type casAttributes struct {
	Entries []casAttributeEntry `xml:",any"`
}

type casAttributeEntry struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

type casPrincipal struct {
	Username  string
	Name      string
	Email     string
	ClassName string
	StudentNo string
	TeacherNo string
}

func NewCASProvider(cfg config.CASConfig, repo Repository, tokenService TokenService, log *zap.Logger, httpClient *http.Client) CASProvider {
	if log == nil {
		log = zap.NewNop()
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultCASValidateTimeout}
	}

	return &casProvider{
		config:       cfg,
		repo:         repo,
		tokenService: tokenService,
		log:          log,
		httpClient:   httpClient,
	}
}

func (p *casProvider) Status() *dto.CASStatusResp {
	return &dto.CASStatusResp{
		Provider:      casProviderName,
		Enabled:       p.config.Enabled,
		Configured:    p.isConfigured(),
		AutoProvision: p.config.AutoProvision,
		LoginPath:     casLoginPath,
		CallbackPath:  casCallbackPath,
	}
}

func (p *casProvider) BuildLogin(context.Context) (*dto.CASLoginResp, error) {
	if !p.config.Enabled {
		return nil, errcode.ErrCASDisabled
	}
	if !p.isConfigured() {
		return nil, errcode.ErrCASNotConfigured
	}

	loginURL, err := p.buildLoginURL()
	if err != nil {
		return nil, errcode.ErrCASNotConfigured.WithCause(err)
	}
	return &dto.CASLoginResp{
		Provider:    casProviderName,
		RedirectURL: loginURL,
		CallbackURL: p.config.ServiceURL,
	}, nil
}

func (p *casProvider) Authenticate(ctx context.Context, ticket string) (*dto.LoginResp, *TokenPair, error) {
	if !p.config.Enabled {
		return nil, nil, errcode.ErrCASDisabled
	}
	if !p.isConfigured() {
		return nil, nil, errcode.ErrCASNotConfigured
	}
	if p.repo == nil || p.tokenService == nil {
		return nil, nil, errcode.ErrCASNotImplemented
	}

	principal, err := p.validateTicket(ctx, ticket)
	if err != nil {
		return nil, nil, err
	}

	user, err := p.syncUser(ctx, principal)
	if err != nil {
		return nil, nil, err
	}

	return p.issueLoginResp(user)
}

func (p *casProvider) validateTicket(ctx context.Context, ticket string) (*casPrincipal, error) {
	validateURL, err := p.buildValidateURL(ticket)
	if err != nil {
		return nil, errcode.ErrCASNotConfigured.WithCause(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, validateURL, nil)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.log.Warn("auth_cas_validate_request_failed", zap.String("ticket", ticket), zap.Error(err))
		return nil, errcode.ErrServiceUnavailable.WithCause(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("cas validate status %d", resp.StatusCode)
		p.log.Warn("auth_cas_validate_http_error", zap.Int("status", resp.StatusCode), zap.ByteString("body", body))
		return nil, errcode.ErrServiceUnavailable.WithCause(err)
	}

	var result casValidateResponse
	if err := xml.Unmarshal(body, &result); err != nil {
		p.log.Warn("auth_cas_validate_decode_failed", zap.Error(err), zap.ByteString("body", body))
		return nil, errcode.ErrServiceUnavailable.WithCause(err)
	}
	if result.AuthenticationFailure != nil {
		p.log.Info(
			"auth_cas_validate_rejected",
			zap.String("code", strings.TrimSpace(result.AuthenticationFailure.Code)),
			zap.String("message", strings.TrimSpace(result.AuthenticationFailure.Message)),
		)
		return nil, errcode.ErrCASTicketInvalid
	}
	if result.AuthenticationSuccess == nil {
		return nil, errcode.ErrCASTicketInvalid
	}

	principal := &casPrincipal{
		Username:  strings.TrimSpace(result.AuthenticationSuccess.User),
		Name:      result.AuthenticationSuccess.Attributes.pick("name", "displayName", "realName", "cn"),
		Email:     result.AuthenticationSuccess.Attributes.pick("email", "mail"),
		ClassName: result.AuthenticationSuccess.Attributes.pick("class_name", "className", "class"),
		StudentNo: result.AuthenticationSuccess.Attributes.pick("student_no", "studentNo", "studentId", "studentNumber"),
		TeacherNo: result.AuthenticationSuccess.Attributes.pick("teacher_no", "teacherNo", "teacherId", "teacherNumber"),
	}
	if principal.Username == "" || !validation.IsValidUsername(principal.Username) {
		p.log.Warn("auth_cas_invalid_username", zap.String("username", principal.Username))
		return nil, errcode.ErrCASTicketInvalid
	}
	return principal, nil
}

func (p *casProvider) syncUser(ctx context.Context, principal *casPrincipal) (*model.User, error) {
	user, err := p.repo.FindByUsername(ctx, principal.Username)
	if err != nil {
		if !errors.Is(err, ErrUserNotFound) {
			p.log.Error("auth_cas_find_user_failed", zap.String("username", principal.Username), zap.Error(err))
			return nil, errcode.ErrInternal.WithCause(err)
		}
		if !p.config.AutoProvision {
			return nil, errcode.ErrCASUserNotProvisioned
		}

		user = &model.User{
			Username:  principal.Username,
			Name:      principal.Name,
			Email:     principal.Email,
			StudentNo: principal.StudentNo,
			TeacherNo: principal.TeacherNo,
			Role:      model.RoleStudent,
			ClassName: principal.ClassName,
			Status:    model.UserStatusActive,
		}
		if err := user.SetPassword(randomPassword()); err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		if err := p.repo.Create(ctx, user); err != nil {
			return nil, p.mapUserSyncError(err)
		}
		return user, nil
	}

	if user.Status == model.UserStatusBanned {
		return nil, errcode.ErrAccountDisabled
	}
	if user.Status == model.UserStatusLocked && (user.LockedUntil == nil || time.Now().Before(*user.LockedUntil)) {
		return nil, errcode.ErrAccountLocked
	}

	changed := p.mergePrincipal(user, principal)
	if user.Status == model.UserStatusLocked || user.FailedLoginAttempts > 0 || user.LastFailedLoginAt != nil || user.LockedUntil != nil {
		user.Status = model.UserStatusActive
		user.FailedLoginAttempts = 0
		user.LastFailedLoginAt = nil
		user.LockedUntil = nil
		changed = true
	}
	if !changed {
		return user, nil
	}
	if err := p.repo.UpdateCASProfile(ctx, user); err != nil {
		return nil, p.mapUserSyncError(err)
	}
	return user, nil
}

func (p *casProvider) mergePrincipal(user *model.User, principal *casPrincipal) bool {
	changed := false
	if principal.Name != "" && user.Name != principal.Name {
		user.Name = principal.Name
		changed = true
	}
	if principal.Email != "" && user.Email != principal.Email {
		user.Email = principal.Email
		changed = true
	}
	if principal.ClassName != "" && user.ClassName != principal.ClassName {
		user.ClassName = principal.ClassName
		changed = true
	}
	if principal.StudentNo != "" && user.StudentNo != principal.StudentNo {
		user.StudentNo = principal.StudentNo
		changed = true
	}
	if principal.TeacherNo != "" && user.TeacherNo != principal.TeacherNo {
		user.TeacherNo = principal.TeacherNo
		changed = true
	}
	return changed
}

func (p *casProvider) issueLoginResp(user *model.User) (*dto.LoginResp, *TokenPair, error) {
	tokens, err := p.tokenService.IssueTokens(user.ID, user.Username, user.Role)
	if err != nil {
		p.log.Error("auth_cas_issue_token_failed", zap.String("username", user.Username), zap.Int64("user_id", user.ID), zap.Error(err))
		return nil, nil, errcode.ErrInternal.WithCause(err)
	}

	return &dto.LoginResp{
		AccessToken: tokens.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(tokens.AccessTokenTTL.Seconds()),
		User:        buildAuthUser(user),
	}, tokens, nil
}

func (p *casProvider) mapUserSyncError(err error) error {
	switch {
	case errors.Is(err, ErrUsernameExists):
		return errcode.ErrUsernameExists
	case errors.Is(err, ErrEmailExists):
		return errcode.ErrEmailExists
	case errors.Is(err, ErrStudentNoExists):
		return errcode.ErrStudentNoExists
	case errors.Is(err, ErrTeacherNoExists):
		return errcode.ErrTeacherNoExists
	case errors.Is(err, ErrRoleNotFound):
		return errcode.ErrInternal.WithCause(err)
	default:
		return errcode.ErrInternal.WithCause(err)
	}
}

func (p *casProvider) isConfigured() bool {
	return strings.TrimSpace(p.config.BaseURL) != "" && strings.TrimSpace(p.config.ServiceURL) != ""
}

func (p *casProvider) buildLoginURL() (string, error) {
	loginPath := strings.TrimSpace(p.config.LoginPath)
	if loginPath == "" {
		loginPath = defaultCASLoginPath
	}
	return p.buildCASURL(loginPath, "")
}

func (p *casProvider) buildValidateURL(ticket string) (string, error) {
	validatePath := strings.TrimSpace(p.config.ValidatePath)
	if validatePath == "" {
		validatePath = defaultCASValidatePath
	}
	return p.buildCASURL(validatePath, ticket)
}

func (p *casProvider) buildCASURL(pathValue, ticket string) (string, error) {
	base, err := url.Parse(strings.TrimRight(p.config.BaseURL, "/"))
	if err != nil {
		return "", err
	}
	base.Path = strings.TrimRight(base.Path, "/") + pathValue

	query := base.Query()
	query.Set("service", p.config.ServiceURL)
	if ticket != "" {
		query.Set("ticket", ticket)
	}
	base.RawQuery = query.Encode()
	return base.String(), nil
}

func (a casAttributes) pick(names ...string) string {
	for _, entry := range a.Entries {
		key := normalizeCASAttributeName(entry.XMLName.Local)
		for _, candidate := range names {
			if key == normalizeCASAttributeName(candidate) {
				value := strings.TrimSpace(entry.Value)
				if value != "" {
					return value
				}
			}
		}
	}
	return ""
}

func normalizeCASAttributeName(value string) string {
	replacer := strings.NewReplacer("_", "", "-", "", ":", "", ".", "")
	return strings.ToLower(replacer.Replace(strings.TrimSpace(value)))
}

func randomPassword() string {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("cas_%d", time.Now().UnixNano())
	}
	return "cas_" + hex.EncodeToString(buf)
}
