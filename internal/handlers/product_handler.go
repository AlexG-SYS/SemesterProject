package handlers

import (
	"net/http"
	"strconv"

	"github.com/AlexG-SYS/semesterproject/internal/data"
)

func (h *Handler) CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	h.App.Logger.Info("ECO PROMPT: Initiating new product entry into the master catalog.")
	var input struct {
		CategoryID    int64  `json:"category_id"`
		Name          string `json:"name"`
		Description   string `json:"description"`
		IsGstEligible bool   `json:"is_gst_eligible"`
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	product := &data.Product{
		CategoryID:    input.CategoryID,
		Name:          input.Name,
		Description:   input.Description,
		IsGstEligible: input.IsGstEligible,
	}

	// Validate (Ensure you update your ValidateProduct function to check CategoryID)
	if err := h.Models.Products.InsertProduct(product); err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusCreated, map[string]any{"product": product}, nil)
}

func (h *Handler) GetProductHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Extract the ID from the URL path
	id, err := strconv.ParseInt(r.URL.Path[len("/v1/products/"):], 10, 64)
	if err != nil || id < 1 {
		h.App.ErrorJSON(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	// 2. Call the Get method from the model
	product, err := h.Models.Products.GetProduct(id)
	if err != nil {
		h.App.ErrorJSON(w, http.StatusNotFound, "product not found")
		return
	}

	// 3. Return the single product wrapped in a JSON object
	h.App.WriteJSON(w, http.StatusOK, map[string]any{"product": product}, nil)
}

func (h *Handler) ListProductsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Updated input struct: Category is now an int to match the DB Foreign Key
	var input struct {
		Name       string
		CategoryID int
		data.Filters
	}

	qs := r.URL.Query()

	// 2. Extract values
	input.Name = qs.Get("name")

	// Use your readInt helper for the CategoryID as well
	// If "category_id" isn't in the URL, it defaults to 0 (which our SQL handles)
	input.CategoryID = h.readInt(qs, "category_id", 0)

	input.Page = h.readInt(qs, "page", 1)
	input.PageSize = h.readInt(qs, "page_size", 10)

	// Handle sort (default to "product_id" to match your schema)
	input.Sort = qs.Get("sort")
	if input.Sort == "" {
		input.Sort = "product_id"
	}

	// 3. Update the Safelist with your ACTUAL database column names
	// Note: I added "is_gst_eligible" since that's a field you might want to sort by!
	input.SortSafelist = []string{
		"product_id", "-product_id",
		"name", "-name",
		"category_id", "-category_id",
		"is_gst_eligible", "-is_gst_eligible",
	}

	// 4. Call the modified GetAll method
	// Now passing input.CategoryID (int) instead of a string
	products, metadata, err := h.Models.Products.GetAllProducts(input.Name, input.CategoryID, input.Filters)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	// 5. Return the JSON
	h.App.WriteJSON(w, http.StatusOK, map[string]any{
		"products": products,
		"metadata": metadata,
	}, nil)
}

func (h *Handler) UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Extract the ID from the URL
	id, err := strconv.ParseInt(r.URL.Path[len("/v1/products/"):], 10, 64)
	if err != nil || id < 1 {
		h.App.ErrorJSON(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	// 2. Fetch the existing product from the database
	product, err := h.Models.Products.GetProduct(id)
	if err != nil {
		h.App.ErrorJSON(w, http.StatusNotFound, "product not found")
		return
	}

	// 3. Read the new data from the request body
	var input struct {
		Name          *string `json:"name"`
		Description   *string `json:"description"`
		CategoryID    *int64  `json:"category_id"`
		IsGstEligible *bool   `json:"is_gst_eligible"`
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// 4. Update the fields on our product pointer
	if input.Name != nil {
		product.Name = *input.Name
	}
	if input.Description != nil {
		product.Description = *input.Description
	}
	if input.CategoryID != nil {
		product.CategoryID = *input.CategoryID
	}
	if input.IsGstEligible != nil {
		product.IsGstEligible = *input.IsGstEligible
	}

	// 5. Validate the updated object
	if errs := data.ValidateProduct(product); len(errs) > 0 {
		h.App.WriteJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs}, nil)
		return
	}

	// 6. Save the changes back to the database
	err = h.Models.Products.UpdateProduct(product)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	// 7. Return the updated product
	h.App.WriteJSON(w, http.StatusOK, map[string]any{"product": product}, nil)
}

func (h *Handler) CreateVariantHandler(w http.ResponseWriter, r *http.Request) {
	h.App.Logger.Info("ECO PROMPT: Adding a new variant for an existing product.")
	var input struct {
		ProductID    int64   `json:"product_id"`
		SKU          string  `json:"sku"`
		SizeAttr     string  `json:"size_attr"`
		ColorAttr    string  `json:"color_attr"`
		CostPrice    float64 `json:"cost_price"`
		SellingPrice float64 `json:"selling_price"`
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	variant := &data.Variant{
		ProductID:    input.ProductID,
		SKU:          input.SKU,
		SizeAttr:     input.SizeAttr,
		ColorAttr:    input.ColorAttr,
		CostPrice:    input.CostPrice,
		SellingPrice: input.SellingPrice,
	}

	// 1. Validate the variant data
	if errs := data.ValidateVariant(variant); len(errs) > 0 {
		h.App.WriteJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs}, nil)
		return
	}
	// 2. Insert the variant into the database
	if err := h.Models.Products.InsertVariant(variant); err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusCreated, map[string]any{"variant": variant}, nil)
}

func (h *Handler) ListVariantsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Correctly extract the {id} from the registered path segment
	idParam := r.PathValue("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || id < 1 {
		h.App.ErrorJSON(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	// 2. Call the Model method (this part is already perfect)
	variants, err := h.Models.Products.GetVariantByProduct(id)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	// 3. Return the JSON
	h.App.WriteJSON(w, http.StatusOK, map[string]any{"variants": variants}, nil)
}

func (h *Handler) UpdateVariantHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Extract Variant ID from URL
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		h.App.ErrorJSON(w, http.StatusBadRequest, "invalid variant ID")
		return
	}

	// 2. Fetch existing variant
	variant, err := h.Models.Products.GetVariant(id)
	if err != nil {
		h.App.ErrorJSON(w, http.StatusNotFound, "variant not found")
		return
	}

	// 3. Define input struct with pointers for PATCH logic
	var input struct {
		SKU          *string  `json:"sku"`
		SizeAttr     *string  `json:"size_attr"`
		ColorAttr    *string  `json:"color_attr"`
		CostPrice    *float64 `json:"cost_price"`
		SellingPrice *float64 `json:"selling_price"`
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// 4. Update fields only if provided
	if input.SKU != nil {
		variant.SKU = *input.SKU
	}
	if input.SizeAttr != nil {
		variant.SizeAttr = *input.SizeAttr
	}
	if input.ColorAttr != nil {
		variant.ColorAttr = *input.ColorAttr
	}
	if input.CostPrice != nil {
		variant.CostPrice = *input.CostPrice
	}
	if input.SellingPrice != nil {
		variant.SellingPrice = *input.SellingPrice
	}

	if errs := data.ValidateVariant(variant); len(errs) > 0 {
		h.App.WriteJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs}, nil)
		return
	}

	// 5. Save to Database
	err = h.Models.Products.UpdateVariant(variant)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, map[string]any{"variant": variant}, nil)
}
