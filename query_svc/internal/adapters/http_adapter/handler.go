package http_adapter

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/speakr/query_svc/internal/core"
	"github.com/speakr/query_svc/internal/ports"
)

// Handler implements the HTTP API for the query service
type Handler struct {
	service *core.Service
	logger  *slog.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(service *core.Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// SetupRoutes configures the HTTP routes
func (h *Handler) SetupRoutes() http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(h.correlationIDMiddleware)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health endpoint
	r.Get("/health", h.healthHandler)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/query", h.queryHandler)
	})

	return r
}

// correlationIDMiddleware adds correlation ID to context
func (h *Handler) correlationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := middleware.GetReqID(r.Context())
		ctx := context.WithValue(r.Context(), "correlation_id", correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// healthHandler handles health check requests
func (h *Handler) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"service": "query_svc",
	})
}

// QueryResponse represents the API response for search queries
type QueryResponse struct {
	Results []ports.SearchResult `json:"results"`
	Count   int                  `json:"count"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// queryHandler handles search query requests
func (h *Handler) queryHandler(w http.ResponseWriter, r *http.Request) {
	correlationID := r.Context().Value("correlation_id")

	h.logger.Info("Received query request",
		"correlation_id", correlationID,
		"method", r.Method,
		"path", r.URL.Path,
	)

	// Parse request body
	var req core.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to parse request body",
			"correlation_id", correlationID,
			"error", err,
		)
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Parse optional limit from query parameter
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	// Perform search
	results, err := h.service.Search(r.Context(), req)
	if err != nil {
		h.logger.Error("Search operation failed",
			"correlation_id", correlationID,
			"error", err,
		)

		// Determine appropriate HTTP status code based on error type
		statusCode := http.StatusInternalServerError
		switch {
		case err == core.ErrInvalidQuery:
			statusCode = http.StatusBadRequest
		case err == core.ErrDatabaseUnavailable:
			statusCode = http.StatusServiceUnavailable
		}

		h.writeErrorResponse(w, statusCode, "Search failed", err.Error())
		return
	}

	// Return successful response
	response := QueryResponse{
		Results: results,
		Count:   len(results),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response",
			"correlation_id", correlationID,
			"error", err,
		)
	}

	h.logger.Info("Query request completed successfully",
		"correlation_id", correlationID,
		"results_count", len(results),
	)
}

// writeErrorResponse writes a JSON error response
func (h *Handler) writeErrorResponse(w http.ResponseWriter, statusCode int, error, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	errorResp := ErrorResponse{
		Error:   error,
		Message: message,
	}
	
	json.NewEncoder(w).Encode(errorResp)
}