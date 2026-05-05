package dto

import "time"

// CreateTeamReq 创建队伍请求
type CreateTeamReq struct {
	Name       string `json:"name" binding:"required,min=2,max=50"`
	MaxMembers int    `json:"max_members" binding:"omitempty,min=2,max=10"`
}

// TeamResp 队伍响应
type TeamResp struct {
	ID          int64     `json:"id"`
	ContestID   int64     `json:"contest_id"`
	Name        string    `json:"name"`
	CaptainID   int64     `json:"captain_id"`
	InviteCode  string    `json:"invite_code"`
	MaxMembers  int       `json:"max_members"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// TeamMemberResp 队伍成员响应
type TeamMemberResp struct {
	UserID   int64     `json:"user_id"`
	Username string    `json:"username"`
	JoinedAt time.Time `json:"joined_at"`
}

type MyTeamResp struct {
	ID         int64             `json:"id"`
	Name       string            `json:"name"`
	InviteCode string            `json:"invite_code"`
	CaptainID  int64             `json:"captain_user_id"`
	Members    []*TeamMemberResp `json:"members"`
}
