package dto

type RegisterReq struct {
	Username  string `json:"username" binding:"required,min=3,max=64,ctf_username"`
	Password  string `json:"password" binding:"required,min=8,max=72"`
	Email     string `json:"email" binding:"omitempty,email,max=255"`
	ClassName string `json:"class_name" binding:"omitempty,max=128"`
}

type LoginReq struct {
	Username string `json:"username" binding:"required,min=3,max=64,ctf_username"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type RefreshReq struct{}

type AuthUser struct {
	ID        int64   `json:"id"`
	Username  string  `json:"username"`
	Role      string  `json:"role"`
	Avatar    *string `json:"avatar,omitempty"`
	Name      *string `json:"name,omitempty"`
	ClassName *string `json:"class_name,omitempty"`
}

type LoginResp struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int64    `json:"expires_in"`
	User        AuthUser `json:"user"`
}

type RefreshResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type WSTicketResp struct {
	Ticket    string `json:"ticket"`
	ExpiresAt string `json:"expires_at"`
}
