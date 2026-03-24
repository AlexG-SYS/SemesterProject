package handlers

import (
	"net/http"
	"strconv"

	"github.com/AlexG-SYS/semesterproject/internal/data"
)

func (h *Handler) CreateProfileHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		FullName    string `json:"full_name"`
		Phone       string `json:"phone"`
		Address     string `json:"address"`
		District    string `json:"district"`
		TownVillage string `json:"town_village"`
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	user := &data.User{Email: input.Email, Password: input.Password, Role: "Customer"}
	profile := &data.Profile{
		FullName:    input.FullName,
		Phone:       input.Phone,
		Address:     input.Address,
		District:    input.District,
		TownVillage: input.TownVillage,
	}

	//validate the input
	if errs := data.ValidateUser(user); len(errs) > 0 {
		h.App.ErrorJSON(w, http.StatusBadRequest, errs["error"])
		return
	}
	if errs := data.ValidateProfile(profile); len(errs) > 0 {
		h.App.ErrorJSON(w, http.StatusBadRequest, errs["error"])
		return
	}

	// Call the new Insert method that handles the transaction
	if err := h.Models.Users.Insert(user, profile); err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusCreated, map[string]any{"user": user, "profile": profile}, nil)
}

func (h *Handler) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from /v1/profiles/{id}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		h.App.ErrorJSON(w, http.StatusNotFound, "invalid record id")
		return
	}

	// Fetch the combined user/profile object
	user, err := h.Models.Profile.Get(id)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, map[string]any{"user": user}, nil)
}

func (h *Handler) UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get ID from URL
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		h.App.ErrorJSON(w, http.StatusNotFound, "invalid id")
		return
	}

	// 2. Fetch existing profile first
	// Note: You'll need a GetProfileByID method in your ProfileModel
	user, err := h.Models.Profile.Get(id)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}
	profile := user.Profile

	// 3. Read JSON into a temporary anonymous struct with pointers
	var input struct {
		FullName    *string `json:"full_name"`
		Phone       *string `json:"phone"`
		Address     *string `json:"address"`
		District    *string `json:"district"`
		TownVillage *string `json:"town_village"`
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// 4. Update only the fields that were provided in the JSON
	if input.FullName != nil {
		profile.FullName = *input.FullName
	}
	if input.Phone != nil {
		profile.Phone = *input.Phone
	}
	if input.Address != nil {
		profile.Address = *input.Address
	}
	if input.District != nil {
		profile.District = *input.District
	}
	if input.TownVillage != nil {
		profile.TownVillage = *input.TownVillage
	}

	// 5. Save the updated version
	if err := h.Models.Profile.Update(profile); err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, map[string]any{"user": user}, nil)
}
