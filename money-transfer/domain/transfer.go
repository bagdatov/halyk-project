package domain

import (
	"context"
)

type Transfer interface {
	CreateAccount(ctx context.Context, account *Account) error
	FindAccount(ctx context.Context, ID int64) (*Account, error)
	GetAccounts(ctx context.Context, OwnerID int64) ([]*Account, error)
	ChangeAccountSum(ctx context.Context, accountID, newValue int64) error
	CreateTransaction(ctx context.Context, requester, SenderID, ReceiverID, Value int64) error
	Close()
}

// Repository is representing database.
// At this moment it is PostgreSQL, but can be updated anytime,
// so bussiness logic will not be affected.
type Repository interface {
	CreateAccount(ctx context.Context, account *Account) error
	FindAccount(ctx context.Context, ID int64) (*Account, error)
	GetAccounts(ctx context.Context, OwnerID int64) ([]*Account, error)
	GetLastTransaction(ctx context.Context, accountID int64) (*Transaction, error)
	ChangeAccountSum(ctx context.Context, accountID, newValue int64) error
	CreateTransaction(ctx context.Context, SenderID, ReceiverID, Value int64) error
	AccountExists(ctx context.Context, accountID int64) bool
	CloseConnection()
}
