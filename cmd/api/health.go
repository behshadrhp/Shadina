package main

import (
	"net/http"

	"github.com/behshadrhp/Shadina/internal/dto"
	"github.com/behshadrhp/Shadina/internal/utils"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	payload := dto.HealthStatusPayload{
		Status:      "ok",
		Version:     app.config.version,
		Environment: app.config.environment,
	}

	utils.WriteJson(w, http.StatusOK, payload)
}
