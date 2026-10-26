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

// listPostsHandler godoc
//
//	@Summary		List Posts
//	@Description	Retrieve all posts with their associated comments.
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response	"Posts retrieved successfully"
//	@Failure		500	{object}	response.Response	"Internal server error"
//	@Router			/posts [get]
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

// createPostHandler godoc
//
//	@Summary		Create Post
//	@Description	Create a new post.
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.CreatePostPayload	true	"Create Post Payload"
//	@Success		201		{object}	response.Response		"Post created successfully"
//	@Failure		400		{object}	response.Response		"Validation error"
//	@Failure		500		{object}	response.Response		"Internal server error"
//	@Router			/posts [post]
func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload dto.CreatePostPayload
	user := app.GetUserCTX(r)

	if err := utils.ReadJson(w, r, &payload); err != nil {
		response.InternalErrorResponse(w, err.Error(), 5002)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		response.InternalErrorResponse(w, err.Error(), 5006)
		return
	}

	post := storage.Post{
		UserID:  user.ID,
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

// retrievePostHandler godoc
//
//	@Summary		Retrieve Post
//	@Description	Retrieve a single post by its ID including comments.
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		int					true	"Post ID"
//	@Success		200		{object}	response.Response	"Post retrieved successfully"
//	@Failure		404		{object}	response.Response	"Post not found"
//	@Failure		500		{object}	response.Response	"Internal server error"
//	@Router			/posts/{postID} [get]
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

// deletePostHandler godoc
//
//	@Summary		Delete Post
//	@Description	Delete a post by its ID.
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		int					true	"Post ID"
//	@Success		200		{object}	response.Response	"Post deleted successfully"
//	@Failure		404		{object}	response.Response	"Post not found"
//	@Failure		500		{object}	response.Response	"Internal server error"
//	@Router			/posts/{postID} [delete]
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

// updatePostHandler godoc
//
//	@Summary		Update Post
//	@Description	Partially update an existing post.
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		int						true	"Post ID"
//	@Param			request	body		dto.UpdatePostPayload	true	"Update Post Payload"
//	@Success		200		{object}	response.Response		"Post updated successfully"
//	@Failure		400		{object}	response.Response		"Invalid request"
//	@Failure		404		{object}	response.Response		"Post not found"
//	@Failure		500		{object}	response.Response		"Internal server error"
//	@Router			/posts/{postID} [patch]
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
