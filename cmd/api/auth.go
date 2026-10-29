package main

import (
	"net/http"
	"time"

	"github.com/behshadrhp/Shadina/internal/consts"
	"github.com/behshadrhp/Shadina/internal/dto"
	"github.com/behshadrhp/Shadina/internal/response"
	"github.com/behshadrhp/Shadina/internal/storage"
	"github.com/behshadrhp/Shadina/internal/utils"
	"github.com/golang-jwt/jwt/v5"
)

func (app *application) userRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	var payload dto.UserRegisterPayload

	ctx := r.Context()

	if err := utils.ReadJson(w, r, &payload); err != nil {
		response.BadRequestResponse(w, "invalid request", 4011)
		return
	}

	if err := utils.Validate.Struct(&payload); err != nil {
		response.BadRequestResponse(w, err.Error(), 4012)
		return
	}

	// create user
	user := &storage.User{
		Email:    payload.Email,
		Username: payload.Username,
		RoleID:   1,
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

// generateTokenHandler godoc
//
//	@Summary		Generate an access token
//	@Description	Authenticates a user using email and password and returns a signed JWT access token.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		dto.GenerateTokenPayload	true	"Login credentials"
//	@Success		200		{object}	object						"Token generated successfully"
//	@Failure		400		{object}	object						"Invalid request or credentials"
//	@Failure		500		{object}	object						"Token generation failed"
//	@Router			/auth/token [post]
func (app *application) generateTokenHandler(w http.ResponseWriter, r *http.Request) {
	var payload dto.GenerateTokenPayload

	ctx := r.Context()

	if err := utils.ReadJson(w, r, &payload); err != nil {
		response.BadRequestResponse(w, "invalid request", 4013)
		return
	}

	if err := utils.Validate.Struct(&payload); err != nil {
		response.BadRequestResponse(w, "invalid request", 4014)
		return
	}

	// fetch user
	user, err := app.storage.Users.GetByEmail(ctx, payload.Email)
	if err != nil {
		switch err {
		case consts.NOT_FOUND:
			response.BadRequestResponse(w, "invalid request", 4015)
			return
		default:
			response.BadRequestResponse(w, "invalid request", 4016)
			return
		}
	}

	// check password
	if ok := user.ComparePassword(payload.Password); !ok {
		response.BadRequestResponse(w, "invalid email or password", 4017)
		return
	}

	// generate token
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.token.iss,
		"aud": app.config.auth.token.aud,
	}

	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		response.BadRequestResponse(w, "invalid request", 4018)
		return
	}

	response.SuccessResponse(w, "generate token successfully", 1011, token)
}
