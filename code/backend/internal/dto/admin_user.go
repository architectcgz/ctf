package dto

import "time"

type AdminUserQuery struct {
	Keyword   string `form:"keyword" binding:"omitempty,max=128"`
	Role      string `form:"role" binding:"omitempty,oneof=student teacher admin"`
	Status    string `form:"status" binding:"omitempty,oneof=active inactive locked banned"`
	ClassName string `form:"class_name" binding:"omitempty,max=128"`
	Page      int    `form:"page" binding:"omitempty,min=1"`
	Size      int    `form:"size" binding:"omitempty,min=1,max=100"`
}

type AdminUserResp struct {
	ID        int64      `json:"id"`
	Username  string     `json:"username"`
	Email     *string    `json:"email,omitempty"`
	ClassName *string    `json:"class_name,omitempty"`
	Status    string     `json:"status"`
	Roles     []string   `json:"roles"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type CreateAdminUserReq struct {
	Username  string `json:"username" binding:"required,min=3,max=64,ctf_username"`
	Password  string `json:"password" binding:"required,min=8,max=72"`
	Email     string `json:"email" binding:"omitempty,email,max=255"`
	ClassName string `json:"class_name" binding:"omitempty,max=128"`
	Role      string `json:"role" binding:"required,oneof=student teacher admin"`
	Status    string `json:"status" binding:"omitempty,oneof=active inactive locked banned"`
}

type UpdateAdminUserReq struct {
	Password  *string `json:"password" binding:"omitempty,min=8,max=72"`
	Email     *string `json:"email" binding:"omitempty,email,max=255"`
	ClassName *string `json:"class_name" binding:"omitempty,max=128"`
	Role      *string `json:"role" binding:"omitempty,oneof=student teacher admin"`
	Status    *string `json:"status" binding:"omitempty,oneof=active inactive locked banned"`
}

type ImportUsersResp struct {
	Created int               `json:"created"`
	Updated int               `json:"updated"`
	Failed  int               `json:"failed"`
	Errors  []ImportUserError `json:"errors,omitempty"`
}

type ImportUserError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}
