package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"auth-service/domain"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
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

func (db *sqlRepository) CreateUser(ctx context.Context, u *domain.User) error {

	_, err := db.Exec(ctx,
		`INSERT INTO users(Email, FirstName, LastName, Password, IIN, Phone, Registered, Role) 
		VALUES($1, $2, $3, $4, $5, $6, $7, $8)`,
		u.Email, u.FirstName, u.LastName, u.Password, u.IIN, u.Phone, u.Registered, u.Role,
	)

	var pgErr *pgconn.PgError
	if err != nil && errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrExists
		}
	}
	return err
}

func (db *sqlRepository) FindUser(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}

	err := db.QueryRow(ctx,
		`SELECT ID, FirstName, LastName, Password, Role, IIN
		FROM users 
		WHERE Email=$1`,
		email,
	).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Password, &user.Role, &user.IIN)

	if err == pgx.ErrNoRows {
		return nil, err
	}

	return user, err
}

func (db *sqlRepository) GetUserData(ctx context.Context, ID int64) (*domain.User, error) {
	user := &domain.User{}

	err := db.QueryRow(ctx,
		`SELECT ID, Email, IIN, Registered, Role
		FROM users 
		WHERE ID=$1`,
		ID,
	).Scan(&user.ID, &user.Email, &user.IIN, &user.Registered, &user.Role)

	if err == pgx.ErrNoRows {
		return nil, err
	}

	return user, err
}
