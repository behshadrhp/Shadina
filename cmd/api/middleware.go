package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		start := time.Now()

		next.ServeHTTP(ww, r)

		app.logger.Infow(
			"http request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"status_code", ww.Status(),
			"bytes", ww.BytesWritten(),
			"duration", time.Since(start),
		)
	})
}
