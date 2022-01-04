package pg

import (
	"context"
	"fmt"
	"time"

	"money-transfer/domain"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type sqlRepository struct {
	*pgxpool.Pool
}

func NewSQLRepository(c *domain.Config) (domain.Repository, error) {

	DSN := fmt.Sprintf("postgres://%s:%s@%s%s/%s", c.UserDB, c.PasswordDB, c.HostDB, c.PortDB, c.NameDB)

	config, err := pgxpool.ParseConfig(DSN)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 25
	config.MaxConnLifetime = 5 * time.Minute

	db, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, err
	}
	return &sqlRepository{db}, nil
}

func (db *sqlRepository) CloseConnection() {
	db.Close()
}

func (db *sqlRepository) CreateAccount(ctx context.Context, account *domain.Account) error {
	_, err := db.Exec(ctx, `
	INSERT INTO accounts(ID, OwnerID, IIN, Amount, Registered) 
	VALUES ($1,$2,$3,$4,$5)`,
		account.ID, account.OwnerID, account.IIN, 0, account.Registered,
	)
	return err
}

func (db *sqlRepository) FindAccount(ctx context.Context, ID int64) (*domain.Account, error) {
	acc := &domain.Account{}
	err := db.QueryRow(ctx,
		`SELECT ID, OwnerID, IIN, Amount, Registered FROM accounts WHERE ID=$1`,
		ID).Scan(&acc.ID, &acc.OwnerID, &acc.IIN, &acc.Amount, &acc.Registered)
	return acc, err
}

func (db *sqlRepository) GetAccounts(ctx context.Context, OwnerID int64) ([]*domain.Account, error) {
	accounts := make([]*domain.Account, 0)
	rows, err := db.Query(ctx, `SELECT ID, Amount, Registered FROM accounts WHERE OwnerID = $1`, OwnerID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		a := &domain.Account{}

		err = rows.Scan(&a.ID, &a.Amount, &a.Registered)
		if err != nil {
			return nil, err
		}

		t, err := db.GetLastTransaction(ctx, a.ID)
		if err != nil {
			return nil, err
		}

		if t != nil {
			a.LastTransaction = *t
		}

		accounts = append(accounts, a)
	}

	return accounts, rows.Err()
}

func (db *sqlRepository) GetLastTransaction(ctx context.Context, accountID int64) (*domain.Transaction, error) {
	t := &domain.Transaction{}

	err := db.QueryRow(ctx,
		`SELECT ID, SenderID, ReceiverID, Amount, Date 
		FROM transactions 
		WHERE SenderID = $1 
		OR ReceiverID = $2 
		ORDER BY Date DESC`,
		accountID, accountID,
	).Scan(&t.ID, &t.SenderID, &t.ReceiverID, &t.Amount, &t.Date)

	if err != nil {
		if err == pgx.ErrNoRows {
			err = nil
		}
		return nil, err
	}

	return t, nil
}

func (db *sqlRepository) ChangeAccountSum(ctx context.Context, accountID, newValue int64) error {
	_, err := db.Exec(ctx, `UPDATE accounts SET Amount = Amount + $1 WHERE ID = $2`, newValue, accountID)
	return err
}

func (db *sqlRepository) CreateTransaction(ctx context.Context, SenderID, ReceiverID, Value int64) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	_, err = tx.Exec(ctx, `UPDATE accounts SET Amount = Amount - $1 WHERE ID = $2`, Value, SenderID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `UPDATE accounts SET Amount = Amount + $1 WHERE ID = $2`, Value, ReceiverID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `INSERT INTO transactions(SenderID, ReceiverID, Amount, Date) VALUES($1, $2, $3, $4)`, SenderID, ReceiverID, Value, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func (db *sqlRepository) AccountExists(ctx context.Context, accountID int64) bool {
	var id int64
	err := db.QueryRow(ctx, `SELECT OwnerID FROM accounts WHERE ID = $1`, accountID).Scan(&id)
	if err != nil && err == pgx.ErrNoRows {
		return false
	}
	return true
}
