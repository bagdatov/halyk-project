package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"net/http"

	"auth-service/domain"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type token struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type AuthHanlder struct {
	au domain.AuthUseCase
}

type ctxKey int8

const CtxKeyUser ctxKey = iota

func NewAuthHandler(router *chi.Mux, au domain.AuthUseCase) {
	handler := &AuthHanlder{
		au: au,
	}

	router.Post("/signup", handler.SignUpHanlder())
	router.Post("/login", handler.LoginHandler())
	router.Post("/update-token", handler.UpdateTokenHanlder())

	router.With(handler.CheckAuthMiddleware).Post("/user-data", handler.UserDataHandler())

}

func (s *AuthHanlder) CheckAuthMiddleware(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {

		header := r.Header.Get("Authorization")

		token, err := s.au.ExtractToken(header)
		if err != nil {
			log.Debug().Err(err).Msgf("Extract token error: %v", err)
			http.Redirect(w, r, "/login", http.StatusMovedPermanently)
			return
		}

		user, err := s.au.ParseToken(token, true)
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

func (s *AuthHanlder) SignUpHanlder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		user := &domain.User{
			Email:      r.FormValue("Email"),
			Password:   r.FormValue("Password"),
			FirstName:  r.FormValue("FirstName"),
			LastName:   r.FormValue("LastName"),
			IIN:        r.FormValue("IIN"),
			Phone:      r.FormValue("Phone"),
			Role:       "user",
			Registered: time.Now(),
		}

		if !user.Valid() {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid credentials"))
			return
		}

		if err := s.au.CreateUser(r.Context(), user); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Could not create user"))
			log.Info().Err(err).Msg("Invalid user")
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Successfully registered user"))
	}
}

func (s *AuthHanlder) LoginHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		email, password := extractCredentials(r)

		u, err := s.au.FindUser(r.Context(), email, password)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Invalid credentials"))
			return
		}

		accessToken, refreshToken, err := s.au.GenerateAndSendTokens(u)
		if err != nil {

			log.Debug().Err(err).Msgf("GenerateAndSendTokens: %v", err)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error proceeding tokens"))
			return
		}

		t := &token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}

		reply, err := json.Marshal(t)
		if err != nil {
			log.Debug().Err(err).Msgf("Login: Marshal token: %v", err)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error proceeding tokens"))
			return
		}

		expiration := time.Now().Add(s.au.GetRefreshTokenTTL())
		cookie := http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			Expires:  expiration,
			HttpOnly: true,
		}

		http.SetCookie(w, &cookie)

		expiration = time.Now().Add(s.au.GetAccessTokenTTL())
		cookie = http.Cookie{
			Name:     "access_token",
			Value:    accessToken,
			Expires:  expiration,
			HttpOnly: true,
		}

		http.SetCookie(w, &cookie)

		w.Header().Set("Content-Type", "application/json")
		w.Write(reply)
	}
}

func (s *AuthHanlder) UpdateTokenHanlder() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		c, err := r.Cookie("refresh_token")
		if err == http.ErrNoCookie {
			http.Redirect(w, r, "/login", http.StatusMovedPermanently)
			return
		}

		accessToken, refreshToken, err := s.au.UpdateToken(c.Value)

		if err != nil {
			log.Debug().Err(err).Msgf("GenerateAndSendTokens: %v", err)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error proceeding tokens"))
			return
		}

		t := &token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}

		reply, err := json.Marshal(t)
		if err != nil {
			log.Debug().Err(err).Msgf("Update: Marshal token: %v", err)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error proceeding tokens"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(reply)
	}
}

func (s *AuthHanlder) UserDataHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		u, ok := r.Context().Value(CtxKeyUser).(*domain.User)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userData, err := s.au.GetUserData(r.Context(), u.ID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("No user found"))
			return
		}

		req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/accounts", nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		c, _ := r.Cookie("access_token")
		cookie := fmt.Sprintf("access_token=%s", c.Value)

		req.Header.Set("Cookie", cookie)

		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer resp.Body.Close()

		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		reply, err := json.Marshal(userData)
		if err != nil {
			log.Debug().Err(err).Msgf("UserData: Marshal user: %v", err)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error proceeding data"))
			return
		}
		reply = append(reply, bytes...)

		w.Header().Set("Content-Type", "application/json")
		w.Write(reply)
	}
}

func extractCredentials(r *http.Request) (login, pass string) {
	return string(r.FormValue("Email")), string(r.FormValue("Password"))
}
