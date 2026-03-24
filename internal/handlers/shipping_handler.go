package handlers

import (
	"net/http"
	"strconv"

	"github.com/AlexG-SYS/semesterproject/internal/data"
)

func (h *Handler) CreateShippingHandler(w http.ResponseWriter, r *http.Request) {
	h.App.Logger.Info("Creating a new shipping method")

	var input struct {
		ProviderName string  `json:"provider_name"`
		ServiceType  string  `json:"service_type"`
		BaseRate     float64 `json:"base_rate"`
		ContactPhone *string `json:"contact_phone,omitempty"`
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	shipping := &data.Shipping{
		ProviderName: input.ProviderName,
		ServiceType:  input.ServiceType,
		BaseRate:     input.BaseRate,
		ContactPhone: input.ContactPhone,
	}

	// 1. Validate (Assuming this function exists in data/shipping.go)
	if errs := data.ValidateShipping(shipping); len(errs) > 0 {
		h.App.WriteJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs}, nil)
		return
	}

	// 2. Create in DB (Passing pointer so s.ID and s.CreatedAt are populated)
	err := h.Models.Shipping.CreateShipping(r.Context(), shipping)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusCreated, map[string]any{"shipping": shipping}, nil)
}

func (h *Handler) GetShippingHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		h.App.ErrorJSON(w, http.StatusBadRequest, "Invalid shipping method ID")
		return
	}

	shipping, err := h.Models.Shipping.GetShipping(r.Context(), id)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}
	if shipping == nil {
		h.App.ErrorJSON(w, http.StatusNotFound, "Shipping method not found")
		return
	}

	h.App.WriteJSON(w, http.StatusOK, map[string]any{"shipping": shipping}, nil)
}

func (h *Handler) UpdateShippingHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		h.App.ErrorJSON(w, http.StatusBadRequest, "Invalid shipping method ID")
		return
	}

	// 1. Fetch existing
	shipping, err := h.Models.Shipping.GetShipping(r.Context(), id)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}
	if shipping == nil {
		h.App.ErrorJSON(w, http.StatusNotFound, "Shipping method not found")
		return
	}

	// 2. Partial Input (Pointers allow us to see what was actually sent)
	var input struct {
		ProviderName *string  `json:"provider_name"`
		ServiceType  *string  `json:"service_type"`
		BaseRate     *float64 `json:"base_rate"`
		ContactPhone *string  `json:"contact_phone"` // Changed from **string
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Update only provided fields
	if input.ProviderName != nil {
		shipping.ProviderName = *input.ProviderName
	}
	if input.ServiceType != nil {
		shipping.ServiceType = *input.ServiceType
	}
	if input.BaseRate != nil {
		shipping.BaseRate = *input.BaseRate
	}
	if input.ContactPhone != nil {
		shipping.ContactPhone = input.ContactPhone // Already a pointer
	}

	// 4. Validate updated state
	if errs := data.ValidateShipping(shipping); len(errs) > 0 {
		h.App.WriteJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs}, nil)
		return
	}

	// 5. Save (Pass as pointer to stay consistent with your model)
	err = h.Models.Shipping.UpdateShipping(r.Context(), *shipping)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, map[string]any{"shipping": shipping}, nil)
}
