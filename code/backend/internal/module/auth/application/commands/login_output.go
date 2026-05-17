package commands

type AuthUser struct {
	ID        int64   `json:"id"`
	Username  string  `json:"username"`
	Role      string  `json:"role"`
	Avatar    *string `json:"avatar,omitempty"`
	Name      *string `json:"name,omitempty"`
	ClassName *string `json:"class_name,omitempty"`
}

type LoginResp struct {
	User AuthUser `json:"user"`
}
