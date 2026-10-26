package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/behshadrhp/Shadina/internal/consts"
	"github.com/behshadrhp/Shadina/internal/response"
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
	user := app.GetUserCTX(r)

	followedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		response.InternalErrorResponse(w, err.Error(), 5020)
		return
	}

	ctx := r.Context()

	if err := app.storage.Followers.Follow(ctx, user.ID, followedID); err != nil {
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
	user := app.GetUserCTX(r)

	unfollowedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		response.InternalErrorResponse(w, err.Error(), 5021)
		return
	}

	ctx := r.Context()

	if err := app.storage.Followers.Unfollow(ctx, user.ID, unfollowedID); err != nil {
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
