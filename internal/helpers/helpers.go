package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
)

type Application struct {
	Logger        *slog.Logger
	TotalRequests atomic.Uint64
	TotalErrors   atomic.Uint64
	TotalLatency  atomic.Uint64
	RouteHits     sync.Map // map[string]*atomic.Uint64
}

// WriteJSON sends a structured JSON response
func (app *Application) WriteJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	jsResponse, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(jsResponse)
	return err
}

// ReadJSON decodes a JSON request body into a destination struct
func (app *Application) ReadJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	maxBytes := 1_048_576 // 1MB limit
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(destination)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}

	// Ensure there is only one JSON object in the body
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// ErrorJSON sends a consistent error format
func (app *Application) ErrorJSON(w http.ResponseWriter, status int, message string) {
	payload := map[string]string{"error": message}
	app.WriteJSON(w, status, payload, nil)
}

// ServerError handles 500 errors and logs them
func (app *Application) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.TotalErrors.Add(1)
	app.Logger.Error("server error", "method", r.Method, "uri", r.URL.RequestURI(), "error", err.Error())

	message := "the server encountered a problem"
	app.ErrorJSON(w, http.StatusInternalServerError, message)
}
