package utils

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func ResponseJSON(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

func ResponseError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Status:  "error",
		Message: message,
		Data:    nil,
	}

	json.NewEncoder(w).Encode(response)
}
