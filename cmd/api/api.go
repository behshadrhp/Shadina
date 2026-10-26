package main

import (
	"net/http"
	"time"

	"github.com/behshadrhp/Shadina/docs"
	"github.com/behshadrhp/Shadina/internal/auth"
	"github.com/behshadrhp/Shadina/internal/consts"
	"github.com/behshadrhp/Shadina/internal/storage"
	"go.uber.org/zap"
)

type application struct {
	config        config
	storage       storage.Storage
	logger        *zap.SugaredLogger
	authenticator auth.Authenticator
}

type config struct {
	addr        string
	url         string
	version     string
	environment string
	db          dbConfig
	auth        authConfig
}

type dbConfig struct {
	addr           string
	maxOpenConns   int
	maxIdleConns   int
	maxIdleTimeout string
}

type authConfig struct {
	token tokenConfig
}

type tokenConfig struct {
	secretKey string
	exp       time.Duration
	iss       string
	aud       string
}

func (app *application) run(mux *http.Handler) error {

	docs.SwaggerInfo.Version = app.config.version
	docs.SwaggerInfo.Host = app.config.url
	docs.SwaggerInfo.BasePath = "/v1/"

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      *mux,
		ReadTimeout:  consts.READ_TIMEOUT,
		WriteTimeout: consts.WRITE_TIMEOUT,
		IdleTimeout:  consts.IDLE_TIMEOUT,
	}

	app.logger.Infow("server stable", "address", app.config.url, "environment", app.config.environment)

	return srv.ListenAndServe()
}
