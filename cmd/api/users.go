package main

import (
	"net/http"
	"social/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{id} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	app.logger.Infow("user id parsed from param ", "user", userID)
	if err != nil {
		app.statusBadRequest(w, r, err)
		return
	}

	user, err := app.getUser(r.Context(), userID)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.statusNotFound(w, r, err)
		default:
			app.statusInternalServerError(w, r, err)
		}
		return
	}

	app.logger.Infow("user response", "user", user)

	if err = app.JSONResponse(w, http.StatusOK, user); err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}
}

// FollowUser godoc
//
//	@Summary		Follows a user
//	@Description	Follows a user by ID
//	@Tags			users
//	@Param			userID	path		int		true	"User ID"
//	@Success		200		{string}	string	"User followed"
//	@Failure		902		{object}	error	"User already followed!"
//	@Failure		404		{object}	error	"User not found"
//	@Failure		500		{object}	error
//	@Failure		400		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/follow [put]
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	follower := ctx.Value("userCtx").(*store.User)
	followedId, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.statusBadRequest(w, r, err)
	}

	if err := app.store.Follower.Follow(ctx, follower.ID, followedId); err != nil {
		switch err {
		case store.ErrNotFound:
			app.statusNotFound(w, r, err)
		case store.ErrDuplicatedKey:
			app.statusConflict(w, r, err)
		default:
			app.statusInternalServerError(w, r, err)
		}
		return
	}

	if err := app.JSONResponse(w, http.StatusNoContent, nil); err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	follower := ctx.Value("userCtx").(*store.User)
	unfollowedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.statusBadRequest(w, r, err)
	}

	if err = app.store.Follower.Unfollow(ctx, follower.ID, unfollowedID); err != nil {
		switch err {
		case store.ErrNotFound:
			app.statusNotFound(w, r, err)
		default:
			app.statusInternalServerError(w, r, err)
		}
		return
	}

	response := map[string]string{
		"success": "successfully unfollowed",
	}

	if err := app.JSONResponse(w, http.StatusOK, response); err != nil {
		app.statusInternalServerError(w, r, err)
		return
	}
}

// ActivateUser godoc
//
//	@Summary		Activates/Register a user
//	@Description	Activates/Register a user by invitation token
//	@Tags			users
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		204		{string}	string	"User activated"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	if err := app.store.User.Activate(r.Context(), token); err != nil {
		switch err {
		case store.ErrNotFound:
			app.statusNotFound(w, r, err)
		default:
			app.statusInternalServerError(w, r, err)
		}
		return
	}

	if err := app.JSONResponse(w, http.StatusNoContent, ""); err != nil {
		app.statusInternalServerError(w, r, err)
	}
}
