package contracts

import "time"

const OAuthScopeMCPChallengeRead = "mcp:challenge:read"

type OAuthClient struct {
	ID                      int64
	ClientID                string
	ClientName              string
	ClientURI               string
	RedirectURIs            []string
	GrantTypes              []string
	ResponseTypes           []string
	Scope                   string
	TokenEndpointAuthMethod string
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

func (c OAuthClient) AllowsRedirectURI(redirectURI string) bool {
	for _, registered := range c.RedirectURIs {
		if registered == redirectURI {
			return true
		}
	}
	return false
}

type OAuthConsent struct {
	ID        int64
	UserID    int64
	ClientID  string
	Scope     string
	GrantedAt time.Time
	ExpiresAt *time.Time
	RevokedAt *time.Time
}

type OAuthAuthorizationCode struct {
	UserID              int64     `json:"user_id"`
	Username            string    `json:"username"`
	Role                string    `json:"role"`
	ClientID            string    `json:"client_id"`
	RedirectURI         string    `json:"redirect_uri"`
	Scope               string    `json:"scope"`
	CodeChallenge       string    `json:"code_challenge"`
	CodeChallengeMethod string    `json:"code_challenge_method"`
	SessionVersion      int64     `json:"session_version"`
	IssuedAt            time.Time `json:"issued_at"`
	ExpiresAt           time.Time `json:"expires_at"`
}

type OAuthTokenClaims struct {
	UserID         int64     `json:"user_id"`
	Username       string    `json:"username"`
	Role           string    `json:"role"`
	ClientID       string    `json:"client_id"`
	Scope          string    `json:"scope"`
	SessionVersion int64     `json:"session_version"`
	IssuedAt       time.Time `json:"issued_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}
