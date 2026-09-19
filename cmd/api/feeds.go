package main

import (
	"net/http"
	"strconv"

	"github.com/behshadrhp/Shadina/internal/response"
	"github.com/go-chi/chi/v5"
)

func (app *application) listFeedsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		response.InternalErrorResponse(w, err.Error(), 5016)
		return
	}

	ctx := r.Context()

	feeds, err := app.storage.Posts.GetFeed(ctx, userID)
	if err != nil {
		response.InternalErrorResponse(w, err.Error(), 5017)
		return
	}

	response.SuccessResponse(w, "feeds retrieved successfully", 1009, feeds)
}
