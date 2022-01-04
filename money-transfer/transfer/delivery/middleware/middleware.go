package delivery

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"money-transfer/domain"

	"github.com/dgrijalva/jwt-go"
	"github.com/rs/zerolog/log"
)

type MiddleWare struct {
	accessSecret    string
	refreshSecret   string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type ctxKey int8

const CtxKeyUser ctxKey = iota

func New(c *domain.Config) *MiddleWare {

	return &MiddleWare{
		accessSecret:    c.AccessTokenSecret,
		refreshSecret:   c.RefreshTokenSecret,
		accessTokenTTL:  c.AccessTokenTTL.Duration,
		refreshTokenTTL: c.RefreshTokenTTL.Duration,
	}
}

func (m *MiddleWare) CheckAuthMiddleware(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {

		// header := r.Header.Get("Authorization")

		// token, err := m.ExtractToken(header)
		// if err != nil {
		// 	log.Debug().Err(err).Msgf("Extract token error: %v", err)
		// 	http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		// 	return
		// }

		c, err := r.Cookie("access_token")

		if err != nil {
			log.Debug().Err(err).Msgf("Extract token error: %v", err)
			http.Redirect(w, r, "/login", http.StatusMovedPermanently)
			return
		}

		user, err := m.ParseToken(c.Value, true)
		if err != nil {
			if err == domain.ErrExpiredToken {
				http.Redirect(w, r, "/update-token", http.StatusMovedPermanently)
				return
			}

			log.Debug().Err(err).Msgf("Parse access token error: %v", err)
			http.Redirect(w, r, "/login", http.StatusMovedPermanently)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CtxKeyUser, user)))
	}
	return http.HandlerFunc(fn)
}

func (m *MiddleWare) ExtractToken(AuthorizationHeader string) (string, error) {
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

func (m *MiddleWare) ParseToken(token string, isAccess bool) (*domain.User, error) {

	JWTToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("failed to extract token metadata, unexpected signing method: %v", token.Header["alg"])
		}
		if isAccess {
			return []byte(m.accessSecret), nil
		}
		return []byte(m.refreshSecret), nil
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

		iin, ok := claims["iin"].(string)
		if !ok {
			return nil, domain.ErrInvalidToken
		}

		expiredTime := time.Unix(int64(exp), 0)

		if time.Now().After(expiredTime) {
			return nil, domain.ErrExpiredToken
		}
		return &domain.User{
			ID:   int64(userID),
			Role: role,
			IIN:  iin,
		}, nil
	}

	return nil, domain.ErrInvalidToken
}
