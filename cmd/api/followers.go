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
