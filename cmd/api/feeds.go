package main

import (
	"net/http"
	"strconv"

	"github.com/behshadrhp/Shadina/internal/response"
	"github.com/behshadrhp/Shadina/internal/storage"
	"github.com/behshadrhp/Shadina/internal/utils"
	"github.com/go-chi/chi/v5"
)

// listFeedsHandler godoc
//
//	@Summary		User Feed
//	@Description	Retrieve a paginated feed for a user based on followed users and feed rules.
//	@Tags			Feeds
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int					true	"User ID"
//	@Param			page	query		int					false	"Page number"		default(1)
//	@Param			limit	query		int					false	"Items per page"	default(20)
//	@Param			sort	query		string				false	"Sort order"		Enums(asc,desc)	default(desc)
//	@Success		200		{object}	response.Response	"Feed retrieved successfully"
//	@Failure		400		{object}	response.Response	"Invalid request"
//	@Failure		500		{object}	response.Response	"Internal server error"
//	@Router			/users/{userID}/feed [get]
func (app *application) listFeedsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		response.InternalErrorResponse(w, err.Error(), 5016)
		return
	}

	// pagination & search & filter
	pq := storage.CreatePaginatedQuery(20, 10, "desc")
	pq, err = pq.Parse(r)
	if err != nil {
		response.BadRequestResponse(w, err.Error(), 4009)
		return
	}

	if err := utils.Validate.Struct(pq); err != nil {
		response.BadRequestResponse(w, err.Error(), 4010)
		return
	}

	ctx := r.Context()

	feeds, err := app.storage.Posts.GetFeed(ctx, userID, pq)
	if err != nil {
		response.InternalErrorResponse(w, err.Error(), 5017)
		return
	}

	response.SuccessResponse(w, "feeds retrieved successfully", 1009, feeds)
}
