package handlers

import (
	"encoding/json"
	"net/http"
)

func sendErrorResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := Response{
		Success: false,
		Error:   message,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"success":false,"error":"Internal server error"}`, http.StatusInternalServerError)
	}
}
