package main

import (
	"net/http"
	"strconv"

	"github.com/behshadrhp/Shadina/internal/response"
	"github.com/go-chi/chi/v5"
)

func (app *application) listUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := app.storage.Users.List(ctx)
	if err != nil {
		response.InternalErrorResponse(w, err.Error(), 5009)
		return
	}

	response.SuccessResponse(w, "users retrieved successfully", 1005, users)
}

func (app *application) retrieveUsersHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		response.InternalErrorResponse(w, err.Error(), 5010)
		return
	}

	ctx := r.Context()

	users, err := app.storage.Users.GetByID(ctx, userID)
	if err != nil {
		response.InternalErrorResponse(w, err.Error(), 5010)
		return
	}

	response.SuccessResponse(w, "user retrieved successfully", 1006, users)
}
