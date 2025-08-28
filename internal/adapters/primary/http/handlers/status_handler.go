package handlers

import (
	"email-parser-poc/internal/ports/incoming"
	"encoding/json"
	"net/http"
	"time"
)

type HealthHandler struct {
	HealthService incoming.HealthPort
}

func NewHealthHandler(healthService incoming.HealthPort) *HealthHandler {
	return &HealthHandler{
		HealthService: healthService,
	}
}

func (h *HealthHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	health, err := h.HealthService.CheckHealth(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get health status", err)
		return
	}

	statusCode := http.StatusOK
	if !health.IsHealthy() {
		statusCode = http.StatusServiceUnavailable
	}

	h.writeJsonResponse(w, statusCode, health)
}

func (h *HealthHandler) writeJsonResponse(w http.ResponseWriter, statuscode int, data interface{}) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statuscode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *HealthHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]interface{}{
		"error":     message,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if err != nil {
		errorResponse["detail"] = err.Error()
	}

	json.NewEncoder(w).Encode(errorResponse)
}
