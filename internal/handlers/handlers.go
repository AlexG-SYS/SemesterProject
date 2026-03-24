package handlers

import (
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/AlexG-SYS/semesterproject/internal/data"
	"github.com/AlexG-SYS/semesterproject/internal/helpers"
)

// Use the Application struct from helpers as a receiver
type Handler struct {
	App    *helpers.Application
	Models data.Models
}

func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "available",
		"message": "Welcome to the Home Page!",
	}
	h.App.WriteJSON(w, http.StatusOK, data, nil)
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := h.App.ReadJSON(w, r, &input)
	if err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Logic example: check if fields are missing
	if input.Email == "" || input.Password == "" {
		h.App.ErrorJSON(w, http.StatusUnprocessableEntity, "missing credentials")
		return
	}

	h.App.WriteJSON(w, http.StatusOK, map[string]string{"message": "login successful"}, nil)
}

// Helper to read integer values from query string
func (h *Handler) readInt(qs url.Values, key string, defaultValue int) int {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return i
}

// MetricsHandler returns the current performance metrics of the system
func (h *Handler) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	total := h.App.TotalRequests.Load()

	// Calculate Average Latency
	var avgLatency string
	if total > 0 {
		avg := time.Duration(h.App.TotalLatency.Load() / total)
		avgLatency = avg.String()
	} else {
		avgLatency = "0s"
	}

	// Prepare Route Hits map
	routes := make(map[string]uint64)
	h.App.RouteHits.Range(func(key, value any) bool {
		routes[key.(string)] = value.(*atomic.Uint64).Load()
		return true
	})

	metrics := map[string]any{
		"total_requests":     total,
		"total_errors":       h.App.TotalErrors.Load(),
		"average_latency":    avgLatency,
		"requests_per_route": routes,
	}

	h.App.WriteJSON(w, http.StatusOK, metrics, nil)
}
