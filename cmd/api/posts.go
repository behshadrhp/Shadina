package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/behshadrhp/Shadina/internal/consts"
	"github.com/behshadrhp/Shadina/internal/dto"
	"github.com/behshadrhp/Shadina/internal/response"
	"github.com/behshadrhp/Shadina/internal/storage"
	"github.com/behshadrhp/Shadina/internal/utils"
	"github.com/go-chi/chi/v5"
)

func (app *application) listPostsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	posts, err := app.storage.Posts.List(ctx)
	if err != nil {
		response.InternalErrorResponse(w, err.Error(), 5001)
		return
	}

	for i := range posts {

		comments, err := app.storage.Comments.GetByPostID(ctx, posts[i].ID)
		if err != nil {
			response.InternalErrorResponse(w, err.Error(), 5009)
			return
		}

		posts[i].Comments = comments
	}

	response.SuccessResponse(w, "posts retrieved successfully", 1001, posts)
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload dto.CreatePostPayload

	if err := utils.ReadJson(w, r, &payload); err != nil {
		response.InternalErrorResponse(w, err.Error(), 5002)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		response.InternalErrorResponse(w, err.Error(), 5006)
		return
	}

	post := storage.Post{
		UserID:  1,
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
	}

	ctx := r.Context()

	if err := app.storage.Posts.Create(ctx, &post); err != nil {
		response.InternalErrorResponse(w, err.Error(), 5003)
		return
	}

	response.SuccessResponse(w, "post created successfully", 1002, post)
}

func (app *application) retrievePostHandler(w http.ResponseWriter, r *http.Request) {
	paramID := chi.URLParam(r, "postID")
	postID, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil {
		response.InternalErrorResponse(w, err.Error(), 5004)
		return
	}

	ctx := r.Context()

	post, err := app.storage.Posts.GetByID(ctx, postID)
	if err != nil {
		switch {
		case errors.Is(err, consts.NOT_FOUND):
			response.NotFoundResponse(w, "post not found", 4001)
		default:
			response.InternalErrorResponse(w, err.Error(), 5005)
		}
		return
	}

	comments, err := app.storage.Comments.GetByPostID(ctx, post.ID)
	if err != nil {
		response.InternalErrorResponse(w, err.Error(), 5009)
		return
	}

	post.Comments = comments

	response.SuccessResponse(w, "post retrieved successfully", 1003, post)
}

func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	paramID := chi.URLParam(r, "postID")
	postID, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil {
		response.InternalErrorResponse(w, err.Error(), 5007)
		return
	}

	ctx := r.Context()

	err = app.storage.Posts.Delete(ctx, postID)
	if err != nil {
		switch {
		case errors.Is(err, consts.NOT_FOUND):
			response.NotFoundResponse(w, "post not found", 4001)
		default:
			response.InternalErrorResponse(w, err.Error(), 5008)
		}
		return
	}

	response.SuccessResponse(w, "post deleted successfully", 1004, nil)
}

func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	paramID := chi.URLParam(r, "postID")
	postID, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil {
		response.InternalErrorResponse(w, err.Error(), 5007)
		return
	}

	ctx := r.Context()

	post, err := app.storage.Posts.GetByID(ctx, postID)
	if err != nil {
		switch {
		case errors.Is(err, consts.NOT_FOUND):
			response.NotFoundResponse(w, "not found", 4002)
			return
		default:
			response.InternalErrorResponse(w, err.Error(), 5008)
			return
		}
	}

	var payload dto.UpdatePostPayload
	if err := utils.ReadJson(w, r, &payload); err != nil {
		response.BadRequestResponse(w, "request invalid", 4003)
		return
	}

	if err := utils.Validate.StructPartial(payload); err != nil {
		response.BadRequestResponse(w, "request invalid", 4004)
		return
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}

	if payload.Content != nil {
		post.Content = *payload.Content
	}

	if err := app.storage.Posts.Update(ctx, post); err != nil {
		switch {
		case errors.Is(err, consts.NOT_FOUND):
			response.NotFoundResponse(w, "post not found", 4001)
		default:
			response.InternalErrorResponse(w, err.Error(), 5008)
		}
		return
	}

	response.SuccessResponse(w, "post updated successfully", 1004, nil)
}
