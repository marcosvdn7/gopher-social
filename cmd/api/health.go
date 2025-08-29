package main

import (
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":      "ok",
		"environment": app.config.env,
		"version":     version,
	}

	if err := app.JSONResponse(w, http.StatusOK, data); err != nil {
		app.statusInternalServerError(w, r, err)
	}
}
