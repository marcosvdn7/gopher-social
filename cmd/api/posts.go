package main

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"social/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.statusBadRequest(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.statusBadRequest(w, r, err)
		return
	}

	ctx := r.Context()
	user := ctx.Value(userCtxKey).(store.User)

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  user.ID,
	}

	if err := app.store.Post.Create(ctx, post); err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}

	if err := app.JSONResponse(w, http.StatusOK, post); err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	ctx := r.Context()

	comments, err := app.store.Comment.GetByPostId(ctx, post.ID)
	if err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}

	post.Comments = comments

	if err := app.JSONResponse(w, http.StatusOK, post); err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}
}

func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "postID")
	postID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}
	ctx := r.Context()
	_, err = app.store.Comment.DeleteByPostId(ctx, postID)
	if err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}

	rowsAffected, err := app.store.Post.Delete(ctx, postID)
	if err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}
	if rowsAffected == 0 {
		app.statusNotFound(w, r, errors.New("no id found to delete"))
		return
	}
	response := map[string]string{
		"success": "successfully deleted",
	}

	app.JSONResponse(w, http.StatusOK, response)
}

type PatchPostPayload struct {
	ID      int64    `json:"id"`
	Title   string   `json:"title" validate:"max=100"`
	Content string   `json:"content" validate:"max=1000"`
	Tags    []string `json:"tags"`
}

func (p *PatchPostPayload) mapUpdatedFields(post *store.Post) {
	if p.Title != post.Title {
		post.Title = p.Title
	}
	if p.Content != post.Content {
		post.Content = p.Content
	}
	for _, newTag := range p.Tags {
		if !slices.Contains(post.Tags, newTag) {
			post.Tags = append(post.Tags, newTag)
		}
	}
}

func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	var payload PatchPostPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}
	persistedPost := getPostFromCtx(r)
	ctx := r.Context()

	payload.mapUpdatedFields(persistedPost)

	if err := app.store.Post.Update(ctx, persistedPost); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.statusNotFound(w, r, err)
		default:
			app.statusInternalServerError(w, r, err)
		}
		return
	}

	app.JSONResponse(w, http.StatusOK, persistedPost)
}

func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paramId := chi.URLParam(r, "postID")
		postId, err := strconv.ParseInt(paramId, 10, 64)
		if err != nil {
			app.statusInternalServerError(w, r, err)
			return
		}

		ctx := r.Context()

		post, err := app.store.Post.GetById(ctx, postId)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.statusNotFound(w, r, err)
			default:
				app.statusInternalServerError(w, r, err)
			}
			return
		}
		ctx = context.WithValue(ctx, postCtx, &post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post := r.Context().Value(postCtx).(*store.Post)
	return post
}
