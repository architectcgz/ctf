package infrastructure

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	authports "ctf-platform/internal/module/auth/ports"
	"ctf-platform/internal/validation"
)

const defaultCASValidateTimeout = 5 * time.Second

type casTicketValidator struct {
	client *http.Client
	log    *zap.Logger
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

func NewCASTicketValidator(log *zap.Logger, httpClient *http.Client) authports.CASTicketValidator {
	if log == nil {
		log = zap.NewNop()
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultCASValidateTimeout}
	}
	return &casTicketValidator{
		client: httpClient,
		log:    log,
	}
}

func (v *casTicketValidator) ValidateTicket(ctx context.Context, validateURL string) (*authports.CASPrincipal, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, validateURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build cas validate request: %w", err)
	}

	resp, err := v.client.Do(req)
	if err != nil {
		v.log.Warn("auth_cas_validate_request_failed", zap.String("validate_url", validateURL), zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read cas validate response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("cas validate status %d", resp.StatusCode)
		v.log.Warn("auth_cas_validate_http_error", zap.Int("status", resp.StatusCode), zap.ByteString("body", body))
		return nil, err
	}

	var result casValidateResponse
	if err := xml.Unmarshal(body, &result); err != nil {
		v.log.Warn("auth_cas_validate_decode_failed", zap.Error(err), zap.ByteString("body", body))
		return nil, err
	}
	if result.AuthenticationFailure != nil {
		v.log.Info(
			"auth_cas_validate_rejected",
			zap.String("code", strings.TrimSpace(result.AuthenticationFailure.Code)),
			zap.String("message", strings.TrimSpace(result.AuthenticationFailure.Message)),
		)
		return nil, authports.ErrCASTicketInvalid
	}
	if result.AuthenticationSuccess == nil {
		return nil, authports.ErrCASTicketInvalid
	}

	principal := &authports.CASPrincipal{
		Username:  strings.TrimSpace(result.AuthenticationSuccess.User),
		Name:      result.AuthenticationSuccess.Attributes.pick("name", "displayName", "realName", "cn"),
		Email:     result.AuthenticationSuccess.Attributes.pick("email", "mail"),
		ClassName: result.AuthenticationSuccess.Attributes.pick("class_name", "className", "class"),
		StudentNo: result.AuthenticationSuccess.Attributes.pick("student_no", "studentNo", "studentId", "studentNumber"),
		TeacherNo: result.AuthenticationSuccess.Attributes.pick("teacher_no", "teacherNo", "teacherId", "teacherNumber"),
	}
	if principal.Username == "" || !validation.IsValidUsername(principal.Username) {
		v.log.Warn("auth_cas_invalid_username", zap.String("username", principal.Username))
		return nil, authports.ErrCASTicketInvalid
	}
	return principal, nil
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
