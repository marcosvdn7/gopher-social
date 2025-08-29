package main

import (
	"net/http"
	"social/internal/store"
)

func (app *application) statusInternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal server error", "method", r.Method, "path", r.URL.Path,
		"error", err.Error())

	writeJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *application) statusBadRequest(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("bad request", "method", r.Method, "path", r.URL.Path,
		"error", err.Error())

	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) statusNotFound(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("not found", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusNotFound, store.ErrNotFound.Error())
}

func (app *application) statusConflict(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("conflict", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusConflict, store.ErrDuplicatedKey.Error())
}

func (app *application) statusUnauthorized(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("unauthorized", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusUnauthorized, err.Error())
}

func (app *application) statusBasicUnauthorized(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Add("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	app.statusUnauthorized(w, r, err)
}

func (app *application) forbiddenResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("forbidden", "method", r.Method, "path", r.URL.Path, "error", err)

	writeJSONError(w, http.StatusForbidden, err.Error())
}

func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	app.logger.Warnw("rate limit exceeded", r.Method, "path", r.URL.Path)

	w.Header().Set("Retry-After", retryAfter)

	writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded, retry after: "+retryAfter)
}
