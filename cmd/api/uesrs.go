package main

import (
	"net/http"
	"strconv"

	"github.com/behshadrhp/Shadina/internal/response"
	"github.com/go-chi/chi/v5"
)

// listUsersHandler godoc
//
//	@Summary		List Users
//	@Description	Retrieve all users from the system.
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{data=[]storage.User}	"Users retrieved successfully"
//	@Failure		500	{object}	response.Response						"Internal server error"
//	@Router			/users [get]
func (app *application) listUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := app.storage.Users.List(ctx)
	if err != nil {
		response.InternalErrorResponse(w, err.Error(), 5009)
		return
	}

	response.SuccessResponse(w, "users retrieved successfully", 1005, users)
}

// retrieveUsersHandler godoc
//
//	@Summary		Retrieve User
//	@Description	Retrieve a user by their unique identifier.
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int										true	"User ID"
//	@Success		200		{object}	response.Response{data=storage.User}	"User retrieved successfully"
//	@Failure		400		{object}	response.Response						"Invalid user ID"
//	@Failure		404		{object}	response.Response						"User not found"
//	@Failure		500		{object}	response.Response						"Internal server error"
//	@Router			/users/{userID} [get]
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
