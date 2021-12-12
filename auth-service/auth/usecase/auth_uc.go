package authUseCase

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"auth-service/auth/repository/pg"
	rd "auth-service/auth/repository/redis"

	"auth-service/domain"

	"github.com/dgrijalva/jwt-go"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type authUseCase struct {
	accessSecret  string
	refreshSecret string

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration

	cache domain.CaсheStore
	db    domain.Repository
}

func New(c *domain.Config) (domain.AuthUseCase, error) {

	r, err := rd.NewRedisClient(c)
	if err != nil {
		return nil, err
	}

	d, err := pg.NewSQLRepository(c)
	if err != nil {
		return nil, err
	}

	return &authUseCase{
		accessSecret:  c.AccessTokenSecret,
		refreshSecret: c.RefreshTokenSecret,

		accessTokenTTL:  c.AccessTokenTTL.Duration,
		refreshTokenTTL: c.RefreshTokenTTL.Duration,
		cache:           r,
		db:              d,
	}, nil
}

func (a *authUseCase) GetUserData(ctx context.Context, ID int64) (*domain.User, error) {

	u, err := a.db.GetUserData(ctx, ID)
	if err != nil {
		return nil, err
	}

	http.Get("localhost:8080/accounts")

	if u == nil {
		return nil, domain.ErrNotFound
	}

	return u, nil
}

func (a *authUseCase) FindUser(ctx context.Context, email, password string) (*domain.User, error) {

	u, err := a.db.FindUser(ctx, email)
	if err != nil {
		return nil, err
	}

	if u == nil {
		return nil, domain.ErrNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		log.Warn().Err(err).Msgf("incorrect password for email: %s", email)
		return nil, domain.ErrNotFound
	}

	return u, nil
}

func (a *authUseCase) CreateUser(ctx context.Context, u *domain.User) error {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Warn().Err(err).Msgf("incorrect encryption for email: %s", u.Email)
		return domain.ErrNotCreated
	}

	u.Password = string(hashedPassword)

	return a.db.CreateUser(ctx, u)
}

func (a *authUseCase) Close() {
	a.db.CloseConnection()
}

func (a *authUseCase) ExtractToken(AuthorizationHeader string) (string, error) {
	// header := r.Header.Get("Authorization")
	if AuthorizationHeader == "" {
		return "", domain.ErrNoHeader
	}

	parsedHeader := strings.Split(AuthorizationHeader, " ")
	if len(parsedHeader) != 2 || parsedHeader[0] != "Bearer" {
		return "", domain.ErrInvalidHeader
	}

	return parsedHeader[1], nil
}

func (a *authUseCase) ParseToken(token string, isAccess bool) (*domain.User, error) {

	JWTToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("failed to extract token metadata, unexpected signing method: %v", token.Header["alg"])
		}
		if isAccess {
			return []byte(a.accessSecret), nil
		}
		return []byte(a.refreshSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := JWTToken.Claims.(jwt.MapClaims)

	if ok && JWTToken.Valid {

		var userID float64

		userID, ok = claims["id"].(float64)
		if !ok {
			return nil, domain.ErrInvalidToken
		}

		exp, ok := claims["exp"].(float64)
		if !ok {
			return nil, domain.ErrInvalidToken
		}

		role, ok := claims["role"].(string)
		if !ok || (role != "user" && role != "admin") {
			return nil, domain.ErrInvalidToken
		}

		expiredTime := time.Unix(int64(exp), 0)

		if time.Now().After(expiredTime) {
			return nil, domain.ErrExpiredToken
		}
		return &domain.User{
			ID:   int64(userID),
			Role: role,
		}, nil
	}

	return nil, domain.ErrInvalidToken
}

// GenerateAndSendTokens return access token and refresh token in that order.
func (a *authUseCase) GenerateAndSendTokens(u *domain.User) (string, string, error) {

	accessTokenExp := time.Now().Add(a.accessTokenTTL).Unix()

	accessTokenClaims := jwt.MapClaims{}
	accessTokenClaims["id"] = u.ID
	accessTokenClaims["iat"] = time.Now().Unix()
	accessTokenClaims["exp"] = accessTokenExp
	accessTokenClaims["role"] = u.Role
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)

	accessSignedToken, err := accessToken.SignedString([]byte(a.accessSecret))
	if err != nil {
		return "", "", domain.ErrTokenNotCreated
	}

	refreshTokenExp := time.Now().Add(a.refreshTokenTTL).Unix()
	refreshTokenClaims := jwt.MapClaims{}
	refreshTokenClaims["id"] = u.ID
	refreshTokenClaims["iat"] = time.Now().Unix()
	refreshTokenClaims["exp"] = refreshTokenExp
	refreshTokenClaims["role"] = u.Role
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)

	refreshSignedToken, err := refreshToken.SignedString([]byte(a.refreshSecret))
	if err != nil {
		return "", "", domain.ErrTokenNotCreated
	}

	if err := a.cache.InsertToken(int64(u.ID), refreshSignedToken); err != nil {
		return "", "", domain.ErrTokenNotCreated
	}
	return accessSignedToken, refreshSignedToken, nil
}

// UpdateToken returns access token and refresh token (in that order).
func (a *authUseCase) UpdateToken(refreshToken string) (string, string, error) {

	//checking update token
	user, err := a.ParseToken(refreshToken, false)
	if err != nil {
		return "", "", err
	}

	// find token in Redis
	ok := a.cache.FindToken(int64(user.ID), refreshToken)
	if !ok {
		return "", "", domain.ErrInvalidToken
	}

	return a.GenerateAndSendTokens(user)
}

func getRandomSecret() string {
	var symbols = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789")
	rand.Seed(time.Now().UnixNano())

	str := make([]rune, 32)

	for i := range str {
		str[i] = symbols[rand.Intn(len(symbols))]
	}
	return string(str)
}

func (a *authUseCase) GetAccessTokenTTL() time.Duration {
	return a.accessTokenTTL
}

func (a *authUseCase) GetRefreshTokenTTL() time.Duration {
	return a.refreshTokenTTL
}
