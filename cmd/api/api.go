package main

import (
	"log"
	"net/http"

	"github.com/behshadrhp/Shadina/internal/consts"
)

type application struct {
	config config
}

type config struct {
	addr string
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
