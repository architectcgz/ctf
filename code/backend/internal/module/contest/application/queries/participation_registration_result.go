package queries

import "time"

type ContestRegistrationQueryInput struct {
	Status *string
	Page   int
	Size   int
}

type RegistrationPageResult[T any] struct {
	List  []T
	Total int64
	Page  int
	Size  int
}

type ContestRegistrationResult struct {
	ID         int64
	ContestID  int64
	UserID     int64
	Username   string
	TeamID     *int64
	Status     string
	ReviewedBy *int64
	ReviewedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
