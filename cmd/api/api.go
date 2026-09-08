package main

import (
	"log"
	"net/http"

	"github.com/behshadrhp/Shadina/internal/consts"
	"github.com/behshadrhp/Shadina/internal/storage"
)

type application struct {
	config  config
	storage storage.Storage
}

type config struct {
	addr        string
	version     string
	environment string
	db          dbConfig
}

type dbConfig struct {
	addr           string
	maxOpenConns   int
	maxIdleConns   int
	maxIdleTimeout string
}

func (app *application) run(mux *http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      *mux,
		ReadTimeout:  consts.READ_TIMEOUT,
		WriteTimeout: consts.WRITE_TIMEOUT,
		IdleTimeout:  consts.IDLE_TIMEOUT,
	}

	log.Printf("run server at http://127.0.0.1%v", app.config.addr)

	return srv.ListenAndServe()
}
