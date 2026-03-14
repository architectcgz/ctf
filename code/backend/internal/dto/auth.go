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

type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required,min=8,max=72"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
}

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

type CASStatusResp struct {
	Provider      string `json:"provider"`
	Enabled       bool   `json:"enabled"`
	Configured    bool   `json:"configured"`
	AutoProvision bool   `json:"auto_provision"`
	LoginPath     string `json:"login_path"`
	CallbackPath  string `json:"callback_path"`
}

type CASLoginResp struct {
	Provider    string `json:"provider"`
	RedirectURL string `json:"redirect_url"`
	CallbackURL string `json:"callback_url"`
}
