package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"money-transfer/domain"
	"money-transfer/transfer/delivery"

	"github.com/BurntSushi/toml"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "configs/server.toml", "path to config file")
}

func main() {
	flag.Parse()

	// getting server configurations
	config := domain.NewConfig()

	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot parse config file")
	}

	// setting up bussiness logic
	// usecase, err := transferUseCase.New(config)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("cannot start app")
	// }

	// connecting delivery layer
	router := chi.NewRouter()

	if err := delivery.NewTransactionHandler(config, router, nil); err != nil {
		log.Fatal().Err(err).Msg("cannot start app")
	}

	// server configurations
	server := &http.Server{
		Addr:         config.BindAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	log.Info().Msg("databases connected")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("something went wrong")
	}
}
