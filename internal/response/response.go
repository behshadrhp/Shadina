package response

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	InternalCode int    `json:"internal_code"`
	Data         any    `json:"data"`
}

func Respond(w http.ResponseWriter, success bool, message string, statusCode int, internalCode int, data any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Success:      success,
		Message:      message,
		InternalCode: internalCode,
		Data:         data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Fatalf("error encoding response: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func SuccessResponse(w http.ResponseWriter, message string, internalCode int, data any) {
	Respond(w, true, message, http.StatusOK, internalCode, data)
}

func CreatedResponse(w http.ResponseWriter, message string, internalCode int, data any) {
	Respond(w, true, message, http.StatusCreated, internalCode, data)
}

func BadRequestResponse(w http.ResponseWriter, message string, internalCode int) {
	Respond(w, false, message, http.StatusBadRequest, internalCode, nil)
}

func UnauthorizedResponse(w http.ResponseWriter, message string, internalCode int) {
	Respond(w, false, message, http.StatusUnauthorized, internalCode, nil)
}

func ForbiddenResponse(w http.ResponseWriter, message string, internalCode int) {
	Respond(w, false, message, http.StatusForbidden, internalCode, nil)
}

func NotFoundResponse(w http.ResponseWriter, message string, internalCode int) {
	Respond(w, false, message, http.StatusNotFound, internalCode, nil)
}

func InternalErrorResponse(w http.ResponseWriter, message string, internalCode int) {
	Respond(w, false, message, http.StatusInternalServerError, internalCode, nil)
}
