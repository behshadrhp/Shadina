package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/behshadrhp/Shadina/internal/consts"
	"github.com/behshadrhp/Shadina/internal/dto"
	"github.com/behshadrhp/Shadina/internal/response"
	"github.com/behshadrhp/Shadina/internal/utils"
	"github.com/go-chi/chi/v5"
)

// UserFollowHandler godoc
//
//	@Summary		Follow User
//	@Description	Create a follower relationship between two users.
//	@Tags			Followers
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int					true	"User ID"
//	@Param			request	body		dto.FollowPayload	true	"Follow Payload"
//	@Success		200		{object}	response.Response	"User followed successfully"
//	@Failure		400		{object}	response.Response	"Invalid request"
//	@Failure		404		{object}	response.Response	"User not found"
//	@Failure		500		{object}	response.Response	"Internal server error"
//	@Router			/users/{userID}/follow [post]
func (app *application) UserFollowHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		response.InternalErrorResponse(w, err.Error(), 5011)
		return
	}

	var payload dto.FollowPayload
	if err := utils.ReadJson(w, r, &payload); err != nil {
		response.BadRequestResponse(w, err.Error(), 4005)
		return
	}

	ctx := r.Context()

	err = app.storage.Followers.Follow(ctx, userID, payload.FollowerID)
	if err != nil {
		switch {
		case errors.Is(err, consts.NOT_FOUND):
			response.NotFoundResponse(w, "not found", 4006)
			return
		default:
			response.InternalErrorResponse(w, err.Error(), 5013)
			return
		}
	}

	response.SuccessResponse(w, "user followed successfully", 1007, nil)
}

// UserUnfollowHandler godoc
//
//	@Summary		Unfollow User
//	@Description	Remove a follower relationship between two users.
//	@Tags			Followers
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int					true	"User ID"
//	@Param			request	body		dto.UnFollowPayload	true	"Unfollow Payload"
//	@Success		200		{object}	response.Response	"User unfollowed successfully"
//	@Failure		400		{object}	response.Response	"Invalid request"
//	@Failure		404		{object}	response.Response	"User not found"
//	@Failure		500		{object}	response.Response	"Internal server error"
//	@Router			/users/{userID}/unfollow [delete]
func (app *application) UserUnfollowHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		response.InternalErrorResponse(w, err.Error(), 5014)
		return
	}

	var payload dto.UnFollowPayload
	if err := utils.ReadJson(w, r, &payload); err != nil {
		response.BadRequestResponse(w, err.Error(), 4007)
		return
	}

	ctx := r.Context()

	err = app.storage.Followers.Unfollow(ctx, userID, payload.FollowerID)
	if err != nil {
		switch {
		case errors.Is(err, consts.NOT_FOUND):
			response.NotFoundResponse(w, "not found", 4008)
			return
		default:
			response.InternalErrorResponse(w, err.Error(), 5015)
			return
		}
	}

	response.SuccessResponse(w, "user unfollowed successfully", 1008, nil)
}
