package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) mount() http.Handler {
	router := chi.NewRouter()

	// middleware
	router.Use(middleware.Heartbeat("/ping"))
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// handlers
	router.Route("/v1", func(r chi.Router) {

		// public routes
		r.Get("/health", app.healthCheckHandler)
	})

	return router
}
