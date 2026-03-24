package handlers

import (
	"net/http"
	"strconv"

	"github.com/AlexG-SYS/semesterproject/internal/data"
)

func (h *Handler) CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	h.App.Logger.Info("ECO PROMPT: Received request to define a new product category.")
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	category := &data.Category{
		Name:        input.Name,
		Description: input.Description,
	}

	if errs := data.ValidateCategory(category); len(errs) > 0 {
		h.App.WriteJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs}, nil)
		return
	}

	if err := h.Models.Categories.Insert(category); err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusCreated, map[string]any{"category": category}, nil)
}

func (h *Handler) ListCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	h.App.Logger.Info("ECO PROMPT: Fetching the list of all product categories.")
	categories, err := h.Models.Categories.GetAll()
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, map[string]any{"categories": categories}, nil)
}

func (h *Handler) UpdateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Extract the ID from the URL
	id, err := strconv.ParseInt(r.URL.Path[len("/v1/categories/"):], 10, 64)
	if err != nil || id < 1 {
		h.App.ErrorJSON(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	// 2. Fetch the existing category from the database
	category, err := h.Models.Categories.GetCategory(id)
	if err != nil {
		h.App.ErrorJSON(w, http.StatusNotFound, "category not found")
		return
	}

	// 3. Read the new data from the request body
	var input struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// 4. Update the fields on our category pointer
	if input.Name != nil {
		category.Name = *input.Name
	}
	if input.Description != nil {
		category.Description = *input.Description
	}

	// 5. Validate the updated object
	if errs := data.ValidateCategory(category); len(errs) > 0 {
		h.App.WriteJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs}, nil)
		return
	}

	// 6. Save the changes back to the database
	err = h.Models.Categories.UpdateCategory(category)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	// 7. Return the updated category so the user sees the changes
	h.App.WriteJSON(w, http.StatusOK, map[string]any{"category": category}, nil)
}
