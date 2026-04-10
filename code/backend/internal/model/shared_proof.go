package model

import "time"

const (
	SharedProofStatusActive   = "active"
	SharedProofStatusConsumed = "consumed"
	SharedProofStatusExpired  = "expired"
)

type SharedProof struct {
	ID          int64      `gorm:"column:id;primaryKey"`
	UserID      int64      `gorm:"column:user_id;not null;index:idx_shared_proofs_user_challenge_contest_status,priority:1"`
	ChallengeID int64      `gorm:"column:challenge_id;not null;index:idx_shared_proofs_user_challenge_contest_status,priority:2"`
	ContestID   *int64     `gorm:"column:contest_id;index:idx_shared_proofs_user_challenge_contest_status,priority:3"`
	InstanceID  int64      `gorm:"column:instance_id;not null;index"`
	ProofHash   string     `gorm:"column:proof_hash;size:64;not null;uniqueIndex:uk_shared_proofs_hash"`
	Status      string     `gorm:"column:status;size:16;not null;default:'active';index:idx_shared_proofs_user_challenge_contest_status,priority:4"`
	ExpiresAt   time.Time  `gorm:"column:expires_at;not null;index"`
	ConsumedAt  *time.Time `gorm:"column:consumed_at"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null"`
}

func (SharedProof) TableName() string {
	return "shared_proofs"
}
