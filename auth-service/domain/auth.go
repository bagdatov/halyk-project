package domain

import (
	"context"
	"time"
)

// AuthUseCase is bussiness logic over our auth service.
type AuthUseCase interface {
	// Login(username, password string) (*User, error)
	GenerateAndSendTokens(u *User) (string, string, error)
	ExtractToken(authorizationHeader string) (string, error)
	ParseToken(token string, isAccess bool) (*User, error)
	UpdateToken(refreshToken string) (string, string, error)

	FindUser(ctx context.Context, email, password string) (*User, error)
	CreateUser(ctx context.Context, u *User) error
	GetUserData(ctx context.Context, ID int64) (*User, error)

	GetAccessTokenTTL() time.Duration
	GetRefreshTokenTTL() time.Duration

	Close()
}

// CaсheStore is representing NoSQL database for cache.
// In our case, it is used to keep refresh tokens.
// At this moment it is Redis, but can be updated anytime,
// so bussiness logic will not be affected.
type CaсheStore interface {
	FindToken(id int64, token string) bool
	InsertToken(id int64, token string) error
}

// Repository is representing database.
// At this moment it is PostgreSQL, but can be updated anytime,
// so bussiness logic will not be affected.
type Repository interface {
	FindUser(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, u *User) error
	GetUserData(ctx context.Context, ID int64) (*User, error)
	CloseConnection()
}
