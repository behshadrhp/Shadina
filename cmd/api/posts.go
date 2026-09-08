package main

import (
	"net/http"

	"github.com/behshadrhp/Shadina/internal/dto"
	"github.com/behshadrhp/Shadina/internal/response"
	"github.com/behshadrhp/Shadina/internal/storage"
	"github.com/behshadrhp/Shadina/internal/utils"
)

func (app *application) listPostsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	posts, err := app.storage.Posts.List(ctx)
	if err != nil {
		response.InternalErrorResponse(w, err.Error(), 5001)
		return
	}

	response.SuccessResponse(w, "posts retrieved successfully", 1001, posts)
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload dto.CreatePostPayload

	if err := utils.ReadJson(w, r, &payload); err != nil {
		response.InternalErrorResponse(w, err.Error(), 5002)
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
