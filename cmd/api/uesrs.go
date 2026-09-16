package main

import (
	"net/http"

	"github.com/behshadrhp/Shadina/internal/response"
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
