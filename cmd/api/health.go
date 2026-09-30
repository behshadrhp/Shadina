package main

import (
	"net/http"

	"github.com/behshadrhp/Shadina/internal/dto"
	"github.com/behshadrhp/Shadina/internal/utils"
)

// healthCheckHandler godoc
//
//	@Summary		Health Check
//	@Description	Check API status, version and environment information.
//	@Tags			System
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.HealthStatusPayload	"API is healthy"
//	@Router			/health [get]
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	payload := dto.HealthStatusPayload{
		Status:      "ok",
		Version:     app.config.version,
		Environment: app.config.environment,
	}

	utils.WriteJson(w, http.StatusOK, payload)
}
