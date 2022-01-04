package transferUseCase

import (
	"context"
	"math/rand"
	"time"

	"money-transfer/domain"
	"money-transfer/transfer/repository/pg"
)

const (
	Min int64 = 1000000000000000
	Max int64 = 9999999999999999
)

type transferUseCase struct {
	db domain.Repository
}

func New(c *domain.Config) (domain.Transfer, error) {
	repo, err := pg.NewSQLRepository(c)
	if err != nil {
		return nil, err
	}

	return &transferUseCase{
		db: repo,
	}, nil
}

func (tu *transferUseCase) CreateAccount(ctx context.Context, account *domain.Account) error {
	rand.Seed(time.Now().UnixNano())

	for {
		randomID := rand.Int63n(Max-Min+1) + Min

		if !tu.db.AccountExists(ctx, randomID) {
			account.ID = randomID
			break
		}
	}

	return tu.db.CreateAccount(ctx, account)
}

func (tu *transferUseCase) FindAccount(ctx context.Context, ID int64) (*domain.Account, error) {
	return tu.db.FindAccount(ctx, ID)
}

func (tu *transferUseCase) GetAccounts(ctx context.Context, OwnerID int64) ([]*domain.Account, error) {
	return tu.db.GetAccounts(ctx, OwnerID)
}

func (tu *transferUseCase) ChangeAccountSum(ctx context.Context, accountID, newValue int64) error {
	if newValue <= 150 {
		return domain.ErrTransSum
	}

	return tu.db.ChangeAccountSum(ctx, accountID, newValue)
}

func (tu *transferUseCase) CreateTransaction(ctx context.Context, requester, SenderID, ReceiverID, Value int64) error {
	account, err := tu.db.FindAccount(ctx, SenderID)
	if err != nil || account.OwnerID != requester {
		return domain.ErrTransSender
	}

	if SenderID == ReceiverID {
		return domain.ErrTransReceiver
	}

	if Value <= 150 {
		return domain.ErrTransSum
	}

	return tu.db.CreateTransaction(ctx, SenderID, ReceiverID, Value)
}

func (tu *transferUseCase) Close() {
	tu.db.CloseConnection()
}
