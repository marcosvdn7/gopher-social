package main

import (
	"net/http"
	"social/internal/store"
)

// GetFeed godoc
//
//	@Summary		Fetches posts
//	@Description	Fetches posts related to the user
//	@Tags			feed
//	@Produce		json
//	@Success		200	{object}	store.PostWithMetadata
//	@Failure		400	{object}	error
//	@Failure		401	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/feed [get]
func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	fq := store.PaginatedFeedQuery{
		Limit:  10,
		Offset: 2,
		Sort:   "desc",
	}
	fq, err := fq.Parse(r)
	if err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.statusBadRequest(w, r, err)
		return
	}

	ctx := r.Context()

	feed, err := app.store.Post.GetUserFeed(ctx, int64(1), fq)
	if err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}

	if err := app.JSONResponse(w, http.StatusOK, feed); err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}
}
