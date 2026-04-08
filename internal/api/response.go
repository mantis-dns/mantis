package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type successResponse struct {
	Data interface{} `json:"data"`
	Meta meta        `json:"meta"`
}

type paginatedResponse struct {
	Data interface{} `json:"data"`
	Meta pageMeta    `json:"meta"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

type meta struct {
	RequestID string `json:"requestId"`
}

type pageMeta struct {
	RequestID string `json:"requestId"`
	Page      int    `json:"page"`
	PerPage   int    `json:"perPage"`
	Total     int    `json:"total"`
}

type errorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
}

func requestID() string {
	return "req_" + uuid.New().String()[:8]
}

// Success writes a JSON success response.
func Success(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, successResponse{
		Data: data,
		Meta: meta{RequestID: requestID()},
	})
}

// Paginated writes a JSON paginated response.
func Paginated(w http.ResponseWriter, data interface{}, page, perPage, total int) {
	writeJSON(w, http.StatusOK, paginatedResponse{
		Data: data,
		Meta: pageMeta{RequestID: requestID(), Page: page, PerPage: perPage, Total: total},
	})
}

// Error writes a JSON error response.
func Error(w http.ResponseWriter, code string, message string, status int) {
	writeJSON(w, status, errorResponse{
		Error: errorBody{Code: code, Message: message, RequestID: requestID()},
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
