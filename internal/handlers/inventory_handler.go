package handlers

import (
	"net/http"
	"strconv"

	"github.com/AlexG-SYS/semesterproject/internal/data"
)

func (h *Handler) CreateInventoryHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		VariantID   int64 `json:"variant_id"`
		LocationID  int64 `json:"location_id"`
		StockOnHand int   `json:"stock_on_hand"`
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Call your model's Insert method
	inventory := &data.Inventory{
		VariantID:   input.VariantID,
		LocationID:  input.LocationID,
		StockOnHand: input.StockOnHand,
	}

	if errs := data.ValidateInventory(inventory); len(errs) > 0 {
		h.App.WriteJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs}, nil)
		return
	}

	if err := h.Models.Inventory.InsertInventory(inventory); err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusCreated, map[string]any{"inventory": inventory}, nil)
}

func (h *Handler) GetInventoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		h.App.ErrorJSON(w, http.StatusBadRequest, "invalid variant ID")
		return
	}

	inventory, err := h.Models.Inventory.GetInventoryByVariant(id)
	if err != nil {
		h.App.ErrorJSON(w, http.StatusNotFound, "inventory not found")
		return
	}

	h.App.WriteJSON(w, http.StatusOK, map[string]any{"inventory": inventory}, nil)
}

func (h *Handler) UpdateInventoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		h.App.ErrorJSON(w, http.StatusBadRequest, "invalid inventory ID")
		return
	}

	var input struct {
		StockOnHand   *int `json:"stock_on_hand"`
		StockReserved *int `json:"stock_reserved"`
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	inventory, err := h.Models.Inventory.GetInventoryByID(id)
	if err != nil {
		h.App.ErrorJSON(w, http.StatusNotFound, "inventory not found")
		return
	}

	if input.StockOnHand != nil {
		inventory.StockOnHand = *input.StockOnHand
	}
	if input.StockReserved != nil {
		inventory.StockReserved = *input.StockReserved
	}

	if errs := data.ValidateInventory(inventory); len(errs) > 0 {
		h.App.WriteJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs}, nil)
		return
	}

	err = h.Models.Inventory.UpdateInventory(inventory)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, map[string]any{"inventory": inventory}, nil)
}
