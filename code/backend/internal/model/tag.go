package model

import "time"

const (
	TagTypeVulnerability = "vulnerability" // 漏洞类型
	TagTypeTechStack     = "tech_stack"    // 技术栈
	TagTypeKnowledge     = "knowledge"     // 知识点
)

type Tag struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	Name        string    `gorm:"column:name"`        // 标签名称
	Type        string    `gorm:"column:type"`        // 标签类型：vulnerability/tech_stack/knowledge
	Description string    `gorm:"column:description"` // 标签描述
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (Tag) TableName() string {
	return "tags"
}

type ChallengeTag struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	ChallengeID int64     `gorm:"column:challenge_id"`
	TagID       int64     `gorm:"column:tag_id"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (ChallengeTag) TableName() string {
	return "challenge_tags"
}
