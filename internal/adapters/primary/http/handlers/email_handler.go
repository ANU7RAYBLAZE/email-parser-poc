package handlers

import (
	"email-parser-poc/internal/domain/entities"
	"email-parser-poc/internal/ports/incoming"
	"encoding/json"
	"net/http"
	"strconv"
)

type EmailHandler struct {
	emailService incoming.EmailService
}

func NewEmailHandler(emailService incoming.EmailService) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
	}
}

func (h *EmailHandler) GetAllEmails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	pageToken := r.URL.Query().Get("page_token")

	filter := entities.EmailFilter{
		MaxResults: limit,
		PageToken:  pageToken,
	}

	emailList, s3Filename, err := h.emailService.GetEmails(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respone := map[string]interface{}{
		"emails":      emailList,
		"s3_filename": s3Filename,
		"message":     "Emails successfully stored in S3",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respone)
}
