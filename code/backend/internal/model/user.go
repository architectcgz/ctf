package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusLocked   = "locked"
	UserStatusBanned   = "banned"
)

type User struct {
	ID           int64          `gorm:"column:id;primaryKey"`
	Username     string         `gorm:"column:username"`
	Name         string         `gorm:"column:name"`
	PasswordHash string         `gorm:"column:password_hash"`
	Email        string         `gorm:"column:email"`
	StudentNo    string         `gorm:"column:student_no"`
	TeacherNo    string         `gorm:"column:teacher_no"`
	Role         string         `gorm:"column:role"`
	ClassName    string         `gorm:"column:class_name"`
	Status       string         `gorm:"column:status"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

type UserRole struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	UserID    int64     `gorm:"column:user_id"`
	RoleID    int64     `gorm:"column:role_id"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (UserRole) TableName() string {
	return "user_roles"
}
