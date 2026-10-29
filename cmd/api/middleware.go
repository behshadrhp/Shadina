package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/behshadrhp/Shadina/internal/consts"
	"github.com/behshadrhp/Shadina/internal/response"
	"github.com/behshadrhp/Shadina/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
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

func (app *application) JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			response.UnauthorizedResponse(w, "authorization header is missing", 4019)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.UnauthorizedResponse(w, "authorization header is malformed", 4020)
			return
		}

		token := parts[1]
		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			response.InternalErrorResponse(w, err.Error(), 5019)
			return
		}

		claims, _ := jwtToken.Claims.(jwt.MapClaims)

		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			response.UnauthorizedResponse(w, err.Error(), 4021)
			return
		}

		ctx := r.Context()

		user, err := app.storage.Users.GetByID(ctx, userID)
		if err != nil {
			response.UnauthorizedResponse(w, err.Error(), 4022)
			return
		}

		ctx = context.WithValue(ctx, consts.USER_CTX, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app application) GetUserCTX(r *http.Request) *storage.User {
	user, _ := r.Context().Value("user").(*storage.User)

	return user
}

func (app *application) UserPostAccessMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := app.GetUserCTX(r)

		paramID := chi.URLParam(r, "postID")
		postID, err := strconv.ParseInt(paramID, 10, 64)
		if err != nil {
			response.InternalErrorResponse(w, err.Error(), 5004)
			return
		}

		ctx := r.Context()

		post, err := app.storage.Posts.GetByID(ctx, postID)
		if err != nil {
			switch {
			case errors.Is(err, consts.NOT_FOUND):
				response.NotFoundResponse(w, "post not found", 4001)
			default:
				response.InternalErrorResponse(w, err.Error(), 5005)
			}
			return
		}

		if user.ID != post.UserID {

			if user.RoleID == 3 {
				next.ServeHTTP(w, r)
			}

			response.UnauthorizedResponse(w, "access denied", 4023)
			return
		}

		next.ServeHTTP(w, r)
	})
}
