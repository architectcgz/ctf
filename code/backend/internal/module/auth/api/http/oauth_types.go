package http

type OAuthClientRegistrationReq struct {
	ClientName              string   `json:"client_name" binding:"required,max=128"`
	ClientURI               string   `json:"client_uri" binding:"omitempty,max=512"`
	RedirectURIs            []string `json:"redirect_uris" binding:"required,min=1,dive,required,max=512"`
	GrantTypes              []string `json:"grant_types" binding:"omitempty,dive,required"`
	ResponseTypes           []string `json:"response_types" binding:"omitempty,dive,required"`
	Scope                   string   `json:"scope" binding:"omitempty,max=256"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method" binding:"omitempty,max=64"`
}

type OAuthClientRegistrationResp struct {
	ClientID                string   `json:"client_id"`
	ClientName              string   `json:"client_name"`
	ClientURI               string   `json:"client_uri,omitempty"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	Scope                   string   `json:"scope"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
}

type OAuthTokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
}

type OAuthErrorResp struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}
