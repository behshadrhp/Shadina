package main

import (
	"net/http"

	"github.com/behshadrhp/Shadina/internal/dto"
	"github.com/behshadrhp/Shadina/internal/response"
	"github.com/behshadrhp/Shadina/internal/storage"
	"github.com/behshadrhp/Shadina/internal/utils"
)

func (app *application) userRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	var payload dto.UserRegisterPayload

	ctx := r.Context()

	if err := utils.ReadJson(w, r, &payload); err != nil {
		response.BadRequestResponse(w, "invalid request", 4011)
		return
	}

	if err := utils.Validate.Struct(&payload); err != nil {
		response.BadRequestResponse(w, "invalid request", 4012)
		return
	}

	// create user
	user := &storage.User{
		Email:    payload.Email,
		Username: payload.Username,
	}

	// hashed password
	if err := user.HashedPassword(payload.Password); err != nil {
		response.InternalErrorResponse(w, "failed to hash password", 5018)
		return
	}

	// create user
	if err := app.storage.Users.Create(ctx, user); err != nil {
		response.InternalErrorResponse(w, err.Error(), 5018)
		return
	}

	response.SuccessResponse(w, "user registered successfully", 1010, user)
}
