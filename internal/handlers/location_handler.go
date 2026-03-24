package handlers

import (
	"net/http"
	"strconv"

	"github.com/AlexG-SYS/semesterproject/internal/data"
)

func (h *Handler) CreateLocationHandler(w http.ResponseWriter, r *http.Request) {
	h.App.Logger.Info("ECO PROMPT: Processing new physical warehouse location registration.")
	var input struct {
		Name    string `json:"name"`
		Address string `json:"address"`
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	location := &data.Location{
		Name:    input.Name,
		Address: input.Address,
	}

	if err := h.Models.Locations.Insert(location); err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusCreated, map[string]any{"location": location}, nil)
}

func (h *Handler) ListLocationsHandler(w http.ResponseWriter, r *http.Request) {
	h.App.Logger.Info("ECO PROMPT: Retrieving all registered warehouse locations.")
	locations, err := h.Models.Locations.GetAll()
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, map[string]any{"locations": locations}, nil)
}

func (h *Handler) UpdateLocationHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Extract the ID from the URL
	id, err := strconv.ParseInt(r.URL.Path[len("/v1/locations/"):], 10, 64)
	if err != nil || id < 1 {
		h.App.ErrorJSON(w, http.StatusBadRequest, "invalid location ID")
		return
	}

	// 2. Fetch the existing location from the database
	location, err := h.Models.Locations.GetLocation(id)
	if err != nil {
		h.App.ErrorJSON(w, http.StatusNotFound, "location not found")
		return
	}

	// 3. Read the new data from the request body
	var input struct {
		Name     *string `json:"name"`
		Address  *string `json:"address"`
		IsActive *bool   `json:"is_active"`
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// 4. Update the fields on our location pointer
	if input.Name != nil {
		location.Name = *input.Name
	}
	if input.Address != nil {
		location.Address = *input.Address
	}
	if input.IsActive != nil {
		location.IsActive = *input.IsActive
	}

	// 5. Validate the updated object
	if errs := data.ValidateLocation(location); len(errs) > 0 {
		h.App.WriteJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs}, nil)
		return
	}

	// 6. Save the changes back to the database
	err = h.Models.Locations.UpdateLocation(location)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	// 7. Return the updated location so the user sees the changes
	h.App.WriteJSON(w, http.StatusOK, map[string]any{"location": location}, nil)
}
