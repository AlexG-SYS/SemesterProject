package handlers

import (
	"net/http"
	"strconv"

	"github.com/AlexG-SYS/semesterproject/internal/data"
)

func (h *Handler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		CustomerID       int64 `json:"customer_id"`
		LocationID       int64 `json:"location_id"`
		ShippingMethodID int64 `json:"shipping_method_id"`
		Items            []struct {
			VariantID int64 `json:"variant_id"`
			Quantity  int   `json:"quantity"`
		} `json:"items"`
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	order := &data.Order{
		CustomerID:       input.CustomerID,
		ShippingMethodID: input.ShippingMethodID,
		Status:           "Pending",
	}

	for _, item := range input.Items {
		order.OrderItems = append(order.OrderItems, data.OrderItem{
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
		})
	}

	if err := h.Models.Orders.Insert(r.Context(), order, input.LocationID); err != nil {
		// If the error contains "insufficient stock", return 409 Conflict or 422
		h.App.ErrorJSON(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.App.WriteJSON(w, http.StatusCreated, map[string]any{"order": order}, nil)
}

func (h *Handler) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		h.App.ErrorJSON(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	order, err := h.Models.Orders.GetByID(r.Context(), id)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}
	if order == nil {
		h.App.ErrorJSON(w, http.StatusNotFound, "Order not found")
		return
	}

	h.App.WriteJSON(w, http.StatusOK, map[string]any{"order": order}, nil)
}

func (h *Handler) UpdateOrderHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		h.App.ErrorJSON(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	var input struct {
		Status     string `json:"status"`
		LocationID int64  `json:"location_id"` // Needed to find the right inventory row
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate status
	validStatuses := map[string]bool{"Pending": true, "Paid": true, "Cancelled": true, "Shipped": true}
	if !validStatuses[input.Status] {
		h.App.ErrorJSON(w, http.StatusBadRequest, "Invalid status value")
		return
	}

	err = h.Models.Orders.UpdateStatus(r.Context(), id, input.Status, input.LocationID)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, map[string]any{"message": "order updated successfully"}, nil)
}
