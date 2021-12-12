package main

import (
	"flag"
	"net/http"
	"time"

	authUseCase "auth-service/auth/usecase"

	"auth-service/auth/delivery"

	"github.com/go-chi/chi/v5"

	"auth-service/domain"

	"github.com/BurntSushi/toml"

	"github.com/rs/zerolog/log"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "configs/server.toml", "path to config file")
}

func main() {
	flag.Parse()

	config := domain.NewConfig()

	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot parse config file")
	}

	router := chi.NewRouter()

	authUseCase, err := authUseCase.New(config)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start")
	}

	delivery.NewAuthHandler(router, authUseCase)

	server := &http.Server{
		Addr:         config.BindAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("something went wrong")
	}
}
