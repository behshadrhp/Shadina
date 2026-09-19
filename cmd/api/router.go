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

		r.Route("/users", func(r chi.Router) {
			// crud routes
			r.Get("/", app.listUsersHandler)

			// retrieve
			r.Route("/{userID}", func(r chi.Router) {
				r.Get("/", app.retrieveUsersHandler)

				// followers
				r.Put("/follow", app.UserFollowHandler)
				r.Put("/unfollow", app.UserUnfollowHandler)

				// feeds
				r.Get("/feeds", app.listFeedsHandler)
			})
		})

		r.Route("/posts", func(r chi.Router) {

			// crud routes
			r.Get("/", app.listPostsHandler)
			r.Post("/", app.createPostHandler)

			// retrieve
			r.Route("/{postID}", func(r chi.Router) {
				r.Get("/", app.retrievePostHandler)
				r.Delete("/", app.deletePostHandler)
				r.Patch("/", app.updatePostHandler)
			})
		})

		// public routes
		r.Get("/health", app.healthCheckHandler)
	})

	return router
}
