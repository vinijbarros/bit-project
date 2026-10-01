package domain

import "time"

type User struct {
	ID          int64
	Username    string
	DisplayName string
}

type UserCredentials struct {
	User
	PasswordHash string
}

type Session struct {
	TokenHash string
	UserID    int64
	CreatedAt time.Time
	ExpiresAt time.Time
}
