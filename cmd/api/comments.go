package main

import (
	"net/http"
	"social/internal/store"
)

type CreateCommentPayload struct {
	Content string `json:"content" validate:"required,max=200"`
}

func (app *application) postCommentHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateCommentPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.statusBadRequest(w, r, err)
		return
	}

	post := getPostFromCtx(r)
	ctx := r.Context()

	comment := store.Comment{
		UserID:  1,
		PostID:  post.ID,
		Content: payload.Content,
	}

	if err := app.store.Comment.Create(ctx, &comment); err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}

	if err := app.JSONResponse(w, http.StatusOK, comment); err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}
}
